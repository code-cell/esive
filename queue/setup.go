package queue

import (
	"github.com/nats-io/jsm.go"
	"github.com/nats-io/nats.go"
)

// Setup prepares the queue backend with all prerequisites to run the service.
func SetupNats(natsUrl string) error {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return err
	}

	mgr, err := jsm.New(nc)
	if err != nil {
		return err
	}

	mgr.LoadOrNewStream("tick",
		jsm.Subjects("tick"),
	)

	mgr.LoadOrNewConsumer("tick", "systems",
		jsm.AcknowledgeExplicit(),
	)

	return nil
}
