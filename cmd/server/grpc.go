package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"

	components "github.com/code-cell/esive/components"
	esive_grpc "github.com/code-cell/esive/grpc"
	"github.com/code-cell/esive/systems"
	"github.com/code-cell/esive/tick"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

type PlayerData struct {
	Entity  components.Entity
	Updater *updater
	Name    string
}

type server struct {
	registry *components.Registry
	vision   *systems.VisionSystem
	movement *systems.MovementSystem
	chat     *systems.ChatSystem
	tick     *tick.Tick

	playersMtx sync.Mutex
	players    map[string]*PlayerData
}

func (s *server) move(ctx context.Context, tick, offsetX, offsetY int64) error {
	playerID := ctx.Value("playerID").(string)
	fmt.Printf("Player %v moved. Offset (%d, %d)\n", playerID, offsetX, offsetY)

	playerData := s.playerData(ctx)
	return s.movement.QueueMove(ctx, playerData.Entity, tick, offsetX, offsetY)
}

func getTickFromCtx(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, errors.New("The context doesn't contain any metadata")
	}
	tick := md["tick"]
	if len(tick) != 1 {
		return 0, errors.New("No tick found in the context")
	}
	return strconv.ParseInt(tick[0], 10, 64)
}

func (s *server) MoveUp(ctx context.Context, req *esive_grpc.MoveReq) (*esive_grpc.MoveRes, error) {
	tick, err := getTickFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	err = s.move(ctx, tick, 0, -1)
	if err != nil {
		panic(err)
	}
	return &esive_grpc.MoveRes{}, nil
}

func (s *server) MoveDown(ctx context.Context, req *esive_grpc.MoveReq) (*esive_grpc.MoveRes, error) {
	tick, err := getTickFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = s.move(ctx, tick, 0, 1)
	if err != nil {
		panic(err)
	}
	return &esive_grpc.MoveRes{}, nil
}

func (s *server) MoveLeft(ctx context.Context, req *esive_grpc.MoveReq) (*esive_grpc.MoveRes, error) {
	tick, err := getTickFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = s.move(ctx, tick, -1, 0)
	if err != nil {
		panic(err)
	}
	return &esive_grpc.MoveRes{}, nil
}

func (s *server) MoveRight(ctx context.Context, req *esive_grpc.MoveReq) (*esive_grpc.MoveRes, error) {
	tick, err := getTickFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = s.move(ctx, tick, 1, 0)
	if err != nil {
		panic(err)
	}
	return &esive_grpc.MoveRes{}, nil
}

func (s *server) Build(ctx context.Context, _ *esive_grpc.BuildReq) (*esive_grpc.BuildRes, error) {
	playerID := ctx.Value("playerID").(string)
	fmt.Printf("Player %v build a dot\n", playerID)

	playerData := s.playerData(ctx)
	pos := &components.Position{}
	err := s.registry.LoadComponents(ctx, playerData.Entity, pos)
	if err != nil {
		panic(err)
	}

	entity, err := s.registry.NewEntity(ctx)
	if err != nil {
		panic(err)
	}
	err = s.registry.CreateComponents(ctx, entity,
		&components.Render{Char: ".", Color: 0x00FF00},
		pos,
	)
	if err != nil {
		panic(err)
	}

	return &esive_grpc.BuildRes{}, nil
}

func (s *server) Say(ctx context.Context, req *esive_grpc.SayReq) (*esive_grpc.SayRes, error) {
	playerID := ctx.Value("playerID").(string)
	fmt.Printf("Player %v say: %v\n", playerID, req.Text)
	playerData := s.playerData(ctx)
	err := s.chat.Say(ctx, playerData.Entity, req.Text)
	if err != nil {
		panic(err)
	}
	return &esive_grpc.SayRes{}, nil
}

