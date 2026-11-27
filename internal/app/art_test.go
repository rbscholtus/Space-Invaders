package app

import (
	"errors"
	"image"
	"image/color"
	"testing"

	"github.com/rbscholtus/Space-Invaders/assets"
)

var errGone = errors.New("gone")

func TestLoadArt(t *testing.T) {
	t.Parallel()

	cab := realCabinet(t)
	res := <-loadArt(assets.ReadArt, cab, 1, func() {})

	if res.err != nil {
		t.Fatal(res.err)
	}

	if b := res.cabinet.Bounds(); b.Dx() != int(cab.Width)/3 && b.Dx() != int(cab.Width+2)/3 {
		t.Errorf("cabinet@1x is %v wide, want a third of %g", b, cab.Width)
	}

	if res.backdrop == nil || len(res.pressed) != 5 {
		t.Errorf("backdrop %v, %d pressed buttons; want it and 5", res.backdrop != nil, len(res.pressed))
	}
}

func TestLoadArtFails(t *testing.T) {
	t.Parallel()

	res := <-loadArt(func(string) ([]byte, error) { return nil, errGone }, realCabinet(t), 2, func() {})

	if !errors.Is(res.err, errGone) {
		t.Errorf("err = %v, want the fetcher's", res.err)
	}
}

// toRGBA converts in bands, pausing between them, and keeps every pixel.
func TestToRGBA(t *testing.T) {
	t.Parallel()

	m := image.NewNRGBA(image.Rect(10, 10, 15, 22))
	m.SetNRGBA(12, 21, color.NRGBA{R: 200, G: 100, B: 50, A: 0xFF})

	pauses := 0
	out := toRGBA(m, 5, func() { pauses++ })

	if out.Bounds() != image.Rect(0, 0, 5, 12) || pauses != 3 {
		t.Errorf("bounds %v, %d pauses; want 5×12 at 0,0 and 3", out.Bounds(), pauses)
	}

	if c := out.RGBAAt(2, 11); c != (color.RGBA{R: 200, G: 100, B: 50, A: 0xFF}) {
		t.Errorf("pixel = %v", c)
	}
}
