package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/rbscholtus/Space-Invaders/internal/app/cabinet"
)

// Preview colours and the darkening of the picture's window.
var (
	outlineColour = color.NRGBA{R: 0xFF, G: 0x20, B: 0x20, A: 0xFF} //nolint:gochecknoglobals // read-only colour
	marqueeColour = color.NRGBA{R: 0x20, G: 0xD0, B: 0xFF, A: 0xFF} //nolint:gochecknoglobals // read-only colour
)

const windowShade = 0.55 // about how the backdrop shows behind the picture

// writePreview composes the @3x outputs as the app would draw them, for
// checking alignment by eye: the backdrop darkened in the picture's
// window, every button pressed, and the screen, the marquee and each
// control outlined.
func writePreview(path string, out *outputs) error {
	m := toNRGBA(out.cabinet)
	c := &out.cab

	scr := rectOf(c.Screen)
	scaleInto(m, scr, out.backdrop)

	for y := scr.Min.Y; y < scr.Max.Y; y++ {
		for x := scr.Min.X; x < scr.Max.X; x++ {
			p := m.NRGBAAt(x, y)
			p.R, p.G, p.B = uint8(float64(p.R)*windowShade), uint8(float64(p.G)*windowShade), uint8(float64(p.B)*windowShade)
			m.SetNRGBA(x, y, p)
		}
	}

	for _, ctl := range c.Controls {
		b := rectOf(ctl.Bounds())

		if s := out.pressed[ctl.Pressed]; s != nil {
			draw.Draw(m, b, s, image.Point{}, draw.Over)
		}

		outline(m, b, outlineColour)
	}

	outline(m, scr, outlineColour)
	outline(m, rectOf(c.Marquee), marqueeColour)

	return writePNG(path, m)
}

func rectOf(r cabinet.Rect) image.Rectangle {
	return image.Rect(int(r.X), int(r.Y), int(r.X+r.W), int(r.Y+r.H))
}

func outline(m *image.NRGBA, r image.Rectangle, c color.NRGBA) {
	for x := r.Min.X; x < r.Max.X; x++ {
		m.SetNRGBA(x, r.Min.Y, c)
		m.SetNRGBA(x, r.Max.Y-1, c)
	}

	for y := r.Min.Y; y < r.Max.Y; y++ {
		m.SetNRGBA(r.Min.X, y, c)
		m.SetNRGBA(r.Max.X-1, y, c)
	}
}
