package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/code-cell/esive/actions"
	"github.com/code-cell/esive/components"
	"github.com/code-cell/esive/queue"
	"github.com/code-cell/esive/systems"
	"github.com/code-cell/esive/tick"
	"github.com/go-redis/redis/extra/redisotel"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var (
	viewRadius          = flag.Int("radius", 15, "Radius used for visibility and for chunk size")
	initialTestEntities = flag.Int("test-entities", 100, "Amount of test entities (a #). This will only trigger if redis database is flushed.")
	redisAddr           = flag.String("redis-addr", "localhost:6379", "Redis address")
	redisUsername       = flag.String("redis-username", "", "Redis username")
	redisPassword       = flag.String("redis-password", "", "Redis password")
	redisFlush          = flag.Bool("redis-flush", true, "If enabled, it empties all database when the server starts")
	jeagerEndpoint      = flag.String("jaeger-endpoint", "http://localhost:14268/api/traces", "Jaeger collector endpoint")
	natsURL             = flag.String("nats-url", "", "NATS server url")
	tickDuration        = flag.Duration("tick", 300*time.Millisecond, "Tick duration")
)

func main() {
	flag.Parse()

	flush := initTracer()
	defer flush()

	logConfig := zap.NewDevelopmentConfig()
	logConfig.OutputPaths = []string{
		"server.log",
	}
	logger, err := logConfig.Build()
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Username: *redisUsername,
		Password: *redisPassword,
	})
	if *redisFlush == true {
		rdb.FlushAll(context.Background())
	}
	rdb.AddHook(redisotel.TracingHook{})

	actionsQueue := actions.NewActionsQueue()
	store := components.NewRedisStore(rdb, logger)
	registry := components.NewRegistry(store, logger)
	geo := components.NewGeo(registry, store, *viewRadius, logger)
	systems.SetRegistry(registry)
	systems.SetGeo(geo)

	vision := systems.NewVisionSystem()
	movement := systems.NewMovementSystem(vision)
	chat := systems.NewChatSystem(actionsQueue, movement, registry)

	err = queue.SetupNats(*natsURL)
	if err != nil {
		panic(err)
	}

	q := queue.NewQueue(*natsURL)
	if err := q.Connect(); err != nil {
		panic(err)
	}

	t := tick.NewTick(0, *tickDuration)

	t.AddSubscriber(q.HandleTick)

	go q.Consume("tick", "actions", &queue.Tick{}, func(m proto.Message) {
		tickMessage := m.(*queue.Tick)
		actionsQueue.CallActions(tickMessage.Tick, context.Background())
		movement.MoveAllMoveables(context.Background(), tickMessage.Tick)
	})
	// go q.Consume("tick", "systems", &queue.Tick{}, movement.OnTick)
	go t.Start()

	registry.OnCreateComponent(func(ctx context.Context, entity components.Entity, component proto.Message) {
		componentType := component.ProtoReflect().Descriptor().FullName().Name()
		vision.HandleNewComponent(ctx, t.Current(), string(componentType), entity)
	})

	registry.OnDeleteComponent(func(ctx context.Context, entity components.Entity, component proto.Message) {
		componentType := component.ProtoReflect().Descriptor().FullName().Name()
		vision.HandleRemovedComponent(ctx, t.Current(), string(componentType), entity)
	})

	if *redisFlush == true {
		// Only create initial test entities if database was flushed
		go func() {
			for i := 0; i < *initialTestEntities; i++ {
				entity, err := registry.NewEntity(context.Background())
				if err != nil {
					panic(err)
				}
				err = registry.CreateComponents(context.Background(), entity,
					&components.Position{
						X: rand.Int63n(60) - 30,
						Y: rand.Int63n(60) - 30,
					},
					&components.Render{Char: "#", Color: 0xaf8769ff},
				)
				if err != nil {
					panic(err)
				}
			}
		}()
	}

	s := newServer(logger, actionsQueue, registry, geo, vision, movement, chat, t)
	go s.Serve()

	repl := NewRepl(s, t, movement)
	repl.Run()
}

// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer() func() {
	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint(*jeagerEndpoint),
		jaeger.WithSDKOptions(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.ServiceNameKey.String("server"),
			)),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	return flush
}
