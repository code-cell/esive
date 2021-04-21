package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"time"

	esive_grpc "github.com/code-cell/esive/grpc"
	"github.com/code-cell/esive/tick"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	// This is configured when building releases to point to the test server.
	defaultAddr = "localhost:9000"

	addr = flag.String("addr", defaultAddr, "Server address")
	name = flag.String("name", "", "Your name. Optional.")
)

func main() {
	log, err := zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		panic(err)
	}
	flag.Parse()

	var t *tick.Tick

	name := askName()
	conn, err := grpc.Dial(*addr,
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			var md metadata.MD
			opts = append(opts, grpc.Header(&md))
			err := invoker(ctx, method, req, reply, cc, opts...)
			serverTick, found := getTickFromMD(md)
			if found && t != nil {
				log := log.With(zap.Int64("serverTick", serverTick), zap.Int64("clientTick", t.Current()))
				if serverTick != t.Current() {
					log.Warn("adjusting tick")
					t.Adjust(serverTick)
				}
				log.Debug("received tick from the server")
			}
			return err
		}),
	)

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := esive_grpc.NewEsiveClient(conn)

	var md metadata.MD
	joinRes, err := client.Join(context.Background(), &esive_grpc.JoinReq{
		Name: name,
	}, grpc.Header(&md))
	if err != nil {
		panic(err)
	}
	playerID := joinRes.PlayerId
	serverTick, found := getTickFromMD(md)
	if !found {
		panic("Didn't receive a tick from the server on the Join call.")
	}
	t = tick.NewTick(serverTick, time.Duration(joinRes.TickMilliseconds)*time.Millisecond)
	go t.Start()

	visStream, err := client.VisibilityUpdates(context.Background(), &esive_grpc.VisibilityUpdatesReq{})
	if err != nil {
		panic(err)
	}

	chatStream, err := client.ChatUpdates(context.Background(), &esive_grpc.ChatUpdatesReq{})
	if err != nil {
		panic(err)
	}

	app := tview.NewApplication()
	gameView := NewGameView(playerID, client, app)

	go func() {
		for {
			e, err := visStream.Recv()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			switch e.Action {
			case esive_grpc.VisibilityUpdatesRes_ADD:
				gameView.WorldView.AddRenderable(e.Renderable)
			case esive_grpc.VisibilityUpdatesRes_REMOVE:
				gameView.WorldView.DeleteRenderable(e.Renderable.Id)
			}
		}
	}()

	go func() {
		for {
			e, err := chatStream.Recv()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			gameView.ChatView.Append(fmt.Sprintf("%v: %v", e.Message.From, e.Message.Text))
		}
	}()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.Stop()
		}
		return event
	})
	if err := app.SetRoot(gameView, true).SetFocus(gameView).Run(); err != nil {
		panic(err)
	}
}

func askName() string {
	if *name != "" {
		return *name
	}
	n := ""
	app := tview.NewApplication()
	inputField := tview.NewInputField().
		SetLabel("Enter your name: ").
		SetFieldWidth(0)
	inputField.
		SetDoneFunc(func(key tcell.Key) {
			n = inputField.GetText()
			app.Stop()
		})
	if err := app.SetRoot(inputField, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
	return n
}

func getTickFromMD(md metadata.MD) (int64, bool) {
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
