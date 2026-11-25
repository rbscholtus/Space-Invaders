package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	xdraw "golang.org/x/image/draw"
)

// This file holds the image operations, kept free of file I/O so they can
// be tested on small synthetic images.

const (
	// pressedShift is how far a pressed button's cap sinks, as a share of
	// its radius, and pressedShade how much darker it gets.
	pressedShift = 0.1
	pressedShade = 0.82
)

func lum(c color.NRGBA) int { return (int(c.R)*3 + int(c.G)*6 + int(c.B)) / 10 }

// toNRGBA copies any image into a fresh NRGBA starting at (0, 0).
func toNRGBA(src image.Image) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)

	return dst
}

// paintOut fills r with a horizontal blend of the columns just left and
// right of it: enough to remove a small mark from a smooth background.
func paintOut(m *image.NRGBA, r image.Rectangle) {
	r = r.Intersect(m.Bounds())
	if r.Empty() {
		return
	}

	left, right := max(r.Min.X-1, 0), min(r.Max.X, m.Bounds().Max.X-1)

	for y := r.Min.Y; y < r.Max.Y; y++ {
		a, b := m.NRGBAAt(left, y), m.NRGBAAt(right, y)

		for x := r.Min.X; x < r.Max.X; x++ {
			t := float64(x-left) / float64(right-left)
			m.SetNRGBA(x, y, mix(a, b, t))
		}
	}
}

func mix(a, b color.NRGBA, t float64) color.NRGBA {
	f := func(x, y uint8) uint8 { return uint8(math.Round(float64(x)*(1-t) + float64(y)*t)) }

	return color.NRGBA{R: f(a.R, b.R), G: f(a.G, b.G), B: f(a.B, b.B), A: f(a.A, b.A)}
}

// fadeEdges blends the outer feather pixels of m into fill (smoothly), so
// the cabinet's cut edges disappear into the page around it.
func fadeEdges(m *image.NRGBA, feather int, fill color.NRGBA) {
	b := m.Bounds()

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			d := min(x-b.Min.X, y-b.Min.Y, b.Max.X-1-x, b.Max.Y-1-y)
			if d >= feather {
				continue
			}

			t := float64(d) / float64(feather)
			m.SetNRGBA(x, y, mix(fill, m.NRGBAAt(x, y), t*t*(3-2*t))) // smoothstep
		}
	}
}

// borderColour is the average colour of m's outermost pixels.
func borderColour(m *image.NRGBA) color.NRGBA {
	b := m.Bounds()

	var r, g, bl, n int

	add := func(x, y int) {
		c := m.NRGBAAt(x, y)
		r, g, bl, n = r+int(c.R), g+int(c.G), bl+int(c.B), n+1
	}

	for x := b.Min.X; x < b.Max.X; x++ {
		add(x, b.Min.Y)
		add(x, b.Max.Y-1)
	}

	for y := b.Min.Y; y < b.Max.Y; y++ {
		add(b.Min.X, y)
		add(b.Max.X-1, y)
	}

	avg := func(sum int) uint8 { return uint8(sum / n) } //nolint:gosec // G115: an average of bytes

	return color.NRGBA{R: avg(r), G: avg(g), B: avg(bl), A: 0xFF}
}

// pressedSprite cuts a 2r×2r square centred on (cx, cy) out of m, with the
// button's cap pushed in: everything inside the circle of radius r moves
// down by pressedShift·r and darkens a little. The square is transparent
// outside the circle, with an anti-aliased edge, so it lies over the
// cabinet's own (unpressed) button.
func pressedSprite(m *image.NRGBA, cx, cy, r int) *image.NRGBA {
	s := image.NewNRGBA(image.Rect(0, 0, 2*r, 2*r))
	shift := int(math.Round(pressedShift * float64(r)))

	for y := range 2 * r {
		for x := range 2 * r {
			d := math.Hypot(float64(x)+0.5-float64(r), float64(y)+0.5-float64(r))
			a := min(max(float64(r)-d, 0), 1)

			c := m.NRGBAAt(cx-r+x, cy-r+y-shift)
			c.R = uint8(float64(c.R) * pressedShade)
			c.G = uint8(float64(c.G) * pressedShade)
			c.B = uint8(float64(c.B) * pressedShade)
			c.A = uint8(math.Round(float64(c.A) * a))
			s.SetNRGBA(x, y, c)
		}
	}

	return s
}

// scaleInto draws src scaled to fill r of dst, with Catmull-Rom filtering.
func scaleInto(dst *image.NRGBA, r image.Rectangle, src image.Image) {
	xdraw.CatmullRom.Scale(dst, r, src, src.Bounds(), xdraw.Src, nil)
}

// scaleTo resizes src to w×h with Catmull-Rom filtering.
func scaleTo(src image.Image, w, h int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, max(w, 1), max(h, 1)))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Src, nil)

	return dst
}
