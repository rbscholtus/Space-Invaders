package main

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/app/cabinet"
)

var (
	grey  = color.NRGBA{R: 200, G: 200, B: 200, A: 0xFF}
	black = color.NRGBA{R: 0, G: 0, B: 0, A: 0xFF}
	white = color.NRGBA{R: 250, G: 250, B: 250, A: 0xFF}
	blue  = color.NRGBA{R: 10, G: 20, B: 120, A: 0xFF}
)

func filled(w, h int, c color.NRGBA) *image.NRGBA {
	m := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(m, m.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)

	return m
}

func fillRect(m *image.NRGBA, r image.Rectangle, c color.NRGBA) {
	draw.Draw(m, r, image.NewUniform(c), image.Point{}, draw.Src)
}

func TestPaintOut(t *testing.T) {
	t.Parallel()

	m := filled(20, 10, grey)
	fillRect(m, image.Rect(8, 2, 12, 6), white) // the mark

	paintOut(m, image.Rect(7, 1, 13, 7))

	for y := range 10 {
		for x := range 20 {
			if c := m.NRGBAAt(x, y); c != grey {
				t.Fatalf("(%d,%d) = %v, want the background", x, y, c)
			}
		}
	}
}

func TestFadeEdges(t *testing.T) {
	t.Parallel()

	m := filled(40, 40, white)
	fadeEdges(m, 8, black)

	if c := m.NRGBAAt(0, 20); c != black {
		t.Errorf("edge = %v, want the fill", c)
	}

	if c := m.NRGBAAt(20, 20); c != white {
		t.Errorf("middle = %v, want untouched", c)
	}

	if a, b := lum(m.NRGBAAt(2, 20)), lum(m.NRGBAAt(5, 20)); a >= b || b >= lum(white) {
		t.Errorf("fade not rising inwards: %d, %d", a, b)
	}
}

func TestPressedSprite(t *testing.T) {
	t.Parallel()

	m := filled(100, 100, grey)
	fillRect(m, image.Rect(30, 30, 70, 70), white) // the cap, r = 20 around (50, 50)

	s := pressedSprite(m, 50, 50, 20)

	if b := s.Bounds(); b.Dx() != 40 || b.Dy() != 40 {
		t.Fatalf("size %v, want 40×40", b)
	}

	if c := s.NRGBAAt(0, 0); c.A != 0 {
		t.Errorf("corner alpha = %d, want transparent outside the circle", c.A)
	}

	// The cap sinks by 2 pixels: the sprite's top rows show what was above it.
	if c := s.NRGBAAt(20, 0); lum(c) >= lum(white)*9/10 {
		t.Errorf("top = %v, want the grey from above the cap", c)
	}

	if c := s.NRGBAAt(20, 20); c.A != 0xFF || c.R >= white.R || c.R < white.R*3/4 {
		t.Errorf("middle = %v, want the cap a little darker", c)
	}
}

func TestPutChecksShape(t *testing.T) {
	t.Parallel()

	master, art := filled(100, 100, grey), filled(40, 20, blue)

	in := insert{Source: "art", From: rectJSON{W: 40, H: 20}, To: rectJSON{X: 10, Y: 10, W: 20, H: 10}}
	if err := put(master, art, in); err != nil {
		t.Fatal(err)
	}

	if c := master.NRGBAAt(20, 15); c != blue {
		t.Errorf("inside = %v, want the art", c)
	}

	if c := master.NRGBAAt(5, 5); c != grey {
		t.Errorf("outside = %v, want untouched", c)
	}

	for name, in := range map[string]insert{
		"stretched":      {From: rectJSON{W: 40, H: 20}, To: rectJSON{W: 20, H: 20}},
		"from outside":   {From: rectJSON{W: 50, H: 25}, To: rectJSON{W: 20, H: 10}},
		"to outside":     {From: rectJSON{W: 40, H: 20}, To: rectJSON{X: 90, W: 20, H: 10}},
		"empty from box": {From: rectJSON{W: 0, H: 0}, To: rectJSON{W: 20, H: 10}},
	} {
		if err := put(master, art, in); !errors.Is(err, errRecipe) {
			t.Errorf("%s: %v, want errRecipe", name, err)
		}
	}
}

func TestBackdropOf(t *testing.T) {
	t.Parallel()

	// The art is twice the glass's size; the screen is the glass's middle.
	art := filled(200, 100, black)
	fillRect(art, image.Rect(50, 25, 150, 75), white)

	glass := insert{From: rectJSON{W: 200, H: 100}, To: rectJSON{X: 10, Y: 10, W: 100, H: 50}}
	b := backdropOf(art, glass, image.Rect(35, 22, 85, 47))

	if got := b.Bounds(); got.Dx() != 100 || got.Dy() != 50 {
		t.Fatalf("size %v, want the screen at the art's resolution, 100×50", got)
	}

	if c := b.NRGBAAt(50, 25); c != white {
		t.Errorf("middle = %v, want the art under the screen", c)
	}
}

