package systems

import (
	"context"

	"github.com/code-cell/esive/components"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var movementTracer = otel.Tracer("systems/movement")

type MovementSystem struct {
	visionSystem *VisionSystem
}

func NewMovementSystem(visionSystem *VisionSystem) *MovementSystem {
	return &MovementSystem{
		visionSystem: visionSystem,
	}
}

func (s *MovementSystem) Move(parentContext context.Context, entity components.Entity, offsetX, offsetY int64) error {
	ctx, span := movementTracer.Start(parentContext, "movement.Move")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
		attribute.Int64("offsetX", offsetX),
		attribute.Int64("offsetY", offsetY),
	)
	defer span.End()

	pos := &components.Position{}
	err := registry.LoadComponent(ctx, entity, pos)
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

	err = s.visionSystem.HandleMovement(ctx, entity, oldPos, newPos)
	if err != nil {
		panic(err)
	}
	registry.UpdateComponents(ctx, entity, pos)
	geo.OnMovePosition(ctx, entity, oldPos, newPos)

	return err
}
