package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// workerHealth is the response from a worker's GET /health endpoint.
type workerHealth struct {
	Status     string  `json:"status"`
	Worker     string  `json:"worker"`
	ActiveReqs int64   `json:"active_requests"`
	Load       float64 `json:"load"`
}

// HealthChecker polls every worker periodically and updates WorkerRegistry.
type HealthChecker struct {
	registry *WorkerRegistry
	interval time.Duration
	client   *http.Client
	log      *slog.Logger
}

func newHealthChecker(registry *WorkerRegistry, interval time.Duration, log *slog.Logger) *HealthChecker {
	return &HealthChecker{
		registry: registry,
		interval: interval,
		client:   &http.Client{Timeout: 3 * time.Second},
		log:      log,
	}
}

// Run starts continuous health polling. Call in a goroutine.
func (h *HealthChecker) Run(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.checkAll()
		}
	}
}

func (h *HealthChecker) checkAll() {
	workers := h.registry.All()
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()
			h.checkOne(w)
		}(workers[i])
	}
	wg.Wait()
}

func (h *HealthChecker) checkOne(w *Worker) {
	resp, err := h.client.Get(w.Address + "/health")
	if err != nil {
		h.log.Warn("health check failed", "worker", w.Name, "err", err)
		h.registry.SetStatus(w.Name, "offline", 0, 0)
		return
	}
	defer resp.Body.Close()

	var health workerHealth
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		h.registry.SetStatus(w.Name, "degraded", 0, 0)
		return
	}
	h.registry.SetStatus(w.Name, "healthy", health.Load, health.ActiveReqs)
}
