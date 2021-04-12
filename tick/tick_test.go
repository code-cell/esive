package tick

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTick(t *testing.T) {
	tick := NewTick(10 * time.Millisecond)
	require.Equal(t, int64(0), tick.Current())

	go tick.Start()
	time.Sleep(21 * time.Millisecond) // Force 2 ticks
	tick.Stop()

	require.Equal(t, int64(2), tick.Current())
}

func TestTickSubscribers(t *testing.T) {
	tick := NewTick(10 * time.Millisecond)
	startAt := time.Now()
	called := false
	tick.AddSubscriber(func(ctx context.Context, tick int64, dt time.Duration) {
		require.Equal(t, int64(1), tick)
		require.WithinDuration(t, time.Now(), startAt.Add(dt), 2*time.Millisecond)
		dl, ok := ctx.Deadline()
		require.True(t, ok)
		require.WithinDuration(t, time.Now().Add(10*time.Millisecond), dl, 2*time.Millisecond)
		called = true
	})
	go tick.Start()
	time.Sleep(11 * time.Millisecond) // Force 1 tick
	tick.Stop()
	require.True(t, called)
}

func TestTickSubscribers_TakeTooLong(t *testing.T) {
	// TODO: Test what happens when a subscriber takes to long to process. The ticker should still tick. We can assume the missbehaving call will not work because the context will be cancelled.
}
