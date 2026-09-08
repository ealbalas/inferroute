package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/inferroute/inferroute/internal/observability"
)

var (
	requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "inferroute_requests_total",
		Help: "Total inference requests processed.",
	}, []string{"worker", "status"})

	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "inferroute_request_duration_ms",
		Help:    "Inference request latency in milliseconds.",
		Buckets: []float64{10, 25, 50, 100, 200, 500, 1000, 2000},
	}, []string{"worker", "cached"})

	cacheHits = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "inferroute_cache_hits_total",
		Help: "Total cache hits.",
	}, []string{"result"}) // hit | miss

	workerLoad = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "inferroute_worker_load",
		Help: "Current worker load 0.0–1.0.",
	}, []string{"worker"})
)

func main() {
	log := observability.NewLogger()

	prometheus.MustRegister(requestsTotal, requestDuration, cacheHits, workerLoad)

	port := getEnv("METRICS_PORT", "9100")

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Background: poll SQS for async metric events (stub — wire up AWS SDK in Phase 3)
	go pollSQS(context.Background(), log)

	log.Info("metrics service starting", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Error("metrics server error", "err", err)
		os.Exit(1)
	}
}

// pollSQS is a stub for the SQS consumer that will process async metric events.
// In Phase 3: use github.com/aws/aws-sdk-go-v2 to receive messages from the queue,
// unmarshal them, and update Prometheus counters / write rollups to PostgreSQL.
func pollSQS(ctx context.Context, log *slog.Logger) {
	log.Info("SQS poller started (stub — no-op until Phase 3)")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Debug("SQS poll tick (stub)")
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
