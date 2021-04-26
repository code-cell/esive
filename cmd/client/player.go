package main

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

type PlayerMovements struct {
	lastProcessedTick  int64
	queuedMovementsMtx sync.Mutex
	queuedMovements    map[int64]*playerMovement
}
type playerMovement struct {
	offsetX int
	offsetY int
}

var ErrAlreadyMoved = errors.New("The player has already moved on that tick")

func NewPlayerMovements() *PlayerMovements {
	return &PlayerMovements{
		queuedMovements: make(map[int64]*playerMovement),
	}
}

func (p *PlayerMovements) CanMove(tick int64) bool {
	p.queuedMovementsMtx.Lock()
	defer p.queuedMovementsMtx.Unlock()
	_, found := p.queuedMovements[tick]
	return !found
}

func (p *PlayerMovements) AddMovement(tick int64, offsetX, offsetY int) error {
	os.Stderr.WriteString(fmt.Sprintf("queue movement for tick: %v\n", tick))
	p.queuedMovementsMtx.Lock()
	defer p.queuedMovementsMtx.Unlock()
	_, found := p.queuedMovements[tick]
	if found {
		return ErrAlreadyMoved
	}
	p.queuedMovements[tick] = &playerMovement{
		offsetX: offsetX,
		offsetY: offsetY,
	}
	return nil
}

func (p *PlayerMovements) GetPlayerPos(newTick, x, y int64) (int64, int64) {
	p.queuedMovementsMtx.Lock()
	defer p.queuedMovementsMtx.Unlock()

	// delete the past
	for i := p.lastProcessedTick; i <= newTick; i++ {
		delete(p.queuedMovements, i)
	}

	// process the rest of the queue, in order
	pending := len(p.queuedMovements)
	for i := newTick + 1; pending > 0; i++ {
		movement, found := p.queuedMovements[i]
		if found {
			log.Warn("Applied movement for tick", zap.Int64("tick", i))
			x += int64(movement.offsetX)
			y += int64(movement.offsetY)
			pending--
		}
	}
	return x, y
}
