package systems

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	components "github.com/code-cell/esive/components"
)

type ChatCommand struct {
	Command string
	Help    string
	Action  func(context.Context, components.Entity, ChatListener, []string, *MovementSystem)
}

var ChatCommands []*ChatCommand

const CommandSender = "<SYSTEM>"

func init() {
	ChatCommands = []*ChatCommand{
		{
			Command: "help",
			Help:    "Displays this help",
			Action:  chatCommandHelp,
		}, {
			Command: "tp",
			Help:    "Teleports you to the given coordinates. Eg: /tp 0 0",
			Action:  chatCommandTp,
		},
	}
}

func chatCommandHelp(_ context.Context, _ components.Entity, listener ChatListener, _ []string, _ *MovementSystem) {
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

func chatCommandTp(ctx context.Context, entity components.Entity, listener ChatListener, args []string, movement *MovementSystem) {
	if len(args) != 2 {
		listener.HandleChatMessage(&ChatMessage{
			FromName: CommandSender,
			Message:  "Invalid syntax.",
		})
		return
	}

	x, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		listener.HandleChatMessage(&ChatMessage{
			FromName: CommandSender,
			Message:  "Invalid syntax.",
		})
		return
	}
	y, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		listener.HandleChatMessage(&ChatMessage{
			FromName: CommandSender,
			Message:  "Invalid syntax.",
		})
		return
	}

	listener.HandleChatMessage(&ChatMessage{
		FromName: CommandSender,
		Message:  fmt.Sprintf("Teleporting to [%d %d].", x, y),
	})

	movement.Teleport(ctx, entity, x, y)

}
