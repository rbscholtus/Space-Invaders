package app

import (
	"errors"
	"testing"

	"github.com/rbscholtus/Space-Invaders/assets"
	"github.com/rbscholtus/Space-Invaders/internal/app/cabinet"
)

func realCabinet(t *testing.T) *cabinet.Cabinet {
	t.Helper()

	c, err := cabinet.Parse(assets.CabinetJSON)
	if err != nil {
		t.Fatal(err)
	}

	return c
}

func TestParseView(t *testing.T) {
	t.Parallel()

	for _, v := range []View{ViewAuto, ViewCabinet, ViewScreen} {
		if got, err := ParseView(v.String()); err != nil || got != v {
			t.Errorf("ParseView(%q) = %v, %v", v, got, err)
		}
	}

	if _, err := ParseView("marquee"); !errors.Is(err, ErrView) {
		t.Errorf("ParseView(marquee): %v, want ErrView", err)
	}
}

// The view by window (DESIGN §11.1): the cabinet's frame, centred, unless
// the picture would be too small in it.
func TestComputeLayout(t *testing.T) {
	t.Parallel()

	cab := realCabinet(t)

	for _, tc := range []struct {
		name   string
		w, h   int // CSS pixels
		scale  float64
		choice View
		view   View
	}{
		{"laptop", 1440, 820, 2, ViewAuto, ViewCabinet},
		{"tablet upright", 820, 1180, 2, ViewAuto, ViewCabinet},
		{"phone", 390, 844, 3, ViewAuto, ViewCabinet},
		{"phone sideways: the picture alone", 844, 390, 3, ViewAuto, ViewScreen},
		{"phone sideways, cabinet chosen", 844, 390, 3, ViewCabinet, ViewCabinet},
		{"laptop, screen chosen", 1440, 820, 2, ViewScreen, ViewScreen},
	} {
		w, h := int(float64(tc.w)*tc.scale), int(float64(tc.h)*tc.scale)
		l := computeLayout(cab, w, h, tc.scale, tc.choice)

		if l.view != tc.view {
			t.Errorf("%s: view %v, want %v", tc.name, l.view, tc.view)

			continue
		}

		if l.tube.Min.X < 0 || l.tube.Min.Y < 0 || l.tube.Max.X > w || l.tube.Max.Y > h {
			t.Errorf("%s: the picture %v is outside the %dx%d window", tc.name, l.tube, w, h)
		}

		// The picture keeps the game's shape, give or take rounding.
		if got, want := float64(l.tube.Dx())/float64(l.tube.Dy()), float64(screenW)/screenH; got < want*0.99 ||
			got > want*1.01 {
			t.Errorf("%s: the picture's shape is %.3f, want %.3f", tc.name, got, want)
		}

		if l.view != ViewCabinet {
			continue
		}

		// The whole frame is in the window, and it fills its width or height.
		if f := l.rect(cab.Frame); f.Min.X < -1 || f.Min.Y < -1 || f.Max.X > w+1 || f.Max.Y > h+1 ||
			(f.Dx() < w-2 && f.Dy() < h-2) {
			t.Errorf("%s: the frame %v doesn't fit the %dx%d window", tc.name, f, w, h)
		}
	}

	if l := computeLayout(nil, 800, 600, 1, ViewCabinet); l.view != ViewScreen {
		t.Errorf("without a cabinet: view %v, want the screen", l.view)
	}
}

// As the window gets narrower, the frame's top stays at the window's top
// and more of the cabinet shows below it, until all of it shows; then the
// room left is shared equally above and below.
func TestFrameInNarrowWindows(t *testing.T) {
	t.Parallel()

	cab := realCabinet(t)
	frameTop := func(l layout) int { return l.rect(cab.Frame).Min.Y }

	if l := fitFrame(cab, 2000, 1000); frameTop(l) != 0 || l.rect(cab.Frame).Max.Y != 1000 {
		t.Errorf("wide window: the frame is at %v, want it from top to bottom", l.rect(cab.Frame))
	}

	// A little narrower than the frame: its top stays put, the cabinet's
	// bottom is below the window.
	if l := fitFrame(cab, 760, 1000); frameTop(l) != 0 || l.cabinet.Max.Y <= 1000 {
		t.Errorf("a little narrower: frame top %d, cabinet bottom %d", frameTop(l), l.cabinet.Max.Y)
	}

	// Much narrower: the whole cabinet below the frame's top shows,
	// centred.
	l := fitFrame(cab, 400, 1000)
	above, below := frameTop(l), 1000-l.cabinet.Max.Y

	if above <= 0 || below <= 0 || above-below > 1 || below-above > 1 {
		t.Errorf("much narrower: %d above the frame, %d below the cabinet; want equal", above, below)
	}
}

func TestSavedView(t *testing.T) {
	t.Parallel()

	p := memPrefs{}
	if v := savedView(p, ViewAuto, false); v != ViewAuto {
		t.Errorf("nothing saved: %v", v)
	}

	p[prefView] = "screen"
	if v := savedView(p, ViewAuto, false); v != ViewScreen {
		t.Errorf("saved screen: %v", v)
	}

	if v := savedView(p, ViewCabinet, true); v != ViewCabinet {
		t.Errorf("given cabinet wins over the saved view: %v", v)
	}

	p[prefView] = "nonsense"
	if v := savedView(p, ViewAuto, false); v != ViewAuto {
		t.Errorf("a bad saved view is ignored: %v", v)
	}
}
