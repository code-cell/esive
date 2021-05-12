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
	prediction     *Prediction
	input          *Input
	worldView      *WorldView
	worldViewImage *ebiten.Image

	lastTick int64

	// When the player changes velocity twice in a tick, we 'plan' it for next tick. For example, moving just one tile requires setting velocity to the direction, and setting it back to 0 on the next tick.
	nextSetVelocity bool
	nextVelocityX   int
	nextVelocityY   int
}

var (
	// This is configured when building releases to point to the test server.
	defaultAddr = "localhost:9000"

	addr = flag.String("addr", defaultAddr, "Server address")
	name = flag.String("name", "", "Your name. Optional.")
)

func NewGame() *Game {
	flag.Parse()
	prediction := NewPrediction()

	client := NewClient(*addr, *name, prediction)
	if err := client.Connect(); err != nil {
		panic(err)
	}

	return &Game{
		ScreenWidth:     800,
		ScreenHeight:    467,
		backgroundColor: color.RGBA{0xfa, 0xf8, 0xef, 0xff},

		client:         client,
		input:          NewInput(),
		prediction:     prediction,
		worldView:      NewWorldView(31, 31, client, prediction, 15),
		worldViewImage: ebiten.NewImage(31*15, 31*15),
	}
}

func (g *Game) Update() error {
	clientTick := g.client.tick.Current()
	if clientTick != g.lastTick && g.nextSetVelocity {
		g.prediction.AddVelocity(clientTick, g.nextVelocityX, g.nextVelocityY)
		go g.client.SetVelocity(g.nextVelocityX, g.nextVelocityY)
		g.nextSetVelocity = false
	}
	g.input.Update()
	x, y, changed := g.input.Dir()
	if changed {
		if g.prediction.CanMove(clientTick) {
			g.prediction.AddVelocity(clientTick, x, y)
			go g.client.SetVelocity(x, y)
		} else {
			// The player already moved this tick. Plan movement for next tick.
			g.nextSetVelocity = true
			g.nextVelocityX = x
			g.nextVelocityY = y
		}
	}

	g.lastTick = clientTick
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
