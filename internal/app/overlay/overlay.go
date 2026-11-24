// Package overlay colours the monochrome picture as the coloured strips on
// the real monitor do (DESIGN §11.2). It is plain Go, so tests need no
// Ebitengine; M5 moves the colouring into a shader.
package overlay

import "github.com/rbscholtus/Space-Invaders/internal/invaders"

// Rect is a tinted rectangle in screen pixels: columns [X0, X1) and rows
// [Y0, Y1), multiplying lit pixels by R, G, B (0–255).
type Rect struct {
	X0, Y0, X1, Y1 int
	R, G, B        byte
}

// tint255 is 0.125 of full brightness: the strips let a little of the other
// colours through.
const tint255 = 32

// Midway is the overlay of the Midway upright cabinet, from MAME's
// layout/invaders.lay (CC0): a red strip for the UFO, a green one for the
// shields and the player, and green for the lives at the bottom left. Rows
// 256–259 in MAME are the 4-pixel border below our picture (DESIGN §10).
func Midway() []Rect {
	const w, h = invaders.ScreenWidth, invaders.ScreenHeight

	return []Rect{
		{0, 32, w, 62, 255, tint255, tint255},    // red
		{0, 184, w, 240, tint255, 255, tint255},  // green
		{16, 240, 134, h, tint255, 255, tint255}, // green, lives
	}
}

// Painter turns the 1-byte-per-pixel picture from invaders.Screen into
// RGBA pixels, with the overlay's colours. Its tables are built once.
type Painter struct {
	tint []byte // RGBA per pixel: the colour of a lit pixel
}

// NewPainter returns a Painter for the overlay rects; where they overlap,
// the last one wins, and pixels outside them are white.
func NewPainter(rects []Rect) *Painter {
	const w, h = invaders.ScreenWidth, invaders.ScreenHeight

	p := &Painter{tint: make([]byte, w*h*4)}
	for i := range w * h {
		p.tint[i*4], p.tint[i*4+1], p.tint[i*4+2], p.tint[i*4+3] = 255, 255, 255, 255
	}

	for _, r := range rects {
		for y := max(r.Y0, 0); y < min(r.Y1, h); y++ {
			for x := max(r.X0, 0); x < min(r.X1, w); x++ {
				i := (y*w + x) * 4
				p.tint[i], p.tint[i+1], p.tint[i+2] = r.R, r.G, r.B
			}
		}
	}

	return p
}

// Paint writes screen (one byte per pixel, 1 = lit) into dst as RGBA, 4
// bytes per pixel: lit pixels take the overlay's colour, dark ones are
// black.
func (p *Painter) Paint(screen, dst []byte) {
	for i, lit := range screen {
		if lit != 0 {
			copy(dst[i*4:i*4+4], p.tint[i*4:i*4+4])
		} else {
			dst[i*4], dst[i*4+1], dst[i*4+2], dst[i*4+3] = 0, 0, 0, 255
		}
	}
}
