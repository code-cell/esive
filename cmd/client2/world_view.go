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

	face   font.Face
	client *Client

	mtx         sync.Mutex
	renderables map[int64]*esive_grpc.Renderable
	playerX     int64
	playerY     int64
	visibility  int64
}

func NewWorldView(width, height int, client *Client, visibility int64) *WorldView {
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
		visibility: visibility,

		renderables: make(map[int64]*esive_grpc.Renderable),
	}
}

func (g *WorldView) Draw(screen *ebiten.Image) {
	screen.Fill(color.Transparent)
	cellWidth := float64(screen.Bounds().Dx()) / float64(g.width)
	cellHeight := float64(screen.Bounds().Dy()) / float64(g.height)

	renderables := g.client.Renderables()
	// First readjust playerX/playerY
	for id, r := range renderables {
		if id == g.client.PlayerID {
			g.playerX = r.Position.X
			g.playerY = r.Position.Y
		}
	}

	for _, r := range renderables {
		text.Draw(screen,
			r.Char,
			g.face,
			int((r.Position.X-g.playerX)+g.visibility)*int(cellWidth),
			int((r.Position.Y-g.playerY)+g.visibility)*int(cellHeight),
			color.Black)
	}

	// for x := 0; x < g.width; x++ {
	// 	for y := 0; y < g.height; y++ {
	// 		var col color.Color
	// 		if (x+y)%2 == 0 {
	// 			col = color.RGBA{0x80, 0, 0, 0xff}
	// 		} else {
	// 			col = color.RGBA{0, 0x80, 0, 0xff}
	// 		}
	// 		text.Draw(screen, "@", g.face, int(cellWidth)*x, (int(cellHeight)*(y+1))-2, col)
	// 	}
	// }
}
