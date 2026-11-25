package crt_test

import (
	"errors"
	"image"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/app/crt"
	"github.com/rbscholtus/Space-Invaders/internal/app/overlay"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

// TestShadersCompile compiles both Kage shaders. Compiling happens on the
// CPU, so no window or GPU is needed.
func TestShadersCompile(t *testing.T) {
	t.Parallel()

	r, err := crt.New(crt.ModeCRT, overlay.Midway())
	if err != nil {
		t.Fatal(err)
	}

	if r.Mode() != crt.ModeCRT {
		t.Errorf("mode %v, want crt", r.Mode())
	}
}

func TestParseMode(t *testing.T) {
	t.Parallel()

	for _, m := range []crt.Mode{crt.ModeCRT, crt.ModeClean} {
		if got, err := crt.ParseMode(m.String()); err != nil || got != m {
			t.Errorf("ParseMode(%q) = %v, %v", m.String(), got, err)
		}
	}

	if _, err := crt.ParseMode("vector"); !errors.Is(err, crt.ErrMode) {
		t.Errorf("ParseMode(vector): %v, want ErrMode", err)
	}
}

// TestRenderSize: the shaders work at the tube's size up to 1280 pixels
// high, and at a smaller size of the same shape above that.
func TestRenderSize(t *testing.T) {
	t.Parallel()

	for tube, want := range map[image.Point]image.Point{
		{448, 512}:   {448, 512},
		{1120, 1280}: {1120, 1280},
		{1792, 2048}: {1120, 1280},
		{2, 3000}:    {1, 1280},
	} {
		if got := crt.RenderSize(tube); got != want {
			t.Errorf("RenderSize(%v) = %v, want %v", tube, got, want)
		}
	}
}

// TestScanStrength: no lines below 2 pixels per line (they'd make a
// moiré), all of them from 4.
func TestScanStrength(t *testing.T) {
	t.Parallel()

	for px, want := range map[float64]float64{1: 0, 2: 0, 3: 0.5, 4: 1, 9: 1} {
		if got := crt.ScanStrength(px); got != want {
			t.Errorf("ScanStrength(%v) = %v, want %v", px, got, want)
		}
	}
}

// TestTint: the tint texture has the overlay's colours inside its strips
// and white outside.
func TestTint(t *testing.T) {
	t.Parallel()

	strips := overlay.Midway()
	tint := crt.Tint(strips)

	at := func(x, y int) [3]byte {
		i := (y*invaders.ScreenWidth + x) * 4

		return [3]byte{tint[i], tint[i+1], tint[i+2]}
	}

	for _, s := range strips {
		if got, want := at(s.X0, s.Y0), [3]byte{s.R, s.G, s.B}; got != want {
			t.Errorf("tint at (%d,%d) = %v, want %v", s.X0, s.Y0, got, want)
		}
	}

	if got := at(0, 0); got != [3]byte{255, 255, 255} {
		t.Errorf("tint outside the strips = %v, want white", got)
	}
}
