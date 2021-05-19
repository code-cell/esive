package main

import (
	"image/color"

	"github.com/blizzy78/ebitenui/image"
	"github.com/blizzy78/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/font"
)

type Menu struct {
	*widget.Container

	chatList  *widget.List
	textInput *widget.TextInput
}

func NewMenu(ff font.Face) *Menu {
	menu := &Menu{}
	idle, _, err := ebitenutil.NewImageFromFile("./cmd/client2/graphics/text-input-idle.png")
	if err != nil {
		panic(err)
	}
	listIdle, _, err := ebitenutil.NewImageFromFile("./cmd/client2/graphics/list-idle.png")
	if err != nil {
		panic(err)
	}

	disabled, _, err := ebitenutil.NewImageFromFile("./cmd/client2/graphics/list-disabled.png")
	if err != nil {
		panic(err)
	}

	mask, _, err := ebitenutil.NewImageFromFile("./cmd/client2/graphics/list-mask.png")
	if err != nil {
		panic(err)
	}
	trackIdle, _, err := ebitenutil.NewImageFromFile("./cmd/client2/graphics/list-track-idle.png")
	if err != nil {
		panic(err)
	}

	trackDisabled, _, err := ebitenutil.NewImageFromFile("./cmd/client2/graphics/list-track-disabled.png")
	if err != nil {
		panic(err)
	}

	handleIdle, _, err := ebitenutil.NewImageFromFile("./cmd/client2/graphics/slider-handle-idle.png")
	if err != nil {
		panic(err)
	}

	handleHover, _, err := ebitenutil.NewImageFromFile("./cmd/client2/graphics/slider-handle-hover.png")
	if err != nil {
		panic(err)
	}
	menu.chatList = widget.NewList(
		widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
		}))),
		widget.ListOpts.ScrollContainerOpts(widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle:     image.NewNineSlice(listIdle, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(disabled, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Mask:     image.NewNineSlice(mask, [3]int{26, 10, 23}, [3]int{26, 10, 26}),
		})),
		widget.ListOpts.SliderOpts(
			widget.SliderOpts.Images(&widget.SliderTrackImage{
				Idle:     image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
				Hover:    image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
				Disabled: image.NewNineSlice(trackDisabled, [3]int{0, 5, 0}, [3]int{25, 12, 25}),
			}, &widget.ButtonImage{
				Idle:     image.NewNineSliceSimple(handleIdle, 0, 5),
				Hover:    image.NewNineSliceSimple(handleHover, 0, 5),
				Pressed:  image.NewNineSliceSimple(handleHover, 0, 5),
				Disabled: image.NewNineSliceSimple(handleIdle, 0, 5),
			}),
			widget.SliderOpts.HandleSize(5),
			widget.SliderOpts.TrackPadding(widget.Insets{
				Top:    5,
				Bottom: 24,
			}),
		),
		widget.ListOpts.HideHorizontalSlider(),
		widget.ListOpts.Entries([]interface{}{"One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine", "Ten", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine", "Ten"}),
		widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
			return e.(string)
		}),
		widget.ListOpts.EntryFontFace(ff),
		widget.ListOpts.EntryColor(&widget.ListEntryColor{
			Unselected:         hexToColor(textIdleColor),
			DisabledUnselected: hexToColor(textDisabledColor),

			Selected:         hexToColor(textIdleColor),
			DisabledSelected: hexToColor(textDisabledColor),

			SelectedBackground:         hexToColor(listSelectedBackground),
			DisabledSelectedBackground: hexToColor(listDisabledSelectedBackground),
		}),
		widget.ListOpts.EntryTextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    2,
			Bottom: 2,
		}),
	)

	menu.textInput = widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle: image.NewNineSlice(idle, [3]int{9, 14, 6}, [3]int{9, 14, 6}),
			// Disabled: image.NewNineSlice(disabled, [3]int{9, 14, 6}, [3]int{9, 14, 6}),
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

	menu.Container.AddChild(menu.chatList)
	menu.Container.AddChild(menu.textInput)

	return menu
}
