package app

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/rbscholtus/Space-Invaders/internal/app/cabinet"
	"github.com/rbscholtus/Space-Invaders/internal/app/input"
)

const (
	// fadeStep is how much the cabinet fades in per frame once loaded.
	fadeStep = 1.0 / 20
	// backdropLevel is the moon backdrop's brightness behind the tube
	// (DESIGN §11.3; ⚠ tune by eye): dim enough for the game to read
	// over it.
	backdropLevel = 0.35
)

// cabinetView is the cabinet: its layout, art and pressed buttons.
type cabinetView struct {
	cab   *cabinet.Cabinet
	fetch ArtFetcher // nil: no art (tests), just the picture
	pause func()     // between loading steps

	art     *cabinetArt
	loading <-chan decodedArt
	fade    float32 // the cabinet fading in once loaded, 0 to 1
	failed  bool    // loading failed: don't try again

	bg color.RGBA
	op ebiten.DrawImageOptions
}

func newCabinetView(cab *cabinet.Cabinet, fetch ArtFetcher) *cabinetView {
	v := &cabinetView{cab: cab, fetch: fetch, pause: artPause}

	if cab != nil {
		c := cab.Colour()
		v.bg = color.RGBA{R: c[0], G: c[1], B: c[2], A: 0xFF}
	}

	return v
}

// want starts loading art big enough for the layout, if it isn't loaded
// or loading already. The screen view needs only the backdrop, so it loads
// the smallest cabinet with it.
func (v *cabinetView) want(l layout) {
	if v.cab == nil || v.fetch == nil || v.failed || v.loading != nil {
		return
	}

	n := 1
	if l.view == ViewCabinet {
		n = v.cab.ScaleFor(float64(l.cabinet.Dx()))
	}

	if v.art != nil && v.art.scale >= n {
		return
	}

	v.loading = loadArt(v.fetch, v.cab, n, v.pause)
}

// update collects loaded art and fades the cabinet in. It returns the new
// backdrop when art arrived, and an error when loading failed. The
// Ebitengine images are made here, on the game goroutine.
func (v *cabinetView) update() (*ebiten.Image, error) {
	var backdrop *ebiten.Image

	select {
	case res := <-v.loading:
		v.loading = nil

		if res.err != nil {
			v.failed = true

			return nil, res.err
		}

		if v.art != nil {
			v.art.dispose()
		}

		v.art = newCabinetArt(res)
		backdrop = v.art.backdrop
	default:
	}

	if v.art != nil && v.fade < 1 {
		v.fade = min(1, v.fade+fadeStep)
	}

	return backdrop, nil
}

// loaded reports whether the art is there.
func (v *cabinetView) loaded() bool { return v.art != nil }

// backdropLevels are the moon backdrop's brightness behind the tube and at
// its edge, for the view: in the cabinet the edge matches the glass around
// it; in the screen view there's no glass, so it stays dim.
func (v *cabinetView) backdropLevels(view View) (float32, float32) {
	if view == ViewCabinet {
		return backdropLevel * v.fade, v.fade
	}

	return backdropLevel * v.fade, backdropLevel * v.fade
}

// draw draws the cabinet, fading in, with the buttons held down pressed
// in. The monitor is drawn over it afterwards.
func (v *cabinetView) draw(dst *ebiten.Image, l layout, keys input.Keys) {
	if v.art == nil || l.view != ViewCabinet {
		return
	}

	dst.Fill(color.RGBA{
		R: uint8(float32(v.bg.R) * v.fade), G: uint8(float32(v.bg.G) * v.fade),
		B: uint8(float32(v.bg.B) * v.fade), A: 0xFF,
	})

	op := &v.op
	op.GeoM.Reset()
	op.ColorScale.Reset()
	op.ColorScale.ScaleAlpha(v.fade)
	op.Filter = ebiten.FilterLinear

	b := v.art.cabinet.Bounds()
	op.GeoM.Scale(float64(l.cabinet.Dx())/float64(b.Dx()), float64(l.cabinet.Dy())/float64(b.Dy()))
	op.GeoM.Translate(float64(l.cabinet.Min.X), float64(l.cabinet.Min.Y))
	dst.DrawImage(v.art.cabinet, op)

	for _, c := range v.cab.Controls {
		img, ok := v.art.pressed[c.ID]
		if !ok || !held(keys, c.ID) {
			continue
		}

		r := l.rect(c.Bounds())
		s := img.Bounds()

		op.GeoM.Reset()
		op.GeoM.Scale(float64(r.Dx())/float64(s.Dx()), float64(r.Dy())/float64(s.Dy()))
		op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
		dst.DrawImage(img, op)
	}
}

// held reports whether the button with the id is held down.
func held(keys input.Keys, id string) bool {
	switch id {
	case cabinet.Left:
		return keys.Left
	case cabinet.Right:
		return keys.Right
	case cabinet.Fire:
		return keys.Fire
	case cabinet.Start1:
		return keys.Start1
	case cabinet.Start2:
		return keys.Start2
	}

	return false
}
