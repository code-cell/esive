// Copyright 2016 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
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
