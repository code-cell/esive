package systems

import (
	"context"
	"sync"

	"github.com/code-cell/esive/components"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var visionTracer = otel.Tracer("systems/vision")

type VisionSystemLookItem struct {
	ID    int64
	X     int64
	Y     int64
	Char  string
	VelX  int64
	VelY  int64
	Color uint32
}

type VisionSystemUpdater interface {
	HandleVisibilityLostSight(components.Entity, int64)
	HandleTickUpdate(*VisionSystemLookItem, int64)
}

type VisionSystem struct {
	radius      int
	updaters    map[components.Entity]VisionSystemUpdater
	updatersMtx sync.Mutex
}

func NewVisionSystem(radius int) *VisionSystem {
	return &VisionSystem{
		radius:   radius,
		updaters: map[components.Entity]VisionSystemUpdater{},
	}
}

func (s *VisionSystem) AddUpdater(entity components.Entity, updater VisionSystemUpdater) error {
	s.updatersMtx.Lock()
	defer s.updatersMtx.Unlock()
	s.updaters[entity] = updater
	return nil
}

func (s *VisionSystem) LookAll(ctx context.Context, entity components.Entity) ([]*VisionSystemLookItem, error) {
	ctx, span := visionTracer.Start(ctx, "vision.LookAll")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
	)
	defer span.End()

	looker := &components.Looker{}
	lookerPos := &components.Position{}
	err := registry.LoadComponents(ctx, entity, looker, lookerPos)
	if err != nil {
		return nil, err
	}

	entitiesInRange, positions, extras, err := geo.FindInRange(ctx, lookerPos.X, lookerPos.Y, float32(s.radius), &components.Render{}, &components.Moveable{})
	if err != nil {
		return nil, err
	}

	res := make([]*VisionSystemLookItem, 0)
	for i, cmp := range entitiesInRange {
		pos := positions[i]
		render := extras[i][0].(*components.Render)
		mov := extras[i][1].(*components.Moveable)
		res = append(res, &VisionSystemLookItem{
			X:     pos.X,
			Y:     pos.Y,
			VelX:  mov.VelX,
			VelY:  mov.VelY,
			ID:    int64(cmp),
			Char:  render.Char,
			Color: render.Color,
		})
	}

	return res, nil
}

