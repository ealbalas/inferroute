package algorithms

import (
	"testing"
)

func TestScore_AllZero(t *testing.T) {
	got := Score(WorkerStats{})
	if got != 0.0 {
		t.Errorf("expected 0.0, got %f", got)
	}
}

func TestScore_AllOne(t *testing.T) {
	got := Score(WorkerStats{NormalizedLatency: 1.0, Load: 1.0, ErrorRate: 1.0, Cost: 1.0})
	const eps = 1e-9
	if got < 1.0-eps || got > 1.0+eps {
		t.Errorf("expected 1.0, got %v", got)
	}
}

func TestScore_WeightContributions(t *testing.T) {
	cases := []struct {
		name   string
		stats  WorkerStats
		expect float64
	}{
		{"latency only", WorkerStats{NormalizedLatency: 1.0}, 0.40},
		{"load only", WorkerStats{Load: 1.0}, 0.30},
		{"error rate only", WorkerStats{ErrorRate: 1.0}, 0.20},
		{"cost only", WorkerStats{Cost: 1.0}, 0.10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Score(tc.stats)
			if got != tc.expect {
				t.Errorf("expected %f, got %f", tc.expect, got)
			}
		})
	}
}

func TestBestWorker_Empty(t *testing.T) {
	if idx := BestWorker(nil); idx != -1 {
		t.Errorf("expected -1 for empty slice, got %d", idx)
	}
	if idx := BestWorker([]WorkerStats{}); idx != -1 {
		t.Errorf("expected -1 for empty slice, got %d", idx)
	}
}

func TestBestWorker_Single(t *testing.T) {
	if idx := BestWorker([]WorkerStats{{Load: 0.5}}); idx != 0 {
		t.Errorf("expected 0 for single-element slice, got %d", idx)
	}
}

func TestBestWorker_WinnerNotIndexZero(t *testing.T) {
	stats := []WorkerStats{
		{NormalizedLatency: 0.9, Load: 0.8, ErrorRate: 0.7, Cost: 0.6}, // index 0: high score
		{NormalizedLatency: 0.1, Load: 0.1, ErrorRate: 0.1, Cost: 0.1}, // index 1: low score - winner
		{NormalizedLatency: 0.5, Load: 0.5, ErrorRate: 0.5, Cost: 0.5}, // index 2: mid
	}
	if idx := BestWorker(stats); idx != 1 {
		t.Errorf("expected index 1 to win, got %d", idx)
	}
}

func TestNormalizeLatencies_ZeroMax(t *testing.T) {
	stats := []WorkerStats{
		{NormalizedLatency: 5.0},
		{NormalizedLatency: 3.0},
	}
	NormalizeLatencies(stats, 0)
	if stats[0].NormalizedLatency != 5.0 || stats[1].NormalizedLatency != 3.0 {
		t.Error("expected no-op when maxLatencyMS=0")
	}
}

func TestNormalizeLatencies_Proportional(t *testing.T) {
	stats := []WorkerStats{
		{NormalizedLatency: 100.0},
		{NormalizedLatency: 50.0},
		{NormalizedLatency: 25.0},
	}
	NormalizeLatencies(stats, 100.0)
	expect := []float64{1.0, 0.5, 0.25}
	for i, w := range stats {
		if w.NormalizedLatency != expect[i] {
			t.Errorf("stats[%d]: expected %f, got %f", i, expect[i], w.NormalizedLatency)
		}
	}
}

func TestNormalizeLatencies_ClampsAtOne(t *testing.T) {
	stats := []WorkerStats{
		{NormalizedLatency: 200.0},
	}
	NormalizeLatencies(stats, 100.0)
	if stats[0].NormalizedLatency != 1.0 {
		t.Errorf("expected 1.0 after clamp, got %f", stats[0].NormalizedLatency)
	}
}
