package systems

import (
	"container/list"
	"fmt"
	"sync"

	components "github.com/code-cell/esive/components"
	"github.com/code-cell/esive/queue"
	"google.golang.org/protobuf/proto"
)

type TickHandler struct {
	futureMovementsMtx sync.Mutex
	futureMovements    map[int64]*list.List
}

type scheduledMovement struct {
	entity  components.Entity
	offsetX int64
	offsetY int64
}

func NewTickHandler() *TickHandler {
	return &TickHandler{
		futureMovements: make(map[int64]*list.List),
	}
}

func (h *TickHandler) OnTick(message proto.Message) {
	tickMessage := message.(*queue.Tick)
	fmt.Printf("Tick %d\n", tickMessage.Tick)
	h.futureMovementsMtx.Lock()
	movements, found := h.futureMovements[tickMessage.Tick]
	if found {
		delete(h.futureMovements, tickMessage.Tick)
	}
	h.futureMovementsMtx.Unlock()
	if !found {
		return
	}
	for e := movements.Front(); e != nil; e = e.Next() {
		movement := e.Value.(*scheduledMovement)
		fmt.Printf("Scheduled movement: %+v\n", movement)
	}
}

func (h *TickHandler) QueueMovement(tick int64, entity components.Entity, offsetX, offsetY int64) {
	h.futureMovementsMtx.Lock()
	defer h.futureMovementsMtx.Unlock()
	movements, found := h.futureMovements[tick]
	if !found {
		movements = list.New()
		h.futureMovements[tick] = movements
	}
	movements.PushBack(&scheduledMovement{
		entity:  entity,
		offsetX: offsetX,
		offsetY: offsetY,
	})
}
