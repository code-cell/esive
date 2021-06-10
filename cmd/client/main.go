package main

import (
	"flag"

	_ "image/png"

	"github.com/code-cell/esive/client"
	"github.com/code-cell/esive/cmd/client/ui"
)

var (
	// This is configured when building releases to point to the test server.
	defaultAddr = "localhost:9000"
	addr        = flag.String("addr", defaultAddr, "Server address")
	name        = flag.String("name", "", "Your name. Required.")
)

func main() {
	flag.Parse()

	if *name == "" {
		panic("the `name` flag is required.")
	}

	c := client.NewClient(*addr, *name)
	if err := c.Connect(); err != nil {
		panic(err)
	}

	game := ui.NewGame(c)
	game.Run()
}
