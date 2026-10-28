package main

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const jwtBearerGrant = "urn:ietf:params:oauth:grant-type:jwt-bearer"

type unicoConfig struct {
	Account     string
	TenantID    string
	PrivateKey  []byte
	APIBase     string
	TokenURL    string
	Flow        string
	CallbackURI string
}

type unicoClient struct {
	config      unicoConfig
	httpClient  *http.Client
	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
}

type upstreamError struct {
	message string
}

func (e *upstreamError) Error() string {
	return e.message
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type unicoProcess struct {
	ID              string `json:"id"`
	State           string `json:"state"`
	Result          string `json:"result"`
	UserRedirectURL string `json:"userRedirectUrl"`
	Person          struct {
		DUIValue     string `json:"duiValue"`
		FriendlyName string `json:"friendlyName"`
	} `json:"person"`
	Services []struct {
		Documents []struct {
			Document struct {
				Data map[string]any `json:"data"`
			} `json:"doc"`
		} `json:"documents"`
	} `json:"services"`
}

func unicoConfigFromEnv() unicoConfig {
	account := envOrDefault("UNICO_ACCOUNT", "megwin_prod_")
	tenantID := envOrDefault("UNICO_TENANT_ID", "52bb482c-7ca1-4bf8-b56e-7965a63b4304")
	keyFile := strings.TrimSpace(envOrDefault("UNICO_PRIVATE_KEY_FILE", "./secrets/unico-uat.pem"))
	var privateKey []byte
	if keyFile != "" {
		data, err := os.ReadFile(keyFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read Unico private key file: %v\n", err)
		} else {
			privateKey = data
		}
	}
	return unicoConfig{
		Account:     account,
		TenantID:    tenantID,
		PrivateKey:  privateKey,
		APIBase:     strings.TrimRight(envOrDefault("UNICO_API_BASE", "https://api.idcloud.uat.unico.app"), "/"),
		TokenURL:    envOrDefault("UNICO_TOKEN_URL", "https://identityhomolog.acesso.io/oauth2/token"),
		Flow:        envOrDefault("UNICO_FLOW", "idtrustdocs_r2-idgov"),
		CallbackURI: envOrDefault("UNICO_CALLBACK_URI", "http://localhost:8080/"),
	}
}

func newUnicoClient(config unicoConfig) *unicoClient {
	return &unicoClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *unicoClient) ready() bool {
	return c.configurationError() == ""
}

func (c *unicoClient) configurationError() string {
	if c.config.Flow != "idtrustdocs_r2-idgov" {
		return "UNICO_FLOW must be idtrustdocs_r2-idgov for the configured Mexico verification journey."
	}
	if strings.TrimSpace(c.config.Account) == "" || strings.TrimSpace(c.config.TenantID) == "" {
		return "Unico account and tenant settings are required."
	}
	_, keyErr := parseRSAPrivateKey(c.config.PrivateKey)
	if keyErr != nil {
		return "Set UNICO_PRIVATE_KEY_FILE to the UAT service-account RSA private key PEM."
	}
	if !validHTTPSURL(c.config.APIBase) || !validHTTPSURL(c.config.TokenURL) {
		return "Unico API and token endpoints must use HTTPS."
	}
	if !validCallbackURI(c.config.CallbackURI) {
		return "UNICO_CALLBACK_URI must be an HTTPS URL (localhost HTTP is allowed for local testing)."
	}
	return ""
}

func validCallbackURI(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "https" || (parsed.Scheme == "http" && parsed.Hostname() == "localhost")
}

func validHTTPSURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}

func (c *unicoClient) createProcess(ctx context.Context, clientReference string) (unicoProcess, error) {
	if !c.ready() {
		return unicoProcess{}, &upstreamError{message: c.configurationError()}
	}
	body := map[string]string{
		"flow":            c.config.Flow,
		"purpose":         "biometryonboarding",
		"clientReference": clientReference,
		"callbackUri":     c.config.CallbackURI,
	}
	var response struct {
		Process unicoProcess `json:"process"`
	}
	if err := c.call(ctx, http.MethodPost, "/client/v1/process", body, &response); err != nil {
		return unicoProcess{}, err
	}
	return response.Process, nil
}

func (c *unicoClient) getProcess(ctx context.Context, processID string) (unicoProcess, error) {
	var response struct {
		Process unicoProcess `json:"process"`
	}
	err := c.call(ctx, http.MethodGet, "/client/v1/process/"+url.PathEscape(processID), nil, &response)
	return response.Process, err
}

