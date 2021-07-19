package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	components "github.com/code-cell/esive/components"
	"github.com/code-cell/esive/systems"
	"github.com/code-cell/esive/tick"
	"github.com/peterh/liner"
)

var (
	history_fn = ".esive_history"
)

type replCommand struct {
	keyword string
	help    string
	action  func([]string)
}

type Repl struct {
	exit     bool
	commands []replCommand

	grpcServer *server
	tick       *tick.Tick
	movement   *systems.MovementSystem
}

func NewRepl(grpcServer *server, tick *tick.Tick, movement *systems.MovementSystem) *Repl {
	r := &Repl{
		grpcServer: grpcServer,
		tick:       tick,
		movement:   movement,
	}

	r.commands = append(r.commands, replCommand{
		keyword: "help",
		help:    "Displays this help",
		action: func(_ []string) {
			for _, cmd := range r.commands {
				fmt.Printf("%v: %v\n", cmd.keyword, cmd.help)
			}
		},
	})

	r.commands = append(r.commands, replCommand{
		keyword: "exit",
		help:    "Shuts down the server",
		action: func(_ []string) {
			r.exit = true
		},
	})

	r.commands = append(r.commands, replCommand{
		keyword: "info",
		help:    "Displays server information",
		action: func(_ []string) {
			players := []string{}
			for _, player := range r.grpcServer.players {
				players = append(players, fmt.Sprintf("%v", player.Entity))
			}
			fmt.Printf("Players:\n\t%v\n", strings.Join(players, "\n\t"))
		},
	})

	r.commands = append(r.commands, replCommand{
		keyword: "tp",
		help:    "`tp PLAYER_ID X Y`. Teleports the player PLAYER_ID to [X,Y]",
		action: func(args []string) {
			entity, err := argInt64(args, 0)
			if err != nil {
				fmt.Printf("Error: %v\n", err.Error())
				return
			}
			x, err := argInt64(args, 1)
			if err != nil {
				fmt.Printf("Error: %v\n", err.Error())
				return
			}
			y, err := argInt64(args, 2)
			if err != nil {
				fmt.Printf("Error: %v\n", err.Error())
				return
			}
			err = r.movement.Teleport(context.TODO(), r.tick.Current(), components.Entity(entity), x, y)
			if err != nil {
				fmt.Printf("Error: %v\n", err.Error())
				return
			}
		},
	})

	return r
}

func (r *Repl) Run() {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	line.SetCompleter(func(line string) (c []string) {
		for _, cmd := range r.commands {
			if strings.HasPrefix(cmd.keyword, line) {
				c = append(c, cmd.keyword)
			}
		}
		return
	})

	if f, err := os.Open(history_fn); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	for !r.exit {
		if input, err := line.Prompt("> "); err == nil {
			err = r.runInput(input)
			if err == nil {
				line.AppendHistory(input)
			} else {
				log.Printf("Error running line: %v", err.Error())
			}
		} else if err == liner.ErrPromptAborted {
			log.Print("Aborted")
			return
		} else if err == io.EOF { // Ctrl + D
			return
		} else {
			log.Print("Error reading line: ", err)
			return
		}

		if f, err := os.Create(history_fn); err != nil {
			log.Print("Error writing history file: ", err)
			return
		} else {
			line.WriteHistory(f)
			f.Close()
		}
	}
}

func (r *Repl) runInput(input string) error {
	parts := strings.Split(input, " ")
	for _, cmd := range r.commands {
		if cmd.keyword == parts[0] {
			cmd.action(parts[1:])
			return nil
		}
	}
	return fmt.Errorf("command `%v` not found", input)
}

func argInt64(args []string, i int) (int64, error) {
	if len(args) <= i {
		return 0, errors.New("Missing arguments")
	}
	return strconv.ParseInt(args[i], 10, 64)
}
