package systems

import (
	"context"
	"sync"

	"github.com/code-cell/esive/components"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var movementTracer = otel.Tracer("systems/movement")

type MovementSystem struct {
	visionSystem *VisionSystem

	forceVelocityUpdatesMtx sync.Mutex
	forceVelocityUpdates    map[int64][]components.Entity
}

func NewMovementSystem(visionSystem *VisionSystem) *MovementSystem {
	return &MovementSystem{
		visionSystem:         visionSystem,
		forceVelocityUpdates: map[int64][]components.Entity{},
	}
}

func (s *MovementSystem) SetVelocity(parentContext context.Context, tick int64, entity components.Entity, velX, velY int64) error {
	ctx, span := movementTracer.Start(parentContext, "movement.SetVelocity")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
		attribute.Int64("velX", velX),
		attribute.Int64("velY", velY),
	)
	defer span.End()

	mov := &components.Moveable{
		VelX: velX,
		VelY: velY,
	}
	err := registry.UpdateComponents(ctx, entity, mov)
	if err != nil {
		panic(err)
	}
	if velX == 0 && velY == 0 {

		s.forceVelocityUpdatesMtx.Lock()
		defer s.forceVelocityUpdatesMtx.Unlock()
		l, found := s.forceVelocityUpdates[tick]
		if !found {
			l = []components.Entity{}
		}
		l = append(l, entity)
		s.forceVelocityUpdates[tick] = l

	}

	return nil
}

func (s *MovementSystem) Move(parentContext context.Context, tick int64, entity components.Entity, offsetX, offsetY int64) error {
	ctx, span := movementTracer.Start(parentContext, "movement.Move")
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
	// TODO: Do something better with the moveable.
	err = s.visionSystem.HandleMovement(ctx, tick, entity, &components.Moveable{}, oldPos, newPos)
	if err != nil {
		panic(err)
	}
	geo.OnMovePosition(ctx, entity, oldPos, newPos)

	return err
}

func (s *MovementSystem) MoveAllMoveables(parentContext context.Context, tick int64) error {
	// TODO: The core logic of this is duplicated
	ctx, span := movementTracer.Start(parentContext, "movement.MoveAll")
	defer span.End()

	entities, extras, err := registry.EntitiesWithComponentType(ctx, &components.Moveable{}, &components.Moveable{}, &components.Position{})
	if err != nil {
		panic(err)
	}

	for i, entity := range entities {
		mov := extras[i][0].(*components.Moveable)
		pos := extras[i][1].(*components.Position)
		if mov.VelX == 0 && mov.VelY == 0 {
			continue
		}

		newPos := &components.Position{
			X: pos.X + mov.VelX,
			Y: pos.Y + mov.VelY,
		}
		oldPos := &components.Position{
			X: pos.X,
			Y: pos.Y,
		}

		pos.X += mov.VelX
		pos.Y += mov.VelY

		registry.UpdateComponents(ctx, entity, pos)
		err = s.visionSystem.HandleMovement(ctx, tick, entity, mov, oldPos, newPos)
		if err != nil {
			panic(err)
		}
		geo.OnMovePosition(ctx, entity, oldPos, newPos)
	}

	s.forceVelocityUpdatesMtx.Lock()

	l, found := s.forceVelocityUpdates[tick]
	if !found {
		s.forceVelocityUpdatesMtx.Unlock()
		return nil
	}
	delete(s.forceVelocityUpdates, tick)
	s.forceVelocityUpdatesMtx.Unlock()

	// TODO: This is O(n), not ideal
	for _, entity := range l {
		pos := &components.Position{}
		mov := &components.Moveable{}
		registry.LoadComponents(ctx, entity, pos, mov)

		newPos := &components.Position{
			X: pos.X + mov.VelX,
			Y: pos.Y + mov.VelY,
		}
		oldPos := &components.Position{
			X: pos.X,
			Y: pos.Y,
		}

		pos.X += mov.VelX
		pos.Y += mov.VelY

		registry.UpdateComponents(ctx, entity, pos)
		err := s.visionSystem.HandleMovement(ctx, tick, entity, mov, oldPos, newPos)
		if err != nil {
			panic(err)
		}
		geo.OnMovePosition(ctx, entity, oldPos, newPos)
	}

	return nil
}