// TODO: Too iterative.
// LookAll process:
// 1. Find entities in range from the old position
// 2. Find entities in range from the new position
// 3. Figure out which ones lost visibility (old - new) and send LostSight to both parties (if the other is a looker)
// 4. For all new ones, send update of the new position to both parties too
func (s *VisionSystem) HandleMovement(parentContext context.Context, tick int64, entity components.Entity, mov *components.Moveable, oldPos, newPos *components.Position) error {
	ctx, span := visionTracer.Start(parentContext, "vision.HandleMovement")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
	)
	defer span.End()

	s.updatersMtx.Lock()
	updater, updaterFound := s.updaters[entity]
	s.updatersMtx.Unlock()

	errGr := &errgroup.Group{}

	render := &components.Render{}
	var oldEntities []components.Entity
	var newEntities []components.Entity
	var newEntitiesPos []*components.Position
	var newEntitiesExtras [][]protoreflect.ProtoMessage

	errGr.Go(func() error {
		err := registry.LoadComponents(ctx, entity, render)
		if err != nil && err == redis.Nil {
			// The entity has no renderer, nothing to do
			return nil
		}
		return err
	})
	errGr.Go(func() error {
		entities, _, _, err := geo.FindInRange(ctx, oldPos.X, oldPos.Y, float32(s.radius))
		if err != nil {
			return err
		}
		oldEntities = entities
		return nil
	})

	errGr.Go(func() error {
		entities, entitiesPos, entitiesExtras, err := geo.FindInRange(ctx, newPos.X, newPos.Y, float32(s.radius), &components.Render{}, &components.Moveable{})
		if err != nil {
			return err
		}
		newEntities = entities
		newEntitiesPos = entitiesPos
		newEntitiesExtras = entitiesExtras
		return nil
	})
	if err := errGr.Wait(); err != nil {
		return err
	}

	// Find lookers that lost visibility
	for _, oldEntity := range oldEntities {
		found := false
		for _, newEntity := range newEntities {
			if oldEntity == newEntity {
				found = true
				break
			}
		}
		if !found {
			if updaterFound {
				updater.HandleVisibilityLostSight(oldEntity, tick)
			}
			s.updatersMtx.Lock()
			externalUpdater, externalUpdaterFound := s.updaters[oldEntity]
			s.updatersMtx.Unlock()
			if externalUpdaterFound {
				externalUpdater.HandleVisibilityLostSight(entity, tick)
			}
		}
	}

	for i, newEntity := range newEntities {
		newEntityPos := newEntitiesPos[i]
		newEntityRender := newEntitiesExtras[i][0].(*components.Render)
		newEntityMov := newEntitiesExtras[i][1].(*components.Moveable)

		if updaterFound {
			updater.HandleTickUpdate(&VisionSystemLookItem{
				ID:    int64(newEntity),
				X:     newEntityPos.X,
				Y:     newEntityPos.Y,
				VelX:  newEntityMov.VelX,
				VelY:  newEntityMov.VelY,
				Char:  newEntityRender.Char,
				Color: newEntityRender.Color,
			}, tick)
		}

		s.updatersMtx.Lock()
		externalUpdater, externalUpdaterFound := s.updaters[newEntity]
		s.updatersMtx.Unlock()
		if externalUpdaterFound {
			externalUpdater.HandleTickUpdate(&VisionSystemLookItem{
				ID:    int64(entity),
				X:     newPos.X,
				Y:     newPos.Y,
				VelX:  mov.VelX,
				VelY:  mov.VelY,
				Char:  render.Char,
				Color: render.Color,
			}, tick)
		}
	}

	return nil
}

func (s *VisionSystem) HandleNewComponent(ctx context.Context, tick int64, t string, entity components.Entity) error {
	ctx, span := visionTracer.Start(ctx, "vision.HandleNewComponent")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
		attribute.String("type", t),
	)
	defer span.End()

	if t != "Position" {
		return nil
	}

	pos := &components.Position{}
	render := &components.Render{}
	err := registry.LoadComponents(ctx, entity, pos, render)
	if err != nil {
		return err
	}

	lookers, extras, err := registry.EntitiesWithComponentType(ctx, &components.Looker{}, &components.Position{})
	if err != nil {
		return err
	}
	for i, lookerE := range lookers {
		if lookerE == entity {
			continue
		}
		lookerPos := extras[i][0].(*components.Position)

		dist := pos.Distance(lookerPos)

		s.updatersMtx.Lock()
		updater, found := s.updaters[lookerE]
		s.updatersMtx.Unlock()
		if !found {
			continue
		}

		if dist <= float32(s.radius) {
			updater.HandleTickUpdate(&VisionSystemLookItem{
				ID:    int64(entity),
				X:     pos.X,
				Y:     pos.Y,
				Char:  render.Char,
				Color: render.Color,
			}, tick)
		}
	}
	return nil
}

func (s *VisionSystem) HandleRemovedComponent(ctx context.Context, tick int64, t string, entity components.Entity) error {
	ctx, span := visionTracer.Start(ctx, "vision.HandleRemovedComponent")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
		attribute.String("type", t),
	)
	defer span.End()

	if t != "Position" {
		return nil
	}

	// TODO: This broadcasts the lost component to all lookers in the game. Maybe we should do it just for the ones in sight, but the component is deleted already.
	lookerEntities, _, err := registry.EntitiesWithComponentType(ctx, &components.Looker{})
	if err != nil {
		return err
	}

	for _, lookerEntity := range lookerEntities {
		s.updatersMtx.Lock()
		updater, found := s.updaters[lookerEntity]
		s.updatersMtx.Unlock()
		if found {
			updater.HandleVisibilityLostSight(entity, tick)
		}
	}

	return nil
}
