package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"time"

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
	initialTestEntities = flag.Int("test-entities", 100, "Amount of test entities (a #)")
	redisAddr           = flag.String("redis-addr", "localhost:6379", "Redis address")
	redisUsername       = flag.String("redis-username", "", "Redis username")
	redisPassword       = flag.String("redis-password", "", "Redis password")
	jeagerEndpoint      = flag.String("jaeger-endpoint", "http://localhost:14268/api/traces", "Jaeger collector endpoint")
	natsURL             = flag.String("nats-url", "", "NATS server url")
	tickDuration        = flag.Duration("tick", 1*time.Second, "Tick duration")
)

func main() {
	flag.Parse()

	flush := initTracer()
	defer flush()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Username: *redisUsername,
		Password: *redisPassword,
	})
	rdb.FlushAll(context.Background())
	rdb.AddHook(redisotel.TracingHook{})

	store := components.NewRedisStore(rdb, logger)
	registry := components.NewRegistry(store, logger)
	geo := components.NewGeo(registry, store, *viewRadius, logger)
	systems.SetRegistry(registry)
	systems.SetGeo(geo)

	vision := systems.NewVisionSystem()
	movement := systems.NewMovementSystem(vision)
	chat := systems.NewChatSystem(movement)

	err = queue.SetupNats(*natsURL)
	if err != nil {
		panic(err)
	}

	q := queue.NewQueue(*natsURL)
	if err := q.Connect(); err != nil {
		panic(err)
	}

	t := tick.NewTick(1 * time.Second)

	t.AddSubscriber(q.HandleTick)

	go q.Consume("tick", "systems", &queue.Tick{}, movement.OnTick)
	go t.Start()

	registry.OnCreateComponent(func(ctx context.Context, entity components.Entity, component proto.Message) {
		t := component.ProtoReflect().Descriptor().FullName().Name()
		vision.HandleNewComponent(ctx, string(t), entity)
	})

	registry.OnDeleteComponent(func(ctx context.Context, entity components.Entity, component proto.Message) {
		t := component.ProtoReflect().Descriptor().FullName().Name()
		vision.HandleRemovedComponent(ctx, string(t), entity)
	})

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
				&components.Render{Char: "#", Color: 0xff7f00},
			)
			if err != nil {
				panic(err)
			}
		}
	}()

	grpcServer(registry, vision, movement, chat, t)
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