func (s *server) Join(ctx context.Context, req *esive_grpc.JoinReq) (*esive_grpc.JoinRes, error) {
	playerID := ctx.Value("playerID").(string)
	fmt.Printf("Player %v joined\n", playerID)

	for _, d := range s.players {
		if d.Name == req.Name {
			return nil, errors.New("Name taken")
		}
	}

	entity, err := s.registry.NewEntity(ctx)
	if err != nil {
		panic(err)
	}

	err = s.registry.CreateComponents(ctx, entity,
		&components.Named{Name: req.Name},
		// &components.Position{X: rand.Int63n(1000) - 500, Y: rand.Int63n(1000) - 500},
		&components.Position{X: rand.Int63n(10) - 5, Y: rand.Int63n(10) - 5},
		&components.Speaker{Range: float32(*viewRadius)},
		&components.Render{Char: "@", Color: 0xFF0000},
		&components.Looker{Range: float32(*viewRadius)},
	)
	if err != nil {
		panic(err)
	}

	updater := newUpdater()
	s.vision.AddUpdater(entity, updater)
	s.chat.AddListener(entity, updater)

	s.players[playerID] = &PlayerData{
		Entity:  entity,
		Updater: updater,
		Name:    req.Name,
	}

	return &esive_grpc.JoinRes{
		PlayerId:         int64(entity),
		TickMilliseconds: int32(s.tick.Delay.Milliseconds()),
	}, nil
}

func (s *server) ChatUpdates(req *esive_grpc.ChatUpdatesReq, stream esive_grpc.Esive_ChatUpdatesServer) error {
	ctx := stream.Context()
	playerID := ctx.Value("playerID").(string)
	fmt.Printf("Player %v subscribed to chat updates\n", playerID)
	playerData := s.playerData(ctx)

	for message := range playerData.Updater.Chats {
		stream.Send(&esive_grpc.ChatUpdatesRes{
			Message: message,
		})
	}
	return nil
}

func (s *server) VisibilityUpdates(req *esive_grpc.VisibilityUpdatesReq, stream esive_grpc.Esive_VisibilityUpdatesServer) error {
	ctx := stream.Context()
	playerID := ctx.Value("playerID").(string)
	fmt.Printf("Player %v subscribed to visibility updates\n", playerID)
	playerData := s.playerData(ctx)

	viewItems, err := s.vision.LookAll(ctx, playerData.Entity)
	if err != nil {
		panic(err)
	}

	for _, viewItem := range viewItems {
		stream.Send(&esive_grpc.VisibilityUpdatesRes{
			Action: esive_grpc.VisibilityUpdatesRes_ADD,
			Renderable: &esive_grpc.Renderable{
				Char:  viewItem.Char,
				Color: viewItem.Color,
				Id:    viewItem.ID,
				Position: &esive_grpc.Position{
					X: viewItem.X,
					Y: viewItem.Y,
				},
			},
		})
	}

	for update := range playerData.Updater.Updates {
		stream.Send(update)
	}
	return nil
}

type serverStats struct {
	server *server
}

func (h *serverStats) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}

func (h *serverStats) HandleRPC(ctx context.Context, s stats.RPCStats) {}

func (h *serverStats) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	playerID := uuid.New().String()
	fmt.Printf("Player %v connected\n", playerID)
	return context.WithValue(ctx, "playerID", playerID)
}

func (h *serverStats) HandleConn(ctx context.Context, s stats.ConnStats) {
	switch s.(type) {
	case *stats.ConnEnd:
		playerID := ctx.Value("playerID").(string)
		fmt.Printf("client %v disconnected\n", playerID)
		playerData, ok := h.server.players[playerID]
		if ok {
			err := h.server.registry.DeleteEntity(ctx, playerData.Entity)
			if err != nil {
				panic(err)
			}
			h.server.playersMtx.Lock()
			delete(h.server.players, playerID)
			h.server.playersMtx.Unlock()
		}

		break
	}
}

func (s *server) playerData(ctx context.Context) *PlayerData {
	s.playersMtx.Lock()
	defer s.playersMtx.Unlock()
	playerID := ctx.Value("playerID").(string)
	return s.players[playerID]
}

func grpcServer(registry *components.Registry, vision *systems.VisionSystem, movement *systems.MovementSystem, chat *systems.ChatSystem, t *tick.Tick) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9000))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := &server{
		registry: registry,
		vision:   vision,
		movement: movement,
		chat:     chat,
		tick:     t,
		players:  map[string]*PlayerData{},
	}
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(&serverStats{s}),
		grpc.ChainUnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(),
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				grpc.SetHeader(ctx, metadata.Pairs("tick", strconv.FormatInt(s.tick.Current()+1, 10)))
				return handler(ctx, req)
			},
		),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	esive_grpc.RegisterEsiveServer(grpcServer, s)
	log.Println("Running...")
	grpcServer.Serve(lis)
}
