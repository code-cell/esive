package main

import (
	components "github.com/code-cell/esive/components"
	"github.com/code-cell/esive/models"
	"github.com/code-cell/esive/systems"
)

type updater struct {
	Updates chan *models.VisibilityUpdatesRes
	Chats   chan *models.ChatMessage
}

func newUpdater() *updater {
	res := &updater{
		Updates: make(chan *models.VisibilityUpdatesRes),
		Chats:   make(chan *models.ChatMessage),
	}
	return res
}

func (u *updater) HandleVisibilityLostSight(entity components.Entity) {
	u.Updates <- &models.VisibilityUpdatesRes{
		Action: models.VisibilityUpdatesRes_REMOVE,
		Renderable: &models.Renderable{
			Id: int64(entity),
		},
	}

}
func (u *updater) HandleVisibilityUpdate(item *systems.VisionSystemLookItem) {
	u.Updates <- &models.VisibilityUpdatesRes{
		Action: models.VisibilityUpdatesRes_ADD,
		Renderable: &models.Renderable{
			Char:  item.Char,
			Color: item.Color,
			Id:    item.ID,
			Position: &models.Position{
				X: item.X,
				Y: item.Y,
			},
		},
	}

}
func (u *updater) HandleChatMessage(message *systems.ChatMessage) {
	u.Chats <- &models.ChatMessage{
		From: message.FromName,
		Text: message.Message,
	}
}
