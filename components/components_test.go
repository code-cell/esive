package components

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func TestRegistrySaveLoad(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	redisStore := NewRedisStore(rdb, zap.NewNop())
	registry := NewRegistry(redisStore, zap.NewNop())

	entity, err := registry.NewEntity(context.Background())
	require.NoError(t, err)

	position := &Position{X: 10, Y: 20}
	require.NoError(t, registry.CreateComponents(context.Background(), entity, position))

	loadedPosition := &Position{}
	require.NoError(t, registry.LoadComponents(context.Background(), entity, loadedPosition))

	require.True(t, proto.Equal(position, loadedPosition))
}

func TestRegistryLoadMissingComponentType(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	redisStore := NewRedisStore(rdb, zap.NewNop())
	registry := NewRegistry(redisStore, zap.NewNop())

	entity, err := registry.NewEntity(context.Background())
	require.NoError(t, err)

	loadedPosition := &Position{}
	require.Error(t, registry.LoadComponents(context.Background(), entity, loadedPosition))
}

func TestRegistryLoadComponents(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	rdb.FlushAll(context.Background())
	redisStore := NewRedisStore(rdb, zap.NewNop())
	registry := NewRegistry(redisStore, zap.NewNop())

	entityWithPosition, err := registry.NewEntity(context.Background())
	require.NoError(t, err)
	entityWithPositionAndRender, err := registry.NewEntity(context.Background())
	require.NoError(t, err)
	entityWithLooker, err := registry.NewEntity(context.Background())
	require.NoError(t, err)

	position := &Position{X: 10, Y: 20}
	require.NoError(t, registry.CreateComponents(context.Background(), entityWithPosition, position))
	require.NoError(t, registry.CreateComponents(context.Background(), entityWithPositionAndRender, position))

	render := &Render{Char: "@", Color: 0xFF0000FF}
	require.NoError(t, registry.CreateComponents(context.Background(), entityWithPositionAndRender, render))

	looker := &Looker{}
	require.NoError(t, registry.CreateComponents(context.Background(), entityWithLooker, looker))

	ids, _, err := registry.EntitiesWithComponentType(context.Background(), &Looker{})
	require.NoError(t, err)
	require.Equal(t, []Entity{entityWithLooker}, ids)

	ids, _, err = registry.EntitiesWithComponentType(context.Background(), &Position{})
	require.NoError(t, err)
	require.Equal(t, []Entity{entityWithPosition, entityWithPositionAndRender}, ids)

	ids, _, err = registry.EntitiesWithComponentType(context.Background(), &Render{})
	require.NoError(t, err)
	require.Equal(t, []Entity{entityWithPositionAndRender}, ids)
}

func TestRegistryCreateCallback(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	redisStore := NewRedisStore(rdb, zap.NewNop())
	registry := NewRegistry(redisStore, zap.NewNop())

	received := false
	registry.OnCreateComponent(func(ctx context.Context, entity Entity, component proto.Message) {
		received = true
	})

	entity, err := registry.NewEntity(context.Background())
	require.NoError(t, err)

	position := &Position{X: 10, Y: 20}
	require.NoError(t, registry.CreateComponents(context.Background(), entity, position))

	require.True(t, received)
}

func TestRegistryCreateCallback_DoesNotTriggerOnUpdate(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	redisStore := NewRedisStore(rdb, zap.NewNop())
	registry := NewRegistry(redisStore, zap.NewNop())

	received := 0
	registry.OnCreateComponent(func(ctx context.Context, entity Entity, component proto.Message) {
		received++
	})

	entity, err := registry.NewEntity(context.Background())
	require.NoError(t, err)

	position := &Position{X: 10, Y: 20}
	require.NoError(t, registry.CreateComponents(context.Background(), entity, position))
	position.X++
	require.NoError(t, registry.UpdateComponents(context.Background(), entity, position))

	require.Equal(t, 1, received)
}

func TestRegistryDeleteCallback(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	redisStore := NewRedisStore(rdb, zap.NewNop())
	registry := NewRegistry(redisStore, zap.NewNop())

	entity, err := registry.NewEntity(context.Background())
	require.NoError(t, err)

	position := &Position{X: 10, Y: 20}
	require.NoError(t, registry.CreateComponents(context.Background(), entity, position))

	received := false
	registry.OnDeleteComponent(func(ctx context.Context, entity Entity, component proto.Message) {
		received = true
	})

	require.NoError(t, registry.DeleteComponent(context.Background(), entity, &Position{}))

	require.True(t, received)
}

func TestRegistryDeleteCallback_FromEntity(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	redisStore := NewRedisStore(rdb, zap.NewNop())
	registry := NewRegistry(redisStore, zap.NewNop())

	entity, err := registry.NewEntity(context.Background())
	require.NoError(t, err)

	position := &Position{X: 10, Y: 20}
	require.NoError(t, registry.CreateComponents(context.Background(), entity, position))

	received := false
	registry.OnDeleteComponent(func(ctx context.Context, entity Entity, component proto.Message) {
		received = true
	})

	require.NoError(t, registry.DeleteEntity(context.Background(), entity))

	require.True(t, received)
}
