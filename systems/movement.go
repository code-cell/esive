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
		// send movement because the movement system skips non-moving entities
		pos := &components.Position{}
		err := registry.LoadComponents(ctx, entity, pos)
		if err != nil {
			panic(err)
		}
		err = s.visionSystem.HandleMovement(ctx, tick, entity, mov, pos, pos)
	}
	return nil
}

func (s *MovementSystem) Teleport(parentContext context.Context, tick int64, entity components.Entity, newX, newY int64) error {
	ctx, span := movementTracer.Start(parentContext, "movement.Move")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
		attribute.Int64("newX", newX),
		attribute.Int64("newY", newY),
	)
	defer span.End()

	pos := &components.Position{}
	err := registry.LoadComponents(ctx, entity, pos)
	if err != nil {
		panic(err)
	}
	newPos := &components.Position{
		X: newX,
		Y: newY,
	}
	oldPos := &components.Position{
		X: pos.X,
		Y: pos.Y,
	}

	pos.X += newX
	pos.Y += newY

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
	ctx, span := movementTracer.Start(parentContext, "movement.MoveAllMoveables")
	defer span.End()

	// Phase 1: Plan movements
	movingEntities, movingEntitiesExtras, err := registry.EntitiesWithComponentType(ctx, &components.Moveable{}, &components.Moveable{}, &components.Position{})
	if err != nil {
		panic(err)
	}

	plannedMovements := map[int64]map[int64]components.Entity{}
	plannedMovingEntities := map[components.Entity]int{}

	for i, entity := range movingEntities {
		mov := movingEntitiesExtras[i][0].(*components.Moveable)
		pos := movingEntitiesExtras[i][1].(*components.Position)
		if mov.VelX == 0 && mov.VelY == 0 {
			continue
		}
		newX := pos.X + mov.VelX
		newY := pos.Y + mov.VelY

		_, found := plannedMovements[newX]
		if !found {
			plannedMovements[newX] = map[int64]components.Entity{}
		}
		_, found = plannedMovements[newX][newY]
		if found {
			// Another entity is moving there
			// TODO: Revert client action
			continue
		}
		plannedMovements[newX][newY] = entity
		plannedMovingEntities[entity] = i
	}

	// Phase 2: Check collisions
	allEntities, allExtras, err := registry.EntitiesWithComponentType(ctx, &components.Position{}, &components.Position{})
	if err != nil {
		panic(err)
	}

	for i, entity := range allEntities {
		pos := allExtras[i][0].(*components.Position)
		if _, found := plannedMovingEntities[entity]; found {
			// The entity is moving somewhere else so we don't need to check anyting at this iteration.
			// We'll check the collision when we process the other entity
			continue
		}

		_, found := plannedMovements[pos.X]
		if !found {
			// No entities are moving here
			continue
		}
		movingEntity, found := plannedMovements[pos.X][pos.Y]
		if !found {
			// No entities are moving here
			continue
		}
		registry.UpdateComponents(ctx, movingEntity, &components.Moveable{})
		err = s.visionSystem.HandleMovement(ctx, tick, movingEntity, &components.Moveable{}, movingEntitiesExtras[plannedMovingEntities[movingEntity]][1].(*components.Position), movingEntitiesExtras[plannedMovingEntities[movingEntity]][1].(*components.Position))
		if err != nil {
			panic(err)
		}
		delete(plannedMovements[pos.X], pos.Y)
		delete(plannedMovingEntities, movingEntity)
	}

	// Phase 3: Apply movements
	for x, row := range plannedMovements {
		for y, entity := range row {
			mov := movingEntitiesExtras[plannedMovingEntities[entity]][0].(*components.Moveable)
			oldPos := movingEntitiesExtras[plannedMovingEntities[entity]][1].(*components.Position)
			newPos := &components.Position{X: x, Y: y}

			registry.UpdateComponents(ctx, entity, newPos)
			err = s.visionSystem.HandleMovement(ctx, tick, entity, mov, oldPos, newPos)
			if err != nil {
				panic(err)
			}
			geo.OnMovePosition(ctx, entity, oldPos, newPos)
		}
	}

	return nil
}
