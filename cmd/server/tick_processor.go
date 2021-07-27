package main

import (
	"context"

	"github.com/code-cell/esive/actions"
	components "github.com/code-cell/esive/components"
	"github.com/code-cell/esive/queue"
	"github.com/code-cell/esive/systems"
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
	go t.q.Consume("tick", "actions", &queue.Tick{}, func(m proto.Message) {
		tickMessage := m.(*queue.Tick)
		t.actionsQueue.CallActions(tickMessage.Tick, context.Background())

		chunks, err := t.movement.ChunksWithMovingEntities(context.Background())
		if err != nil {
			panic(err)
		}

		across := []components.Entity{}
		for x, c := range chunks {
			for y := range c {
				entities, err := t.movement.MoveAllEntitiesInChunk(context.Background(), x, y, tickMessage.Tick)
				if err != nil {
					panic(err)
				}
				across = append(across, entities...)
			}
		}
		if err := t.movement.MoveEntitiesAcrossChunks(context.Background(), across, tickMessage.Tick); err != nil {
			panic(err)
		}

		t.q.HandleTickServicesDone(context.Background(), tickMessage.Tick)
	})
}
