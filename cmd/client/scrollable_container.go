package main

import (
	"math"

	"github.com/blizzy78/ebitenui/image"
	"github.com/blizzy78/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func NewScrollableContainer(content widget.PreferredSizeLocateableWidget) widget.PreferredSizeLocateableWidget {
	listIdle, _, err := ebitenutil.NewImageFromFile("./cmd/client/graphics/list-idle.png")
	if err != nil {
		panic(err)
	}

	mask, _, err := ebitenutil.NewImageFromFile("./cmd/client/graphics/list-mask.png")
	if err != nil {
		panic(err)
	}
	trackIdle, _, err := ebitenutil.NewImageFromFile("./cmd/client/graphics/slider-track-idle.png")
	if err != nil {
		panic(err)
	}
	handleIdle, _, err := ebitenutil.NewImageFromFile("./cmd/client/graphics/slider-handle-idle.png")
	if err != nil {
		panic(err)
	}
	handleHover, _, err := ebitenutil.NewImageFromFile("./cmd/client/graphics/slider-handle-hover.png")
	if err != nil {
		panic(err)
	}

	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true, false}),
		// widget.GridLayoutOpts.Spacing(l.controlWidgetSpacing, l.controlWidgetSpacing),
		),
		),
	)
	chatCont := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout(widget.AnchorLayoutOpts.Padding(widget.Insets{
		Left:   30,
		Right:  30,
		Top:    2,
		Bottom: 2,
	}))))
	chatCont.AddChild(content)

	scrollContainer := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(chatCont),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: image.NewNineSlice(listIdle, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Mask: image.NewNineSlice(mask, [3]int{26, 10, 23}, [3]int{26, 10, 26}),
		}),
	)
	container.AddChild(scrollContainer)

	pageSizeFunc := func() int {
		return int(math.Round(float64(scrollContainer.ContentRect().Dy()) / float64(content.GetWidget().Rect.Dy()) * 1000))
	}

	vSlider := widget.NewSlider(
		widget.SliderOpts.Direction(widget.DirectionVertical),
		widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.PageSizeFunc(pageSizeFunc),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			scrollContainer.ScrollTop = float64(args.Slider.Current) / 1000
		}),

		widget.SliderOpts.Images(&widget.SliderTrackImage{
			Idle:  image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Hover: image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
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
	)
	container.AddChild(vSlider)

	scrollContainer.GetWidget().ScrolledEvent.AddHandler(func(args interface{}) {
		a := args.(*widget.WidgetScrolledEventArgs)
		p := pageSizeFunc() / 3
		if p < 1 {
			p = 1
		}
		vSlider.Current -= int(math.Round(a.Y * float64(p)))
	})

	hSlider := widget.NewSlider(
		widget.SliderOpts.Direction(widget.DirectionHorizontal),
		widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.PageSizeFunc(func() int {
			return int(math.Round(float64(scrollContainer.ContentRect().Dx()) / float64(content.GetWidget().Rect.Dx()) * 1000))
		}),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			scrollContainer.ScrollLeft = float64(args.Slider.Current) / 1000
		}),
		widget.SliderOpts.Images(&widget.SliderTrackImage{
			Idle:  image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Hover: image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
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
	)
	container.AddChild(hSlider)
	return container
}
