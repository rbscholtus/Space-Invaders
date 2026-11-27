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

// The view and how much of the cabinet it shows, by window (DESIGN §11.1).
func TestComputeLayout(t *testing.T) {
	t.Parallel()

	cab := realCabinet(t)

	for _, tc := range []struct {
		name   string
		w, h   int // CSS pixels
		scale  float64
		choice View
		view   View
		upper  bool
	}{
		{"laptop: the upper part", 1440, 820, 2, ViewAuto, ViewCabinet, true},
		{"tablet upright: the whole cabinet", 820, 1180, 2, ViewAuto, ViewCabinet, false},
		{"phone: the whole cabinet fits as big", 390, 844, 3, ViewAuto, ViewCabinet, false},
		{"phone sideways: the picture alone", 844, 390, 3, ViewAuto, ViewScreen, false},
		{"phone sideways, cabinet chosen", 844, 390, 3, ViewCabinet, ViewCabinet, true},
		{"laptop, screen chosen", 1440, 820, 2, ViewScreen, ViewScreen, false},
	} {
		w, h := int(float64(tc.w)*tc.scale), int(float64(tc.h)*tc.scale)
		l := computeLayout(cab, w, h, tc.scale, tc.choice)

		if l.view != tc.view || l.upper != tc.upper {
			t.Errorf("%s: view %v, upper %v; want %v, %v", tc.name, l.view, l.upper, tc.view, tc.upper)

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

		if l.view == ViewCabinet && l.upper {
			// The control panel is in the window; the coin door is below it.
			fire, _ := cab.Control(cabinet.Fire)
			lock, _ := cab.Control(cabinet.Door)

			if r := l.rect(fire.Bounds()); r.Max.Y > h {
				t.Errorf("%s: the fire button at %v is below the window", tc.name, r)
			}

			if r := l.rect(lock.Bounds()); r.Min.Y < h {
				t.Errorf("%s: the lock at %v is in the window", tc.name, r)
			}
		}
	}

	if l := computeLayout(nil, 800, 600, 1, ViewCabinet); l.view != ViewScreen {
		t.Errorf("without a cabinet: view %v, want the screen", l.view)
	}
}

// The upper part makes the picture bigger in a landscape window than the
// whole cabinet would.
func TestUpperIsBigger(t *testing.T) {
	t.Parallel()

	cab := realCabinet(t)
	whole, upper := fitCabinet(cab, 1440, 820, false), fitCabinet(cab, 1440, 820, true)

	if ratio := float64(upper.tube.Dy()) / float64(whole.tube.Dy()); ratio < 1.3 {
		t.Errorf("the upper part's picture is %.2f× the whole cabinet's, want at least 1.3×", ratio)
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
