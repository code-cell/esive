package client

import (
	"sync"
	"sync/atomic"
	"time"
)

type latencyTracker struct {
	mtx  sync.Mutex
	data []time.Duration
	i    int
	avg  time.Duration
}

func newLatencyTracker(length int) *latencyTracker {
	return &latencyTracker{
		data: make([]time.Duration, length),
	}
}

func (lt *latencyTracker) addLatency(l time.Duration) {
	lt.mtx.Lock()
	defer lt.mtx.Unlock()
	lt.data[lt.i] = l
	lt.i++
	if lt.i >= len(lt.data) {
		lt.i = 0
	}
	avg := time.Duration(0)
	n := int64(0)
	for _, d := range lt.data {
		if d != 0 {
			avg += d
			n++
		}
	}
	atomic.StoreInt64((*int64)(&lt.avg), int64(avg)/n)
}
