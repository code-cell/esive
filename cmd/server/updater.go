package main

import (
	components "github.com/code-cell/esive/components"
	esive_grpc "github.com/code-cell/esive/grpc"
	"github.com/code-cell/esive/systems"
)

type updater struct {
	Updates chan *esive_grpc.VisibilityUpdate
	Chats   chan *esive_grpc.ChatMessage
}

func newUpdater() *updater {
	res := &updater{
		Updates: make(chan *esive_grpc.VisibilityUpdate),
		Chats:   make(chan *esive_grpc.ChatMessage),
	}
	return res
}

func (u *updater) HandleVisibilityLostSight(entity components.Entity, tick int64) {
	u.Updates <- &esive_grpc.VisibilityUpdate{
		Action: esive_grpc.VisibilityUpdate_REMOVE,
		Tick:   tick,
		Renderable: &esive_grpc.Renderable{
			Id: int64(entity),
		},
	}

}
func (u *updater) HandleTickUpdate(item *systems.VisionSystemLookItem, tick int64) {
	u.Updates <- &esive_grpc.VisibilityUpdate{
		Action: esive_grpc.VisibilityUpdate_ADD,
		Tick:   tick,
		Renderable: &esive_grpc.Renderable{
			Char:  item.Char,
			Color: item.Color,
			Id:    item.ID,
			Position: &esive_grpc.Position{
				X: item.X,
				Y: item.Y,
			},
			Velocity: &esive_grpc.Velocity{
				X: item.VelX,
				Y: item.VelY,
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
