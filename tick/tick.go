package tick

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Tick struct {
	Delay time.Duration

	current    int64
	adjustment int64

	ticker *time.Ticker
	done   chan struct{}

	subscribersMtx sync.Mutex
	subscribers    []func(context.Context, int64)
}

func NewTick(current int64, delay time.Duration) *Tick {
	return &Tick{
		Delay:       delay,
		current:     current,
		done:        make(chan struct{}),
		subscribers: make([]func(context.Context, int64), 0),
	}
}

func (tick *Tick) Current() int64 {
	return atomic.LoadInt64(&tick.current)
}

// Adjust makes the next tick be `newVal+1`. It doesn't change the current tick because other processes might need to be consistent with the tick number.
func (tick *Tick) Adjust(newVal int64) {
	atomic.SwapInt64(&tick.adjustment, newVal)
}

func (tick *Tick) Start() {
	tick.ticker = time.NewTicker(tick.Delay)

	for {
		select {
		case <-tick.done:
			return
		case <-tick.ticker.C:
			tick.tickOnce()
		}
	}
}

func (tick *Tick) tickOnce() {
	var next int64
	adj := atomic.LoadInt64(&tick.adjustment)
	if adj > 0 {
		next = adj + 1
		atomic.SwapInt64(&tick.current, next)
		atomic.SwapInt64(&tick.adjustment, 0)
	} else {
		next = atomic.AddInt64(&tick.current, 1)
	}

	now := time.Now()
	deadline := now.Add(tick.Delay)

	tick.subscribersMtx.Lock()
	for _, sub := range tick.subscribers {
		sub := sub
		go func() {
			ctx, cancel := context.WithDeadline(context.Background(), deadline)
			defer cancel()
			sub(ctx, next)
		}()
	}
	tick.subscribersMtx.Unlock()
}

func (tick *Tick) Stop() {
	tick.ticker.Stop()
	tick.done <- struct{}{}
}

func (tick *Tick) AddSubscriber(fn func(context.Context, int64)) {
	tick.subscribersMtx.Lock()
	defer tick.subscribersMtx.Unlock()
	tick.subscribers = append(tick.subscribers, fn)
}
