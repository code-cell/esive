package tick

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTick(t *testing.T) {
	tick := NewTick(100 * time.Millisecond)
	require.Equal(t, int64(0), tick.Current())

	go tick.Start()
	time.Sleep(201 * time.Millisecond) // Force 2 ticks
	tick.Stop()

	require.Equal(t, int64(2), tick.Current())
}

func TestTickSubscribers(t *testing.T) {
	tick := NewTick(100 * time.Millisecond)
	startAt := time.Now()
	called := false
	tick.AddSubscriber(func(ctx context.Context, tick int64, dt time.Duration) {
		require.Equal(t, int64(1), tick)
		require.WithinDuration(t, time.Now(), startAt.Add(dt), 2*time.Millisecond)
		dl, ok := ctx.Deadline()
		require.True(t, ok)
		require.WithinDuration(t, time.Now().Add(100*time.Millisecond), dl, 20*time.Millisecond)
		called = true
	})
	go tick.Start()
	time.Sleep(101 * time.Millisecond) // Force 1 tick
	tick.Stop()
	require.True(t, called)
}

func TestTickSubscribers_TakeTooLong_CallsAllTicksAnyway(t *testing.T) {
	tick := NewTick(100 * time.Millisecond)
	called := int32(0)
	tick.AddSubscriber(func(ctx context.Context, tick int64, dt time.Duration) {
		atomic.AddInt32(&called, 1)
		time.Sleep(1000 * time.Millisecond)
	})
	go tick.Start()
	time.Sleep(201 * time.Millisecond) // Force 2 ticks
	tick.Stop()
	require.Equal(t, int32(2), called)
}
