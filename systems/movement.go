package systems

import (
	"context"

	"github.com/code-cell/esive/components"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"
)

var movementTracer = otel.Tracer("systems/movement")

// Movement steps:
// 1. Figure out which chunks need processing, send a queue message for each
// 2. For each chunk:
// 2.1. Find all entities moving
// 2.2. Plan movements (if it moves to another chunk, save this entity somewhere for inter-chunk collisions check)
// 2.3. Handle collisions in-chunk
// 2.4. Save new positions (in-chunk only)
// 3. Handle inter-chunk collisions. If an entity lands on another chunk, it will only work if the end position is empty after handling in-chunk movement. So in-chunk has preference over inter-chunk.

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

	pos.X = newX
	pos.Y = newY

	registry.UpdateComponents(ctx, entity, pos)
	// TODO: Do something better with the moveable.
	err = s.visionSystem.HandleMovement(ctx, tick, entity, &components.Moveable{}, oldPos, newPos)
	if err != nil {
		panic(err)
	}
	geo.OnMovePosition(ctx, entity, oldPos, newPos)

	return err
}

func (m *MovementSystem) ChunksWithMovingEntities(parentContext context.Context) (map[int64]map[int64]struct{}, error) {
	ctx, span := movementTracer.Start(parentContext, "movement.ChunksWithMovingEntities")
	defer span.End()

	res := map[int64]map[int64]struct{}{}

	movingEntities, movingEntitiesExtras, err := registry.EntitiesWithComponentType(ctx, &components.Moveable{}, &components.Moveable{}, &components.Position{})
	if err != nil {
		return nil, err
	}

	for i, _ := range movingEntities {
		mov := movingEntitiesExtras[i][0].(*components.Moveable)
		pos := movingEntitiesExtras[i][1].(*components.Position)
		if mov.VelX == 0 && mov.VelY == 0 {
			continue
		}
		cx, cy := geo.Chunk(pos.X, pos.Y)
		ymap, found := res[cx]
		if !found {
			ymap = map[int64]struct{}{}
			res[cx] = ymap
		}
		ymap[cy] = struct{}{}
	}

	return res, nil
}

// MoveAllEntitiesInChunk performs all movements within a chunk, and returns the entities that move across chunks for further processing.
func (m *MovementSystem) MoveAllEntitiesInChunk(parentContext context.Context, chunkX, chunkY int64, tick int64) ([]components.Entity, error) {
	ctx, span := movementTracer.Start(parentContext, "movement.MoveAllEntitiesInChunk")
	defer span.End()

	errGr := &errgroup.Group{}
	res := []components.Entity{}

	entities, positions, extras, err := geo.FindInChunk(ctx, chunkX, chunkY, &components.Moveable{})
	if err != nil {
		return nil, err
	}

	// Phase 1: Plan movements

	plannedMovements := map[int64]map[int64]components.Entity{}
	plannedMovingEntities := map[components.Entity]int{}

	for i, entity := range entities {
		pos := positions[i]
		mov := extras[i][0].(*components.Moveable)

		newChunkX, newChunkY := geo.Chunk(pos.X+mov.VelX, pos.Y+mov.VelY)
		if newChunkX != chunkX || newChunkY != chunkY {
			// The entity is moving to a different chunk, we handle this later.
			// TODO: Other entities collide with this entity even though it might move out.
			//   This is currently a tradeoff introduced to favor concurrency without adding a lot of complexity to the system
			res = append(res, entity)
			continue
		}

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

	// Phase 2: Check collisions in chunk against non-moving entities.
	// At this point we've planned movements making sure only one entity moves to a given position.

	for i, entity := range entities {
		pos := positions[i]
		if _, found := plannedMovingEntities[entity]; found {
			// The target entity is moving too so we don't need to check anyting at this iteration.
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
		// There's an entity planning to move here. We make it stop and remove it from the plan
		errGr.Go(func() error {
			registry.UpdateComponents(ctx, movingEntity, &components.Moveable{})
			return m.visionSystem.HandleMovement(ctx, tick, movingEntity, &components.Moveable{}, pos, pos)
		})
		delete(plannedMovements[pos.X], pos.Y)
		delete(plannedMovingEntities, movingEntity)
	}

	// Phase 3: Apply movements

	for x, row := range plannedMovements {
		for y, entity := range row {
			entity := entity
			oldPos := positions[plannedMovingEntities[entity]]
			mov := extras[plannedMovingEntities[entity]][0].(*components.Moveable)
			newPos := &components.Position{X: x, Y: y}

			errGr.Go(func() error {
				registry.UpdateComponents(ctx, entity, newPos)
				err = m.visionSystem.HandleMovement(ctx, tick, entity, mov, oldPos, newPos)
				if err != nil {
					return err
				}
				geo.OnMovePosition(ctx, entity, oldPos, newPos)
				return nil
			})
		}
	}
	if err := errGr.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

func (m *MovementSystem) MoveEntitiesAcrossChunks(parentContext context.Context, entities []components.Entity, tick int64) error {
	ctx, span := movementTracer.Start(parentContext, "movement.MoveEntitiesAcrossChunks")
	defer span.End()

	for _, entity := range entities {
		pos := &components.Position{}
		mov := &components.Moveable{}
		if err := registry.LoadComponents(ctx, entity, pos, mov); err != nil {
			return err
		}

		targetEntities, _, _, err := geo.FindInRange(ctx, pos.X+mov.VelX, pos.Y+mov.VelY, 0)
		if err != nil {
			return err
		}
		if len(targetEntities) > 0 {
			// Something found in destination, can't move the entity.
			// We stop it too.
			registry.UpdateComponents(ctx, entity, &components.Moveable{})
			err = m.visionSystem.HandleMovement(ctx, tick, entity, &components.Moveable{}, pos, pos)
			if err != nil {
				panic(err)
			}
			continue
		}

		newPos := &components.Position{X: pos.X + mov.VelX, Y: pos.Y + mov.VelY}
		registry.UpdateComponents(ctx, entity, newPos)
		err = m.visionSystem.HandleMovement(ctx, tick, entity, mov, pos, newPos)
		if err != nil {
			panic(err)
		}
		geo.OnMovePosition(ctx, entity, pos, newPos)
	}
	return nil
}
