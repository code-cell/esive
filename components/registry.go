package components

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const keyEntitiesIDSeq = "entity_id_seq"

var registryTracer = otel.Tracer("registry")

type Registry struct {
	redisStore *RedisStore
	logger     *zap.Logger

	onCreateComponent []func(context.Context, Entity, proto.Message)
	onDeleteComponent []func(context.Context, Entity, proto.Message)
}

func NewRegistry(redisStore *RedisStore, logger *zap.Logger) *Registry {
	return &Registry{
		redisStore:        redisStore,
		logger:            logger.With(zap.String("service", "registry")),
		onCreateComponent: make([]func(context.Context, Entity, proto.Message), 0),
		onDeleteComponent: make([]func(context.Context, Entity, proto.Message), 0),
	}
}

func (b *Registry) OnCreateComponent(cb func(context.Context, Entity, proto.Message)) {
	b.logger.Debug("registered onCreate callback")
	b.onCreateComponent = append(b.onCreateComponent, cb)
}

func (b *Registry) OnDeleteComponent(cb func(context.Context, Entity, proto.Message)) {
	b.logger.Debug("registered onDelete callback")
	b.onDeleteComponent = append(b.onDeleteComponent, cb)
}

func (b *Registry) NewEntity(parentCtx context.Context) (Entity, error) {
	b.logger.Debug("creating new entity")
	ctx, span := registryTracer.Start(parentCtx, "NewEntity")
	defer span.End()

	id, err := b.redisStore.NextInt64(ctx, keyEntitiesIDSeq)
	b.logger.Debug("created new entity", zap.Int64("entity_id", id))
	return Entity(id), err
}

func (b *Registry) CreateComponents(parentCtx context.Context, entity Entity, components ...proto.Message) error {
	logger := b.logger.With(zap.Int64("entity_id", int64(entity)))
	ctx, span := registryTracer.Start(parentCtx, "CreateComponents")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
	)
	defer span.End()

	logger.Debug("saving components")
	idStr := strconv.FormatInt(int64(entity), 10)
	err := b.redisStore.HSaveProto(ctx, idStr, components...)
	if err != nil {
		logger.Error("error saving proto", zap.Error(err))
		return err
	}

	for _, component := range components {
		_, err := b.redisStore.SAdd(ctx, b.keyEntitiesWithComponentType(component), idStr)
		if err != nil {
			logger.Error("error adding component to the index per component type", zap.Error(err))
			return err
		}
	}
	for _, component := range components {
		logger.Debug("calling onCreate callbacks")
		for _, cb := range b.onCreateComponent {
			cb(ctx, entity, component)
		}
	}
	return nil
}

func (b *Registry) UpdateComponents(parentCtx context.Context, entity Entity, components ...proto.Message) error {
	logger := b.logger.With(zap.Int64("entity_id", int64(entity)))
	ctx, span := registryTracer.Start(parentCtx, "UpdateComponents")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
	)
	defer span.End()

	logger.Debug("saving components")
	idStr := strconv.FormatInt(int64(entity), 10)
	err := b.redisStore.HSaveProto(ctx, idStr, components...)
	if err != nil {
		logger.Error("error saving proto", zap.Error(err))
		return err
	}

	return nil
}

func (b *Registry) DeleteComponent(parentCtx context.Context, entity Entity, component proto.Message) error {
	componentType := string(component.ProtoReflect().Descriptor().FullName().Name())
	logger := b.logger.With(zap.Int64("entity_id", int64(entity)), zap.String("component_type", componentType))
	ctx, span := registryTracer.Start(parentCtx, "DeleteComponent")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
		attribute.String("component_type", componentType),
	)
	defer span.End()

	logger.Debug("deleting component")
	idStr := strconv.FormatInt(int64(entity), 10)
	err := b.redisStore.HDelProto(ctx, idStr, component)
	if err != nil {
		logger.Error("error deleting component", zap.Error(err))
		return err
	}
	err = b.redisStore.SRem(ctx, b.keyEntitiesWithComponentType(component), idStr)
	if err != nil {
		logger.Error("error deleting component from index per component type", zap.Error(err))
		return err
	}
	logger.Debug("calling onDelete callbacks")
	for _, cb := range b.onDeleteComponent {
		cb(ctx, entity, component)
	}
	return nil
}

