package app

import (
	"errors"
	"fmt"
	"image"
	"math"

	"github.com/rbscholtus/Space-Invaders/internal/app/cabinet"
)

// View is what the window shows (DESIGN §11.1): the cabinet with the
// picture in it, or only the picture.
type View uint8

// The views. Auto picks one by the window's size.
const (
	ViewAuto View = iota
	ViewCabinet
	ViewScreen
)

// ErrView is returned for an unknown view name.
var ErrView = errors.New("app: the view must be auto, cabinet or screen")

// ParseView reads a view's name, as the -view flag and ?view= give it.
func ParseView(s string) (View, error) {
	switch s {
	case "auto":
		return ViewAuto, nil
	case "cabinet":
		return ViewCabinet, nil
	case "screen":
		return ViewScreen, nil
	}

	return ViewAuto, fmt.Errorf("%w, not %q", ErrView, s)
}

// String names the view, as ParseView reads it.
func (v View) String() string {
	switch v {
	case ViewCabinet:
		return "cabinet"
	case ViewScreen:
		return "screen"
	case ViewAuto:
	}

	return "auto"
}

// minCabinetPicture is the smallest picture, in CSS pixels, for which auto
// shows the cabinet: below it, as on a phone held sideways, the picture
// alone (DESIGN §11.1; ⚠ tune on devices). The game's own height is 256.
const minCabinetPicture = 160

// layout is where things go in the window, in real pixels.
type layout struct {
	view View // ViewCabinet or ViewScreen: what's shown
	// cabinet is the whole cabinet image, which may reach past the window.
	cabinet image.Rectangle
	k       float64         // real pixels per cabinet pixel
	tube    image.Rectangle // the picture
}

// computeLayout lays out a w×h window (real pixels, scale of them per CSS
// pixel) for the chosen view. Without a cabinet it's the screen view.
func computeLayout(cab *cabinet.Cabinet, w, h int, scale float64, choice View) layout {
	screen := layout{view: ViewScreen, tube: fitTube(w, h)}
	if cab == nil || choice == ViewScreen || w <= 0 || h <= 0 {
		return screen
	}

	l := fitFrame(cab, w, h)
	if choice == ViewAuto && float64(l.tube.Dy())/scale < minCabinetPicture {
		return screen
	}

	return l
}

// fitFrame shows the cabinet's frame (from its top to the control panel)
// as big as fits, centred across, with its top at the window's top. In a
// window narrower than the frame, more of the cabinet shows below it; once
// all of it shows, the cabinet from the frame's top down is centred, with
// as much room above as below.
func fitFrame(cab *cabinet.Cabinet, w, h int) layout {
	f := cab.Frame
	k := max(min(float64(w)/f.W, float64(h)/f.H), 0)
	x := (float64(w)-f.W*k)/2 - f.X*k
	y := max(0, (float64(h)-(cab.Height-f.Y)*k)/2) - f.Y*k

	l := layout{view: ViewCabinet, k: k}
	l.cabinet = roundRect(x, y, x+cab.Width*k, y+cab.Height*k)
	l.tube = l.rect(cab.Screen)

	return l
}

// rect converts a rectangle in cabinet pixels to window pixels.
func (l layout) rect(r cabinet.Rect) image.Rectangle {
	x, y := float64(l.cabinet.Min.X), float64(l.cabinet.Min.Y)

	return roundRect(x+r.X*l.k, y+r.Y*l.k, x+(r.X+r.W)*l.k, y+(r.Y+r.H)*l.k)
}

func roundRect(x0, y0, x1, y1 float64) image.Rectangle {
	return image.Rect(int(math.Round(x0)), int(math.Round(y0)), int(math.Round(x1)), int(math.Round(y1)))
}

// fitTube returns the biggest rectangle of the picture's shape that fits a
// w×h window, centred: the screen view.
func fitTube(w, h int) image.Rectangle {
	tw, th := w, w*screenH/screenW
	if th > h {
		tw, th = h*screenW/screenH, h
	}

	x, y := (w-tw)/2, (h-th)/2

	return image.Rect(x, y, x+tw, y+th)
}
