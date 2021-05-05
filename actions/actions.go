package actions

import (
	"container/list"
	"context"
	"sync"
)

type Action func(context.Context)

type ActionsQueue struct {
	queueMtx  sync.Mutex
	queue     map[int64]*list.List
	inmediate *list.List
}

func NewActionsQueue() *ActionsQueue {
	return &ActionsQueue{
		queue:     make(map[int64]*list.List),
		inmediate: list.New(),
	}
}

func (q *ActionsQueue) QueueAction(action Action, tick int64) {
	q.queueMtx.Lock()
	defer q.queueMtx.Unlock()
	l, found := q.queue[tick]
	if !found {
		l = list.New()
		q.queue[tick] = l
	}
	l.PushBack(action)
}

func (q *ActionsQueue) QueueInmediate(action Action) {
	q.queueMtx.Lock()
	defer q.queueMtx.Unlock()
	q.inmediate.PushBack(action)
}

func (q *ActionsQueue) CallActions(tick int64, ctx context.Context) {
	q.queueMtx.Lock()
	defer q.queueMtx.Unlock()

	l, found := q.queue[tick]
	delete(q.queue, tick)

	if found {
		for e := l.Front(); e != nil; e = e.Next() {
			action := e.Value.(Action)
			action(ctx)
		}
	}
	for e := q.inmediate.Front(); e != nil; e = e.Next() {
		action := e.Value.(Action)
		action(ctx)
	}
	q.inmediate.Init()
}
