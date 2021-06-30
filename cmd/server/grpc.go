package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"

	"github.com/code-cell/esive/actions"
	components "github.com/code-cell/esive/components"
	esive_grpc "github.com/code-cell/esive/grpc"
	"github.com/code-cell/esive/systems"
	"github.com/code-cell/esive/tick"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
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
	actionsQueue *actions.ActionsQueue
	registry     *components.Registry
	geo          *components.Geo
	vision       *systems.VisionSystem
	movement     *systems.MovementSystem
	chat         *systems.ChatSystem
	tick         *tick.Tick
	logger       *zap.Logger

	playersMtx sync.Mutex
	players    map[string]*PlayerData
}

func newServer(logger *zap.Logger, actionsQueue *actions.ActionsQueue, registry *components.Registry, geo *components.Geo, vision *systems.VisionSystem, movement *systems.MovementSystem, chat *systems.ChatSystem, t *tick.Tick) *server {
	s := &server{
		actionsQueue: actionsQueue,
		registry:     registry,
		geo:          geo,
		vision:       vision,
		movement:     movement,
		chat:         chat,
		tick:         t,
		players:      map[string]*PlayerData{},
		logger:       logger,
	}
	return s
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

func (s *server) SetVelocity(ctx context.Context, v *esive_grpc.Velocity) (*esive_grpc.MoveRes, error) {
	tick, err := getTickFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	playerData := s.playerData(ctx)
	s.actionsQueue.QueueAction(ctx, tick, func(ctx context.Context) {
		s.movement.SetVelocity(ctx, tick, playerData.Entity, int64(v.X), int64(v.Y))
	})
	return &esive_grpc.MoveRes{}, nil
}

func (s *server) Read(ctx context.Context, req *esive_grpc.ReadReq) (*esive_grpc.ReadRes, error) {
	playerID := ctx.Value("playerID").(string)
	s.logger.Debug("Player read", zap.String("playerID", playerID))

	playerData := s.playerData(ctx)
	pos := &components.Position{}
	err := s.registry.LoadComponents(ctx, playerData.Entity, pos)
	if err != nil {
		panic(err)
	}

	if components.Distance(req.Position.X, req.Position.Y, pos.X, pos.Y) > 5 {
		playerData.Updater.Chats <- &esive_grpc.ChatMessage{
			From: "<SYSTEM>",
			Text: "You can read only up to 5 tiles from you. Get closer and try again.",
		}
		return &esive_grpc.ReadRes{}, nil
	}

	_, _, extras, err := s.geo.FindInRange(ctx, req.Position.X, req.Position.Y, 0, &components.Readable{})
	if err != nil {
		panic(err)
	}
	for _, entityExtras := range extras {
		readable := entityExtras[0].(*components.Readable)
		if readable.Text == "" {
			continue
		}
		playerData.Updater.Chats <- &esive_grpc.ChatMessage{
			From: "<SYSTEM>",
			Text: readable.Text,
		}
	}

	return &esive_grpc.ReadRes{}, nil
}

func (s *server) Say(ctx context.Context, req *esive_grpc.SayReq) (*esive_grpc.SayRes, error) {
	playerID := ctx.Value("playerID").(string)
	s.logger.Debug("Player say", zap.String("playerID", playerID), zap.String("text", req.Text))
	tick, err := getTickFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	playerData := s.playerData(ctx)
	err = s.chat.Say(ctx, tick, playerData.Entity, req.Text)
	if err != nil {
		panic(err)
	}
	return &esive_grpc.SayRes{}, nil
}

func (s *server) Join(ctx context.Context, req *esive_grpc.JoinReq) (*esive_grpc.JoinRes, error) {
	playerID := ctx.Value("playerID").(string)
	s.logger.Debug("Player joined", zap.String("playerID", playerID))

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
		&components.Position{X: rand.Int63n(10) - 5, Y: rand.Int63n(10) - 5},
		&components.Moveable{},
		&components.Speaker{Range: float32(*viewRadius)},
		&components.Render{Char: "@", Color: 0x5bd54dff},
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
	s.logger.Debug("Player subscribed to chat updates", zap.String("playerID", playerID))
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
	s.logger.Debug("Player subscribed to visibility updates", zap.String("playerID", playerID))
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
				Velocity: &esive_grpc.Velocity{
					X: viewItem.VelX,
					Y: viewItem.VelY,
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
	logger *zap.Logger
}

func (h *serverStats) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}

func (h *serverStats) HandleRPC(ctx context.Context, s stats.RPCStats) {}

func (h *serverStats) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	playerID := uuid.New().String()
	h.logger.Debug("Player connected", zap.String("playerID", playerID))
	return context.WithValue(ctx, "playerID", "playerID")
}

func (h *serverStats) HandleConn(ctx context.Context, s stats.ConnStats) {
	switch s.(type) {
	case *stats.ConnEnd:
		playerID := ctx.Value("playerID").(string)
		h.logger.Debug("Player disconnected", zap.String("playerID", playerID))
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

func (s *server) Serve() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9000))
	if err != nil {
		s.logger.Fatal("failed to listen", zap.Error(err))
	}
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(&serverStats{s, s.logger}),
		grpc.ChainUnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(),
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				t, err := getTickFromCtx(ctx)
				cur := s.tick.Current()
				if t != 0 && t <= cur {
					return nil, errors.New("can't send requests for current or past ticks")
				}
				grpc.SetHeader(ctx, metadata.Pairs("tick", strconv.FormatInt(s.tick.Current(), 10)))
				return handler(ctx, req)
			},
		),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	esive_grpc.RegisterEsiveServer(grpcServer, s)
	s.logger.Info("Running...")
	grpcServer.Serve(lis)
}
