package main

import (
	"sync"
	"time"

	"github.com/inferroute/inferroute/services/router/algorithms"
)

// Worker represents a model worker known to the router.
type Worker struct {
	Name           string
	Address        string
	Status         string  // healthy | degraded | offline | unknown
	Load           float64 // 0.0–1.0
	ActiveRequests int64
	AvgLatencyMS   float64
	ErrorRate      float64
	LastSeen       time.Time

	// Circuit breaker state
	failures    int
	openUntil   time.Time
}

func (w *Worker) isOpen() bool {
	return w.failures >= circuitBreakerThreshold && time.Now().Before(w.openUntil)
}

const circuitBreakerThreshold = 5
const circuitBreakerCooldown = 30 * time.Second

// WorkerRegistry is a thread-safe store of all known workers.
type WorkerRegistry struct {
	mu      sync.RWMutex
	workers map[string]*Worker
	rr      algorithms.RoundRobin
}

func newRegistry(addresses []string) *WorkerRegistry {
	r := &WorkerRegistry{workers: make(map[string]*Worker)}
	for _, addr := range addresses {
		name := addr // use address as name until health check fills it in
		r.workers[name] = &Worker{Name: name, Address: addr, Status: "unknown"}
	}
	return r
}

// All returns a copy of all workers (no lock held by caller needed).
func (r *WorkerRegistry) All() []*Worker {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Worker, 0, len(r.workers))
	for _, w := range r.workers {
		cp := *w
		out = append(out, &cp)
	}
	return out
}

// Healthy returns workers that are healthy and whose circuit breaker is closed.
func (r *WorkerRegistry) Healthy() []*Worker {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Worker
	for _, w := range r.workers {
		if w.Status == "healthy" && !w.isOpen() {
			cp := *w
			out = append(out, &cp)
		}
	}
	return out
}

// SetStatus updates a worker's health information from the health checker.
func (r *WorkerRegistry) SetStatus(name, status string, load float64, activeReqs int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.workers[name]
	if !ok {
		return
	}
	w.Status = status
	w.Load = load
	w.ActiveRequests = activeReqs
	w.LastSeen = time.Now()
	if status == "healthy" {
		w.failures = 0
	}
}

// RecordFailure increments the circuit breaker counter for a worker.
func (r *WorkerRegistry) RecordFailure(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.workers[name]
	if !ok {
		return
	}
	w.failures++
	if w.failures >= circuitBreakerThreshold {
		w.openUntil = time.Now().Add(circuitBreakerCooldown)
	}
}

// PickRoundRobin selects the next healthy worker using round-robin.
func (r *WorkerRegistry) PickRoundRobin() *Worker {
	healthy := r.Healthy()
	idx := r.rr.Pick(len(healthy))
	if idx < 0 {
		return nil
	}
	return healthy[idx]
}

// PickWeightedScore selects the healthy worker with the lowest routing score.
func (r *WorkerRegistry) PickWeightedScore() *Worker {
	healthy := r.Healthy()
	if len(healthy) == 0 {
		return nil
	}

	maxLatency := 0.0
	for _, w := range healthy {
		if w.AvgLatencyMS > maxLatency {
			maxLatency = w.AvgLatencyMS
		}
	}

	stats := make([]algorithms.WorkerStats, len(healthy))
	for i, w := range healthy {
		stats[i] = algorithms.WorkerStats{
			Name:              w.Name,
			NormalizedLatency: w.AvgLatencyMS,
			Load:              w.Load,
			ErrorRate:         w.ErrorRate,
			Cost:              0.5, // placeholder; extend when cost data is available
		}
	}
	algorithms.NormalizeLatencies(stats, maxLatency)
	best := algorithms.BestWorker(stats)
	if best < 0 {
		return nil
	}
	return healthy[best]
}
