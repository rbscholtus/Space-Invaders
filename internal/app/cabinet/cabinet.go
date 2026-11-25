// Package cabinet describes the cabinet artwork: where the picture, the
// marquee and the controls are, in the cabinet image's own pixels (DESIGN
// §12). It parses and checks cabinet.json and does the geometry (which
// control is under a point, which pre-scaled size to load). It holds no
// Ebitengine state, so cmd/cabinet-gen writes cabinet.json with the same
// types and the maths can be tested directly.
//
// Adapted from Chippie's skin package (github.com/rbscholtus/chippie).
package cabinet

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// Kind is what a control does when it's pressed.
type Kind string

// Control kinds.
const (
	Button Kind = "button" // a push button, with a pressed sprite
	Slot   Kind = "slot"   // a coin slot: a press inserts a coin
	Lock   Kind = "lock"   // the coin door's lock: opens the settings
)

// The controls' ids. Every cabinet has the buttons and the lock, and at
// least one coin slot.
const (
	Left   = "left"
	Right  = "right"
	Fire   = "fire"
	Start1 = "start1"
	Start2 = "start2"
	Coin1  = "coin1"
	Coin2  = "coin2"
	Door   = "lock"
)

// MaxScale is the largest pre-scaled size: @3x is the cabinet at full size.
const MaxScale = 3

// ErrInvalid reports a cabinet.json that doesn't describe a usable cabinet.
var ErrInvalid = errors.New("cabinet: invalid cabinet.json")

// Rect is a rectangle in cabinet pixels (at @3x).
type Rect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

// Contains reports whether (x, y) is inside r.
func (r Rect) Contains(x, y float64) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// Inside reports whether r lies within o.
func (r Rect) Inside(o Rect) bool {
	return r.X >= o.X && r.Y >= o.Y && r.X+r.W <= o.X+o.W && r.Y+r.H <= o.Y+o.H
}

// Overlaps reports whether r and o share any area.
func (r Rect) Overlaps(o Rect) bool {
	return r.X < o.X+o.W && o.X < r.X+r.W && r.Y < o.Y+o.H && o.Y < r.Y+r.H
}

// Control is a button, a coin slot or the lock. Buttons and the lock are
// round (centre and radius); slots are rectangles.
type Control struct {
	ID   string `json:"id"`
	Kind Kind   `json:"kind"`
	// Pressed is a button's pressed sprite: <pressed>@<n>x.png, the size
	// of Bounds.
	Pressed string  `json:"pressed,omitempty"`
	CX      float64 `json:"cx,omitempty"`
	CY      float64 `json:"cy,omitempty"`
	R       float64 `json:"r,omitempty"`
	X       float64 `json:"x,omitempty"`
	Y       float64 `json:"y,omitempty"`
	W       float64 `json:"w,omitempty"`
	H       float64 `json:"h,omitempty"`
}

// Round reports whether the control is a circle.
func (c Control) Round() bool { return c.Kind != Slot }

// Bounds is the control's rectangle (a round control's square).
func (c Control) Bounds() Rect {
	if c.Round() {
		return Rect{X: c.CX - c.R, Y: c.CY - c.R, W: 2 * c.R, H: 2 * c.R}
	}

	return Rect{X: c.X, Y: c.Y, W: c.W, H: c.H}
}

// Centre is the middle of the control.
func (c Control) Centre() (float64, float64) {
	b := c.Bounds()

	return b.X + b.W/2, b.Y + b.H/2
}

// Cabinet is a parsed cabinet.json.
type Cabinet struct {
	Name   string  `json:"name"`
	Width  float64 `json:"width"`  // cabinet image size at @3x
	Height float64 `json:"height"` //
	// Background fills the page around the cabinet: "#rrggbb", the colour
	// the cabinet image's edges fade into.
	Background string `json:"background"`
	// Image is the cabinet's base name: <image>@<n>x.jpg.
	Image string `json:"image"`
	// Backdrop is the moon painting behind the picture, one size: the
	// part of the glass art under Screen, at the art's own resolution
	// (DESIGN §11.3).
	Backdrop string `json:"backdrop"`
	// Screen is where the picture is drawn: the window in the glass.
	Screen Rect `json:"screen"`
	// Marquee is the backlit title panel.
	Marquee  Rect      `json:"marquee"`
	Controls []Control `json:"controls"`
}

// Parse reads and checks cabinet.json.
func Parse(data []byte) (*Cabinet, error) {
	var c Cabinet
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &c, nil
}

