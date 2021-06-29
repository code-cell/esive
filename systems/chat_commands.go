package systems

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/code-cell/esive/actions"
	components "github.com/code-cell/esive/components"
)

type ChatAction func(context.Context, int64, components.Entity, ChatListener, []string)
type ChatCommand struct {
	Command string
	Help    string
	Action  ChatAction
}

type ChatCommands struct {
	Commands map[string]*ChatCommand

	actionQueue *actions.ActionsQueue
	movement    *MovementSystem
	registry    *components.Registry

	systemSender string
}

func NewChatCommands(systemSender string, actionQueue *actions.ActionsQueue, movement *MovementSystem, registry *components.Registry) *ChatCommands {
	cm := &ChatCommands{
		Commands:     make(map[string]*ChatCommand),
		actionQueue:  actionQueue,
		movement:     movement,
		registry:     registry,
		systemSender: systemSender,
	}

	cm.addCommand("help", "Displays this help", cm.helpCommand)
	cm.addCommand("tp", "Teleports you to the given coordinates. Eg: /tp 0 0", cm.teleportCommand)
	cm.addCommand("note", "Leaves a note in the world. Eg: /note Hello world!", cm.noteCommand)

	return cm
}

func (cm *ChatCommands) addCommand(command, help string, action ChatAction) {
	cm.Commands[command] = &ChatCommand{
		Command: command,
		Help:    help,
		Action:  action,
	}
}

func (cm *ChatCommands) helpCommand(_ context.Context, _ int64, _ components.Entity, listener ChatListener, _ []string) {
	message := bytes.NewBufferString("This is the list of commands:\n")
	for _, command := range cm.Commands {
		message.WriteString("  /")
		message.WriteString(command.Command)
		message.WriteString(": ")
		message.WriteString(command.Help)
		message.WriteString("\n")
	}
	listener.HandleChatMessage(&ChatMessage{
		FromName: cm.systemSender,
		Message:  message.String(),
	})
}

func (cm *ChatCommands) teleportCommand(ctx context.Context, tick int64, entity components.Entity, listener ChatListener, args []string) {
	if len(args) != 2 {
		listener.HandleChatMessage(&ChatMessage{
			FromName: cm.systemSender,
			Message:  "Invalid syntax.",
		})
		return
	}

	x, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		listener.HandleChatMessage(&ChatMessage{
			FromName: cm.systemSender,
			Message:  "Invalid syntax.",
		})
		return
	}
	y, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		listener.HandleChatMessage(&ChatMessage{
			FromName: cm.systemSender,
			Message:  "Invalid syntax.",
		})
		return
	}

	listener.HandleChatMessage(&ChatMessage{
		FromName: cm.systemSender,
		Message:  fmt.Sprintf("Teleporting to [%d %d].", x, y),
	})

	cm.actionQueue.QueueInmediate(ctx, func(ctx context.Context) {
		pos := &components.Position{}
		err := cm.registry.LoadComponents(ctx, entity, pos)
		if err != nil {
			panic(err)
		}
		cm.movement.Teleport(ctx, tick, entity, x, y)
	})
}

func (cm *ChatCommands) noteCommand(ctx context.Context, _ int64, entity components.Entity, listener ChatListener, args []string) {
	if len(args) == 0 {
		listener.HandleChatMessage(&ChatMessage{
			FromName: cm.systemSender,
			Message:  "Invalid syntax.",
		})
		return
	}

	pos := &components.Position{}
	name := &components.Named{}
	if err := cm.registry.LoadComponents(ctx, entity, pos, name); err != nil {
		panic(err)
	}

	text := strings.Join(args, " ")

	noteEntity, err := registry.NewEntity(ctx)
	if err != nil {
		panic(err)
	}

	err = cm.registry.CreateComponents(ctx, noteEntity,
		&components.Position{X: pos.X, Y: pos.Y},
		&components.Render{Char: "n", Color: 0x649ce4ff},
		&components.Readable{Text: fmt.Sprintf("Message from %v: %v", name.Name, text)},
	)
	if err != nil {
		panic(err)
	}

	listener.HandleChatMessage(&ChatMessage{
		FromName: cm.systemSender,
		Message:  fmt.Sprintf("Note sent."),
	})
}
