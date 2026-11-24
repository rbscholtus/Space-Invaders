// Package app is the Ebitengine front end: it reads the keyboard, runs the
// board one frame per Update and draws the coloured picture (DESIGN §4). The
// emulation itself lives in internal/invaders.
package app

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/rbscholtus/Space-Invaders/internal/app/input"
	"github.com/rbscholtus/Space-Invaders/internal/app/overlay"
	"github.com/rbscholtus/Space-Invaders/internal/app/sound"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

const (
	screenW = invaders.ScreenWidth
	screenH = invaders.ScreenHeight

	// pausedDim is the brightness of the picture while paused.
	pausedDim = 0.4
	// volumeStep is how much the volume keys change the volume.
	volumeStep = 0.1
)

// Options are the player's settings.
type Options struct {
	Volume float64 // 0–1
	// OnError is told about problems that don't stop the game, such as
	// no audio device (the game then runs silently).
	OnError func(error)
}

// Game implements ebiten.Game.
type Game struct {
	board    *invaders.Board
	controls input.Controls
	painter  *overlay.Painter
	settings Options

	frames     *sound.Frames
	stream     *sound.Stream
	output     sound.Output
	audioTried bool

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
func New(board *invaders.Board, opts Options) *Game {
	return &Game{
		board:    board,
		settings: opts,
		frames:   sound.NewFrames(),
		stream:   sound.NewStream(opts.Volume),
		painter:  overlay.NewPainter(overlay.Midway()),
		screen:   make([]byte, screenW*screenH),
		rgba:     make([]byte, screenW*screenH*4),
		picture:  ebiten.NewImage(screenW, screenH),
		dirty:    true,
		scale:    1,
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

// Update runs one emulated frame (DESIGN §8) and its sound (§14.4),
// unless paused: by the player, or because the window isn't in front.
// While paused nothing is rendered, so the sound falls silent.
func (g *Game) Update() error {
	g.startAudio()
	g.handleKeys()

	g.stopped = g.paused || !ebiten.IsFocused()
	if g.stopped {
		return nil
	}

	g.board.SetInputs(g.controls.Update(keys()))
	g.board.RunFrame()
	g.stream.Push(g.frames.Render(g.board.SoundEvents()))

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

	// A paused game is shown dimmed. (ebitenutil's DebugPrint would pull
	// net/http into the WASM build, CONTRIBUTING.)
	g.opts.ColorScale.Reset()

	if g.stopped {
		g.opts.ColorScale.Scale(pausedDim, pausedDim, pausedDim, 1)
	}

	dst.DrawImage(g.picture, &g.opts)
}

// Layout returns the window's size in real pixels (CONTRIBUTING, HiDPI).
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.scale = ebiten.Monitor().DeviceScaleFactor()
	g.width = int(float64(outsideWidth) * g.scale)
	g.height = int(float64(outsideHeight) * g.scale)

	return g.width, g.height
}

// startAudio starts the audio output once. Without an audio device the
// game runs silently.
func (g *Game) startAudio() {
	if g.audioTried {
		return
	}

	g.audioTried = true

	if err := g.output.Start(g.stream); err != nil && g.settings.OnError != nil {
		g.settings.OnError(err)
	}
}

// handleKeys handles the keys that aren't the cabinet's controls (DESIGN
// §13): P pause, M mute, - and = volume.
func (g *Game) handleKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.paused = !g.paused
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.stream.SetMuted(!g.stream.Muted())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		g.stream.SetVolume(g.stream.Volume() - volumeStep)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		g.stream.SetVolume(g.stream.Volume() + volumeStep)
	}
}
