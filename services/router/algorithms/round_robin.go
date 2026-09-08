package algorithms

import (
	"sync/atomic"
)

// RoundRobin distributes requests evenly across healthy workers in order.
type RoundRobin struct {
	counter atomic.Uint64
}

// Pick returns the index of the next worker to use from the healthy pool.
func (rr *RoundRobin) Pick(workerCount int) int {
	if workerCount == 0 {
		return -1
	}
	n := rr.counter.Add(1)
	return int((n - 1) % uint64(workerCount))
}
