// Package app is the Ebitengine front end: it reads the keyboard, runs the
// board one frame per Update and draws the coloured picture (DESIGN §4).
// Without ROMs it shows the ROM screen (§9.2). The emulation itself lives
// in internal/invaders.
package app

import (
	"encoding/json"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/rbscholtus/Space-Invaders/internal/app/input"
	"github.com/rbscholtus/Space-Invaders/internal/app/overlay"
	"github.com/rbscholtus/Space-Invaders/internal/app/romscreen"
	"github.com/rbscholtus/Space-Invaders/internal/app/sound"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
	"github.com/rbscholtus/Space-Invaders/internal/platform"
	"github.com/rbscholtus/Space-Invaders/internal/romset"
)

const (
	screenW = invaders.ScreenWidth
	screenH = invaders.ScreenHeight

	// pausedDim is the brightness of the picture while paused.
	pausedDim = 0.4
	// volumeStep is how much the volume keys change the volume.
	volumeStep = 0.1
)

// cacheKey is the remembered ROM image's key: the MAME set's name.
const cacheKey = "invaders"

// Options are the player's settings.
type Options struct {
	// ROMs is a checked ROM image (the -roms flag), or nil: the remembered
	// one, or the ROM screen.
	ROMs   []byte
	DIPs   invaders.DIPs
	Volume float64 // 0–1
	// VolumeGiven says the player chose the volume at start (-volume,
	// ?volume=); otherwise the remembered one is used.
	VolumeGiven bool
	// OnError is told about problems that don't stop the game, such as
	// no audio device (the game then runs silently).
	OnError func(error)
}

// restored is the remembered image, read in a goroutine.
type restored struct {
	image []byte
	found bool
	err   error
}

