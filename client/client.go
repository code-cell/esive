package client

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

type ChatMessageHandler func(from, message string)
type UpdateRenderableHandler func(id, tick int64, renderable *esive_grpc.Renderable)
type DeleteRenderableHandler func(id, tick int64)

type ClientOpts struct {
	addr string
	name string
}

type Client struct {
	opts ClientOpts

	Tick *tick.Tick

	grpcConn    *grpc.ClientConn
	esiveClient esive_grpc.EsiveClient

	PlayerID int64

	chatMessageHandlersMtx sync.Mutex
	chatMessageHandlers    []ChatMessageHandler

	updateRenderableHandlersMtx sync.Mutex
	updateRenderableHandlers    []UpdateRenderableHandler

	deleteRenderableHandlersMtx sync.Mutex
	deleteRenderableHandlers    []DeleteRenderableHandler
}

func NewClient(addr, name string) *Client {
	return &Client{
		opts: ClientOpts{
			addr: addr,
			name: name,
		},
		chatMessageHandlers:      make([]ChatMessageHandler, 0),
		updateRenderableHandlers: make([]UpdateRenderableHandler, 0),
		deleteRenderableHandlers: make([]DeleteRenderableHandler, 0),
	}
}

func (c *Client) AddChatHandler(h ChatMessageHandler) {
	c.chatMessageHandlersMtx.Lock()
	defer c.chatMessageHandlersMtx.Unlock()
	c.chatMessageHandlers = append(c.chatMessageHandlers, h)
}

func (c *Client) AddUpdateRenderableHandler(h UpdateRenderableHandler) {
	c.updateRenderableHandlersMtx.Lock()
	defer c.updateRenderableHandlersMtx.Unlock()
	c.updateRenderableHandlers = append(c.updateRenderableHandlers, h)
}

func (c *Client) AddDeleteRenderableHandler(h DeleteRenderableHandler) {
	c.deleteRenderableHandlersMtx.Lock()
	defer c.deleteRenderableHandlersMtx.Unlock()
	c.deleteRenderableHandlers = append(c.deleteRenderableHandlers, h)
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
				if c.Tick != nil {
					ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("tick", strconv.FormatInt(c.Tick.Current(), 10)))
				}
				return invoker(ctx, method, req, reply, cc, opts...)
			},
			// Parse server tick and adjust to it
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				var md metadata.MD
				opts = append(opts, grpc.Header(&md))
				err := invoker(ctx, method, req, reply, cc, opts...)
				receivedTick, found := c.getTickFromMD(md)
				if found && c.Tick != nil {
					log := log.With(zap.Int64("serverTick", receivedTick), zap.Int64("clientTick", c.Tick.Current()))
					current := c.Tick.Current()
					if current < receivedTick+1 || current > receivedTick+5 {
						log.Warn("adjusting tick")
						c.Tick.Adjust(receivedTick + 3)
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

	chatStream, err := c.esiveClient.ChatUpdates(context.Background(), &esive_grpc.ChatUpdatesReq{})
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			e, err := chatStream.Recv()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			c.chatMessageHandlersMtx.Lock()
			for _, h := range c.chatMessageHandlers {
				h(e.Message.From, e.Message.Text)
			}
			c.chatMessageHandlersMtx.Unlock()
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

	c.Tick = tick.NewTick(serverTick+3, time.Duration(durationMs)*time.Millisecond)
	go c.Tick.Start()
}

func (c *Client) Disconnect() error {
	if err := c.grpcConn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdateRenderable(tick int64, renderable *esive_grpc.Renderable) {
	c.updateRenderableHandlersMtx.Lock()
	defer c.updateRenderableHandlersMtx.Unlock()

	for _, h := range c.updateRenderableHandlers {
		h(renderable.Id, tick, renderable)
	}
}

func (c *Client) DeleteRenderable(tick int64, id int64) {
	c.deleteRenderableHandlersMtx.Lock()
	defer c.deleteRenderableHandlersMtx.Unlock()

	for _, h := range c.deleteRenderableHandlers {
		h(id, tick)
	}
}

func (c *Client) SetVelocity(x, y int) {
	c.esiveClient.SetVelocity(context.Background(), &esive_grpc.Velocity{
		X: int64(x),
		Y: int64(y),
	})
}

func (c *Client) SendChatMessage(message string) {
	c.esiveClient.Say(context.Background(), &esive_grpc.SayReq{
		Text: message,
	})
}
