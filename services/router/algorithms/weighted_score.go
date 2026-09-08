package algorithms

// WorkerStats contains the real-time metrics used by the scoring algorithm.
type WorkerStats struct {
	Name           string
	NormalizedLatency float64 // 0.0–1.0 relative to worst in pool
	Load           float64 // 0.0–1.0 CPU/connection load
	ErrorRate      float64 // 0.0–1.0
	Cost           float64 // 0.0–1.0 normalised relative cost
}

// Weights used by the scoring function. Lower score wins.
const (
	weightLatency   = 0.40
	weightLoad      = 0.30
	weightErrorRate = 0.20
	weightCost      = 0.10
)

// Score computes the routing score for a single worker.
// Lower is better — the router should pick the worker with the lowest score.
func Score(s WorkerStats) float64 {
	return weightLatency*s.NormalizedLatency +
		weightLoad*s.Load +
		weightErrorRate*s.ErrorRate +
		weightCost*s.Cost
}

// BestWorker returns the index of the worker with the lowest weighted score.
// Returns -1 if the slice is empty.
func BestWorker(stats []WorkerStats) int {
	if len(stats) == 0 {
		return -1
	}
	best := 0
	bestScore := Score(stats[0])
	for i := 1; i < len(stats); i++ {
		if s := Score(stats[i]); s < bestScore {
			best = i
			bestScore = s
		}
	}
	return best
}

// NormalizeLatencies sets NormalizedLatency relative to the pool maximum.
// Call this before passing stats to BestWorker.
func NormalizeLatencies(stats []WorkerStats, maxLatencyMS float64) {
	if maxLatencyMS == 0 {
		return
	}
	for i := range stats {
		stats[i].NormalizedLatency = stats[i].NormalizedLatency / maxLatencyMS
		if stats[i].NormalizedLatency > 1.0 {
			stats[i].NormalizedLatency = 1.0
		}
	}
}