// Validate checks that everything lies within the cabinet, that controls
// don't overlap each other or the screen, and that every control the game
// needs is there.
func (c *Cabinet) Validate() error {
	whole := Rect{W: c.Width, H: c.Height}

	switch {
	case c.Width <= 0 || c.Height <= 0:
		return fmt.Errorf("%w: cabinet size %gx%g", ErrInvalid, c.Width, c.Height)
	case c.Name == "":
		return fmt.Errorf("%w: no name", ErrInvalid)
	case c.Image == "" || c.Backdrop == "":
		return fmt.Errorf("%w: cabinet or backdrop image missing", ErrInvalid)
	case !validColour(c.Background):
		return fmt.Errorf("%w: background %q is not #rrggbb", ErrInvalid, c.Background)
	case c.Screen.W <= 0 || c.Screen.H <= 0 || !c.Screen.Inside(whole):
		return fmt.Errorf("%w: screen %+v not inside the cabinet", ErrInvalid, c.Screen)
	case c.Marquee.W <= 0 || c.Marquee.H <= 0 || !c.Marquee.Inside(whole):
		return fmt.Errorf("%w: marquee %+v not inside the cabinet", ErrInvalid, c.Marquee)
	}

	if err := c.validateControls(whole); err != nil {
		return err
	}

	for _, id := range []string{Left, Right, Fire, Start1, Start2, Door} {
		if _, ok := c.Control(id); !ok {
			return fmt.Errorf("%w: no %q control", ErrInvalid, id)
		}
	}

	if _, ok := c.Control(Coin1); !ok {
		return fmt.Errorf("%w: no %q coin slot", ErrInvalid, Coin1)
	}

	return nil
}

func (c Control) validate() error {
	switch {
	case c.ID == "":
		return fmt.Errorf("%w: control without an id", ErrInvalid)
	case c.Kind != Button && c.Kind != Slot && c.Kind != Lock:
		return fmt.Errorf("%w: control %q has unknown kind %q", ErrInvalid, c.ID, c.Kind)
	case c.Round() && c.R <= 0:
		return fmt.Errorf("%w: control %q has no radius", ErrInvalid, c.ID)
	case !c.Round() && (c.W <= 0 || c.H <= 0):
		return fmt.Errorf("%w: control %q has no size", ErrInvalid, c.ID)
	case c.Kind == Button && c.Pressed == "":
		return fmt.Errorf("%w: button %q has no pressed sprite", ErrInvalid, c.ID)
	}

	return nil
}

// Control returns the control with the id.
func (c *Cabinet) Control(id string) (Control, bool) {
	for _, ctl := range c.Controls {
		if ctl.ID == id {
			return ctl, true
		}
	}

	return Control{}, false
}

// validColour reports whether s is "#rrggbb".
func validColour(s string) bool {
	if len(s) != len("#rrggbb") || s[0] != '#' {
		return false
	}

	_, err := strconv.ParseUint(s[1:], 16, 32)

	return err == nil
}

// Colour returns the background as red, green and blue bytes; Validate has
// checked it.
func (c *Cabinet) Colour() [3]uint8 {
	v, _ := strconv.ParseUint(c.Background[1:], 16, 32) // checked by Validate

	return [3]uint8{uint8(v >> 16), uint8(v >> 8), uint8(v)} //nolint:gosec // G115: six hex digits fit 24 bits
}

// ScaleFor picks the pre-scaled size (1, 2 or 3) for drawing the cabinet
// width pixels wide, preferring to scale down: the smallest one at least
// that wide, else the largest.
func (c *Cabinet) ScaleFor(width float64) int {
	for n := 1; n < MaxScale; n++ {
		if c.Width*float64(n)/MaxScale >= width {
			return n
		}
	}

	return MaxScale
}

// FileName is a pre-scaled file's name: "cabinet", 2, "jpg" → cabinet@2x.jpg.
func FileName(base string, scale int, ext string) string {
	return fmt.Sprintf("%s@%dx.%s", base, scale, ext)
}

// HitTest returns the control under (x, y), in cabinet pixels. Every
// control's hit area reaches at least minR from its centre (round) or
// edges (slots), so small controls stay easy to touch; where grown areas
// overlap, the nearest centre wins.
func (c *Cabinet) HitTest(x, y, minR float64) (Control, bool) {
	best, bestD := -1, math.Inf(1)

	for i, ctl := range c.Controls {
		cx, cy := ctl.Centre()
		d := math.Hypot(x-cx, y-cy)

		var hit bool

		if ctl.Round() {
			hit = d <= max(ctl.R, minR)
		} else {
			b := ctl.Bounds()
			gx, gy := max(0, minR-b.W/2), max(0, minR-b.H/2)
			hit = (Rect{X: b.X - gx, Y: b.Y - gy, W: b.W + 2*gx, H: b.H + 2*gy}).Contains(x, y)
		}

		if hit && d < bestD {
			best, bestD = i, d
		}
	}

	if best < 0 {
		return Control{}, false
	}

	return c.Controls[best], true
}

func (c *Cabinet) validateControls(whole Rect) error {
	seen := map[string]bool{}

	for i, ctl := range c.Controls {
		if err := ctl.validate(); err != nil {
			return err
		}

		b := ctl.Bounds()

		switch {
		case seen[ctl.ID]:
			return fmt.Errorf("%w: control %q twice", ErrInvalid, ctl.ID)
		case !b.Inside(whole):
			return fmt.Errorf("%w: control %q outside the cabinet", ErrInvalid, ctl.ID)
		case b.Overlaps(c.Screen):
			return fmt.Errorf("%w: control %q overlaps the screen", ErrInvalid, ctl.ID)
		}

		seen[ctl.ID] = true

		for _, o := range c.Controls[:i] {
			if b.Overlaps(o.Bounds()) {
				return fmt.Errorf("%w: controls %q and %q overlap", ErrInvalid, o.ID, ctl.ID)
			}
		}
	}

	return nil
}
