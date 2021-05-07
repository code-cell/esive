package actions

import (
	"container/list"
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var actionsTracer = otel.Tracer("actions/queue")

type Action func(context.Context)

type actionQueueItem struct {
	action     Action
	parentSpan trace.Span
	span       trace.Span
}

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

func (q *ActionsQueue) QueueAction(ctx context.Context, tick int64, action Action) {
	q.queueMtx.Lock()
	defer q.queueMtx.Unlock()
	l, found := q.queue[tick]
	if !found {
		l = list.New()
		q.queue[tick] = l
	}

	parentSpan := trace.SpanFromContext(ctx)
	_, span := actionsTracer.Start(ctx, "QueueAction")

	l.PushBack(&actionQueueItem{
		action:     action,
		span:       span,
		parentSpan: parentSpan,
	})
}

func (q *ActionsQueue) QueueInmediate(ctx context.Context, action Action) {
	q.queueMtx.Lock()
	defer q.queueMtx.Unlock()
	parentSpan := trace.SpanFromContext(ctx)
	_, span := actionsTracer.Start(ctx, "QueueInmediate")
	q.inmediate.PushBack(&actionQueueItem{
		action:     action,
		span:       span,
		parentSpan: parentSpan,
	})
}

func (q *ActionsQueue) CallActions(tick int64, ctx context.Context) {
	q.queueMtx.Lock()
	defer q.queueMtx.Unlock()

	l, found := q.queue[tick]
	delete(q.queue, tick)

	if found {
		for e := l.Front(); e != nil; e = e.Next() {
			actionItem := e.Value.(*actionQueueItem)
			actionItem.span.End()
			ctx := trace.ContextWithSpan(context.Background(), actionItem.parentSpan)
			actionItem.action(ctx)
		}
	}
	for e := q.inmediate.Front(); e != nil; e = e.Next() {
		actionItem := e.Value.(*actionQueueItem)
		actionItem.span.End()
		ctx := trace.ContextWithSpan(context.Background(), actionItem.parentSpan)
		actionItem.action(ctx)
	}
	q.inmediate.Init()
}
