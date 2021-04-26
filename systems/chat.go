package systems

import (
	"context"
	"strings"
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
	registry       *components.Registry
	movementSystem *MovementSystem
	listeners      map[components.Entity]ChatListener
	listenersMtx   sync.Mutex
}

func NewChatSystem(movementSystem *MovementSystem, registry *components.Registry) *ChatSystem {
	return &ChatSystem{
		movementSystem: movementSystem,
		registry:       registry,
		listeners:      map[components.Entity]ChatListener{},
	}
}

func (s *ChatSystem) Say(parentContext context.Context, entity components.Entity, text string) error {
	if text == "" {
		return nil
	}
	ctx, span := chatTracer.Start(parentContext, "chat.Say")
	span.SetAttributes(
		attribute.Int64("entity_id", int64(entity)),
	)
	defer span.End()
	if text[0] == '/' {
		s.listenersMtx.Lock()
		listener, ok := s.listeners[entity]
		s.listenersMtx.Unlock()
		if !ok {
			// TODO: This shouldn't happen. Maybe log an error message?
			return nil
		}

		parts := strings.Split(text[1:], " ")
		cmd := parts[0]
		for _, command := range ChatCommands {
			if command.Command == cmd {
				command.Action(ctx, entity, listener, parts[1:], s.movementSystem, s.registry)
				return nil
			}
		}
		listener.HandleChatMessage(&ChatMessage{
			FromName: CommandSender,
			Message:  "Unknown command. Use `/help` to see the full list.",
		})
	} else {
		speakerPos := &components.Position{}
		speaker := &components.Speaker{}
		name := &components.Named{}
		err := registry.LoadComponents(ctx, entity, speakerPos, speaker, name)
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
	}

	return nil
}

func (s *ChatSystem) AddListener(entity components.Entity, listener ChatListener) error {
	s.listenersMtx.Lock()
	defer s.listenersMtx.Unlock()
	s.listeners[entity] = listener
	return nil
}
