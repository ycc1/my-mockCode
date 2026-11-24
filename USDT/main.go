package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ycc1/my-mockCode/USDT/internal/market"
	"github.com/ycc1/my-mockCode/USDT/internal/signal"
)

//go:embed web
var webFiles embed.FS

type response struct {
	Snapshot  *market.Snapshot `json:"snapshot"`
	Signal    signal.Plan      `json:"signal"`
	FetchedAt int64            `json:"fetchedAtMs"`
}

type monitor struct {
	client      *market.Client
	includeSpot bool
	cacheFor    time.Duration
	mu          sync.Mutex
	cached      *response
	cachedAt    time.Time
}

func environment(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func environmentInt(key string, fallback int) int {
	value := environment(key, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("%s 必須是整數", key)
	}
	return parsed
}

func (m *monitor) snapshotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	m.mu.Lock()
	if m.cached != nil && time.Since(m.cachedAt) < m.cacheFor {
		payload := m.cached
		m.mu.Unlock()
		writeJSON(w, http.StatusOK, payload)
		return
	}
	m.mu.Unlock()

	snapshot, err := m.client.Snapshot(m.includeSpot)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": "行情來源暫時無法使用；拒絕顯示舊報價。" + err.Error(),
		})
		return
	}
	payload := &response{
		Snapshot:  snapshot,
		Signal:    signal.Judge(snapshot),
		FetchedAt: time.Now().UnixMilli(),
	}
	m.mu.Lock()
	m.cached, m.cachedAt = payload, time.Now()
	m.mu.Unlock()
	writeJSON(w, http.StatusOK, payload)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func main() {
	symbol := strings.ToUpper(environment("SYMBOL", "ETHUSDT"))
	maxAge := time.Duration(environmentInt("MAX_QUOTE_AGE_SECONDS", 90)) * time.Second
	cacheFor := time.Duration(environmentInt("CACHE_SECONDS", 15)) * time.Second
	includeSpot := strings.EqualFold(environment("INCLUDE_SPOT", "true"), "true")
	monitor := &monitor{
		client:      market.NewClient(symbol, maxAge),
		includeSpot: includeSpot,
		cacheFor:    cacheFor,
	}

	webRoot, err := fs.Sub(webFiles, "web")
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/snapshot", monitor.snapshotHandler)
	mux.Handle("/", http.FileServer(http.FS(webRoot)))
	address := environment("ADDR", ":8080")
	log.Printf("ETH/USDT read-only monitor listening on http://localhost%s", address)
	log.Fatal(http.ListenAndServe(address, mux))
}
