package main

import (
	"image/color"
	"log"
	"sync"

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
	client     *Client
	prediction *Prediction

	mtx         sync.Mutex
	renderables map[int64]*esive_grpc.Renderable
	playerX     int64
	playerY     int64
	visibility  int64
}

func NewWorldView(width, height int, client *Client, prediction *Prediction, visibility int64) *WorldView {
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

	return &WorldView{
		width:  width,
		height: height,

		face:       ff,
		client:     client,
		prediction: prediction,
		visibility: visibility,

		renderables: make(map[int64]*esive_grpc.Renderable),
	}
}

func (g *WorldView) Draw(screen *ebiten.Image) {
	screen.Fill(color.Transparent)
	cellWidth := float64(screen.Bounds().Dx()) / float64(g.width)
	cellHeight := float64(screen.Bounds().Dy()) / float64(g.height)

	renderables := g.client.Renderables()

	g.playerX, g.playerY = g.prediction.GetPredictedPlayerPosition(g.client.tick.Current())

	for id, r := range renderables {
		x := r.Position.X
		y := r.Position.Y
		if id == g.client.PlayerID {
			x = g.playerX
			y = g.playerY
		}
		text.Draw(screen,
			r.Char,
			g.face,
			int((x-g.playerX)+g.visibility)*int(cellWidth),
			int((y-g.playerY)+g.visibility+1)*int(cellHeight),
			color.Black)
	}
}
