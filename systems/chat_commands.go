package systems

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	components "github.com/code-cell/esive/components"
)

type ChatCommand struct {
	Command string
	Help    string
	Action  func(context.Context, components.Entity, ChatListener, []string, *MovementSystem, *components.Registry)
}

var ChatCommands []*ChatCommand

const CommandSender = "<SYSTEM>"

func init() {
	ChatCommands = []*ChatCommand{
		{
			Command: "help",
			Help:    "Displays this help",
			Action:  chatCommandHelp,
			// }, {
			// 	Command: "tp",
			// 	Help:    "Teleports you to the given coordinates. Eg: /tp 0 0",
			// 	Action:  chatCommandTp,
		}, {
			Command: "note",
			Help:    "Leaves a note in the world. Eg: /note Hello world!",
			Action:  chatCommandNote,
		},
	}
}

func chatCommandHelp(_ context.Context, _ components.Entity, listener ChatListener, _ []string, _ *MovementSystem, registry *components.Registry) {
	message := bytes.NewBufferString("This is the list of commands:\n")
	for _, command := range ChatCommands {
		message.WriteString("  /")
		message.WriteString(command.Command)
		message.WriteString(": ")
		message.WriteString(command.Help)
		message.WriteString("\n")
	}
	listener.HandleChatMessage(&ChatMessage{
		FromName: CommandSender,
		Message:  message.String(),
	})
}

// func chatCommandTp(ctx context.Context, entity components.Entity, listener ChatListener, args []string, actionQueue *actions.ActionsQueue, movement *MovementSystem, registry *components.Registry) {
// 	if len(args) != 2 {
// 		listener.HandleChatMessage(&ChatMessage{
// 			FromName: CommandSender,
// 			Message:  "Invalid syntax.",
// 		})
// 		return
// 	}

// 	x, err := strconv.ParseInt(args[0], 10, 64)
// 	if err != nil {
// 		listener.HandleChatMessage(&ChatMessage{
// 			FromName: CommandSender,
// 			Message:  "Invalid syntax.",
// 		})
// 		return
// 	}
// 	y, err := strconv.ParseInt(args[1], 10, 64)
// 	if err != nil {
// 		listener.HandleChatMessage(&ChatMessage{
// 			FromName: CommandSender,
// 			Message:  "Invalid syntax.",
// 		})
// 		return
// 	}

// 	listener.HandleChatMessage(&ChatMessage{
// 		FromName: CommandSender,
// 		Message:  fmt.Sprintf("Teleporting to [%d %d].", x, y),
// 	})

// 	actionQueue.QueueInmediate(func(c context.Context) {
// 		movement.DoMove(parentContext context.Context, tick int64, entity components.Entity, offsetX int64, offsetY int64)
// 	})
// 	movement.Teleport(ctx, entity, x, y)
// }

func chatCommandNote(ctx context.Context, entity components.Entity, listener ChatListener, args []string, _ *MovementSystem, registry *components.Registry) {
	if len(args) == 0 {
		listener.HandleChatMessage(&ChatMessage{
			FromName: CommandSender,
			Message:  "Invalid syntax.",
		})
		return
	}

	pos := &components.Position{}
	name := &components.Named{}
	if err := registry.LoadComponents(ctx, entity, pos, name); err != nil {
		panic(err)
	}

	text := strings.Join(args, " ")

	noteEntity, err := registry.NewEntity(ctx)
	if err != nil {
		panic(err)
	}

	err = registry.CreateComponents(ctx, noteEntity,
		&components.Position{X: pos.X, Y: pos.Y},
		&components.Render{Char: "n", Color: 0x00c965},
		&components.Readable{Text: fmt.Sprintf("Message from %v: %v", name.Name, text)},
	)
	if err != nil {
		panic(err)
	}

	listener.HandleChatMessage(&ChatMessage{
		FromName: CommandSender,
		Message:  fmt.Sprintf("Note sent."),
	})
}
