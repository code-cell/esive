package systems

import (
	"context"
	"sync"

	components "github.com/code-cell/esive/components"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var chatTracer = otel.Tracer("systems/chat")

type ChatMessage struct {
	From     components.Entity
	FromName string
	Message  string
}

type ChatListener interface {
	HandleChatMessage(*ChatMessage)
}

type ChatSystem struct {
	listeners    map[components.Entity]ChatListener
	listenersMtx sync.Mutex
}

func NewChatSystem() *ChatSystem {
	return &ChatSystem{
		listeners: map[components.Entity]ChatListener{},
	}
}

func (s *ChatSystem) Say(parentContext context.Context, entity components.Entity, text string) error {
	ctx, span := chatTracer.Start(parentContext, "chat.Say")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
	)
	defer span.End()

	speakerPos := &components.Position{}
	err := registry.LoadComponent(ctx, entity, speakerPos)
	if err != nil {
		return err
	}
	speaker := &components.Speaker{}
	err = registry.LoadComponent(ctx, entity, speaker)
	if err != nil {
		return err
	}
	name := &components.Named{}
	err = registry.LoadComponent(ctx, entity, name)
	if err != nil {
		return err
	}

	chatMessage := &ChatMessage{
		From:     entity,
		FromName: name.Name,
		Message:  text,
	}
	entities, _, _, err := geo.FindInRange(ctx, speakerPos.X, speakerPos.Y, speaker.Range)
	for _, entity := range entities {
		s.listenersMtx.Lock()
		listener, ok := s.listeners[entity]
		s.listenersMtx.Unlock()
		if ok {
			listener.HandleChatMessage(chatMessage)
		}
	}
	return nil
}

func (s *ChatSystem) AddListener(entity components.Entity, listener ChatListener) error {
	s.listenersMtx.Lock()
	defer s.listenersMtx.Unlock()
	s.listeners[entity] = listener
	return nil
}
