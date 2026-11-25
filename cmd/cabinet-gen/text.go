package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Our lettering: generated images garble text, and the words must be ours
// (DESIGN §12). Go's fonts are BSD-licensed and used only here, in the
// generator; the app draws nothing with them.

const lineSpacing = 1.2 // line height, as a share of the font size, by default

// textRec is one piece of lettering, centred on (x, y) in the cabinet
// master. Lines are separated by "\n".
type textRec struct {
	Text string `json:"text"`
	// Font is "bold" or "medium".
	Font string  `json:"font"`
	Size float64 `json:"size"` // pixels
	X    int     `json:"x"`
	Y    int     `json:"y"`
	// MaxWidth shrinks the text to fit, when set.
	MaxWidth int    `json:"maxWidth"`
	Colour   string `json:"colour"`
	// Outline, when set, rings the letters OutlineWidth pixels wide.
	Outline      string `json:"outline"`
	OutlineWidth int    `json:"outlineWidth"`
	// Tracking is extra space between letters, in pixels.
	Tracking float64 `json:"tracking"`
	// LineHeight is the distance between lines as a share of Size.
	LineHeight float64 `json:"lineHeight"`
}

// letter draws one piece of lettering onto m.
func letter(m *image.NRGBA, t textRec) error {
	ink, err := parseColour(t.Colour)
	if err != nil {
		return err
	}

	var ring color.NRGBA

	if t.Outline != "" {
		if ring, err = parseColour(t.Outline); err != nil {
			return err
		}
	}

	ttf := map[string][]byte{"bold": gobold.TTF, "medium": gomedium.TTF}[t.Font]
	if ttf == nil || t.Size <= 0 {
		return fmt.Errorf("%w: text %q: font %q, size %g", errRecipe, t.Text, t.Font, t.Size)
	}

	face, err := newFace(ttf, t.Size)
	if err != nil {
		return err
	}

	lines := strings.Split(t.Text, "\n")

	if w := widest(face, lines, t.Tracking); t.MaxWidth > 0 && w > float64(t.MaxWidth) {
		size := t.Size * float64(t.MaxWidth) / w
		if face, err = newFace(ttf, size); err != nil {
			return err
		}

		t.Tracking *= size / t.Size
		t.Size = size
	}

	alpha := textMask(face, lines, t)
	b := alpha.Bounds()
	at := b.Add(image.Pt(t.X-b.Dx()/2, t.Y-b.Dy()/2))

	if t.Outline != "" && t.OutlineWidth > 0 {
		draw.DrawMask(m, at, image.NewUniform(ring), image.Point{}, dilate(alpha, t.OutlineWidth), image.Point{}, draw.Over)
	}

	draw.DrawMask(m, at, image.NewUniform(ink), image.Point{}, alpha, image.Point{}, draw.Over)

	return nil
}

// textMask draws the lines, centred under each other, as coverage, with
// room around them for an outline.
func textMask(face font.Face, lines []string, t textRec) *image.Alpha {
	met := face.Metrics()
	pad := t.OutlineWidth + 1
	step := t.Size * lineSpacing

	if t.LineHeight > 0 {
		step = t.Size * t.LineHeight
	}

	w := int(math.Ceil(widest(face, lines, t.Tracking))) + 2*pad
	h := int(math.Ceil(step*float64(len(lines)-1)+float64((met.Ascent+met.Descent).Ceil()))) + 2*pad
	m := image.NewAlpha(image.Rect(0, 0, w, h))
	d := font.Drawer{Dst: m, Src: image.Opaque, Face: face}

	for i, line := range lines {
		x := float64(pad) + (float64(w-2*pad)-measure(face, line, t.Tracking))/2
		d.Dot = fixed.Point26_6{
			X: fixed.Int26_6(x * 64),
			Y: met.Ascent + fixed.I(pad) + fixed.Int26_6(step*float64(i)*64),
		}

		for _, r := range line {
			d.DrawString(string(r))
			d.Dot.X += fixed.Int26_6(t.Tracking * 64)
		}
	}

	return m
}

// measure is a line's width with tracking between its letters.
func measure(face font.Face, s string, tracking float64) float64 {
	n := len([]rune(s))
	w := float64(font.MeasureString(face, s)) / 64

	return w + tracking*float64(max(n-1, 0))
}

func widest(face font.Face, lines []string, tracking float64) float64 {
	w := 0.0
	for _, l := range lines {
		w = max(w, measure(face, l, tracking))
	}

	return w
}

// dilate grows coverage by r pixels in every direction (a disc), for an
// outline drawn under the letters.
func dilate(m *image.Alpha, r int) *image.Alpha {
	b := m.Bounds()
	out := image.NewAlpha(b)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			var v uint8

			for dy := -r; dy <= r; dy++ {
				for dx := -r; dx <= r; dx++ {
					if dx*dx+dy*dy > r*r || !image.Pt(x+dx, y+dy).In(b) {
						continue
					}

					v = max(v, m.AlphaAt(x+dx, y+dy).A)
				}
			}

			out.SetAlpha(x, y, color.Alpha{A: v})
		}
	}

	return out
}

// newFace loads a TrueType font at size pixels.
func newFace(ttf []byte, size float64) (*opentype.Face, error) {
	f, err := opentype.Parse(ttf)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		return nil, fmt.Errorf("font face: %w", err)
	}

	of, ok := face.(*opentype.Face)
	if !ok {
		return nil, fmt.Errorf("font face: %w: unexpected type %T", errRecipe, face)
	}

	return of, nil
}
