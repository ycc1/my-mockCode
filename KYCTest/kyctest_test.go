package main

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadDotEnvLoadsMissingVariablesAndPreservesExistingValues(t *testing.T) {
	t.Setenv("KYCTEST_DOTENV_LOADED", "")
	if err := os.Unsetenv("KYCTEST_DOTENV_LOADED"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KYCTEST_DOTENV_OVERRIDE", "from-shell")
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("# local config\nKYCTEST_DOTENV_LOADED=from-file\nKYCTEST_DOTENV_OVERRIDE=from-file\n"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := loadDotEnv(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("KYCTEST_DOTENV_LOADED"); got != "from-file" {
		t.Fatalf("loaded value = %q, want from-file", got)
	}
	if got := os.Getenv("KYCTEST_DOTENV_OVERRIDE"); got != "from-shell" {
		t.Fatalf("existing environment value = %q, want from-shell", got)
	}
}

func TestDescribeUpstreamErrorIncludesSafeDiagnostics(t *testing.T) {
	got := describeUpstreamError(http.MethodPost, "/client/v1/process", http.StatusBadRequest, "req-123", []byte(`{"code":"FLOW_NOT_ENABLED","message":"Flow is not enabled for this account."}`))
	for _, want := range []string{
		"HTTP 400",
		"POST /client/v1/process",
		`code="FLOW_NOT_ENABLED"`,
		`message="Flow is not enabled for this account."`,
		`requestId="req-123"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("diagnostic %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "access_token") || strings.Contains(got, "assertion") {
		t.Fatalf("diagnostic contains authentication material: %q", got)
	}
}

func TestRecordStorePersistsRecordsWithoutIdentityFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "records.json")
	store, err := newRecordStore(path)
	if err != nil {
		t.Fatal(err)
	}
	record := kycRecord{
		ID:        "record-1",
		ProcessID: "process-1",
		Status:    "PROCESS_STATE_CREATED",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := store.save(record); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "CURP") || strings.Contains(string(data), "duiValue") {
		t.Fatalf("record store persisted identity fields: %s", data)
	}
	reopened, err := newRecordStore(path)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := reopened.get(record.ID)
	if !ok || got.ProcessID != record.ProcessID || got.Status != record.Status {
		t.Fatalf("reopened record = %#v, found = %v", got, ok)
	}
}

func TestUnicoCreateAndGetProcess(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	mux := http.NewServeMux()
	var upstream *httptest.Server
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("unexpected token request: %s %s", r.Method, r.Header.Get("Content-Type"))
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse token form: %v", err)
			return
		}
		if r.Form.Get("grant_type") != jwtBearerGrant {
			t.Errorf("grant_type = %q", r.Form.Get("grant_type"))
		}
		claims := verifyAssertion(t, r.Form.Get("assertion"), &privateKey.PublicKey)
		if claims["iss"] != "acct@tenant.iam.acesso.io" ||
			claims["aud"] != "https://"+r.Host || claims["scope"] != "*" {
			t.Errorf("unexpected JWT claims: %#v", claims)
		}
		_, _ = io.WriteString(w, `{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`)
	})
	mux.HandleFunc("/client/v1/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected process request: %s, auth %q", r.Method, r.Header.Get("Authorization"))
		}
		var payload map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode process payload: %v", err)
			return
		}
		var flow, purpose, clientReference, callbackURI string
		_ = json.Unmarshal(payload["flow"], &flow)
		_ = json.Unmarshal(payload["purpose"], &purpose)
		_ = json.Unmarshal(payload["clientReference"], &clientReference)
		_ = json.Unmarshal(payload["callbackUri"], &callbackURI)
		if flow != "multiaccounts_idgov" || purpose != "biometryonboarding" ||
			clientReference != "record-1" || callbackURI != "http://localhost:8080/" {
			t.Errorf("unexpected KYC process payload: %#v", payload)
		}
		if _, exists := payload["person"]; exists {
			t.Error("liveness process must not include a person identifier")
		}
		if _, exists := payload["document"]; exists {
			t.Error("process must not include a document before the Unico-hosted journey")
		}
		_, _ = io.WriteString(w, `{"process":{"id":"process-1","state":"PROCESS_STATE_CREATED","userRedirectUrl":"https://cadastro.uat.unico.app/flow"}}`)
	})
	mux.HandleFunc("/client/v1/process/process-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected status request: %s", r.Method)
		}
		_, _ = io.WriteString(w, `{"process":{"id":"process-1","state":"PROCESS_STATE_FINISHED","result":"PROCESS_RESULT_APPROVED","person":{"duiValue":"ABCD123456HDFXXX01","friendlyName":"Ana Perez"},"services":[{"documents":[{"doc":{"data":{"curp":"ABCD123456HDFXXX01","fechaNacimiento":"1990-04-15T00:00:00Z","nombreCompleto":"Ana Perez"}}}]}]}}`)
	})
	upstream = httptest.NewTLSServer(mux)
	defer upstream.Close()
	config := unicoConfig{
		Account:     "acct",
		TenantID:    "tenant",
		PrivateKey:  privateKeyPEM,
		APIBase:     upstream.URL,
		TokenURL:    upstream.URL + "/oauth2/token",
		Flow:        "multiaccounts_idgov",
		CallbackURI: "http://localhost:8080/",
	}
	client := newUnicoClient(config)
	client.httpClient = upstream.Client()
	if !client.ready() {
		t.Fatal("configured UAT client should be ready")
	}
	created, err := client.createProcess(context.Background(), "record-1")
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "process-1" || created.UserRedirectURL != "https://cadastro.uat.unico.app/flow" {
		t.Fatalf("created process = %#v", created)
	}
	checked, err := client.getProcess(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if checked.State != "PROCESS_STATE_FINISHED" || checked.Result != "PROCESS_RESULT_APPROVED" {
		t.Fatalf("checked process = %#v", checked)
	}
	identity := identityFromProcess(checked)
	if identity.CURP != "ABCD123456HDFXXX01" ||
		identity.BirthDate != "1990-04-15" || identity.FullName != "Ana Perez" {
		t.Fatalf("extracted identity = %#v", identity)
	}
}

func TestUnicoConfigurationRequiresMexicoVerificationFlow(t *testing.T) {
	config := unicoConfig{
		Account:     "acct",
		TenantID:    "tenant",
		APIBase:     "https://api.idcloud.uat.unico.app",
		TokenURL:    "https://identityhomolog.acesso.io/oauth2/token",
		Flow:        "idlive_r2",
		CallbackURI: "http://localhost:8080/",
	}
	client := newUnicoClient(config)
	if client.ready() {
		t.Fatal("unsupported Unico flow should not be accepted for this four-step journey")
	}
	if !strings.Contains(client.configurationError(), "multiaccounts_idgov") {
		t.Fatalf("configuration error = %q", client.configurationError())
	}
}

func TestIdentityCanBeEditedAfterUnicoFinishes(t *testing.T) {
	store, err := newRecordStore(filepath.Join(t.TempDir(), "records.json"))
	if err != nil {
		t.Fatal(err)
	}
	record := kycRecord{
		ID:        "record-1",
		ProcessID: "process-1",
		Status:    "PROCESS_STATE_FINISHED",
		Identity: identityData{
			CURP:      "ABCD123456HDFXXX01",
			BirthDate: "1990-04-15",
			FullName:  "Ana Perez",
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := store.save(record); err != nil {
		t.Fatal(err)
	}
	app := &server{store: store}
	request := httptest.NewRequest(http.MethodPatch, "/api/kyc/process/record-1", strings.NewReader(
		`{"curp":"wxyz123456hdfxxx02","birthDate":"1992-06-20","fullName":"Ana Perez Lopez"}`,
	))
	response := httptest.NewRecorder()
	app.processRecord(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("PATCH status=%d body=%s", response.Code, response.Body.String())
	}
	updated, ok := store.get(record.ID)
	if !ok || updated.Identity.CURP != "WXYZ123456HDFXXX02" ||
		updated.Identity.BirthDate != "1992-06-20" ||
		updated.Identity.FullName != "Ana Perez Lopez" || !updated.IdentityEdited {
		t.Fatalf("edited record = %#v", updated)
	}
}

func TestRecordListDoesNotExposeIdentityFields(t *testing.T) {
	store, err := newRecordStore(filepath.Join(t.TempDir(), "records.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.save(kycRecord{
		ID:     "record-1",
		Status: "PROCESS_STATE_FINISHED",
		Identity: identityData{
			CURP:      "ABCD123456HDFXXX01",
			BirthDate: "1990-04-15",
			FullName:  "Ana Perez",
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	app := &server{store: store}
	response := httptest.NewRecorder()
	app.listRecords(response, httptest.NewRequest(http.MethodGet, "/api/kyc/records", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("list status = %d", response.Code)
	}
	body := response.Body.String()
	for _, privateValue := range []string{"ABCD123456HDFXXX01", "1990-04-15", "Ana Perez"} {
		if strings.Contains(body, privateValue) {
			t.Fatalf("record list leaked identity value %q: %s", privateValue, body)
		}
	}
}

func verifyAssertion(t *testing.T, assertion string, publicKey *rsa.PublicKey) map[string]any {
	t.Helper()
	parts := strings.Split(assertion, ".")
	if len(parts) != 3 {
		t.Fatalf("invalid JWT assertion")
	}
	signingInput := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature); err != nil {
		t.Fatalf("verify JWT signature: %v", err)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	return claims
}

func TestValidUnicoRedirect(t *testing.T) {
	for _, value := range []string{
		"https://cadastro.uat.unico.app/flow",
		"https://unico.app/flow",
	} {
		if !validUnicoRedirect(value) {
			t.Errorf("expected valid Unico redirect: %s", value)
		}
	}
	for _, value := range []string{
		"http://cadastro.uat.unico.app/flow",
		"https://unico.app.attacker.example/flow",
		"https://example.com/",
	} {
		if validUnicoRedirect(value) {
			t.Errorf("expected invalid Unico redirect: %s", value)
		}
	}
}

func TestCreateProcessRequiresConsentAndRecordsLivenessAttempt(t *testing.T) {
	store, err := newRecordStore(filepath.Join(t.TempDir(), "records.json"))
	if err != nil {
		t.Fatal(err)
	}
	app := &server{
		unico: newUnicoClient(unicoConfig{}),
		store: store,
	}

	handler := httptest.NewRequest(http.MethodPost, "/api/kyc/process", strings.NewReader(`{"consent":false}`))
	handler.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.createProcess(response, handler)
	if response.Code != http.StatusBadRequest || len(store.list()) != 0 {
		t.Fatalf("without consent: status=%d records=%d", response.Code, len(store.list()))
	}

	handler = httptest.NewRequest(http.MethodPost, "/api/kyc/process", strings.NewReader(`{"consent":true}`))
	handler.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.createProcess(response, handler)
	if response.Code != http.StatusBadGateway {
		t.Fatalf("unconfigured UAT status = %d, body = %s", response.Code, response.Body.String())
	}
	records := store.list()
	if len(records) != 1 || records[0].Status != "FAILED" {
		t.Fatalf("liveness attempt record = %#v", records)
	}
	data, err := os.ReadFile(store.filePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "CURP") || strings.Contains(string(data), "document") {
		t.Fatalf("record contains CURP or document fields before Unico returns them: %s", data)
	}
}
