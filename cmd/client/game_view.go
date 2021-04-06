package main

import (
	"github.com/code-cell/esive/models"
	"github.com/rivo/tview"
)

type GameView struct {
	*tview.Flex
	playerID int64

	WorldView *WorldView
	ChatView  *ChatView

	client models.IcecreamClient
}

func NewGameView(playerID int64, client models.IcecreamClient, app *tview.Application) *GameView {
	chat := NewChatView(client, app)
	world := NewWorldView(playerID, client, app, chat.input)
	chat.SetBackView(world)

	flex := tview.NewFlex().
		AddItem(world, 34, 1, true).
		AddItem(chat, 0, 1, true)

	return &GameView{
		playerID:  playerID,
		Flex:      flex,
		WorldView: world,
		ChatView:  chat,
		client:    client,
	}
}
