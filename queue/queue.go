package queue

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	// "github.com/nats-io/jsm.go"
)

type Queue struct {
	natsUrl string

	nc *nats.Conn
}

func NewQueue(natsUrl string) *Queue {
	return &Queue{
		natsUrl: natsUrl,
	}
}

func (q *Queue) Connect() error {
	nc, err := nats.Connect(q.natsUrl)
	if err != nil {
		return err
	}
	q.nc = nc
	return nil
}

func (q *Queue) HandleTick(ctx context.Context, tick int64) {
	t := &Tick{
		Tick: tick,
	}

	payload, err := proto.Marshal(t)
	if err != nil {
		// TODO: Handle this
		panic(err)
	}

	err = q.nc.Publish("tick", payload)
	if err != nil {
		// TODO: Handle this
		panic(err)
	}
}

func (q *Queue) Consume(subject, consumer string, message proto.Message, cb func(proto.Message)) {
	sub, err := q.nc.QueueSubscribeSync("tick", "systems")
	if err != nil {
		panic(err)
	}

	for {
		msg, err := sub.NextMsg(5 * time.Second)
		if err != nil {
			if err == nats.ErrTimeout {
				continue
			}
			panic(err)
		}

		t := proto.Clone(message)
		proto.Unmarshal(msg.Data, t)
		cb(t)
	}
}
