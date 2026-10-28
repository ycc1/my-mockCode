package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var curpPattern = regexp.MustCompile(`^[A-Z0-9]{18}$`)

type server struct {
	unico *unicoClient
	store *recordStore
}

type apiError struct {
	Error string `json:"error"`
}

type kycRecordSummary struct {
	ID        string    `json:"id"`
	ProcessID string    `json:"processId,omitempty"`
	Status    string    `json:"status"`
	Result    string    `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func main() {
	if err := loadDotEnv(".env"); err != nil {
		log.Fatalf("load local environment: %v", err)
	}

	listenAddr := envOrDefault("LISTEN_ADDR", "127.0.0.1:8080")
	dataDir := envOrDefault("KYCTEST_DATA_DIR", "./data")
	config := unicoConfigFromEnv()

	store, err := newRecordStore(filepath.Join(dataDir, "records.json"))
	if err != nil {
		log.Fatalf("initialize KYC record store: %v", err)
	}

	app := &server{
		unico: newUnicoClient(config),
		store: store,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.frontend)
	mux.HandleFunc("/healthz", app.health)
	mux.HandleFunc("/api/kyc/config", app.config)
	mux.HandleFunc("/api/kyc/process", app.createProcess)
	mux.HandleFunc("/api/kyc/process/", app.processRecord)
	mux.HandleFunc("/api/kyc/records", app.listRecords)

	log.Printf("KYCTest listening on http://%s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, withSecurityHeaders(mux)))
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		separator := strings.IndexByte(line, '=')
		if separator <= 0 {
			return fmt.Errorf("%s:%d: expected NAME=value", path, lineNumber)
		}
		name := strings.TrimSpace(line[:separator])
		value := strings.TrimSpace(line[separator+1:])
		if name == "" {
			return fmt.Errorf("%s:%d: environment variable name is empty", path, lineNumber)
		}
		if _, exists := os.LookupEnv(name); exists {
			continue
		}
		if err := os.Setenv(name, value); err != nil {
			return fmt.Errorf("%s:%d: set environment variable: %w", path, lineNumber, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	return nil
}

func (s *server) frontend(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/Megwin_KYC_New_UI.html" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	http.ServeFile(w, r, "Megwin_KYC_New_UI.html")
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) config(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"provider":    "Unico IDCloud",
		"environment": "UAT",
		"flow":        s.unico.config.Flow,
		"ready":       s.unico.ready(),
		"message":     s.unico.configurationError(),
	})
}

func (s *server) createProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<10)
	var request struct {
		Consent bool `json:"consent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "A valid JSON consent confirmation is required.")
		return
	}
	if !request.Consent {
		writeError(w, http.StatusBadRequest, "Consent is required before starting the liveness check.")
		return
	}

	id, err := newUUID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not create a KYC record.")
		return
	}
	now := time.Now().UTC()
	record := kycRecord{
		ID:        id,
		Status:    "SUBMITTING",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.save(record); err != nil {
		writeError(w, http.StatusInternalServerError, "Could not save the KYC record.")
		return
	}

	process, err := s.unico.createProcess(r.Context(), record.ID)
	if err != nil {
		record.Status = "FAILED"
		record.Error = err.Error()
		record.UpdatedAt = time.Now().UTC()
		if saveErr := s.store.save(record); saveErr != nil {
			log.Printf("save failed KYC record %s: %v", record.ID, saveErr)
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if process.ID == "" || !validUnicoRedirect(process.UserRedirectURL) {
		record.Status = "FAILED"
		record.Error = "Unico response did not include a process ID and secure redirect URL."
		record.UpdatedAt = time.Now().UTC()
		if saveErr := s.store.save(record); saveErr != nil {
			log.Printf("save incomplete KYC record %s: %v", record.ID, saveErr)
		}
		writeError(w, http.StatusBadGateway, record.Error)
		return
	}

	record.ProcessID = process.ID
	record.Status = process.State
	if record.Status == "" {
		record.Status = "PROCESS_STATE_CREATED"
	}
	record.Result = process.Result
	record.UpdatedAt = time.Now().UTC()
	if err := s.store.save(record); err != nil {
		writeError(w, http.StatusInternalServerError, "Unico created the process, but KYCTest could not save its record.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"record":          record,
		"userRedirectUrl": process.UserRedirectURL,
	})
}

func (s *server) processRecord(w http.ResponseWriter, r *http.Request) {
	recordID := strings.TrimPrefix(r.URL.Path, "/api/kyc/process/")
	if recordID == "" || strings.Contains(recordID, "/") {
		writeError(w, http.StatusBadRequest, "A KYC record ID is required.")
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.getProcess(w, r, recordID)
	case http.MethodPatch:
		s.updateIdentity(w, r, recordID)
	default:
		w.Header().Set("Allow", "GET, PATCH")
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed.")
	}
}

func (s *server) getProcess(w http.ResponseWriter, r *http.Request, recordID string) {
	record, ok := s.store.get(recordID)
	if !ok {
		writeError(w, http.StatusNotFound, "KYC record not found.")
		return
	}
	if record.ProcessID == "" {
		writeError(w, http.StatusConflict, "This KYC record does not have an Unico process to check.")
		return
	}

	process, err := s.unico.getProcess(r.Context(), record.ProcessID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	record.Status = process.State
	record.Result = process.Result
	if record.Status == "PROCESS_STATE_FINISHED" && !record.IdentityEdited {
		record.Identity = identityFromProcess(process)
	}
	record.UpdatedAt = time.Now().UTC()
	if err := s.store.save(record); err != nil {
		writeError(w, http.StatusInternalServerError, "Could not update the KYC record.")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (s *server) updateIdentity(w http.ResponseWriter, r *http.Request, recordID string) {
	record, ok := s.store.get(recordID)
	if !ok {
		writeError(w, http.StatusNotFound, "KYC record not found.")
		return
	}
	if record.Status != "PROCESS_STATE_FINISHED" {
		writeError(w, http.StatusConflict, "Identity details can only be edited after the Unico process is finished.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var identity identityData
	if err := json.NewDecoder(r.Body).Decode(&identity); err != nil {
		writeError(w, http.StatusBadRequest, "A valid JSON identity object is required.")
		return
	}
	identity.CURP = strings.ToUpper(strings.TrimSpace(identity.CURP))
	identity.BirthDate = strings.TrimSpace(identity.BirthDate)
	identity.FullName = strings.TrimSpace(identity.FullName)
	if !curpPattern.MatchString(identity.CURP) {
		writeError(w, http.StatusBadRequest, "CURP must contain exactly 18 letters or numbers.")
		return
	}
	if _, err := time.Parse("2006-01-02", identity.BirthDate); err != nil {
		writeError(w, http.StatusBadRequest, "Birth date must use YYYY-MM-DD format.")
		return
	}
	if identity.FullName == "" || len([]rune(identity.FullName)) > 120 {
		writeError(w, http.StatusBadRequest, "Full name is required and must be at most 120 characters.")
		return
	}
	record.Identity = identity
	record.IdentityEdited = true
	record.UpdatedAt = time.Now().UTC()
	if err := s.store.save(record); err != nil {
		writeError(w, http.StatusInternalServerError, "Could not save edited identity details.")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (s *server) listRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	records := s.store.list()
	summaries := make([]kycRecordSummary, 0, len(records))
	for _, record := range records {
		summaries = append(summaries, kycRecordSummary{
			ID:        record.ID,
			ProcessID: record.ProcessID,
			Status:    record.Status,
			Result:    record.Result,
			Error:     record.Error,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"records": summaries})
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func validUnicoRedirect(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" &&
		(parsed.Hostname() == "unico.app" || strings.HasSuffix(parsed.Hostname(), ".unico.app"))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, apiError{Error: message})
}

func methodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	writeError(w, http.StatusMethodNotAllowed, "Method not allowed.")
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func newUUID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	encoded := make([]byte, 32)
	hex.Encode(encoded, value[:])
	return string(encoded[0:8]) + "-" + string(encoded[8:12]) + "-" +
		string(encoded[12:16]) + "-" + string(encoded[16:20]) + "-" +
		string(encoded[20:32]), nil
}
