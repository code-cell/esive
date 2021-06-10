package ui

import (
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Input represents the current key states.
type Input struct {
	prevX int
	prevY int
}

// NewInput generates a new Input object.
func NewInput() *Input {
	return &Input{}
}

// Update updates the current input states.
func (i *Input) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}
}

// Dir returns a currently pressed direction.
// Dir returns false if no direction key is pressed.
func (i *Input) Dir() (int, int, bool) {
	x := 0
	y := 0
	if inpututil.KeyPressDuration(ebiten.KeyArrowUp) > 0 {
		y -= 1
	}
	if inpututil.KeyPressDuration(ebiten.KeyArrowDown) > 0 {
		y += 1
	}
	if inpututil.KeyPressDuration(ebiten.KeyArrowLeft) > 0 {
		x -= 1
	}
	if inpututil.KeyPressDuration(ebiten.KeyArrowRight) > 0 {
		x += 1
	}

	changed := x != i.prevX || y != i.prevY
	i.prevX = x
	i.prevY = y
	return x, y, changed
}
