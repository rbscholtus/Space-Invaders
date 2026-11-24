// Package app is the Ebitengine front end: it reads the keyboard, runs the
// board one frame per Update and draws the coloured picture (DESIGN §4). The
// emulation itself lives in internal/invaders.
package app

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/rbscholtus/Space-Invaders/internal/app/input"
	"github.com/rbscholtus/Space-Invaders/internal/app/overlay"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

const (
	screenW = invaders.ScreenWidth
	screenH = invaders.ScreenHeight
)

// Game implements ebiten.Game.
type Game struct {
	board    *invaders.Board
	controls input.Controls
	painter  *overlay.Painter

	screen  []byte // the picture, 1 byte per pixel
	rgba    []byte // the coloured picture
	picture *ebiten.Image
	dirty   bool // rgba changed since the last Draw

	paused  bool // by the player (P)
	stopped bool // paused, for any reason, in the last Update

	opts   ebiten.DrawImageOptions
	scale  float64 // device scale factor
	width  int     // Layout's size, in real pixels
	height int
}

// New returns a Game running board.
func New(board *invaders.Board) *Game {
	return &Game{
		board:   board,
		painter: overlay.NewPainter(overlay.Midway()),
		screen:  make([]byte, screenW*screenH),
		rgba:    make([]byte, screenW*screenH*4),
		picture: ebiten.NewImage(screenW, screenH),
		dirty:   true,
		scale:   1,
	}
}

// keys reads the keyboard as mapped in DESIGN §13.
func keys() input.Keys {
	down := ebiten.IsKeyPressed

	return input.Keys{
		Coin:   down(ebiten.Key5) || down(ebiten.KeyC),
		Start1: down(ebiten.Key1),
		Start2: down(ebiten.Key2),
		Left:   down(ebiten.KeyArrowLeft) || down(ebiten.KeyA),
		Right:  down(ebiten.KeyArrowRight) || down(ebiten.KeyD),
		Fire:   down(ebiten.KeySpace) || down(ebiten.KeyArrowUp) || down(ebiten.KeyW),
		Tilt:   down(ebiten.KeyT),
	}
}

// Update runs one emulated frame (DESIGN §8), unless paused: by the player,
// or because the window isn't in front.
func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.paused = !g.paused
	}

	g.stopped = g.paused || !ebiten.IsFocused()
	if g.stopped {
		return nil
	}

	g.board.SetInputs(g.controls.Update(keys()))
	g.board.RunFrame()

	invaders.Screen(g.board.Video(), g.screen)
	g.painter.Paint(g.screen, g.rgba)
	g.dirty = true

	return nil
}

// Draw scales the picture to fit the window, keeping its shape, with sharp
// pixels.
func (g *Game) Draw(dst *ebiten.Image) {
	if g.dirty {
		g.picture.WritePixels(g.rgba)
		g.dirty = false
	}

	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	s := min(float64(w)/screenW, float64(h)/screenH)

	g.opts.GeoM.Reset()
	g.opts.GeoM.Scale(s, s)
	g.opts.GeoM.Translate((float64(w)-screenW*s)/2, (float64(h)-screenH*s)/2)
	g.opts.Filter = ebiten.FilterNearest
	dst.DrawImage(g.picture, &g.opts)

	if g.stopped {
		ebitenutil.DebugPrintAt(dst, "PAUSED", int(8*g.scale), int(8*g.scale))
	}
}

// Layout returns the window's size in real pixels (CONTRIBUTING, HiDPI).
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.scale = ebiten.Monitor().DeviceScaleFactor()
	g.width = int(float64(outsideWidth) * g.scale)
	g.height = int(float64(outsideHeight) * g.scale)

	return g.width, g.height
}
