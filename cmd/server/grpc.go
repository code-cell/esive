package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"

	components "github.com/code-cell/esive/components"
	esive_grpc "github.com/code-cell/esive/grpc"
	"github.com/code-cell/esive/systems"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
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

	players map[string]*PlayerData
}

func (s *server) move(ctx context.Context, offsetX, offsetY int64) (*esive_grpc.Position, error) {
	playerID := ctx.Value("playerID").(string)
	fmt.Printf("Player %v moved. Offset (%d, %d)\n", playerID, offsetX, offsetY)

	playerData := s.players[playerID]
	err := s.movement.Move(ctx, playerData.Entity, offsetX, offsetY)
	if err != nil {
		panic(err)
	}

	newPos := &components.Position{}
	err = s.registry.LoadComponents(ctx, playerData.Entity, newPos)
	if err != nil {
		panic(err)
	}

	return &esive_grpc.Position{
		X: newPos.X,
		Y: newPos.Y,
	}, nil
}

func (s *server) MoveUp(ctx context.Context, _ *esive_grpc.MoveUpReq) (*esive_grpc.MoveUpRes, error) {
	pos, err := s.move(ctx, 0, -1)
	if err != nil {
		panic(err)
	}
	return &esive_grpc.MoveUpRes{
		Position: pos,
	}, nil
}

func (s *server) MoveDown(ctx context.Context, _ *esive_grpc.MoveDownReq) (*esive_grpc.MoveDownRes, error) {
	pos, err := s.move(ctx, 0, 1)
	if err != nil {
		panic(err)
	}
	return &esive_grpc.MoveDownRes{
		Position: pos,
	}, nil
}

func (s *server) MoveLeft(ctx context.Context, _ *esive_grpc.MoveLeftReq) (*esive_grpc.MoveLeftRes, error) {
	pos, err := s.move(ctx, -1, 0)
	if err != nil {
		panic(err)
	}
	return &esive_grpc.MoveLeftRes{
		Position: pos,
	}, nil
}

func (s *server) MoveRight(ctx context.Context, _ *esive_grpc.MoveRightReq) (*esive_grpc.MoveRightRes, error) {
	pos, err := s.move(ctx, 1, 0)
	if err != nil {
		panic(err)
	}
	return &esive_grpc.MoveRightRes{
		Position: pos,
	}, nil
}

func (s *server) Build(ctx context.Context, _ *esive_grpc.BuildReq) (*esive_grpc.BuildRes, error) {
	playerID := ctx.Value("playerID").(string)
	fmt.Printf("Player %v build a dot\n", playerID)

	playerData := s.players[playerID]
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
	playerData := s.players[playerID]
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
		&components.Position{X: rand.Int63n(20) - 10, Y: rand.Int63n(20) - 10},
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
		PlayerId: int64(entity),
	}, nil
}

func (s *server) ChatUpdates(req *esive_grpc.ChatUpdatesReq, stream esive_grpc.Esive_ChatUpdatesServer) error {
	ctx := stream.Context()
	playerID := ctx.Value("playerID").(string)
	fmt.Printf("Player %v subscribed to chat updates\n", playerID)
	playerData := s.players[playerID]

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
	playerData := s.players[playerID]

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
			delete(h.server.players, playerID)
		}

		break
	}
}

func grpcServer(registry *components.Registry, vision *systems.VisionSystem, movement *systems.MovementSystem, chat *systems.ChatSystem) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9000))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := &server{
		registry: registry,
		vision:   vision,
		movement: movement,
		chat:     chat,
		players:  map[string]*PlayerData{},
	}
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(&serverStats{s}),
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	esive_grpc.RegisterEsiveServer(grpcServer, s)
	log.Println("Running...")
	grpcServer.Serve(lis)
}
