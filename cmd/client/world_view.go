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
	box.SetBorder(true).SetRect(0, 0, 30, 30)
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

func (r *WorldView) AddRenderable(renderable *esive_grpc.Renderable) {
	r.mtx.Lock()
	r.Renderables[renderable.Id] = renderable
	if renderable.Id == r.playerID {
		r.PlayerX = renderable.Position.X
		r.PlayerY = renderable.Position.Y
	}
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

func (r *WorldView) DeleteRenderable(id int64) {
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
			x+int(renderable.Position.X-r.PlayerX+r.Visibility),
			y+int(renderable.Position.Y-r.PlayerY+r.Visibility),
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
			}
		case tcell.KeyCtrlT:
			_, err := r.client.Say(context.Background(), &esive_grpc.SayReq{Text: "FOO BAR"})
			if err != nil {
				panic(err)
			}
		case tcell.KeyUp:
			res, err := r.client.MoveUp(context.Background(), &esive_grpc.MoveUpReq{})
			if err != nil {
				panic(err)
			}
			r.SetPosition(r.playerID, res.Position)
		case tcell.KeyDown:
			res, err := r.client.MoveDown(context.Background(), &esive_grpc.MoveDownReq{})
			if err != nil {
				panic(err)
			}
			r.SetPosition(r.playerID, res.Position)
		case tcell.KeyLeft:
			res, err := r.client.MoveLeft(context.Background(), &esive_grpc.MoveLeftReq{})
			if err != nil {
				panic(err)
			}
			r.SetPosition(r.playerID, res.Position)
		case tcell.KeyRight:
			res, err := r.client.MoveRight(context.Background(), &esive_grpc.MoveRightReq{})
			if err != nil {
				panic(err)
			}
			r.SetPosition(r.playerID, res.Position)
		}
	})
}
