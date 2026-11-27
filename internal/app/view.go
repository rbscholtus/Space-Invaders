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

// The picture's height in CSS pixels decides the view (DESIGN §11.1;
// ⚠ tune on devices). 256 is the game's own height.
const (
	// minCabinetPicture: auto shows the screen view when the cabinet would
	// make the picture smaller than this, as on a phone held sideways.
	minCabinetPicture = 160
	// minWholePicture: a portrait window (a tablet held upright) shows the
	// whole cabinet when the picture is at least this tall in it;
	// otherwise, and in every landscape window, the cabinet's upper part.
	minWholePicture = 256
	// cabinetMargin is the room around the cabinet, as a share of the
	// window's smaller side.
	cabinetMargin = 0.02
)

// layout is where things go in the window, in real pixels.
type layout struct {
	view  View // ViewCabinet or ViewScreen: what's shown
	upper bool // the cabinet view shows the upper part, running on below the window
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

	css := func(l layout) float64 { return float64(l.tube.Dy()) / scale }

	l := fitCabinet(cab, w, h, true)

	if h > w {
		if whole := fitCabinet(cab, w, h, false); css(whole) >= minWholePicture {
			l = whole
		}
	}

	if choice == ViewAuto && css(l) < minCabinetPicture {
		return screen
	}

	return l
}

// fitCabinet places the whole cabinet, centred, as big as fits; or its
// upper part, as big as fits from the top of the window, the rest of the
// cabinet running on below the window's bottom edge. In a window too narrow
// for the upper part to be bigger, that's the whole cabinet, centred.
func fitCabinet(cab *cabinet.Cabinet, w, h int, upper bool) layout {
	ww, wh := float64(w), float64(h)
	m := min(ww, wh) * cabinetMargin

	if !upper {
		k := max(min((ww-2*m)/cab.Width, (wh-2*m)/cab.Height), 0)
		x, y := (ww-cab.Width*k)/2, (wh-cab.Height*k)/2

		return placed(cab, x, y, k, false)
	}

	part := cab.Upper
	k := max(min((ww-2*m)/part.W, (wh-m)/part.H), 0)
	x := (ww-part.W*k)/2 - part.X*k

	if cab.Height*k <= wh-2*m {
		// A narrow window: the whole cabinet fits at this size anyway.
		return placed(cab, x, (wh-cab.Height*k)/2, k, false)
	}

	return placed(cab, x, m-part.Y*k, k, true)
}

// placed is the cabinet view with the cabinet's top left at (x, y), k real
// pixels per cabinet pixel.
func placed(cab *cabinet.Cabinet, x, y, k float64, upper bool) layout {
	l := layout{view: ViewCabinet, upper: upper, k: k}
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
