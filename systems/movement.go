package systems

import (
	"context"

	"github.com/code-cell/esive/components"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var movementTracer = otel.Tracer("systems/movement")

type queuedMovement struct {
	offsetX int64
	offsetY int64

	span trace.Span
}

type MovementSystem struct {
	visionSystem *VisionSystem
}

func NewMovementSystem(visionSystem *VisionSystem) *MovementSystem {
	return &MovementSystem{
		visionSystem: visionSystem,
	}
}

func (s *MovementSystem) Move(parentContext context.Context, tick int64, entity components.Entity, offsetX, offsetY int64) error {
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

// func (s *MovementSystem) Teleport(parentContext context.Context, entity components.Entity, x, y int64) error {
// 	ctx, span := movementTracer.Start(parentContext, "movement.Teleport")
// 	span.SetAttributes(
// 		attribute.Int64("entity_id", int64(entity)),
// 		attribute.Int64("x", x),
// 		attribute.Int64("y", y),
// 	)
// 	defer span.End()

// 	pos := &components.Position{}
// 	err := registry.LoadComponents(ctx, entity, pos)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return s.QueueMoveNextTick(ctx, entity, x-pos.X, y-pos.Y)
// }
