package main

import (
	"time"

	"github.com/code-cell/esive/components"
	"github.com/code-cell/esive/queue"
	"github.com/code-cell/esive/systems"
	"github.com/code-cell/esive/tick"
)

func main() {
	url := ""
	err := queue.SetupNats(url)
	if err != nil {
		panic(err)
	}

	q := queue.NewQueue(url)
	if err := q.Connect(); err != nil {
		panic(err)
	}

	t := tick.NewTick(1 * time.Second)

	t.AddSubscriber(q.HandleTick)

	go t.Start()

	tickHandler := systems.NewTickHandler()
	tickHandler.QueueMovement(2, components.Entity(123), 8, 5)
	q.Consume("tick", "systems", &queue.Tick{}, tickHandler.OnTick)
}
