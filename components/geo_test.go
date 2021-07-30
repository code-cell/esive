package components_test

import (
	"context"
	"testing"

	"github.com/code-cell/esive/components"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGeo(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	rdb.FlushAll(context.Background())
	redisStore := components.NewRedisStore(rdb, zap.NewNop())
	registry := components.NewRegistry(redisStore, zap.NewNop())
	geo := components.NewGeo(registry, redisStore, 10, zap.NewNop())

	validPositions := []*components.Position{
		{X: 0, Y: 0},
		{X: 10, Y: 0},
		{X: -10, Y: 0},
		{X: 0, Y: 10},
		{X: 0, Y: -10},
	}

	invalidPositions := []*components.Position{
		{X: 11, Y: 0},
		{X: -11, Y: 0},
		{X: 0, Y: 11},
		{X: 0, Y: -11},
		{X: 9, Y: 9},
	}

	validEntities := make([]components.Entity, len(validPositions))
	invalidEntities := make([]components.Entity, len(invalidPositions))
	for i, pos := range validPositions {
		e, err := registry.NewEntity(context.Background())
		require.NoError(t, err)
		require.NoError(t, registry.CreateComponents(context.Background(), e, pos))
		validEntities[i] = e
	}
	for i, pos := range invalidPositions {
		e, err := registry.NewEntity(context.Background())
		require.NoError(t, err)
		require.NoError(t, registry.CreateComponents(context.Background(), e, pos))
		invalidEntities[i] = e
	}

	found, _, _, err := geo.FindInRange(context.Background(), 0, 0, 10)
	require.NoError(t, err)

	require.ElementsMatch(t, validEntities, found)
}

// func TestGeo_HandleUpdates(t *testing.T) {
// 	rdb := redis.NewClient(&redis.Options{
// 		Addr: "localhost:6379",
// 	})
// 	rdb.FlushAll(context.Background())
// 	redisStore := components.NewRedisStore(rdb, zap.NewNop())
// 	registry := components.NewRegistry(redisStore, zap.NewNop())
// 	logger, _ := zap.NewDevelopment()
// 	geo := components.NewGeo(registry, redisStore, 10, logger)
// 	vision := systems.NewVisionSystem()
// 	movement := systems.NewMovementSystem(vision)

// 	systems.SetRegistry(registry)
// 	systems.SetGeo(geo)

// 	entity, err := registry.NewEntity(context.Background())
// 	require.NoError(t, err)
// 	require.NoError(t, registry.CreateComponents(context.Background(), entity, &components.Position{X: 30, Y: 0}))

// 	found, _, _, err := geo.FindInRange(context.Background(), 0, 0, 10)
// 	require.NoError(t, err)
// 	require.Equal(t, []components.Entity{}, found)

// 	require.NoError(t, movement.Move(context.Background(), entity, -20, 0))

// 	found, _, _, err = geo.FindInRange(context.Background(), 0, 0, 10)
// 	require.NoError(t, err)
// 	require.Equal(t, []components.Entity{entity}, found)
// }
