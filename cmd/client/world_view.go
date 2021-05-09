package main

import (
	"context"
	"fmt"
	"sync"

	esive_grpc "github.com/code-cell/esive/grpc"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type WorldView struct {
	*tview.Box

	playerID int64

	app    *tview.Application
	client esive_grpc.EsiveClient
	chat   tview.Primitive

	mtx         sync.Mutex
	Renderables map[int64]*esive_grpc.Renderable
	Visibility  int64
	PlayerX     int64
	PlayerY     int64
}

func NewWorldView(playerID int64, client esive_grpc.EsiveClient, app *tview.Application, chat tview.Primitive) *WorldView {
	box := tview.NewBox()
	box.SetBorder(true).SetRect(0, 0, 34, 34)
	return &WorldView{
		Box:         box,
		Renderables: map[int64]*esive_grpc.Renderable{},
		playerID:    playerID,
		Visibility:  15,
		client:      client,
		chat:        chat,
		app:         app,
	}
}

func (r *WorldView) AddRenderable(tick int64, renderable *esive_grpc.Renderable) {
	r.mtx.Lock()
	if renderable.Id == r.playerID {
		r.PlayerX, r.PlayerY = playerMovements.GetPlayerPos(tick, renderable.Position.X, renderable.Position.Y)
		renderable.Position.X = r.PlayerX
		renderable.Position.Y = r.PlayerY
	}
	r.Renderables[renderable.Id] = renderable
	r.mtx.Unlock()
	go r.app.Draw()
}

func (r *WorldView) SetPosition(id int64, position *esive_grpc.Position) {
	r.mtx.Lock()
	r.Renderables[id].Position = position
	if id == r.playerID {
		r.PlayerX = position.X
		r.PlayerY = position.Y
	}
	r.mtx.Unlock()
	go r.app.Draw()
}

func (r *WorldView) GetPosition(id int64) *esive_grpc.Position {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.Renderables[id].Position
}

func (r *WorldView) DeleteRenderable(tick int64, id int64) {
	r.mtx.Lock()
	delete(r.Renderables, id)
	r.mtx.Unlock()
	go r.app.Draw()
}

func (r *WorldView) Draw(screen tcell.Screen) {
	r.Box.DrawForSubclass(screen, r)
	x, y, _, _ := r.GetInnerRect()

	r.mtx.Lock()
	defer r.mtx.Unlock()
	for _, renderable := range r.Renderables {
		screen.SetContent(
			x+int(renderable.Position.X-r.PlayerX+r.Visibility+2),
			y+int(renderable.Position.Y-r.PlayerY+r.Visibility+2),
			rune(renderable.Char[0]),
			nil,
			tcell.StyleDefault.Foreground(tcell.NewHexColor(renderable.Color)),
		)
	}
	r.SetTitle(fmt.Sprintf(" World view [%d, %d] ", r.PlayerX, r.PlayerY))
}

func (r *WorldView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyEnter:
			_, err := r.client.Build(context.Background(), &esive_grpc.BuildReq{})
			if err != nil {
				panic(err)
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 't':
				setFocus(r.chat)
			case 'i':
				r.client.Inspect(context.Background(), &esive_grpc.InspectReq{})
			}
		case tcell.KeyUp:
			if !playerMovements.CanMove(t.Current()) {
				return
			}
			_, err := r.client.MoveUp(context.Background(), &esive_grpc.MoveReq{})
			if err != nil {
				panic(err)
			}
			pos := r.GetPosition(r.playerID)
			r.SetPosition(r.playerID, &esive_grpc.Position{
				X: pos.X,
				Y: pos.Y - 1,
			})
			playerMovements.AddMovement(t.Current(), 0, -1)
		case tcell.KeyDown:
			if !playerMovements.CanMove(t.Current()) {
				return
			}
			_, err := r.client.MoveDown(context.Background(), &esive_grpc.MoveReq{})
			if err != nil {
				panic(err)
			}
			pos := r.GetPosition(r.playerID)
			r.SetPosition(r.playerID, &esive_grpc.Position{
				X: pos.X,
				Y: pos.Y + 1,
			})
			playerMovements.AddMovement(t.Current(), 0, 1)
		case tcell.KeyLeft:
			if !playerMovements.CanMove(t.Current()) {
				return
			}
			_, err := r.client.MoveLeft(context.Background(), &esive_grpc.MoveReq{})
			if err != nil {
				panic(err)
			}
			pos := r.GetPosition(r.playerID)
			r.SetPosition(r.playerID, &esive_grpc.Position{
				X: pos.X - 1,
				Y: pos.Y,
			})
			playerMovements.AddMovement(t.Current(), -1, 0)
		case tcell.KeyRight:
			if !playerMovements.CanMove(t.Current()) {
				return
			}
			_, err := r.client.MoveRight(context.Background(), &esive_grpc.MoveReq{})
			if err != nil {
				panic(err)
			}
			pos := r.GetPosition(r.playerID)
			r.SetPosition(r.playerID, &esive_grpc.Position{
				X: pos.X + 1,
				Y: pos.Y,
			})
			playerMovements.AddMovement(t.Current(), 1, 0)
		}
	})
}
