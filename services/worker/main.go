package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

// Worker simulates an AI inference server. Configure via env vars:
//
//	WORKER_NAME        display name (e.g. "us-west-1")
//	WORKER_PORT        HTTP port (default 9001)
//	WORKER_LATENCY_MS  base processing latency in ms (default 100)
//	WORKER_ERROR_RATE  fraction 0.0–1.0 of requests to fail (default 0.02)

var (
	workerName  = getenv("WORKER_NAME", "worker")
	port        = getenv("WORKER_PORT", "9001")
	latencyMS   = parseInt(getenv("WORKER_LATENCY_MS", "100"))
	errorRate   = parseFloat(getenv("WORKER_ERROR_RATE", "0.02"))
	activeReqs  atomic.Int64
	totalReqs   atomic.Int64
	errorCount  atomic.Int64
	startTime   = time.Now()
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /generate", handleGenerate(logger))
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /metrics", handleMetrics)

	addr := ":" + port
	logger.Info("worker starting", "name", workerName, "addr", addr,
		"latency_ms", latencyMS, "error_rate", errorRate)

	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}

type generateRequest struct {
	Prompt   string `json:"prompt"`
	Model    string `json:"model"`
	Priority string `json:"priority"`
}

type generateResponse struct {
	RequestID  string  `json:"request_id"`
	Worker     string  `json:"worker"`
	Output     string  `json:"output"`
	Tokens     int     `json:"tokens"`
	LatencyMS  int64   `json:"latency_ms"`
	Cost       float64 `json:"cost"`
}

func handleGenerate(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeReqs.Add(1)
		defer activeReqs.Add(-1)
		totalReqs.Add(1)

		var req generateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Simulate error injection
		if rand.Float64() < errorRate {
			errorCount.Add(1)
			log.Warn("simulated error", "worker", workerName)
			http.Error(w, "upstream inference error", http.StatusServiceUnavailable)
			return
		}

		// Simulate variable processing time (±30% jitter)
		jitter := int(float64(latencyMS) * 0.3 * (rand.Float64()*2 - 1))
		delay := time.Duration(latencyMS+jitter) * time.Millisecond
		start := time.Now()
		time.Sleep(delay)
		elapsed := time.Since(start).Milliseconds()

		tokens := 50 + rand.Intn(200)
		cost := float64(tokens) * 0.000002

		resp := generateResponse{
			RequestID: fmt.Sprintf("req_%d", time.Now().UnixNano()),
			Worker:    workerName,
			Output:    fmt.Sprintf("[%s] Simulated response to: %s", workerName, req.Prompt),
			Tokens:    tokens,
			LatencyMS: elapsed,
			Cost:      cost,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

type healthResponse struct {
	Status     string  `json:"status"`
	Worker     string  `json:"worker"`
	ActiveReqs int64   `json:"active_requests"`
	Load       float64 `json:"load"`
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	active := activeReqs.Load()
	load := float64(active) / 50.0 // assume max 50 concurrent
	if load > 1.0 {
		load = 1.0
	}
	resp := healthResponse{
		Status:     "healthy",
		Worker:     workerName,
		ActiveReqs: active,
		Load:       load,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type metricsResponse struct {
	Worker          string  `json:"worker"`
	UptimeSeconds   float64 `json:"uptime_seconds"`
	TotalRequests   int64   `json:"total_requests"`
	ActiveRequests  int64   `json:"active_requests"`
	ErrorCount      int64   `json:"error_count"`
	ErrorRate       float64 `json:"error_rate"`
	AvgLatencyMS    int     `json:"avg_latency_ms"`
	CPUCores        int     `json:"cpu_cores"`
}

func handleMetrics(w http.ResponseWriter, _ *http.Request) {
	total := totalReqs.Load()
	errors := errorCount.Load()
	rate := 0.0
	if total > 0 {
		rate = float64(errors) / float64(total)
	}
	resp := metricsResponse{
		Worker:         workerName,
		UptimeSeconds:  time.Since(startTime).Seconds(),
		TotalRequests:  total,
		ActiveRequests: activeReqs.Load(),
		ErrorCount:     errors,
		ErrorRate:      rate,
		AvgLatencyMS:   latencyMS,
		CPUCores:       runtime.NumCPU(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
