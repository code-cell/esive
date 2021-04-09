package main

import (
	components "github.com/code-cell/esive/components"
	esive_grpc "github.com/code-cell/esive/grpc"
	"github.com/code-cell/esive/systems"
)

type updater struct {
	Updates chan *esive_grpc.VisibilityUpdatesRes
	Chats   chan *esive_grpc.ChatMessage
}

func newUpdater() *updater {
	res := &updater{
		Updates: make(chan *esive_grpc.VisibilityUpdatesRes),
		Chats:   make(chan *esive_grpc.ChatMessage),
	}
	return res
}

func (u *updater) HandleVisibilityLostSight(entity components.Entity) {
	u.Updates <- &esive_grpc.VisibilityUpdatesRes{
		Action: esive_grpc.VisibilityUpdatesRes_REMOVE,
		Renderable: &esive_grpc.Renderable{
			Id: int64(entity),
		},
	}

}
func (u *updater) HandleVisibilityUpdate(item *systems.VisionSystemLookItem) {
	u.Updates <- &esive_grpc.VisibilityUpdatesRes{
		Action: esive_grpc.VisibilityUpdatesRes_ADD,
		Renderable: &esive_grpc.Renderable{
			Char:  item.Char,
			Color: item.Color,
			Id:    item.ID,
			Position: &esive_grpc.Position{
				X: item.X,
				Y: item.Y,
			},
		},
	}

}
func (u *updater) HandleChatMessage(message *systems.ChatMessage) {
	u.Chats <- &esive_grpc.ChatMessage{
		From: message.FromName,
		Text: message.Message,
	}
}
