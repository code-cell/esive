package systems

import (
	"context"
	"sync"

	"github.com/code-cell/esive/components"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var visionTracer = otel.Tracer("systems/vision")

type VisionSystemLookItem struct {
	ID    int64
	X     int64
	Y     int64
	Char  string
	Color int32
}

type VisionSystemUpdater interface {
	HandleVisibilityLostSight(components.Entity)
	HandleVisibilityUpdate(*VisionSystemLookItem)
}

type VisionSystem struct {
	updaters    map[components.Entity]VisionSystemUpdater
	updatersMtx sync.Mutex
}

func NewVisionSystem() *VisionSystem {
	return &VisionSystem{
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

	lookerPos := &components.Position{}
	err := registry.LoadComponent(ctx, entity, lookerPos)
	if err != nil {
		return nil, err
	}
	looker := &components.Looker{}
	err = registry.LoadComponent(ctx, entity, looker)
	if err != nil {
		return nil, err
	}

	entitiesInRange, positions, extras, err := geo.FindInRange(ctx, lookerPos.X, lookerPos.Y, looker.Range, &components.Render{})
	if err != nil {
		return nil, err
	}

	res := make([]*VisionSystemLookItem, 0)
	for i, cmp := range entitiesInRange {
		pos := positions[i]
		render := extras[i][0].(*components.Render)
		res = append(res, &VisionSystemLookItem{
			X:     pos.X,
			Y:     pos.Y,
			ID:    int64(cmp),
			Char:  render.Char,
			Color: render.Color,
		})
	}

	return res, nil
}

func (s *VisionSystem) HandleMovement(parentContext context.Context, entity components.Entity, oldPos, newPos *components.Position) error {
	ctx, span := visionTracer.Start(parentContext, "vision.HandleMovement")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
	)
	defer span.End()

	render := &components.Render{}
	err := registry.LoadComponent(ctx, entity, render)
	if err != nil {
		if err == redis.Nil {
			// The entity has no renderer, nothing to do
			return nil
		}
		return err
	}

	lookers, extras, err := registry.EntitiesWithComponentType(ctx, &components.Looker{}, &components.Looker{}, &components.Position{})
	if err != nil {
		return err
	}
	for i, lookerE := range lookers {
		looker := extras[i][0].(*components.Looker)
		lookerPos := extras[i][1].(*components.Position)

		s.updatersMtx.Lock()
		updater, found := s.updaters[lookerE]
		s.updatersMtx.Unlock()
		if !found {
			continue
		}

		if lookerE != entity {
			// TODO: Store this data somehow, querying it all every time is a lot.
			oldDist := lookerPos.Distance(oldPos)
			newDist := lookerPos.Distance(newPos)
			if oldDist <= looker.Range && newDist > looker.Range {
				updater.HandleVisibilityLostSight(entity)
			} else if newDist <= looker.Range {
				updater.HandleVisibilityUpdate(&VisionSystemLookItem{
					ID:    int64(entity),
					X:     newPos.X,
					Y:     newPos.Y,
					Char:  render.Char,
					Color: render.Color,
				})
			}
		} else {
			oldEntities, _, _, err := geo.FindInRange(ctx, oldPos.X, oldPos.Y, looker.Range)
			if err != nil {
				return err
			}
			newEntities, newPositions, extras, err := geo.FindInRange(ctx, newPos.X, newPos.Y, looker.Range, &components.Render{})
			if err != nil {
				return err
			}
			oldIdx := map[components.Entity]struct{}{}
			newIdx := map[components.Entity]struct{}{}
			for _, e := range oldEntities {
				oldIdx[e] = struct{}{}
			}
			for _, e := range newEntities {
				newIdx[e] = struct{}{}
			}
			for _, oldEntity := range oldEntities {
				if _, foundInNew := newIdx[oldEntity]; !foundInNew {
					updater.HandleVisibilityLostSight(oldEntity)
				}
			}
			for i, newEntity := range newEntities {
				if _, foundInOld := oldIdx[newEntity]; !foundInOld {
					render := extras[i][0].(*components.Render)
					updater.HandleVisibilityUpdate(&VisionSystemLookItem{
						ID:    int64(newEntity),
						X:     newPositions[i].X,
						Y:     newPositions[i].Y,
						Char:  render.Char,
						Color: render.Color,
					})
				}
			}
		}
	}
	return nil
}

func (s *VisionSystem) HandleNewComponent(ctx context.Context, t string, entity components.Entity) error {
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
	err := registry.LoadComponent(ctx, entity, pos)
	if err != nil {
		return err
	}

	lookers, extras, err := registry.EntitiesWithComponentType(ctx, &components.Looker{}, &components.Looker{}, &components.Position{}, &components.Render{})
	if err != nil {
		return err
	}
	for i, lookerE := range lookers {
		if lookerE == entity {
			continue
		}
		looker := extras[i][0].(*components.Looker)
		lookerPos := extras[i][1].(*components.Position)
		lookerRenderer := extras[i][2].(*components.Render)

		dist := pos.Distance(lookerPos)

		s.updatersMtx.Lock()
		updater, found := s.updaters[lookerE]
		s.updatersMtx.Unlock()
		if !found {
			continue
		}

		if dist <= looker.Range {
			updater.HandleVisibilityUpdate(&VisionSystemLookItem{
				ID:    int64(entity),
				X:     pos.X,
				Y:     pos.Y,
				Char:  lookerRenderer.Char,
				Color: lookerRenderer.Color,
			})
		}
	}
	return nil
}

func (s *VisionSystem) HandleRemovedComponent(ctx context.Context, t string, entity components.Entity) error {
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
			updater.HandleVisibilityLostSight(entity)
		}
	}

	return nil
}
