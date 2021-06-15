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
	adjustment int32

	ticker *time.Ticker
	part   int8
	parts  int8
	done   chan struct{}

	subscribersMtx sync.Mutex
	subscribers    []func(context.Context, int64)
}

func NewTick(current int64, delay time.Duration) *Tick {
	return &Tick{
		Delay:       delay,
		current:     current,
		parts:       4,
		done:        make(chan struct{}),
		subscribers: make([]func(context.Context, int64), 0),
	}
}

func (tick *Tick) Current() int64 {
	return atomic.LoadInt64(&tick.current)
}

// AdjustOnce Speeds up or slows down (by 1/4th of the tick) the tick for one single click towards the desired newVal
func (tick *Tick) AdjustOnce(newVal int64) {
	current := tick.Current()
	if newVal == current {
		return
	}
	if newVal > current {
		atomic.StoreInt32(&tick.adjustment, 1) // Speed up
	} else {
		atomic.StoreInt32(&tick.adjustment, -1) // Slow down
	}
}

func (tick *Tick) Start() {
	tick.ticker = time.NewTicker(tick.Delay / time.Duration(tick.parts))
	for {
		select {
		case <-tick.done:
			return
		case <-tick.ticker.C:
			tick.part += 1 + int8(atomic.LoadInt32(&tick.adjustment))
			atomic.StoreInt32(&tick.adjustment, 0)
			if tick.part >= tick.parts {
				tick.part -= tick.parts
				tick.tickOnce()
			}
		}
	}
}

func (tick *Tick) tickOnce() {
	next := atomic.AddInt64(&tick.current, 1)
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
