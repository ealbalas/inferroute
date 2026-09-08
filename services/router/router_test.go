package main

import (
	"testing"
	"time"
)

func newTestRegistry(names ...string) *WorkerRegistry {
	r := &WorkerRegistry{workers: make(map[string]*Worker)}
	for _, name := range names {
		r.workers[name] = &Worker{Name: name, Address: name, Status: "unknown"}
	}
	return r
}

func TestWorker_FreshBreakerClosed(t *testing.T) {
	w := &Worker{}
	if w.isOpen() {
		t.Error("fresh worker should have circuit breaker closed")
	}
}

func TestCircuitBreaker_BelowThreshold(t *testing.T) {
	r := newTestRegistry("w1")
	for i := 0; i < circuitBreakerThreshold-1; i++ {
		r.RecordFailure("w1")
	}
	if r.workers["w1"].isOpen() {
		t.Errorf("breaker should be closed after %d failures (threshold is %d)",
			circuitBreakerThreshold-1, circuitBreakerThreshold)
	}
}

func TestCircuitBreaker_OpensAtThreshold(t *testing.T) {
	r := newTestRegistry("w1")
	for i := 0; i < circuitBreakerThreshold; i++ {
		r.RecordFailure("w1")
	}
	if !r.workers["w1"].isOpen() {
		t.Errorf("breaker should be open after %d failures", circuitBreakerThreshold)
	}
}

func TestCircuitBreaker_ResetOnHealthy(t *testing.T) {
	r := newTestRegistry("w1")
	for i := 0; i < circuitBreakerThreshold; i++ {
		r.RecordFailure("w1")
	}
	if !r.workers["w1"].isOpen() {
		t.Fatal("precondition: breaker should be open")
	}
	r.SetStatus("w1", "healthy", 0.0, 0)
	w := r.workers["w1"]
	if w.failures != 0 {
		t.Errorf("failures should be reset to 0, got %d", w.failures)
	}
	if w.isOpen() {
		t.Error("breaker should be closed after SetStatus healthy")
	}
}

func TestHealthy_ExcludesOpenBreaker(t *testing.T) {
	r := newTestRegistry("w1")
	r.workers["w1"].Status = "healthy"
	for i := 0; i < circuitBreakerThreshold; i++ {
		r.RecordFailure("w1")
	}
	if len(r.Healthy()) != 0 {
		t.Error("Healthy() should exclude a worker with open circuit breaker")
	}
}

func TestHealthy_IncludesExpiredBreaker(t *testing.T) {
	r := newTestRegistry("w1")
	w := r.workers["w1"]
	w.Status = "healthy"
	w.failures = circuitBreakerThreshold
	w.openUntil = time.Now().Add(-1 * time.Second) // already expired
	if len(r.Healthy()) != 1 {
		t.Error("Healthy() should include a worker whose circuit breaker has expired")
	}
}

func TestPickWeightedScore_AllBreakersOpen(t *testing.T) {
	r := newTestRegistry("w1", "w2")
	for name := range r.workers {
		w := r.workers[name]
		w.Status = "healthy"
		for i := 0; i < circuitBreakerThreshold; i++ {
			r.RecordFailure(name)
		}
	}
	if got := r.PickWeightedScore(); got != nil {
		t.Errorf("expected nil when all breakers open, got %s", got.Name)
	}
}

func TestPickWeightedScore_ReturnsBestWorker(t *testing.T) {
	r := newTestRegistry("high-load", "low-load")

	high := r.workers["high-load"]
	high.Status = "healthy"
	high.Load = 0.9
	high.AvgLatencyMS = 200.0

	low := r.workers["low-load"]
	low.Status = "healthy"
	low.Load = 0.1
	low.AvgLatencyMS = 50.0

	got := r.PickWeightedScore()
	if got == nil {
		t.Fatal("expected a worker, got nil")
	}
	if got.Name != "low-load" {
		t.Errorf("expected low-load to win, got %s", got.Name)
	}
}
