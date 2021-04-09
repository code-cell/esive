package main

import (
	"context"
	"flag"
	"fmt"

	esive_grpc "github.com/code-cell/esive/grpc"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
)

var (
	addr = flag.String("addr", "localhost:9000", "Server address")
	name = flag.String("name", "", "Your name")
)

func main() {
	flag.Parse()
	name := askName()
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := esive_grpc.NewEsiveClient(conn)
	joinRes, err := client.Join(context.Background(), &esive_grpc.JoinReq{
		Name: name,
	})
	if err != nil {
		panic(err)
	}
	playerID := joinRes.PlayerId

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
