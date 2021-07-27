package main

import (
	"context"
	"sync"

	"github.com/code-cell/esive/actions"
	components "github.com/code-cell/esive/components"
	"github.com/code-cell/esive/queue"
	"github.com/code-cell/esive/systems"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type TickProcessor struct {
	logger *zap.Logger
	q      *queue.Queue

	actionsQueue *actions.ActionsQueue
	movement     *systems.MovementSystem
	vision       *systems.VisionSystem
}

func NewTickProcessor(logger *zap.Logger, q *queue.Queue, actionsQueue *actions.ActionsQueue, movement *systems.MovementSystem, vision *systems.VisionSystem) *TickProcessor {
	return &TickProcessor{
		logger:       logger.With(zap.String("service", "tick_processor")),
		q:            q,
		actionsQueue: actionsQueue,
		movement:     movement,
		vision:       vision,
	}
}

func (t *TickProcessor) Init() {
	go t.q.Consume("process-chunk-movements", "worker", &queue.ProcessChunkMovements{}, func(nm *nats.Msg, m proto.Message) {
		data := m.(*queue.ProcessChunkMovements)
		entities, err := t.movement.MoveAllEntitiesInChunk(context.Background(), data.ChunkX, data.ChunkY, data.Tick)
		if err != nil {
			panic(err)
		}
		ids := []int64{}
		for _, id := range entities {
			ids = append(ids, int64(id))
		}
		res := &queue.ProcessChunkMovementsRes{
			Entities: ids,
		}
		payload, err := proto.Marshal(res)
		if err != nil {
			panic(err)
		}
		nm.Respond(payload)
	})

	go t.q.Consume("tick", "actions", &queue.Tick{}, func(_ *nats.Msg, m proto.Message) {
		tickMessage := m.(*queue.Tick)
		t.actionsQueue.CallActions(tickMessage.Tick, context.Background())

		chunks, err := t.movement.ChunksWithMovingEntities(context.Background())
		if err != nil {
			panic(err)
		}

		across := []components.Entity{}
		acrossMtx := &sync.Mutex{}
		wg := &sync.WaitGroup{}

		for x, c := range chunks {
			for y := range c {
				x := x
				y := y
				wg.Add(1)
				go func() {
					entities, err := t.q.ProcessChunkMovements(context.Background(), tickMessage.Tick, x, y)
					if err != nil {
						panic(err)
					}
					acrossMtx.Lock()
					across = append(across, entities...)
					acrossMtx.Unlock()
					wg.Done()
				}()
			}
		}
		wg.Wait()
		if err := t.movement.MoveEntitiesAcrossChunks(context.Background(), across, tickMessage.Tick); err != nil {
			panic(err)
		}

		t.q.HandleTickServicesDone(context.Background(), tickMessage.Tick)
	})
}