func TestLetter(t *testing.T) {
	t.Parallel()

	m := filled(200, 100, blue)
	text := textRec{
		Text: "HI", Font: "bold", Size: 40, X: 100, Y: 50,
		Colour: "#ffffff", Outline: "#ff0000", OutlineWidth: 3,
	}

	if err := letter(m, text); err != nil {
		t.Fatal(err)
	}

	var ink, ring int

	for y := range 100 {
		for x := range 200 {
			switch c := m.NRGBAAt(x, y); {
			case c.R > 240 && c.G > 240:
				ink++

				if x < 60 || x > 140 {
					t.Fatalf("ink at (%d,%d): not centred", x, y)
				}
			case c.R > 240 && c.G < 20:
				ring++
			}
		}
	}

	if ink == 0 || ring == 0 {
		t.Errorf("ink %d, outline %d pixels: want both", ink, ring)
	}

	// MaxWidth shrinks it.
	narrow := filled(200, 100, blue)
	text.MaxWidth, text.Outline = 20, ""

	if err := letter(narrow, text); err != nil {
		t.Fatal(err)
	}

	for y := range 100 {
		for x := range 200 {
			if c := narrow.NRGBAAt(x, y); c.R > 128 && (x < 88 || x > 112) {
				t.Fatalf("ink at (%d,%d): wider than 20 pixels", x, y)
			}
		}
	}

	for name, bad := range map[string]textRec{
		"font":    {Text: "X", Font: "comic", Size: 10, Colour: "#ffffff"},
		"size":    {Text: "X", Font: "bold", Colour: "#ffffff"},
		"colour":  {Text: "X", Font: "bold", Size: 10, Colour: "white"},
		"outline": {Text: "X", Font: "bold", Size: 10, Colour: "#ffffff", Outline: "red"},
	} {
		if err := letter(m, bad); !errors.Is(err, errRecipe) {
			t.Errorf("bad %s: %v, want errRecipe", name, err)
		}
	}
}

func TestBuild(t *testing.T) {
	t.Parallel()

	cab := filled(400, 800, grey)
	fillRect(cab, image.Rect(40, 40, 360, 120), black)  // the marquee
	fillRect(cab, image.Rect(60, 160, 340, 360), black) // the glass

	var rc recipe

	rc.Name = "Test"
	rc.Cabinet.Source, rc.Cabinet.Crop, rc.Cabinet.Feather = "cab", rectJSON{X: 10, Y: 10, W: 380, H: 780}, 4
	rc.Glass = insert{Source: "glass", From: rectJSON{W: 280, H: 200}, To: rectJSON{X: 60, Y: 160, W: 280, H: 200}}
	rc.Marquee = insert{Source: "marquee", From: rectJSON{W: 320, H: 80}, To: rectJSON{X: 40, Y: 40, W: 320, H: 80}}
	rc.Screen = rectJSON{X: 150, Y: 170, W: 100, H: 180}
	rc.Controls = []controlRec{
		{ID: cabinet.Left, Kind: cabinet.Button, CX: 80, CY: 450, R: 15},
		{ID: cabinet.Right, Kind: cabinet.Button, CX: 130, CY: 450, R: 15},
		{ID: cabinet.Fire, Kind: cabinet.Button, CX: 200, CY: 450, R: 15},
		{ID: cabinet.Start1, Kind: cabinet.Button, CX: 300, CY: 430, R: 15},
		{ID: cabinet.Start2, Kind: cabinet.Button, CX: 300, CY: 470, R: 15},
		{ID: cabinet.Coin1, Kind: cabinet.Slot, X: 150, Y: 600, W: 20, H: 30},
		{ID: cabinet.Door, Kind: cabinet.Lock, CX: 250, CY: 650, R: 10},
	}

	masters := map[string]*image.NRGBA{
		"cab": cab, "glass": filled(280, 200, blue), "marquee": filled(320, 80, white),
	}

	out, err := build(&rc, masters)
	if err != nil {
		t.Fatal(err)
	}

	c := out.cab
	if c.Width != 380 || c.Height != 780 || c.Screen != (cabinet.Rect{X: 140, Y: 160, W: 100, H: 180}) {
		t.Errorf("cabinet %gx%g, screen %+v: want everything relative to the crop", c.Width, c.Height, c.Screen)
	}

	if ctl, _ := c.Control(cabinet.Fire); ctl.CX != 190 || ctl.CY != 440 || ctl.Pressed != "fire-pressed" {
		t.Errorf("fire = %+v", ctl)
	}

	if len(out.pressed) != 5 {
		t.Errorf("%d pressed sprites, want one per button", len(out.pressed))
	}

	if got := out.cabinet.NRGBAAt(100, 200); got != blue {
		t.Errorf("glass = %v, want the glass art put in", got)
	}

	if b := out.backdrop.Bounds(); b.Dx() != 100 || b.Dy() != 180 {
		t.Errorf("backdrop %v, want the screen's size", b)
	}

	if c.Background != "#c8c8c8" {
		t.Errorf("background %s, want the cabinet's grey", c.Background)
	}

	rc.Screen.X = 20 // outside the glass

	if _, err := build(&rc, masters); !errors.Is(err, errRecipe) {
		t.Errorf("screen outside the glass: %v", err)
	}
}
