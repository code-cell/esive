package tick

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTick(t *testing.T) {
	tick := NewTick(0, 100*time.Millisecond)
	require.Equal(t, int64(0), tick.Current())

	go tick.Start()
	time.Sleep(201 * time.Millisecond) // 2 ticks
	tick.Stop()

	require.Equal(t, int64(2), tick.Current())
}

func TestTick_Adjust_SlowDown(t *testing.T) {
	tick := NewTick(0, 100*time.Millisecond)
	require.Equal(t, int64(0), tick.Current())

	durations := []time.Duration{}
	lastTick := time.Now()
	tick.AddSubscriber(func(c context.Context, i int64) {
		durations = append(durations, time.Since(lastTick))
		lastTick = time.Now()
		if i == 2 {
			tick.AdjustOnce(1) // slow towards tick 1 by 1/4 once
		}
	})

	go tick.Start()

	time.Sleep(400 * time.Millisecond) // 4 ticks
	tick.Stop()

	require.InDelta(t, int64(100), durations[0].Milliseconds(), float64(5))
	require.InDelta(t, int64(100), durations[1].Milliseconds(), float64(5))
	require.InDelta(t, int64(125), durations[2].Milliseconds(), float64(5))
}

func TestTick_Adjust_SpeedUp(t *testing.T) {
	tick := NewTick(0, 100*time.Millisecond)
	require.Equal(t, int64(0), tick.Current())

	durations := []time.Duration{}
	lastTick := time.Now()
	tick.AddSubscriber(func(c context.Context, i int64) {
		durations = append(durations, time.Since(lastTick))
		lastTick = time.Now()
		if i == 2 {
			tick.AdjustOnce(3) // speed up towards tick 3 by 1/4 once
		}
	})

	go tick.Start()

	time.Sleep(400 * time.Millisecond) // 4 ticks
	tick.Stop()

	require.InDelta(t, int64(100), durations[0].Milliseconds(), float64(5))
	require.InDelta(t, int64(100), durations[1].Milliseconds(), float64(5))
	require.InDelta(t, int64(75), durations[2].Milliseconds(), float64(5))
}

func TestTickSubscribers(t *testing.T) {
	tick := NewTick(0, 100*time.Millisecond)
	called := false
	tick.AddSubscriber(func(ctx context.Context, tick int64) {
		require.Equal(t, int64(1), tick)
		dl, ok := ctx.Deadline()
		require.True(t, ok)
		require.WithinDuration(t, time.Now().Add(100*time.Millisecond), dl, 20*time.Millisecond)
		called = true
	})
	go tick.Start()
	require.Eventually(t, func() bool {
		return called
	}, 200*time.Millisecond, 5*time.Millisecond)
	tick.Stop()
}

func TestTickSubscribers_TakeTooLong_CallsAllTicksAnyway(t *testing.T) {
	tick := NewTick(0, 100*time.Millisecond)
	called := int32(0)
	tick.AddSubscriber(func(ctx context.Context, tick int64) {
		atomic.AddInt32(&called, 1)
		time.Sleep(1000 * time.Millisecond)
	})
	go tick.Start()
	time.Sleep(201 * time.Millisecond) // Force 2 ticks
	tick.Stop()
	require.Equal(t, int32(2), called)
}
