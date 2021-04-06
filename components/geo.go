package components

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var geoTracer = otel.Tracer("geo")

type Geo struct {
	registry  *Registry
	store     *RedisStore
	chunkSize int
	logger    *zap.Logger
}

func NewGeo(registry *Registry, store *RedisStore, chunkSize int, logger *zap.Logger) *Geo {
	g := &Geo{
		registry:  registry,
		store:     store,
		chunkSize: chunkSize,
		logger:    logger.With(zap.String("service", "geo")),
	}
	registry.OnCreateComponent(g.OnCreateComponent)
	registry.OnDeleteComponent(g.OnDeleteComponent)
	return g
}

func (g *Geo) OnCreateComponent(parentCtx context.Context, entity Entity, component proto.Message) {
	componentType := string(component.ProtoReflect().Descriptor().FullName().Name())
	logger := g.logger.With(zap.Int64("entity_id", int64(entity)), zap.String("component_type", componentType))

	ctx, span := geoTracer.Start(parentCtx, "OnCreateComponent")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
		attribute.String("component_type", componentType),
	)
	defer span.End()

	logger.Debug("registering new component")

	if componentType == "Position" {
		pos := component.(*Position)
		chunkX, chunkY := g.chunk(pos.X, pos.Y)
		idStr := strconv.FormatInt(int64(entity), 10)
		_, err := g.registry.redisStore.SAdd(ctx, g.key(chunkX, chunkY), idStr)
		if err != nil {
			g.logger.Error("error saving entity in chunk", zap.Error(err))
		}
	}
}

func (g *Geo) OnDeleteComponent(parentCtx context.Context, entity Entity, component proto.Message) {
	componentType := string(component.ProtoReflect().Descriptor().FullName().Name())
	logger := g.logger.With(zap.Int64("entity_id", int64(entity)), zap.String("component_type", componentType))

	ctx, span := geoTracer.Start(parentCtx, "OnDeleteComponent")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
		attribute.String("component_type", componentType),
	)
	defer span.End()

	logger.Debug("removing component")

	if componentType == "Position" {
		pos := component.(*Position)
		chunkX, chunkY := g.chunk(pos.X, pos.Y)
		idStr := strconv.FormatInt(int64(entity), 10)
		err := g.store.SRem(ctx, g.key(chunkX, chunkY), idStr)
		if err != nil {
			g.logger.Error("error removing entity from chunk", zap.Error(err))
		}
	}
}

func (g *Geo) OnMovePosition(parentCtx context.Context, entity Entity, old, new *Position) {
	logger := g.logger.With(zap.Int64("entity_id", int64(entity)), zap.Any("old", old), zap.Any("new", new))

	ctx, span := geoTracer.Start(parentCtx, "OnMovePosition")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
		attribute.Any("old", old),
		attribute.Any("new", new),
	)
	defer span.End()

	oldChunkX, oldChunkY := g.chunk(old.X, old.Y)
	newChunkX, newChunkY := g.chunk(new.X, new.Y)
	if oldChunkX != newChunkX || oldChunkY != newChunkY {
		logger.Debug("moving component to new chunk")
		idStr := strconv.FormatInt(int64(entity), 10)
		err := g.store.SRem(ctx, g.key(oldChunkX, oldChunkY), idStr)
		if err != nil {
			g.logger.Error("error removing entity from chunk", zap.Error(err))
		}
		_, err = g.registry.redisStore.SAdd(ctx, g.key(newChunkX, newChunkY), idStr)
		if err != nil {
			g.logger.Error("error saving entity in chunk", zap.Error(err))
		}
		logger.Debug("component moved to new chunk")
	} else {
		logger.Debug("the component moved within the same chunk")
	}
}

func (g *Geo) FindInRange(parentCtx context.Context, x, y int64, rng float32, extraComponents ...proto.Message) ([]Entity, []*Position, [][]proto.Message, error) {
	logger := g.logger.With(zap.Int64("x", x), zap.Int64("y", y), zap.Float32("range", rng))

	ctx, span := geoTracer.Start(parentCtx, "FindInRange")
	span.SetAttributes(
		attribute.Int64("x", x),
		attribute.Int64("y", y),
		attribute.Float64("range", float64(rng)),
	)
	defer span.End()

	logger.Debug("finding entities in range")
	chunksInRange := int64(math.Ceil(float64(rng) / float64(g.chunkSize)))
	originChunkX, originChunkY := g.chunk(x, y)

	entities := []Entity{}
	positions := []*Position{}
	extras := [][]proto.Message{}

	queryComponents := append([]proto.Message{&Position{}}, extraComponents...)

	for chunkX := originChunkX - chunksInRange; chunkX <= originChunkX+chunksInRange; chunkX++ {
		for chunkY := originChunkY - chunksInRange; chunkY <= originChunkY+chunksInRange; chunkY++ {
			logger.Debug("checking chunk", zap.Int64("chunkX", chunkX), zap.Int64("chunkY", chunkY))

			e, c, err := g.registry.LoadComponentsFromIndex(ctx, g.key(chunkX, chunkY), queryComponents...)
			if err != nil {
				logger.Error("error finding chunk members", zap.Error(err))
				return nil, nil, nil, err
			}
			for i, entity := range e {
				pos := c[i][0].(*Position)
				if Distance(x, y, pos.X, pos.Y) <= rng {
					entities = append(entities, entity)
					positions = append(positions, pos)
					extras = append(extras, c[i][1:])
				}
			}
		}
	}
	return entities, positions, extras, nil
}

func (g *Geo) chunk(x, y int64) (int64, int64) {
	return x / int64(g.chunkSize), y / int64(g.chunkSize)
}

func (g *Geo) key(chunkX, chunkY int64) string {
	return fmt.Sprintf("chunks:%d:%d", chunkX, chunkY)
}

func Distance(x1, y1, x2, y2 int64) float32 {
	return float32(math.Sqrt(
		math.Pow(float64(x2-x1), 2) + math.Pow(float64(y2-y1), 2),
	))
}
