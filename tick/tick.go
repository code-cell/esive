package tick

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Tick struct {
	current int64

	delay  time.Duration
	ticker *time.Ticker
	done   chan struct{}

	subscribersMtx sync.Mutex
	subscribers    []func(context.Context, int64, time.Duration)
}

func NewTick(delay time.Duration) *Tick {
	return &Tick{
		delay:       delay,
		done:        make(chan struct{}),
		subscribers: make([]func(context.Context, int64, time.Duration), 0),
	}
}

func (tick *Tick) Current() int64 {
	return atomic.LoadInt64(&tick.current)
}

func (tick *Tick) Start() {
	tick.ticker = time.NewTicker(tick.delay)

	lastTickTime := time.Now()
	for {
		select {
		case <-tick.done:
			return
		case <-tick.ticker.C:
			current := atomic.AddInt64(&tick.current, 1)
			now := time.Now()
			dt := now.Sub(lastTickTime)
			deadline := now.Add(tick.delay)

			tick.subscribersMtx.Lock()
			for _, sub := range tick.subscribers {
				sub := sub
				go func() {
					ctx, cancel := context.WithDeadline(context.Background(), deadline)
					defer cancel()
					sub(ctx, current, dt)
				}()
			}
			tick.subscribersMtx.Unlock()
		}
	}
}

func (tick *Tick) Stop() {
	tick.ticker.Stop()
	tick.done <- struct{}{}
}

func (tick *Tick) AddSubscriber(fn func(context.Context, int64, time.Duration)) {
	tick.subscribersMtx.Lock()
	defer tick.subscribersMtx.Unlock()
	tick.subscribers = append(tick.subscribers, fn)
}
