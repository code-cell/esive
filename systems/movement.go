package systems

import (
	"context"
	"sync"

	"github.com/code-cell/esive/components"
	"github.com/code-cell/esive/queue"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

var movementTracer = otel.Tracer("systems/movement")

type queuedMovement struct {
	offsetX int64
	offsetY int64

	span trace.Span
}

type MovementSystem struct {
	visionSystem *VisionSystem

	queuedMovementsMtx sync.Mutex
	queuedMovements    map[int64]map[components.Entity]*queuedMovement
	nextMovements      map[components.Entity]*queuedMovement
}

func NewMovementSystem(visionSystem *VisionSystem) *MovementSystem {
	return &MovementSystem{
		visionSystem:    visionSystem,
		queuedMovements: make(map[int64]map[components.Entity]*queuedMovement),
		nextMovements:   make(map[components.Entity]*queuedMovement),
	}
}

func (s *MovementSystem) QueueMove(parentContext context.Context, entity components.Entity, tick, offsetX, offsetY int64) error {
	span := trace.SpanFromContext(parentContext)

	s.queuedMovementsMtx.Lock()
	defer s.queuedMovementsMtx.Unlock()
	perEntity, ok := s.queuedMovements[tick]
	if !ok {
		perEntity = make(map[components.Entity]*queuedMovement)
		s.queuedMovements[tick] = perEntity
	}
	perEntity[entity] = &queuedMovement{
		offsetX: offsetX,
		offsetY: offsetY,
		span:    span,
	}
	return nil
}

func (s *MovementSystem) QueueMoveNextTick(parentContext context.Context, entity components.Entity, offsetX, offsetY int64) error {
	span := trace.SpanFromContext(parentContext)

	s.queuedMovementsMtx.Lock()
	defer s.queuedMovementsMtx.Unlock()
	s.nextMovements[entity] = &queuedMovement{
		offsetX: offsetX,
		offsetY: offsetY,
		span:    span,
	}
	return nil
}

func (s *MovementSystem) doMove(parentContext context.Context, tick int64, entity components.Entity, offsetX, offsetY int64) error {
	ctx, span := movementTracer.Start(parentContext, "movement.doMove")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
		attribute.Int64("offsetX", offsetX),
		attribute.Int64("offsetY", offsetY),
	)
	defer span.End()

	pos := &components.Position{}
	err := registry.LoadComponents(ctx, entity, pos)
	if err != nil {
		panic(err)
	}
	newPos := &components.Position{
		X: pos.X + offsetX,
		Y: pos.Y + offsetY,
	}
	oldPos := &components.Position{
		X: pos.X,
		Y: pos.Y,
	}

	pos.X += offsetX
	pos.Y += offsetY

	registry.UpdateComponents(ctx, entity, pos)
	err = s.visionSystem.HandleMovement(ctx, tick, entity, oldPos, newPos)
	if err != nil {
		panic(err)
	}
	geo.OnMovePosition(ctx, entity, oldPos, newPos)

	return err
}

func (s *MovementSystem) Teleport(parentContext context.Context, entity components.Entity, x, y int64) error {
	ctx, span := movementTracer.Start(parentContext, "movement.Teleport")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
		attribute.Int64("x", x),
		attribute.Int64("y", y),
	)
	defer span.End()

	pos := &components.Position{}
	err := registry.LoadComponents(ctx, entity, pos)
	if err != nil {
		panic(err)
	}
	return s.QueueMoveNextTick(ctx, entity, x-pos.X, y-pos.Y)
}

func (s *MovementSystem) OnTick(message proto.Message) {
	tickMessage := message.(*queue.Tick)

	s.queuedMovementsMtx.Lock()
	movements, found := s.queuedMovements[tickMessage.Tick]
	if found {
		delete(s.queuedMovements, tickMessage.Tick)
	}
	nextMovements := s.nextMovements
	s.nextMovements = make(map[components.Entity]*queuedMovement)
	s.queuedMovementsMtx.Unlock()

	for entity, movement := range movements {
		ctx := trace.ContextWithSpan(context.Background(), movement.span)
		s.doMove(ctx, tickMessage.Tick, entity, movement.offsetX, movement.offsetY)
	}
	for entity, movement := range nextMovements {
		ctx := trace.ContextWithSpan(context.Background(), movement.span)
		s.doMove(ctx, tickMessage.Tick, entity, movement.offsetX, movement.offsetY)
	}
}
