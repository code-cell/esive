package systems

// type TickHandler struct {
// 	futureMovementsMtx sync.Mutex
// 	futureMovements    map[int64]*list.List
// 	nextMovements      *list.List
// }

// type scheduledMovement struct {
// 	entity  components.Entity
// 	offsetX int64
// 	offsetY int64
// }

// func NewTickHandler() *TickHandler {
// 	return &TickHandler{
// 		futureMovements: make(map[int64]*list.List),
// 		nextMovements:   list.New(),
// 	}
// }

// func (h *TickHandler) OnTick(message proto.Message) {
// 	tickMessage := message.(*queue.Tick)
// 	fmt.Printf("Tick %d\n", tickMessage.Tick)
// 	h.futureMovementsMtx.Lock()
// 	movements, found := h.futureMovements[tickMessage.Tick]
// 	if found {
// 		delete(h.futureMovements, tickMessage.Tick)
// 	}
// 	h.futureMovementsMtx.Unlock()
// 	if !found {
// 		return
// 	}
// 	for e := movements.Front(); e != nil; e = e.Next() {
// 		movement := e.Value.(*scheduledMovement)
// 		fmt.Printf("Scheduled movement: %+v\n", movement)
// 	}
// 	for e := h.nextMovements.Front(); e != nil; e = e.Next() {
// 		movement := e.Value.(*scheduledMovement)
// 		fmt.Printf("Scheduled movement: %+v\n", movement)
// 	}
// }

// func (h *TickHandler) QueueMovement(tick int64, entity components.Entity, offsetX, offsetY int64) {
// 	h.futureMovementsMtx.Lock()
// 	defer h.futureMovementsMtx.Unlock()
// 	movements, found := h.futureMovements[tick]
// 	if !found {
// 		movements = list.New()
// 		h.futureMovements[tick] = movements
// 	}
// 	movements.PushBack(&scheduledMovement{
// 		entity:  entity,
// 		offsetX: offsetX,
// 		offsetY: offsetY,
// 	})
// }

// func (h *TickHandler) QueueMovementNextTick(entity components.Entity, offsetX, offsetY int64) {
// 	h.futureMovementsMtx.Lock()
// 	defer h.futureMovementsMtx.Unlock()
// 	h.nextMovements.PushBack(&scheduledMovement{
// 		entity:  entity,
// 		offsetX: offsetX,
// 		offsetY: offsetY,
// 	})
// }
