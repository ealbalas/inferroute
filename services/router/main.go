package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/inferroute/inferroute/internal/observability"
)

func main() {
	log := observability.NewLogger()
	ctx := context.Background()

	port := getEnv("ROUTER_PORT", "8081")
	addrs := strings.Split(getEnv("WORKER_ADDRESSES", "http://localhost:9001"), ",")
	algorithm := getEnv("ROUTING_ALGORITHM", "weighted_score") // round_robin | weighted_score

	registry := newRegistry(addrs)

	checker := newHealthChecker(registry, 5*time.Second, log)
	go checker.Run(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /route", handleRoute(registry, algorithm, log))
	mux.HandleFunc("GET /workers", handleWorkers(registry))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	log.Info("router starting", "port", port, "algorithm", algorithm, "workers", addrs)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error("router error", "err", err)
		os.Exit(1)
	}
}

type routeRequest struct {
	Prompt   string `json:"prompt"`
	Model    string `json:"model"`
	Priority string `json:"priority"` // latency | cost | balanced
}

type routeResponse struct {
	WorkerName    string  `json:"worker_name"`
	WorkerAddress string  `json:"worker_address"`
	Algorithm     string  `json:"algorithm"`
	Score         float64 `json:"score,omitempty"`
}

func handleRoute(registry *WorkerRegistry, algorithm string, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req routeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Priority hint overrides default algorithm
		algo := algorithm
		if req.Priority == "latency" {
			algo = "weighted_score"
		}

		var selected *Worker
		switch algo {
		case "round_robin":
			selected = registry.PickRoundRobin()
		default:
			selected = registry.PickWeightedScore()
		}

		if selected == nil {
			log.Warn("no healthy workers available")
			http.Error(w, "no healthy workers available", http.StatusServiceUnavailable)
			return
		}

		resp := routeResponse{
			WorkerName:    selected.Name,
			WorkerAddress: selected.Address,
			Algorithm:     algo,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func handleWorkers(registry *WorkerRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		workers := registry.All()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(workers)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
