package tick

import (
	"container/list"
	"context"
	"sync"
	"time"
)

type Queue struct {
	actionsMtx sync.Mutex
	actions    map[int64]*list.List
}

func NewQueue() *Queue {
	return &Queue{
		actions: make(map[int64]*list.List),
	}
}

func (q *Queue) QueueAction(tick int64, action func(context.Context)) {
	q.actionsMtx.Lock()
	defer q.actionsMtx.Unlock()
	l, found := q.actions[tick]
	if !found {
		l = list.New()
		q.actions[tick] = l
	}
	l.PushBack(action)
}

func (q *Queue) onTick(ctx context.Context, tick int64, dt time.Duration) {
	actions := q.popActions(tick)
	if actions == nil {
		return
	}
	for e := actions.Front(); e != nil; e = e.Next() {
		action := e.Value.(func(context.Context))
		action(ctx)
	}
}

func (q *Queue) popActions(tick int64) *list.List {
	q.actionsMtx.Lock()
	defer q.actionsMtx.Unlock()
	actions, found := q.actions[tick]
	if !found {
		return nil
	}
	delete(q.actions, tick)
	return actions
}
