package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"strconv"

	_ "image/png"

	"github.com/blizzy78/ebitenui"
	"github.com/blizzy78/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
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
	menuImage      *ebiten.Image

	menuUI *ebitenui.UI

	lastTick int64

	// When the player changes velocity twice in a tick, we 'plan' it for next tick. For example, moving just one tile requires setting velocity to the direction, and setting it back to 0 on the next tick.
	nextSetVelocity bool
	nextVelocityX   int
	nextVelocityY   int
}

var (
	// This is configured when building releases to point to the test server.
	defaultAddr = "localhost:9000"

	textIdleColor                  = "dff4ff"
	textDisabledColor              = "5a7a91"
	textInputCaretColor            = "e7c34b"
	textInputDisabledCaretColor    = "766326"
	listSelectedBackground         = "4b687a"
	listDisabledSelectedBackground = "2a3944"
)

func NewGame() *Game {
	addr := flag.String("addr", defaultAddr, "Server address")
	name := flag.String("name", "", "Your name. Required.")
	flag.Parse()
	if *name == "" {
		panic("the `name` flag is required.")
	}

	prediction := NewPrediction()

	client := NewClient(*addr, *name, prediction)
	if err := client.Connect(); err != nil {
		panic(err)
	}
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	ff, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    18,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	menu := NewMenu(ff)
	worldView := NewWorldView(31, 31, client, prediction, 15)
	worldView.Focus(true)

	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(0, 5),
		)),
	)
	worldViewContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	worldViewContainer.AddChild(worldView)
	container.AddChild(worldViewContainer)
	container.AddChild(menu)

	return &Game{
		ScreenWidth:     800,
		ScreenHeight:    467,
		backgroundColor: color.RGBA{0x13, 0x1a, 0x22, 0xff},

		client:         client,
		input:          NewInput(),
		prediction:     prediction,
		worldView:      worldView,
		worldViewImage: ebiten.NewImage(31*15, 31*15),
		menuImage:      ebiten.NewImage(800-31*15, 467),

		menuUI: &ebitenui.UI{
			Container: container,
		},
	}
}

func (g *Game) Update() error {
	g.menuUI.Update()

	clientTick := g.client.tick.Current()
	if g.worldView.focused {
		g.input.Update()
		if clientTick != g.lastTick && g.nextSetVelocity {
			g.prediction.AddVelocity(clientTick, g.nextVelocityX, g.nextVelocityY)
			fmt.Printf("[%v] Sending velocity to (%v,%v)\n", clientTick, g.nextVelocityX, g.nextVelocityY)
			go g.client.SetVelocity(g.nextVelocityX, g.nextVelocityY)
			g.nextSetVelocity = false
		}
		x, y, changed := g.input.Dir()
		if changed {
			if g.prediction.CanMove(clientTick) {
				g.prediction.AddVelocity(clientTick, x, y)
				fmt.Printf("[%v] Sending velocity to (%v,%v)\n", clientTick, x, y)
				go g.client.SetVelocity(x, y)
			} else {
				// The player already moved this tick. Plan movement for next tick.
				g.nextSetVelocity = true
				g.nextVelocityX = x
				g.nextVelocityY = y
			}
		}

	}
	g.lastTick = clientTick
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(g.backgroundColor)
	g.menuUI.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	game := NewGame()
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowTitle("Esive")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func hexToColor(h string) color.Color {
	u, err := strconv.ParseUint(h, 16, 0)
	if err != nil {
		panic(err)
	}

	return color.RGBA{
		R: uint8(u & 0xff0000 >> 16),
		G: uint8(u & 0xff00 >> 8),
		B: uint8(u & 0xff),
		A: 255,
	}
}