func (b *Registry) DeleteEntity(parentCtx context.Context, entity Entity) error {
	logger := b.logger.With(zap.Int64("entity_id", int64(entity)))
	ctx, span := registryTracer.Start(parentCtx, "DeleteEntity")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
	)
	defer span.End()

	logger.Debug("deleting entity")
	idStr := strconv.FormatInt(int64(entity), 10)
	keys, err := b.redisStore.HKeys(ctx, idStr)
	if err != nil {
		logger.Error("error finding keys", zap.Error(err))
		return err
	}
	for _, key := range keys {
		switch key {
		case "Position":
			b.DeleteComponent(ctx, entity, &Position{})
		case "Render":
			b.DeleteComponent(ctx, entity, &Render{})
		case "Looker":
			b.DeleteComponent(ctx, entity, &Looker{})
		case "Named":
			b.DeleteComponent(ctx, entity, &Named{})
		case "Speaker":
			b.DeleteComponent(ctx, entity, &Speaker{})
		default:
			panic("Unknown component type: " + key)
		}
	}
	err = b.redisStore.Del(ctx, idStr)
	if err != nil {
		logger.Error("error deleting entity from redis", zap.Error(err))
		return err
	}
	return nil
}

func (b *Registry) LoadComponents(parentCtx context.Context, entity Entity, components ...proto.Message) error {
	componentTypes := make([]string, len(components))
	for i, component := range components {
		componentTypes[i] = string(component.ProtoReflect().Descriptor().FullName().Name())
	}

	logger := b.logger.With(zap.Int64("entity_id", int64(entity)), zap.Strings("component_type", componentTypes))
	ctx, span := registryTracer.Start(parentCtx, "LoadComponents")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
		attribute.Array("componentType", componentTypes),
	)
	defer span.End()

	logger.Debug("loading components")
	err := b.redisStore.HReadProtos(ctx, strconv.FormatInt(int64(entity), 10), components...)
	if err != nil {
		if err == redis.Nil {
			logger.Warn("component not found")
		} else {
			logger.Error("error loading components", zap.Error(err))
		}
		return err
	}

	return nil
}

func (b *Registry) LoadComponentsFromIndex(parentCtx context.Context, indexKey string, componentTypes ...proto.Message) ([]Entity, [][]proto.Message, error) {
	logger := b.logger.With(zap.String("index_key", indexKey))
	ctx, span := registryTracer.Start(parentCtx, "LoadComponentsFromIndex")
	span.SetAttributes(
		attribute.String("index_key", indexKey),
	)
	defer span.End()

	logger.Debug("loading components from index")
	return b.sortWithExtras(ctx, indexKey, componentTypes...)
}

func (b *Registry) EntitiesWithComponentType(parentCtx context.Context, component proto.Message, componentTypes ...proto.Message) ([]Entity, [][]proto.Message, error) {
	componentType := string(component.ProtoReflect().Descriptor().FullName().Name())
	logger := b.logger.With(zap.String("component_type", componentType))
	ctx, span := registryTracer.Start(parentCtx, "EntitiesWithComponentType")
	span.SetAttributes(
		attribute.String("component_type", componentType),
	)
	defer span.End()

	logger.Debug("loading entities with component type")
	return b.sortWithExtras(ctx, b.keyEntitiesWithComponentType(component), componentTypes...)
}

func (b *Registry) keyEntitiesWithComponentType(component proto.Message) string {
	return fmt.Sprintf("by_component:%v", component.ProtoReflect().Descriptor().FullName().Name())
}

func (b *Registry) sortWithExtras(ctx context.Context, key string, componentTypes ...proto.Message) ([]Entity, [][]proto.Message, error) {
	get := []string{"#"}
	for _, componentType := range componentTypes {
		componentType := componentType.ProtoReflect().Descriptor().FullName().Name()
		get = append(get, "*->"+string(componentType))
	}

	res := b.redisStore.client.Sort(ctx, key, &redis.Sort{By: "nosort", Get: get})
	if err := res.Err(); err != nil {
		return nil, nil, err
	}

	entities := make([]Entity, 0)
	components := make([][]proto.Message, 0)

	resStr := res.Val()
	for i := 0; i < len(resStr)/(len(componentTypes)+1); i++ {
		idx := i * (len(componentTypes) + 1)
		parsed, err := strconv.ParseInt(resStr[idx], 10, 64)
		if err != nil {
			return nil, nil, err
		}

		entityComponents := make([]proto.Message, 0)
		allComponents := true
		for i, componentType := range componentTypes {
			err = proto.Unmarshal([]byte(resStr[idx+1+i]), componentType)
			if err != nil {
				return nil, nil, err
			}
			entityComponents = append(entityComponents, proto.Clone(componentType))
		}

		if allComponents {
			entities = append(entities, Entity(parsed))
			components = append(components, entityComponents)
		}
	}

	return entities, components, nil
}
