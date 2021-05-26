package main

import (
	"fmt"
	"image/color"

	"github.com/blizzy78/ebitenui/image"
	"github.com/blizzy78/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/font"
)

type Menu struct {
	*widget.Container

	chatText  *widget.Text
	TextInput *TextInput
}

func NewMenu(ff font.Face) *Menu {
	menu := &Menu{}
	idle, _, err := ebitenutil.NewImageFromFile("./cmd/client2/graphics/text-input-idle.png")
	if err != nil {
		panic(err)
	}

	menu.TextInput = NewTextInput(
		widget.TextInputOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle: image.NewNineSlice(idle, [3]int{9, 14, 6}, [3]int{9, 14, 6}),
		}),
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          hexToColor(textIdleColor),
			Disabled:      hexToColor(textDisabledColor),
			Caret:         hexToColor(textInputCaretColor),
			DisabledCaret: hexToColor(textInputDisabledCaretColor),
		}),
		widget.TextInputOpts.Padding(widget.Insets{
			Left:   13,
			Right:  13,
			Top:    7,
			Bottom: 7,
		}),
		widget.TextInputOpts.Face(ff),
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(ff, 2),
		),
		widget.TextInputOpts.Placeholder("Enter text here"),
	)

	menu.Container = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true, false}),
			widget.GridLayoutOpts.Spacing(0, 5),
		)),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.RGBA{0x13, 0x1a, 0x22, 0xff})),
	)

	menu.chatText = widget.NewText(
		widget.TextOpts.Text("", ff, color.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	menu.Container.AddChild(NewScrollableContainer(menu.chatText))
	menu.Container.AddChild(menu.TextInput)

	menu.chatText.Label += "Welcome to Esive!\n"
	menu.chatText.Label += "\n"
	menu.chatText.Label += "Use the arrows in your keyboard to move around.\n"
	menu.chatText.Label += "Type '/help' in the chat to see the list of commands.\n"
	menu.chatText.Label += "Press 'esc' to close the game.\n"
	menu.chatText.Label += "\n"
	return menu
}

func (m *Menu) HandleChatMessage(from, message string) {
	m.chatText.Label += fmt.Sprintf("%v: %v\n", from, message)
}
