package app

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/rbscholtus/Space-Invaders/internal/app/cabinet"
)

// This file loads the cabinet's art in the background (DESIGN §12):
// fetched in the browser, embedded on the desktop. The picture runs while
// it loads, and the cabinet fades in when it's there.

// ArtFetcher reads one of the cabinet's images by name ("cabinet@2x.jpg").
// It may block: it's only called from a goroutine.
type ArtFetcher func(name string) ([]byte, error)

// artBandRows is how many rows of an image are converted between two
// pauses (artPause).
const artBandRows = 128

// artPause lets the browser draw a frame between two steps of loading.
// Under WebAssembly Go has one thread and doesn't preempt: a goroutine
// busy decoding holds up Update and Draw until it blocks, and only a sleep
// (not runtime.Gosched) hands control back to the page so the next
// animation frame can run. (From Chippie.)
func artPause() { time.Sleep(time.Millisecond) }

// decodedArt is the cabinet's images at one scale, from the loading
// goroutine. They are premultiplied *image.RGBA, which Ebitengine uploads
// as they are; any other type it converts pixel by pixel on the game
// goroutine.
type decodedArt struct {
	scale    int
	cabinet  *image.RGBA
	backdrop *image.RGBA
	pressed  map[string]*image.RGBA // by control id
	err      error
}

// loadArt fetches and decodes the cabinet's images at scale n in a
// goroutine, pausing between steps so the game keeps running (artPause).
// The channel gets one result.
func loadArt(fetch ArtFetcher, cab *cabinet.Cabinet, n int, pause func()) <-chan decodedArt {
	out := make(chan decodedArt, 1) // buffered: a result nobody wants doesn't block

	go func() {
		res := decodedArt{scale: n, pressed: map[string]*image.RGBA{}}

		get := func(name string, decode func([]byte) (image.Image, error)) *image.RGBA {
			if res.err != nil {
				return nil
			}

			data, err := fetch(name)
			if err == nil {
				var m image.Image
				if m, err = decode(data); err == nil {
					pause()

					return toRGBA(m, artBandRows, pause)
				}
			}

			res.err = fmt.Errorf("cabinet art %s: %w", name, err)

			return nil
		}

		res.cabinet = get(cabinet.FileName(cab.Image, n, "jpg"), decodeJPEG)
		res.backdrop = get(cab.Backdrop, decodeJPEG)

		for _, c := range cab.Controls {
			if c.Pressed != "" {
				res.pressed[c.ID] = get(cabinet.FileName(c.Pressed, n, "png"), decodePNG)
			}
		}

		out <- res
	}()

	return out
}

func decodeJPEG(b []byte) (image.Image, error) {
	return jpeg.Decode(bytes.NewReader(b)) //nolint:wrapcheck // loadArt adds the file's name
}

func decodePNG(b []byte) (image.Image, error) {
	return png.Decode(bytes.NewReader(b)) //nolint:wrapcheck // loadArt adds the file's name
}

// toRGBA returns m as a premultiplied *image.RGBA with its origin at 0,0,
// converting bandRows rows at a time and calling pause between bands.
func toRGBA(m image.Image, bandRows int, pause func()) *image.RGBA {
	if rgba, ok := m.(*image.RGBA); ok && rgba.Rect.Min == (image.Point{}) {
		return rgba
	}

	b := m.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

	for y := 0; y < b.Dy(); y += bandRows {
		band := image.Rect(0, y, b.Dx(), min(y+bandRows, b.Dy()))
		draw.Draw(out, band, m, b.Min.Add(band.Min), draw.Src)
		pause()
	}

	return out
}

// cabinetArt is the loaded images at one scale, on the GPU.
type cabinetArt struct {
	scale    int
	cabinet  *ebiten.Image
	backdrop *ebiten.Image
	pressed  map[string]*ebiten.Image
}

func newCabinetArt(d decodedArt) *cabinetArt {
	a := &cabinetArt{
		scale:    d.scale,
		cabinet:  ebiten.NewImageFromImage(d.cabinet),
		backdrop: ebiten.NewImageFromImage(d.backdrop),
		pressed:  map[string]*ebiten.Image{},
	}

	for id, m := range d.pressed {
		a.pressed[id] = ebiten.NewImageFromImage(m)
	}

	return a
}

func (a *cabinetArt) dispose() {
	a.cabinet.Deallocate()
	a.backdrop.Deallocate()

	for _, s := range a.pressed {
		s.Deallocate()
	}
}