func identityFromProcess(process unicoProcess) identityData {
	identity := identityData{
		CURP:     process.Person.DUIValue,
		FullName: process.Person.FriendlyName,
	}
	nameParts := make(map[string]string)
	for _, service := range process.Services {
		for _, document := range service.Documents {
			for key, value := range document.Document.Data {
				text, ok := value.(string)
				if !ok {
					continue
				}
				switch normalizeFieldName(key) {
				case "curp", "curpnumero", "numerocurp":
					identity.CURP = text
				case "birthdate", "dateofbirth", "fechanacimiento", "datanascimento":
					identity.BirthDate = normalizeBirthDate(text)
				case "fullname", "nombre", "nombrecompleto", "nomecivil":
					identity.FullName = text
				case "nombres", "firstname", "givenname":
					nameParts["given"] = text
				case "apellidopaterno", "paternalsurname":
					nameParts["paternal"] = text
				case "apellidomaterno", "maternalsurname":
					nameParts["maternal"] = text
				}
			}
		}
	}
	if strings.TrimSpace(identity.FullName) == "" {
		identity.FullName = strings.TrimSpace(strings.Join([]string{
			nameParts["given"],
			nameParts["paternal"],
			nameParts["maternal"],
		}, " "))
	}
	return identity
}

func normalizeBirthDate(value string) string {
	for _, layout := range []string{time.RFC3339, "2006-01-02", "02/01/2006", "2006/01/02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("2006-01-02")
		}
	}
	return value
}

func normalizeFieldName(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u",
		"ã", "a", "õ", "o", " ", "", "_", "", "-", "",
	)
	return replacer.Replace(value)
}

func (c *unicoClient) call(ctx context.Context, method, path string, body any, result any) error {
	token, err := c.accessToken(ctx)
	if err != nil {
		return err
	}

	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode Unico request: %w", err)
		}
		requestBody = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.config.APIBase+path, requestBody)
	if err != nil {
		return fmt.Errorf("create Unico request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send Unico request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		errorBody, err := io.ReadAll(io.LimitReader(resp.Body, 16<<10))
		if err != nil {
			return fmt.Errorf("read Unico error response: %w", err)
		}
		return &upstreamError{message: describeUpstreamError(method, path, resp.StatusCode, resp.Header.Get("X-Request-ID"), errorBody)}
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(result); err != nil {
		return fmt.Errorf("decode Unico response: %w", err)
	}
	return nil
}

func describeUpstreamError(method, path string, status int, requestID string, body []byte) string {
	message := fmt.Sprintf("Unico returned HTTP %d for %s %s.", status, method, path)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err == nil {
		for _, name := range []string{"code", "error", "error_description", "message", "detail"} {
			value, ok := fields[name]
			if !ok {
				continue
			}
			var text string
			if json.Unmarshal(value, &text) != nil {
				continue
			}
			text = strings.Join(strings.Fields(text), " ")
			if len(text) > 240 {
				text = text[:240] + "..."
			}
			if text != "" {
				message += fmt.Sprintf(" %s=%q.", name, text)
			}
		}
	}
	if requestID != "" {
		message += fmt.Sprintf(" requestId=%q.", requestID)
	}
	return message
}

func (c *unicoClient) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return c.token, nil
	}

	privateKey, err := parseRSAPrivateKey(c.config.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("load Unico RS256 private key: %w", err)
	}
	now := time.Now().UTC()
	issuer := c.config.Account + "@" + c.config.TenantID + ".iam.acesso.io"
	assertion, err := signJWT(privateKey, map[string]any{
		"iss":   issuer,
		"aud":   audienceForTokenURL(c.config.TokenURL),
		"scope": "*",
		"iat":   now.Unix(),
		"exp":   now.Add(30 * time.Minute).Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("sign Unico authentication assertion: %w", err)
	}
	form := url.Values{
		"grant_type": {jwtBearerGrant},
		"assertion":  {assertion},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("create Unico token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send Unico token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", &upstreamError{message: fmt.Sprintf("Unico token endpoint returned HTTP %d.", resp.StatusCode)}
	}
	var token tokenResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&token); err != nil {
		return "", fmt.Errorf("decode Unico token response: %w", err)
	}
	if token.AccessToken == "" {
		return "", errors.New("Unico token response did not contain an access token")
	}
	lifetime := time.Duration(token.ExpiresIn) * time.Second
	if lifetime <= 0 || lifetime > time.Hour {
		lifetime = time.Hour
	}
	c.token = token.AccessToken
	c.tokenExpiry = time.Now().Add(lifetime - time.Minute)
	return c.token, nil
}

func audienceForTokenURL(tokenURL string) string {
	parsed, err := url.Parse(tokenURL)
	if err != nil {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func parseRSAPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("PEM block not found")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		privateKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("PEM key is not RSA")
		}
		return privateKey, nil
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("PEM key is not a supported RSA private key")
	}
	return key, nil
}

func signJWT(key *rsa.PrivateKey, claims map[string]any) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(encoded))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return encoded + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}
