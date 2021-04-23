package main

import (
	"errors"
	"sync"
)

type PlayerMovements struct {
	X                  int64
	Y                  int64
	tick               int64
	queuedMovementsMtx sync.Mutex
	queuedMovements    map[int64]*playerMovement
}
type playerMovement struct {
	offsetX int
	offsetY int
}

func (p *PlayerMovements) AddMovement(tick int64, offsetX, offsetY int) error {
	if tick <= p.tick {
		return errors.New("This tick has been already processed")
	}

	p.queuedMovementsMtx.Lock()
	defer p.queuedMovementsMtx.Unlock()
	p.queuedMovements[tick-p.tick] = &playerMovement{
		offsetX: offsetX,
		offsetY: offsetY,
	}
	return nil
}
