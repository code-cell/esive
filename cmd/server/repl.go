package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

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
}

func NewRepl(grpcServer *server) *Repl {
	r := &Repl{
		grpcServer: grpcServer,
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
				players = append(players, fmt.Sprintf("%d `%s`", player.Entity, player.Name))
			}
			fmt.Printf("Players:\n\t%v\n", strings.Join(players, "\n\t"))
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
