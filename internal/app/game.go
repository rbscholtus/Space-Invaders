// Package app is the Ebitengine front end: it reads the keyboard, runs the
// board one frame per Update and draws the coloured picture (DESIGN §4).
// Without ROMs it shows the ROM screen (§9.2). The emulation itself lives
// in internal/invaders.
package app

import (
	"encoding/json"
	"image"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/rbscholtus/Space-Invaders/assets"
	"github.com/rbscholtus/Space-Invaders/internal/app/cabinet"
	"github.com/rbscholtus/Space-Invaders/internal/app/crt"
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
	// Shader is how the picture is drawn: the CRT monitor, or clean
	// (-shader, ?shader=).
	Shader crt.Mode
	// View is the cabinet, the picture alone, or either by the window's
	// size (-view, ?view=). ViewGiven says the player chose it at start;
	// otherwise the one last chosen with V is used.
	View      View
	ViewGiven bool
	// Art reads the cabinet's images; nil shows the picture alone.
	Art ArtFetcher
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
	monitor  *crt.Renderer
	cabinet  *cabinetView
	settings Options
	view     View       // the view chosen: auto, cabinet or screen
	lay      layout     // where things go, from Layout
	keys     input.Keys // the keys held, for the pressed buttons

	cache    *platform.Cache
	prefs    *platform.Prefs
	restored chan restored
	incoming chan platform.Picked // dropped or picked files, read in goroutines
	state    string               // the last state published for the browser tests
	layout   string               // the last layout published
	hiScore  int                  // the HI-SCORE last remembered

	frames     *sound.Frames
	stream     *sound.Stream
	output     sound.Output
	audioTried bool

	screen []byte // the picture, 1 byte per pixel
	ticks  uint64 // Updates so far

	paused  bool // by the player (P)
	stopped bool // paused, for any reason, in the last Update

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
		screen:   make([]byte, screenW*screenH),
		scale:    1,
	}

	g.monitor = newMonitor(opts)
	g.cabinet = newCabinetView(loadCabinet(opts), opts.Art)

	applySaved(g.stream, g.prefs, opts.VolumeGiven)
	g.view = savedView(g.prefs, opts.View, opts.ViewGiven)

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

// loadCabinet reads the cabinet's layout. It was checked when it was made,
// so it only fails if the build is broken: then the app shows the picture
// alone.
func loadCabinet(opts Options) *cabinet.Cabinet {
	cab, err := cabinet.Parse(assets.CabinetJSON)
	if err != nil {
		if opts.OnError != nil {
			opts.OnError(err)
		}

		return nil
	}

	return cab
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
	g.ticks++
	g.monitor.Tick()
	g.startAudio()
	g.handleKeys()
	g.handleROMs()
	g.updateCabinet()
	g.publishState()
	g.publishLayout()

	g.stopped = g.paused || !platform.InFront(ebiten.IsFocused())

	g.monitor.SetDim(1)

	if g.stopped {
		g.monitor.SetDim(pausedDim)

		return nil
	}

	if g.board == nil {
		clear(g.screen)
		g.flow.Draw(g.screen, screenW)
	} else {
		g.keys = keys()
		g.board.SetInputs(g.controls.Update(g.keys))
		g.board.RunFrame()
		g.hiScore = saveHiScore(g.board, g.prefs, g.hiScore)
		g.stream.Push(g.frames.Render(g.board.SoundEvents()))
		invaders.Screen(g.board.Video(), g.screen)
	}

	g.monitor.Update(g.screen)

	return nil
}

// Draw draws the cabinet, if shown, and the monitor in it (DESIGN §11).
func (g *Game) Draw(dst *ebiten.Image) {
	g.cabinet.draw(dst, g.lay, g.keys)
	g.monitor.SetBackdropLevel(g.cabinet.backdropLevels(g.lay.view))
	g.monitor.SetScreen(g.lay.tube)
	g.monitor.Draw(dst, g.ticks)
}

// newMonitor makes the renderer in the chosen mode. If the CRT shaders
// don't compile on this machine, it reports that and draws the clean
// picture instead.
func newMonitor(opts Options) *crt.Renderer {
	m, err := crt.New(opts.Shader, overlay.Midway())
	if err == nil {
		return m
	}

	if opts.OnError != nil {
		opts.OnError(err)
	}

	return crt.NewClean(overlay.Midway())
}

// Layout returns the window's size in real pixels (CONTRIBUTING, HiDPI).
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.scale = ebiten.Monitor().DeviceScaleFactor()
	g.width = int(float64(outsideWidth) * g.scale)
	g.height = int(float64(outsideHeight) * g.scale)
	g.lay = computeLayout(g.cabinet.cab, g.width, g.height, g.scale, g.view)

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
// §13): P pause, M mute, - and = volume, O open ROM files, F forget them,
// V switch between the cabinet and the picture alone.
func (g *Game) handleKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyV) && g.cabinet.cab != nil {
		g.view = ViewCabinet
		if g.lay.view == ViewCabinet {
			g.view = ViewScreen
		}

		g.prefs.Set(prefView, g.view.String())
	}

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
	g.hiScore = restoreHiScore(board, g.prefs)
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

// forget stops the game, deletes the remembered image and HI-SCORE, and
// shows the ROM screen.
func (g *Game) forget() {
	g.board = nil
	g.hiScore = 0
	g.prefs.Set(prefHiScore, "")
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
	hiScore := 0
	if g.board != nil {
		hiScore = g.board.HiScore()
	}

	state, err := json.Marshal(struct {
		Mode    string `json:"mode"`
		Message string `json:"message"`
		HiScore int    `json:"hiscore"`
		Shader  string `json:"shader"`
	}{g.flow.Mode().String(), g.flow.Message(), hiScore, g.monitor.Mode().String()})
	if err != nil || string(state) == g.state {
		return
	}

	g.state = string(state)
	platform.PublishState(g.state)
}

// cssRect is a rectangle in CSS pixels, for the browser tests.
type cssRect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

func (g *Game) css(r image.Rectangle) cssRect {
	return cssRect{
		float64(r.Min.X) / g.scale, float64(r.Min.Y) / g.scale,
		float64(r.Dx()) / g.scale, float64(r.Dy()) / g.scale,
	}
}

// publishLayout puts the layout on the page for the browser tests, when it
// changes: the view, where the picture and the cabinet are (CSS pixels),
// and whether the cabinet's art has loaded.
func (g *Game) publishLayout() {
	part := ""
	if g.lay.view == ViewCabinet {
		part = "whole"
		if g.lay.upper {
			part = "upper"
		}
	}

	lay, err := json.Marshal(struct {
		View    string  `json:"view"`
		Chosen  string  `json:"chosen"`
		Part    string  `json:"part"`
		Tube    cssRect `json:"tube"`
		Cabinet cssRect `json:"cabinet"`
		Art     bool    `json:"art"`
	}{g.lay.view.String(), g.view.String(), part, g.css(g.lay.tube), g.css(g.lay.cabinet), g.cabinet.loaded()})
	if err != nil || string(lay) == g.layout {
		return
	}

	g.layout = string(lay)
	platform.PublishLayout(g.layout)
}

// updateCabinet loads the cabinet's art as the layout needs it, and gives
// the monitor its backdrop when it arrives.
func (g *Game) updateCabinet() {
	g.cabinet.want(g.lay)

	backdrop, err := g.cabinet.update()
	if err != nil {
		g.report(err)
	}

	if backdrop != nil {
		g.monitor.SetBackdrop(backdrop)
	}
}
