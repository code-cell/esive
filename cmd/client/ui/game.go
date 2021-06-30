package ui

import (
	"fmt"
	"image/color"
	"log"
	"strconv"

	"github.com/blizzy78/ebitenui"
	"github.com/blizzy78/ebitenui/widget"
	"github.com/code-cell/esive/client"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	textIdleColor                  = "dff4ff"
	textDisabledColor              = "5a7a91"
	textInputCaretColor            = "e7c34b"
	textInputDisabledCaretColor    = "766326"
	listSelectedBackground         = "4b687a"
	listDisabledSelectedBackground = "2a3944"
)

type Game struct {
	ScreenWidth  int
	ScreenHeight int

	backgroundColor color.Color

	client         *client.Client
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

func NewGame(c *client.Client) *Game {
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
	c.AddChatHandler(menu.HandleChatMessage)
	prediction := NewPrediction()

	menu.textInput.SendEvent.AddHandler(func(args interface{}) {
		eventArgs := args.(*TextInputSendEventArgs)
		c.SendChatMessage(eventArgs.InputText)
	})

	worldView := NewWorldView(31, 31, c, prediction, 15)
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

		client:         c,
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

	clientTick := g.client.Tick.Current()
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

func (g *Game) Run() {
	ebiten.SetWindowSize(g.ScreenWidth, g.ScreenHeight)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowTitle("Esive")
	if err := ebiten.RunGame(g); err != nil {
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
