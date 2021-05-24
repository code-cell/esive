package main

import (
	"github.com/blizzy78/ebitenui/event"
	"github.com/blizzy78/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type TextInputSendEventArgs struct {
	TextInput *TextInput
	InputText string
}

type TextInput struct {
	*widget.TextInput
	focused bool

	wasEnterPressed bool
	SendEvent       *event.Event
}

func NewTextInput(opts ...widget.TextInputOpt) *TextInput {
	return &TextInput{
		TextInput: widget.NewTextInput(opts...),

		SendEvent: &event.Event{},
	}
}

func (t *TextInput) Render(screen *ebiten.Image, def widget.DeferredRenderFunc) {
	t.TextInput.Render(screen, def)
	if t.focused && ebiten.IsKeyPressed(ebiten.KeyEnter) {
		if t.wasEnterPressed {
			return
		}
		t.wasEnterPressed = true
		t.SendEvent.Fire(&TextInputSendEventArgs{
			TextInput: t,
			InputText: t.InputText,
		})
		t.InputText = ""
	} else {
		t.wasEnterPressed = false
	}
}

func (t *TextInput) Focus(focused bool) {
	t.TextInput.Focus(focused)
	t.focused = focused
}
