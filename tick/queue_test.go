package tick

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	tick := NewTick(10 * time.Millisecond)
	queue := NewQueue()
	tick.AddSubscriber(queue.onTick)

	called := false
	queue.QueueAction(2, func(ctx context.Context) {
		called = true
	})

	tick.tickOnce()
	time.Sleep(5 * time.Millisecond) //Wait a bit so we don't receive false negatives
	require.False(t, called)
	tick.tickOnce()
	require.Eventually(t, func() bool { return called }, 5*time.Millisecond, time.Millisecond)
}
