package ui

import (
	"image"
	"image/color"
	"log"
	"sync"

	"github.com/blizzy78/ebitenui/widget"
	"github.com/code-cell/esive/client"
	esive_grpc "github.com/code-cell/esive/grpc"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type WorldView struct {
	width  int
	height int

	face       font.Face
	client     *client.Client
	prediction *Prediction

	mtx         sync.Mutex
	renderables map[int64]*esive_grpc.Renderable
	playerX     int64
	playerY     int64
	visibility  int64

	widget  *widget.Widget
	focused bool
}

func NewWorldView(width, height int, c *client.Client, prediction *Prediction, visibility int64) *WorldView {
	// tt, err := opentype.Parse(gomono.TTF)
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	ff, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    18,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	wv := &WorldView{
		width:  width,
		height: height,

		face:       ff,
		client:     c,
		prediction: prediction,
		visibility: visibility,

		renderables: make(map[int64]*esive_grpc.Renderable),

		widget: widget.NewWidget(),
	}

	c.AddUpdateRenderableHandler(func(id, tick int64, renderable *esive_grpc.Renderable) {
		wv.mtx.Lock()
		defer wv.mtx.Unlock()
		wv.renderables[id] = renderable
		if id == c.PlayerID {
			prediction.UpdatePlayerPositionFromServer(tick, renderable.Position.X, renderable.Position.Y, renderable.Velocity.X, renderable.Velocity.Y)
		}
	})

	c.AddDeleteRenderableHandler(func(id, tick int64) {
		wv.mtx.Lock()
		defer wv.mtx.Unlock()
		delete(wv.renderables, id)
	})

	return wv
}

func (g *WorldView) Render(screen *ebiten.Image, def widget.DeferredRenderFunc) {
	g.widget.Render(screen, def)
	g.draw(screen)
}

func (g *WorldView) draw(screen *ebiten.Image) {
	r := g.widget.Rect
	cellWidth := float64(r.Bounds().Dx()) / float64(g.width)
	cellHeight := float64(r.Bounds().Dy()) / float64(g.height)

	g.playerX, g.playerY = g.prediction.GetPredictedPlayerPosition(g.client.Tick.Current())

	g.mtx.Lock()
	defer g.mtx.Unlock()

	for id, r := range g.renderables {
		x := r.Position.X
		y := r.Position.Y
		if id == g.client.PlayerID {
			x = g.playerX
			y = g.playerY
		}

		col := color.RGBA{
			R: uint8(r.Color >> 24),
			G: uint8(r.Color >> 16),
			B: uint8(r.Color >> 8),
			A: uint8(r.Color),
		}

		text.Draw(screen,
			r.Char,
			g.face,
			int((x-g.playerX)+g.visibility)*int(cellWidth),
			int((y-g.playerY)+g.visibility+1)*int(cellHeight),
			col)
	}
}

func (g *WorldView) GetWidget() *widget.Widget {
	return g.widget
}

func (g *WorldView) PreferredSize() (int, int) {
	return g.width * 15, g.height * 15
}

func (g *WorldView) SetLocation(rect image.Rectangle) {
	g.widget.Rect = rect
}

func (g *WorldView) Focus(focused bool) {
	g.focused = focused
}
