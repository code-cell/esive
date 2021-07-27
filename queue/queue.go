package queue

import (
	"context"
	"time"

	"github.com/code-cell/esive/components"
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

func (q *Queue) HandleTickServicesDone(ctx context.Context, tick int64) {
	t := &TickServicesFinished{
		Tick: tick,
	}

	payload, err := proto.Marshal(t)
	if err != nil {
		// TODO: Handle this
		panic(err)
	}

	err = q.nc.Publish("tick-services-finished", payload)
	if err != nil {
		// TODO: Handle this
		panic(err)
	}
}

func (q *Queue) ProcessChunkMovements(ctx context.Context, tick int64, cx, cy int64) ([]components.Entity, error) {
	t := &ProcessChunkMovements{
		Tick:   tick,
		ChunkX: cx,
		ChunkY: cy,
	}

	payload, err := proto.Marshal(t)
	if err != nil {
		return nil, err
	}

	msg, err := q.nc.Request("process-chunk-movements", payload, 50*time.Millisecond)
	if err != nil {
		return nil, err
	}
	res := &ProcessChunkMovementsRes{}
	if err := proto.Unmarshal(msg.Data, res); err != nil {
		return nil, err
	}

	entities := []components.Entity{}
	for _, id := range res.Entities {
		entities = append(entities, components.Entity(id))
	}
	return entities, nil
}

func (q *Queue) Consume(subject, consumer string, message proto.Message, cb func(*nats.Msg, proto.Message)) {
	sub, err := q.nc.QueueSubscribeSync(subject, consumer)
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
		cb(msg, t)
	}
}
