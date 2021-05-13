package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	esive_grpc "github.com/code-cell/esive/grpc"
	"github.com/code-cell/esive/tick"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ClientOpts struct {
	addr string
	name string
}

type Client struct {
	opts ClientOpts

	tick *tick.Tick

	grpcConn    *grpc.ClientConn
	esiveClient esive_grpc.EsiveClient

	PlayerID       int64
	renderablesMtx sync.Mutex
	renderables    map[int64]*esive_grpc.Renderable
	prediction     *Prediction
}

func NewClient(addr, name string, prediction *Prediction) *Client {
	return &Client{
		opts: ClientOpts{
			addr: addr,
			name: name,
		},
		renderables: make(map[int64]*esive_grpc.Renderable),
		prediction:  prediction,
	}
}

func (c *Client) Connect() error {
	log, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	conn, err := grpc.Dial(c.opts.addr,
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(
			// Set client tick in the request header
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				if c.tick != nil {
					ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("tick", strconv.FormatInt(c.tick.Current(), 10)))
				}
				return invoker(ctx, method, req, reply, cc, opts...)
			},
			// Parse server tick and adjust to it
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				var md metadata.MD
				opts = append(opts, grpc.Header(&md))
				err := invoker(ctx, method, req, reply, cc, opts...)
				receivedTick, found := c.getTickFromMD(md)
				if found && c.tick != nil {
					log := log.With(zap.Int64("serverTick", receivedTick), zap.Int64("clientTick", c.tick.Current()))
					current := c.tick.Current()
					if current < receivedTick+1 || current > receivedTick+5 {
						log.Warn("adjusting tick")
						c.tick.Adjust(receivedTick + 3)
					}
					// log.Debug("received tick from the server")
				}
				return err
			},
		),
	)
	if err != nil {
		return err
	}
	c.grpcConn = conn

	c.esiveClient = esive_grpc.NewEsiveClient(conn)

	var md metadata.MD
	res, err := c.esiveClient.Join(context.Background(), &esive_grpc.JoinReq{
		Name: c.opts.name,
	}, grpc.Header(&md))
	if err != nil {
		return err
	}
	c.PlayerID = res.PlayerId
	c.initTickFromMD(md, res.TickMilliseconds)

	visStream, err := c.esiveClient.VisibilityUpdates(context.Background(), &esive_grpc.VisibilityUpdatesReq{})
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			e, err := visStream.Recv()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			switch e.Action {
			case esive_grpc.VisibilityUpdatesRes_ADD:
				c.UpdateRenderable(e.Tick, e.Renderable)
			case esive_grpc.VisibilityUpdatesRes_REMOVE:
				c.DeleteRenderable(e.Tick, e.Renderable.Id)
			}
		}
	}()

	return nil
}

func (c *Client) getTickFromMD(md metadata.MD) (int64, bool) {
	str, found := md["tick"]
	if found && len(str) > 0 {
		serverTick, err := strconv.ParseInt(str[0], 10, 64)
		if err != nil {
			return 0, false
		}
		return serverTick, true
	}
	return 0, false
}

func (c *Client) initTickFromMD(md metadata.MD, durationMs int32) {
	serverTick, found := c.getTickFromMD(md)
	if !found {
		panic("Didn't receive a tick from the server on the Join call.")
	}

	c.tick = tick.NewTick(serverTick+3, time.Duration(durationMs)*time.Millisecond)
	go c.tick.Start()
}

func (c *Client) Disonnect() error {
	if err := c.grpcConn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Client) Renderables() map[int64]*esive_grpc.Renderable {
	// c.renderablesMtx.Lock()
	// defer c.renderablesMtx.Unlock()
	return c.renderables
}

func (c *Client) UpdateRenderable(tick int64, renderable *esive_grpc.Renderable) {
	c.renderablesMtx.Lock()
	defer c.renderablesMtx.Unlock()
	c.renderables[renderable.Id] = renderable
	if renderable.Id == c.PlayerID {
		c.prediction.UpdatePlayerPositionFromServer(tick, renderable.Position.X, renderable.Position.Y, renderable.Velocity.X, renderable.Velocity.Y)
	}
}

func (c *Client) DeleteRenderable(tick int64, id int64) {
	c.renderablesMtx.Lock()
	defer c.renderablesMtx.Unlock()
	delete(c.renderables, id)
}

func (c *Client) SetVelocity(x, y int) {
	c.esiveClient.SetVelocity(context.Background(), &esive_grpc.Velocity{
		X: int64(x),
		Y: int64(y),
	})
}
