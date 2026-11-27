// Package crt draws the monitor (DESIGN §11): the picture goes through a
// phosphor afterglow shader, then a CRT shader that adds the colour
// overlay, bloom, the tube's curve and its lines. A clean mode draws the
// plain coloured picture, pixel for pixel, for tests and slow machines.
//
// Adapted from Chippie's renderer (github.com/rbscholtus/chippie).
package crt

import (
	_ "embed"
	"errors"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/rbscholtus/Space-Invaders/internal/app/overlay"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

var (
	//go:embed crt.kage
	crtSource []byte
	//go:embed phosphor.kage
	phosphorSource []byte
)

// Mode selects how the picture is drawn.
type Mode uint8

// The modes.
const (
	ModeCRT   Mode = iota // the monitor, with its shaders
	ModeClean             // the plain picture, pixel for pixel
)

// ErrMode is returned for an unknown mode name.
var ErrMode = errors.New("crt: the mode must be crt or clean")

// ParseMode reads a mode's name, as the -shader flag and ?shader= give it.
func ParseMode(s string) (Mode, error) {
	switch s {
	case "crt":
		return ModeCRT, nil
	case "clean":
		return ModeClean, nil
	}

	return ModeCRT, fmt.Errorf("%w, not %q", ErrMode, s)
}

// String names the mode, as ParseMode reads it.
func (m Mode) String() string {
	if m == ModeClean {
		return "clean"
	}

	return "crt"
}

const (
	// PhosphorDecay is how much of a pixel's glow is left one frame later.
	// The monitor's phosphor is fast: enough to soften the moving shots
	// and the flicker of the sprites, not to smear them (⚠ tune by eye).
	PhosphorDecay = 0.6
	// settleFrames is how many afterglow steps follow the last new
	// picture, so the glow fades out while paused: 0.6^12 × 255 < 1.
	settleFrames = 12

	// maxRenderHeight caps the size the shaders work at; a bigger tube is
	// scaled up from it. Full-screen on a high-density display would
	// otherwise cost too much, on phones above all (as Chippie found).
	maxRenderHeight = 1280
	// The monitor's lines show from 2 pixels per line, fully from 4.
	scanFrom, scanFull = 2.0, 4.0
	// bloomReach is the bloom's reach, in the game's pixels.
	bloomReach = 1.2

	frameRate     = 60
	bytesPerPixel = 4
)

const (
	pictureW = invaders.ScreenWidth
	pictureH = invaders.ScreenHeight
)

// Renderer draws the picture. Images, shaders, options and uniforms are
// made once, or when the size changes, and reused every frame.
type Renderer struct {
	mode Mode

	crt      *ebiten.Shader
	phosphor *ebiten.Shader

	painter *overlay.Painter // clean: the overlay's colours; CRT: white
	pix     []byte           // RGBA staging buffer for fb
	fb      *ebiten.Image    // the picture, 224×256
	tintPix []byte           // the overlay's tint per pixel, RGBA
	tintFb  *ebiten.Image    // the same, 224×256

	up   *ebiten.Image // the picture at the render size, nearest-neighbour
	tint *ebiten.Image // the tint at the render size
	back *ebiten.Image // the backdrop at the render size
	// backdrop is the moon backdrop as given, any size; nil until set.
	backdrop *ebiten.Image
	phos     [2]*ebiten.Image // the afterglow, at the render size
	out      *ebiten.Image    // the CRT pass, when the tube is bigger than the render size
	cur      int              // phos[cur] is the latest afterglow

	screen image.Rectangle // the tube, in destination pixels
	size   image.Point     // the render size

	fresh  bool // a new picture hasn't reached the afterglow yet
	due    bool // an afterglow step is due: once per tick
	settle int  // afterglow steps left after the last new picture
	dim    float32

	upOpts   ebiten.DrawImageOptions
	outOpts  ebiten.DrawImageOptions
	phosOpts ebiten.DrawRectShaderOptions
	crtOpts  ebiten.DrawRectShaderOptions
}

// NewClean returns a renderer in clean mode, for the overlay's strips.
func NewClean(strips []overlay.Rect) *Renderer {
	return &Renderer{
		mode:    ModeClean,
		painter: overlay.NewPainter(strips),
		pix:     make([]byte, pictureW*pictureH*bytesPerPixel),
		fb:      ebiten.NewImage(pictureW, pictureH),
		dim:     1,
	}
}

// New returns a renderer for the overlay's strips, compiling the shaders
// in CRT mode.
func New(mode Mode, strips []overlay.Rect) (*Renderer, error) {
	r := NewClean(strips)
	if mode == ModeClean {
		return r, nil
	}

	r.mode = ModeCRT
	r.painter = overlay.NewPainter(nil)
	r.tintPix = Tint(strips)

	var err error

	if r.crt, err = ebiten.NewShader(crtSource); err != nil {
		return nil, fmt.Errorf("compile the CRT shader: %w", err)
	}

	if r.phosphor, err = ebiten.NewShader(phosphorSource); err != nil {
		return nil, fmt.Errorf("compile the phosphor shader: %w", err)
	}

	r.tintFb = ebiten.NewImage(pictureW, pictureH)
	r.tintFb.WritePixels(r.tintPix)

	r.phosOpts.Uniforms = map[string]any{"Decay": float32(PhosphorDecay)}
	r.crtOpts.Uniforms = map[string]any{
		"Time": float32(0), "Dim": float32(1), "Scan": float32(0),
		"Columns": float32(pictureW), "Bloom": float32(0),
		"Back": float32(0), "BackEdge": float32(0),
	}
	r.outOpts.Filter = ebiten.FilterLinear

	return r, nil
}

// Tint returns the overlay's tint for every pixel of the picture, as RGBA:
// the colour a lit pixel takes there (white outside the strips).
func Tint(strips []overlay.Rect) []byte {
	lit := make([]byte, pictureW*pictureH)
	for i := range lit {
		lit[i] = 1
	}

	tint := make([]byte, pictureW*pictureH*bytesPerPixel)
	overlay.NewPainter(strips).Paint(lit, tint)

	return tint
}

// Mode returns how the picture is drawn.
func (r *Renderer) Mode() Mode { return r.mode }

// SetDim sets the brightness: 1 normally, less while paused.
func (r *Renderer) SetDim(dim float32) { r.dim = dim }

// SetScreen sets where the tube is drawn, in destination pixels. The
// render-sized images are made again only when the size changes.
func (r *Renderer) SetScreen(rect image.Rectangle) {
	if rect.Empty() {
		return
	}

	sizeChanged := rect.Size() != r.screen.Size()
	r.screen = rect

	if !sizeChanged || r.mode != ModeCRT {
		return
	}

	r.size = RenderSize(rect.Size())

	for _, img := range []*ebiten.Image{r.up, r.tint, r.back, r.phos[0], r.phos[1], r.out} {
		if img != nil {
			img.Deallocate()
		}
	}

	w, h := r.size.X, r.size.Y
	r.up, r.tint, r.back = ebiten.NewImage(w, h), ebiten.NewImage(w, h), ebiten.NewImage(w, h)
	r.phos[0], r.phos[1] = ebiten.NewImage(w, h), ebiten.NewImage(w, h)
	r.out = nil

	if r.size != rect.Size() {
		r.out = ebiten.NewImage(w, h)
	}

	var m ebiten.DrawImageOptions

	m.GeoM.Scale(float64(w)/pictureW, float64(h)/pictureH)
	r.tint.DrawImage(r.tintFb, &m)
	r.scaleBackdrop()

	pxPerLine := float64(rect.Dx()) / pictureW
	r.crtOpts.Uniforms["Scan"] = float32(ScanStrength(pxPerLine))
	r.crtOpts.Uniforms["Bloom"] = float32(bloomReach * float64(h) / pictureH)
	r.fresh, r.due = true, true
}

// SetBackdrop sets the moon backdrop seen through the glass (DESIGN
// §11.3): any size, stretched over the tube. The CRT mode only: the clean
// picture has none.
func (r *Renderer) SetBackdrop(img *ebiten.Image) {
	r.backdrop = img
	r.scaleBackdrop()
}

// SetBackdropLevel sets the backdrop's brightness behind the tube, and at
// and beyond its edge: there it should match the glass around the screen.
// It can change every frame, to fade the backdrop in.
func (r *Renderer) SetBackdropLevel(behind, edge float32) {
	if r.mode != ModeCRT {
		return
	}

	r.crtOpts.Uniforms["Back"] = behind
	r.crtOpts.Uniforms["BackEdge"] = edge
}

// Screen returns where the tube is drawn.
func (r *Renderer) Screen() image.Rectangle { return r.screen }

// RenderSize returns the size the shaders work at for a tube of the given
// size: the tube's own, or smaller, keeping its shape, when it's taller
// than maxRenderHeight.
func RenderSize(tube image.Point) image.Point {
	if tube.Y <= maxRenderHeight {
		return tube
	}

	return image.Pt(max(1, tube.X*maxRenderHeight/tube.Y), maxRenderHeight)
}

// ScanStrength returns how strongly the monitor's lines show, 0–1, for
// the given width of a line in screen pixels: thinner lines would only
// make a moiré.
func ScanStrength(pxPerLine float64) float64 {
	return min(max((pxPerLine-scanFrom)/(scanFull-scanFrom), 0), 1)
}

// Update takes a new picture: one byte per pixel, 1 = lit, as
// invaders.Screen writes it. Call it for every emulated frame.
func (r *Renderer) Update(picture []byte) {
	r.painter.Paint(picture, r.pix)
	r.fb.WritePixels(r.pix)
	r.fresh = true
}

// Tick marks the passing of one emulated frame's time, whether or not the
// game runs: the afterglow moves on once per tick, however often the
// display draws (60 or 120 times a second).
func (r *Renderer) Tick() {
	if r.fresh || r.settle > 0 {
		r.due = true
	}
}

// Draw draws the tube into dst. tick is the number of Updates so far, for
// the flicker.
func (r *Renderer) Draw(dst *ebiten.Image, tick uint64) {
	if r.screen.Empty() {
		return
	}

	if r.mode == ModeClean {
		r.drawClean(dst)

		return
	}

	if r.due {
		r.step()
	}

	r.crtOpts.Uniforms["Time"] = float32(tick) / frameRate
	r.crtOpts.Uniforms["Dim"] = r.dim
	r.crtOpts.Images[0], r.crtOpts.Images[1], r.crtOpts.Images[2] = r.phos[r.cur], r.tint, r.back

	if r.out == nil {
		r.crtOpts.GeoM.Reset()
		r.crtOpts.GeoM.Translate(float64(r.screen.Min.X), float64(r.screen.Min.Y))
		dst.DrawRectShader(r.size.X, r.size.Y, r.crt, &r.crtOpts)

		return
	}

	r.crtOpts.GeoM.Reset()
	r.out.DrawRectShader(r.size.X, r.size.Y, r.crt, &r.crtOpts)

	r.outOpts.GeoM.Reset()
	r.outOpts.GeoM.Scale(float64(r.screen.Dx())/float64(r.size.X), float64(r.screen.Dy())/float64(r.size.Y))
	r.outOpts.GeoM.Translate(float64(r.screen.Min.X), float64(r.screen.Min.Y))
	dst.DrawImage(r.out, &r.outOpts)
}

// step moves the afterglow on one frame: new glow = max(the picture, the
// old glow × decay).
func (r *Renderer) step() {
	r.due = false

	if r.fresh {
		r.upOpts.GeoM.Reset()
		r.upOpts.GeoM.Scale(float64(r.size.X)/pictureW, float64(r.size.Y)/pictureH)
		r.up.Clear()
		r.up.DrawImage(r.fb, &r.upOpts)
		r.fresh = false
		r.settle = settleFrames
	}

	r.settle--
	prev := r.phos[r.cur]
	r.cur ^= 1
	r.phosOpts.Images[0], r.phosOpts.Images[1] = r.up, prev
	r.phos[r.cur].DrawRectShader(r.size.X, r.size.Y, r.phosphor, &r.phosOpts)
}

// drawClean draws the coloured picture with nearest-neighbour scaling.
func (r *Renderer) drawClean(dst *ebiten.Image) {
	r.upOpts.GeoM.Reset()
	r.upOpts.GeoM.Scale(float64(r.screen.Dx())/pictureW, float64(r.screen.Dy())/pictureH)
	r.upOpts.GeoM.Translate(float64(r.screen.Min.X), float64(r.screen.Min.Y))
	r.upOpts.Filter = ebiten.FilterNearest
	r.upOpts.ColorScale.Reset()
	r.upOpts.ColorScale.Scale(r.dim, r.dim, r.dim, 1)
	dst.DrawImage(r.fb, &r.upOpts)
}

// scaleBackdrop draws the backdrop into the render-sized image the shader
// reads, as the shader's sources must all be the size it draws.
func (r *Renderer) scaleBackdrop() {
	if r.back == nil {
		return
	}

	r.back.Clear()

	if r.backdrop == nil {
		return
	}

	var m ebiten.DrawImageOptions

	b := r.backdrop.Bounds()
	m.GeoM.Scale(float64(r.size.X)/float64(b.Dx()), float64(r.size.Y)/float64(b.Dy()))
	m.Filter = ebiten.FilterLinear
	r.back.DrawImage(r.backdrop, &m)
}
