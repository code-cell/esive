package main

import (
	"fmt"
	"time"

	"github.com/code-cell/esive/queue"
	"github.com/code-cell/esive/tick"
	"google.golang.org/protobuf/proto"
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

	q.Consume("tick", "systems:movement", &queue.Tick{}, func(m proto.Message) {
		t := m.(*queue.Tick)
		fmt.Printf("Tick %d\n", t.Tick)
	})
}
