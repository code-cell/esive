package main

import (
	"errors"
	"sync"
)

type Prediction struct {
	mtx              sync.Mutex
	queuedVelocities map[int64]*velocity

	serverTick int64
	serverX    int64
	serverY    int64
	serverVX   int64
	serverVY   int64
}
type velocity struct {
	x int
	y int
}

var ErrAlreadyMoved = errors.New("The player has already moved on that tick")

func NewPrediction() *Prediction {
	return &Prediction{
		queuedVelocities: make(map[int64]*velocity),
	}
}

func (p *Prediction) CanMove(tick int64) bool {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.queuedVelocities[tick]
	return !found
}

func (p *Prediction) AddVelocity(tick int64, x, y int) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.queuedVelocities[tick]
	if found {
		return ErrAlreadyMoved
	}
	p.queuedVelocities[tick] = &velocity{
		x: x,
		y: y,
	}
	return nil
}

func (p *Prediction) UpdatePlayerPositionFromServer(tick, x, y, vx, vy int64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	for i := p.serverTick; i <= tick; i++ {
		delete(p.queuedVelocities, i)
	}

	p.serverTick = tick
	p.serverX = x
	p.serverY = y
	p.serverVX = vx
	p.serverVY = vy
}

func (p *Prediction) GetPredictedPlayerPosition(clientTick int64) (int64, int64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	// process the rest of the queue, in order
	x := p.serverX
	y := p.serverY
	vx := p.serverVX
	vy := p.serverVY
	for i := p.serverTick + 1; i <= clientTick; i++ {
		velocity, found := p.queuedVelocities[i]
		if found {
			vx = int64(velocity.x)
			vy = int64(velocity.y)
			x += vx
			y += vy
		}
	}
	return x, y
}
