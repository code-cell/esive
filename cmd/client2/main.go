package main

import (
	"flag"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	ScreenWidth  int
	ScreenHeight int

	backgroundColor color.Color

	client         *Client
	input          *Input
	worldView      *WorldView
	worldViewImage *ebiten.Image
}

var (
	// This is configured when building releases to point to the test server.
	defaultAddr = "localhost:9000"

	addr = flag.String("addr", defaultAddr, "Server address")
	name = flag.String("name", "", "Your name. Optional.")
)

func NewGame() *Game {
	flag.Parse()

	client := NewClient(*addr, *name)
	if err := client.Connect(); err != nil {
		panic(err)
	}

	return &Game{
		ScreenWidth:     800,
		ScreenHeight:    467,
		backgroundColor: color.RGBA{0xfa, 0xf8, 0xef, 0xff},

		client:         client,
		input:          NewInput(),
		worldView:      NewWorldView(31, 31, client, 15),
		worldViewImage: ebiten.NewImage(31*15, 31*15),
	}
}

func (g *Game) Update() error {
	g.input.Update()
	x, y, changed := g.input.Dir()
	if changed {
		g.client.SetVelocity(x, y)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(g.backgroundColor)
	g.worldView.Draw(g.worldViewImage)
	op := &ebiten.DrawImageOptions{}
	// op.GeoM.Translate(10, 10)
	screen.DrawImage(g.worldViewImage, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ScreenWidth, g.ScreenHeight
}

func main() {
	game := NewGame()
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("Esive")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