// Game implements ebiten.Game.
type Game struct {
	board    *invaders.Board // nil until there are ROMs
	flow     *romscreen.Flow
	controls input.Controls
	painter  *overlay.Painter
	settings Options

	cache    *platform.Cache
	prefs    *platform.Prefs
	restored chan restored
	incoming chan platform.Picked // dropped or picked files, read in goroutines
	state    string               // the last state published for the browser tests

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

// New returns a Game: running opts.ROMs if given, else looking for the
// remembered ROMs and showing the ROM screen meanwhile.
func New(opts Options) *Game {
	g := &Game{
		flow:     romscreen.New(platform.CanPickFiles()),
		settings: opts,
		cache:    platform.NewCache(),
		prefs:    platform.NewPrefs(),
		restored: make(chan restored, 1),
		incoming: make(chan platform.Picked, 1),
		frames:   sound.NewFrames(),
		stream:   sound.NewStream(opts.Volume),
		painter:  overlay.NewPainter(overlay.Midway()),
		screen:   make([]byte, screenW*screenH),
		rgba:     make([]byte, screenW*screenH*4),
		picture:  ebiten.NewImage(screenW, screenH),
		dirty:    true,
		scale:    1,
	}

	applySaved(g.stream, g.prefs, opts.VolumeGiven)

	if opts.ROMs != nil {
		g.flow.Start()
		g.startBoard(opts.ROMs)
	} else {
		go func() {
			image, found, err := g.cache.Get(cacheKey)
			g.restored <- restored{image, found, err}
		}()
	}

	return g
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

// Update runs one emulated frame (DESIGN §8) and its sound (§14.4), or the
// ROM screen, unless paused: by the player, or because the app isn't in
// front. While paused nothing is rendered, so the sound falls silent.
func (g *Game) Update() error {
	g.startAudio()
	g.handleKeys()
	g.handleROMs()
	g.publishState()

	g.stopped = g.paused || !platform.InFront(ebiten.IsFocused())
	if g.stopped {
		return nil
	}

	if g.board == nil {
		clear(g.screen)
		g.flow.Draw(g.screen, screenW)
	} else {
		g.board.SetInputs(g.controls.Update(keys()))
		g.board.RunFrame()
		g.stream.Push(g.frames.Render(g.board.SoundEvents()))
		invaders.Screen(g.board.Video(), g.screen)
	}

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

// startAudio starts the audio output once: on the desktop at once, in the
// browser on the first key press, click, touch or drop, as browsers allow
// sound only after a user gesture (DESIGN §14.4). Without an audio device
// the game runs silently.
func (g *Game) startAudio() {
	if g.audioTried || (runtime.GOOS == "js" && !gesture()) {
		return
	}

	g.audioTried = true

	if err := g.output.Start(g.stream); err != nil {
		g.report(err)
	}
}

// gesture reports whether the player just pressed a key, clicked, touched
// or dropped something.
func gesture() bool {
	return len(inpututil.AppendJustPressedKeys(nil)) > 0 ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		len(inpututil.AppendJustPressedTouchIDs(nil)) > 0 ||
		ebiten.DroppedFiles() != nil
}

// handleKeys handles the keys that aren't the cabinet's controls (DESIGN
// §13): P pause, M mute, - and = volume, O open ROM files, F forget them.
func (g *Game) handleKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.paused = !g.paused
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.stream.SetMuted(!g.stream.Muted())
		saveSound(g.stream, g.prefs)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		g.stream.SetVolume(g.stream.Volume() - volumeStep)
		saveSound(g.stream, g.prefs)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		g.stream.SetVolume(g.stream.Volume() + volumeStep)
		saveSound(g.stream, g.prefs)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyO) && g.canOpenFiles() {
		picked := platform.PickFiles()

		go func() { g.incoming <- <-picked }()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) && g.flow.Mode() != romscreen.Loading {
		g.forget()
	}
}

// handleROMs takes the remembered image, and dropped or picked files, as
// they arrive; it never waits.
func (g *Game) handleROMs() {
	select {
	case r := <-g.restored:
		if image := g.flow.Restored(r.image, r.found, r.err); image != nil {
			g.startBoard(image)
		}
	case p := <-g.incoming:
		if p.Err == nil && len(p.Files) == 0 {
			break // the player cancelled the picker
		}

		if p.Err != nil {
			g.flow.Fail(p.Err)

			break
		}

		if image := g.flow.Offer(romset.Files(p.Files)); image != nil {
			g.startBoard(image)
			g.remember(image)
		}
	default:
	}

	if files := ebiten.DroppedFiles(); files != nil && g.flow.Mode() == romscreen.NoROMs {
		go func() { g.incoming <- readDropped(files) }()
	}
}

// startBoard powers on a board with the image, with fresh sound.
func (g *Game) startBoard(image []byte) {
	board, err := invaders.New(image, g.settings.DIPs)
	if err != nil {
		g.flow.Fail(err)

		return
	}

	g.board = board
	g.controls = input.Controls{}
	g.frames = sound.NewFrames()
}

// remember stores the image for the next visit, in the background.
func (g *Game) remember(image []byte) {
	go func() {
		if err := g.cache.Put(cacheKey, image); err != nil {
			g.report(err)
		}
	}()
}

// forget stops the game, deletes the remembered image and shows the ROM
// screen.
func (g *Game) forget() {
	g.board = nil
	g.frames = sound.NewFrames()
	g.flow.Forget()

	go func() {
		if err := g.cache.Delete(cacheKey); err != nil {
			g.report(err)
		}
	}()
}

// canOpenFiles reports whether O opens the file picker now. There's no
// guard against opening it twice: the browser's picker is modal, and a
// guard would get stuck in browsers that don't report a cancel.
func (g *Game) canOpenFiles() bool {
	return g.flow.Mode() == romscreen.NoROMs && platform.CanPickFiles()
}

func (g *Game) report(err error) {
	if g.settings.OnError != nil {
		g.settings.OnError(err)
	}
}

// publishState puts the app's state on the page for the browser tests,
// when it changes.
func (g *Game) publishState() {
	state, err := json.Marshal(struct {
		Mode    string `json:"mode"`
		Message string `json:"message"`
	}{g.flow.Mode().String(), g.flow.Message()})
	if err != nil || string(state) == g.state {
		return
	}

	g.state = string(state)
	platform.PublishState(g.state)
}
