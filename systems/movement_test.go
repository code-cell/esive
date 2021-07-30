package systems

import (
	"context"
	"testing"

	components "github.com/code-cell/esive/components"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type Env struct {
	registry *components.Registry
	geo      *components.Geo
	movement *MovementSystem
	vision   *VisionSystem
}

func Setup(t *testing.T) *Env {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	redisClient.FlushAll(context.Background())
	logger := zap.NewNop()

	// actionsQueue := actions.NewActionsQueue()
	store := components.NewRedisStore(redisClient, logger)
	registry := components.NewRegistry(store, logger)
	geo := components.NewGeo(registry, store, 15, logger)
	SetRegistry(registry)
	SetGeo(geo)

	vision := NewVisionSystem(15)
	movement := NewMovementSystem(vision)

	return &Env{
		registry: registry,
		geo:      geo,
		movement: movement,
		vision:   vision,
	}
}

func TestSimpleMovement(t *testing.T) {
	env := Setup(t)
	entity, err := env.registry.NewEntity(context.Background())
	require.NoError(t, err)

	err = env.registry.CreateComponents(context.Background(), entity,
		&components.Position{X: 10, Y: 20},
		&components.Moveable{VelX: 1, VelY: -1},
	)
	require.NoError(t, err)

	move(t, env.movement)
	require.NoError(t, err)

	pos := &components.Position{}
	require.NoError(t, env.registry.LoadComponents(context.Background(), entity, pos))

	require.Equal(t, int64(11), pos.X)
	require.Equal(t, int64(19), pos.Y)
}

func TestCollision_WithAStaticEntity(t *testing.T) {
	env := Setup(t)
	entity1, err := env.registry.NewEntity(context.Background())
	require.NoError(t, err)
	entity2, err := env.registry.NewEntity(context.Background())
	require.NoError(t, err)

	err = env.registry.CreateComponents(context.Background(), entity1,
		&components.Position{X: 10, Y: 20},
		&components.Moveable{VelX: 1, VelY: 0},
	)
	require.NoError(t, err)

	err = env.registry.CreateComponents(context.Background(), entity2,
		&components.Position{X: 11, Y: 20},
	)
	require.NoError(t, err)

	move(t, env.movement)
	require.NoError(t, err)

	pos := &components.Position{}
	require.NoError(t, env.registry.LoadComponents(context.Background(), entity1, pos))

	require.Equal(t, int64(10), pos.X)
	require.Equal(t, int64(20), pos.Y)
}

func TestCollision_TwoEntitiesMoveToTheSamePlace(t *testing.T) {
	env := Setup(t)
	entity1, err := env.registry.NewEntity(context.Background())
	require.NoError(t, err)
	entity2, err := env.registry.NewEntity(context.Background())
	require.NoError(t, err)

	err = env.registry.CreateComponents(context.Background(), entity1,
		&components.Position{X: 10, Y: 20},
		&components.Moveable{VelX: 1, VelY: 0},
	)
	require.NoError(t, err)

	err = env.registry.CreateComponents(context.Background(), entity2,
		&components.Position{X: 12, Y: 20},
		&components.Moveable{VelX: -1, VelY: 0},
	)
	require.NoError(t, err)

	move(t, env.movement)
	require.NoError(t, err)

	pos := &components.Position{}

	// From the current implementation, we loop through entities by their ID, so entity1 moves.
	require.NoError(t, env.registry.LoadComponents(context.Background(), entity1, pos))
	require.Equal(t, int64(11), pos.X)
	require.Equal(t, int64(20), pos.Y)

	require.NoError(t, env.registry.LoadComponents(context.Background(), entity2, pos))
	require.Equal(t, int64(12), pos.X)
	require.Equal(t, int64(20), pos.Y)
}

func TestCollision_TakingPlaceOfMovingEntity(t *testing.T) {
	env := Setup(t)
	entity1, err := env.registry.NewEntity(context.Background())
	require.NoError(t, err)
	entity2, err := env.registry.NewEntity(context.Background())
	require.NoError(t, err)

	err = env.registry.CreateComponents(context.Background(), entity1,
		&components.Position{X: 10, Y: 20},
		&components.Moveable{VelX: 1, VelY: 0},
	)
	require.NoError(t, err)

	err = env.registry.CreateComponents(context.Background(), entity2,
		&components.Position{X: 11, Y: 20},
		&components.Moveable{VelX: 1, VelY: 0},
	)
	require.NoError(t, err)

	move(t, env.movement)
	require.NoError(t, err)

	pos := &components.Position{}

	// From the current implementation, we loop through entities by their ID, so entity1 moves.
	require.NoError(t, env.registry.LoadComponents(context.Background(), entity1, pos))
	require.Equal(t, int64(11), pos.X)
	require.Equal(t, int64(20), pos.Y)

	require.NoError(t, env.registry.LoadComponents(context.Background(), entity2, pos))
	require.Equal(t, int64(12), pos.X)
	require.Equal(t, int64(20), pos.Y)
}

func move(t *testing.T, m *MovementSystem) {
	chunks, err := m.ChunksWithMovingEntities(context.Background())
	require.NoError(t, err)

	across := []components.Entity{}
	for x, c := range chunks {
		for y := range c {
			entities, err := m.MoveAllEntitiesInChunk(context.Background(), x, y, 0)
			require.NoError(t, err)
			across = append(across, entities...)
		}
	}
	m.MoveEntitiesAcrossChunks(context.Background(), across, 0)
}
