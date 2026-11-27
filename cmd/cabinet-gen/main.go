// Command cabinet-gen turns the generated masters into the cabinet art the
// app loads (DESIGN §12). It reads the masters from assets/source/ (not
// committed: they're large) and the measurements in
// assets/cabinet/recipe.json, and writes into assets/cabinet/:
//
//	cabinet@1x.jpg … @3x.jpg        the cabinet, with the glass art, the
//	                                marquee and our lettering put in, its
//	                                edges faded into the background
//	backdrop.jpg                    the moon painting behind the picture
//	<button>-pressed@1x.png … @3x   each button pressed in, cut from the cabinet
//	cabinet.json                    where everything is, at @3x
//
// @3x is the cropped master's full size; @2x and @1x are two thirds and one
// third of it. Run it again whenever the art or the recipe changes:
//
//	go run ./cmd/cabinet-gen
//
// To check the alignment, also write a preview with every button pressed,
// the picture's window darkened and the controls outlined:
//
//	go run ./cmd/cabinet-gen -preview /tmp/cabinet.png
//
// Adapted from Chippie's cmd/skin-gen (github.com/rbscholtus/chippie).
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rbscholtus/Space-Invaders/internal/app/cabinet"
)

const (
	jpegQuality  = 85
	backdropFile = "backdrop.jpg" // one size, the art's own: the picture is drawn over it
	imageBase    = "cabinet"
	pressedTag   = "-pressed"
	// aspectSlack is how far an insert's source and target shapes may
	// differ before the art would look stretched.
	aspectSlack = 0.01
)

var errRecipe = errors.New("cabinet-gen: bad recipe")

// recipe is assets/cabinet/recipe.json. Coordinates are in the cabinet
// master's pixels, except an insert's From, which is in its own master's.
type recipe struct {
	Name    string `json:"name"`
	Cabinet struct {
		Source   string     `json:"source"`
		Crop     rectJSON   `json:"crop"`
		PaintOut []rectJSON `json:"paintOut"`
		// Feather is how many pixels at the crop's edges fade into the
		// background colour.
		Feather int `json:"feather"`
	} `json:"cabinet"`
	// Glass is the moon painting put into the cabinet's black glass, and
	// Marquee the art put into its black title panel.
	Glass   insert `json:"glass"`
	Marquee insert `json:"marquee"`
	// Screen is the window in the glass where the picture is drawn.
	Screen rectJSON `json:"screen"`
	// Upper is the cabinet from its top to the control panel's front edge,
	// shown when the whole cabinet would make the picture too small.
	Upper    rectJSON     `json:"upper"`
	Controls []controlRec `json:"controls"`
	Texts    []textRec    `json:"texts"`
}

// insert scales the From part of another master into To.
type insert struct {
	Source string   `json:"source"`
	From   rectJSON `json:"from"`
	To     rectJSON `json:"to"`
}

type rectJSON struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

func (r rectJSON) rect() image.Rectangle { return image.Rect(r.X, r.Y, r.X+r.W, r.Y+r.H) }

// controlRec is one control: a circle (cx, cy, r) or a rectangle. A
// button's r is its pressed sprite's radius: the cap that moves.
type controlRec struct {
	ID   string       `json:"id"`
	Kind cabinet.Kind `json:"kind"`
	CX   int          `json:"cx"`
	CY   int          `json:"cy"`
	R    int          `json:"r"`
	X    int          `json:"x"`
	Y    int          `json:"y"`
	W    int          `json:"w"`
	H    int          `json:"h"`
}

func main() {
	src := flag.String("source", "assets/source", "directory of the master images")
	dir := flag.String("dir", "assets/cabinet", "directory with recipe.json, and for the output")
	preview := flag.String("preview", "", "also write a composed preview here, e.g. /tmp/cabinet.png")

	flag.Parse()

	if err := run(*src, *dir, *preview); err != nil {
		log.Fatal(err)
	}

	log.Printf("built %s", *dir)
}

// outputs are everything cabinet-gen writes, at @3x.
type outputs struct {
	cab      cabinet.Cabinet
	cabinet  *image.NRGBA
	backdrop *image.NRGBA
	pressed  map[string]*image.NRGBA
}

func run(srcDir, dir, preview string) error {
	var rc recipe
	if err := readJSON(filepath.Join(dir, "recipe.json"), &rc); err != nil {
		return err
	}

	masters := map[string]*image.NRGBA{}

	for _, name := range []string{rc.Cabinet.Source, rc.Glass.Source, rc.Marquee.Source} {
		if masters[name] != nil {
			continue
		}

		m, err := readPNG(filepath.Join(srcDir, name))
		if err != nil {
			return err
		}

		masters[name] = m
	}

	out, err := build(&rc, masters)
	if err != nil {
		return err
	}

	if err := writeOutputs(dir, &out); err != nil {
		return err
	}

	if preview != "" {
		if err := writePreview(preview, &out); err != nil {
			return err
		}
	}

	return writeCabinet(filepath.Join(dir, "cabinet.json"), &out.cab)
}

// build makes the outputs from the masters, following the recipe.
func build(rc *recipe, masters map[string]*image.NRGBA) (outputs, error) {
	master := masters[rc.Cabinet.Source]
	for _, r := range rc.Cabinet.PaintOut {
		paintOut(master, r.rect())
	}

	crop, screen := rc.Cabinet.Crop.rect(), rc.Screen.rect()

	switch {
	case !crop.In(master.Bounds()):
		return outputs{}, fmt.Errorf("%w: crop %v outside the %v cabinet", errRecipe, crop, master.Bounds())
	case !screen.In(rc.Glass.To.rect()):
		return outputs{}, fmt.Errorf("%w: screen %v outside the glass %v", errRecipe, screen, rc.Glass.To.rect())
	}

	for _, in := range []insert{rc.Glass, rc.Marquee} {
		if err := put(master, masters[in.Source], in); err != nil {
			return outputs{}, err
		}
	}

	for _, t := range rc.Texts {
		if err := letter(master, t); err != nil {
			return outputs{}, err
		}
	}

	out := outputs{pressed: map[string]*image.NRGBA{}}
	out.cab = cabinet.Cabinet{
		Name: rc.Name, Width: float64(crop.Dx()), Height: float64(crop.Dy()),
		Image: imageBase, Backdrop: backdropFile,
		Screen:  relative(screen, crop.Min),
		Marquee: relative(rc.Marquee.To.rect(), crop.Min),
		Upper:   relative(rc.Upper.rect(), crop.Min),
	}

	out.cab.Controls = controls(rc.Controls, master, crop.Min, out.pressed)

	out.cabinet = toNRGBA(master.SubImage(crop))
	fill := borderColour(out.cabinet)
	out.cab.Background = fmt.Sprintf("#%02x%02x%02x", fill.R, fill.G, fill.B)
	fadeEdges(out.cabinet, rc.Cabinet.Feather, fill)

	out.backdrop = backdropOf(masters[rc.Glass.Source], rc.Glass, screen)

	if err := out.cab.Validate(); err != nil {
		return outputs{}, err //nolint:wrapcheck // ErrInvalid says what's wrong
	}

	return out, nil
}

// controls describes the controls relative to the crop's origin, and cuts
// each button's pressed sprite into pressed.
func controls(
	recs []controlRec, master *image.NRGBA, origin image.Point, pressed map[string]*image.NRGBA,
) []cabinet.Control {
	ctls := make([]cabinet.Control, 0, len(recs))

	for _, c := range recs {
		ctl := cabinet.Control{ID: c.ID, Kind: c.Kind}

		if c.Kind == cabinet.Slot {
			r := relative(image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H), origin)
			ctl.X, ctl.Y, ctl.W, ctl.H = r.X, r.Y, r.W, r.H
		} else {
			ctl.CX, ctl.CY, ctl.R = float64(c.CX-origin.X), float64(c.CY-origin.Y), float64(c.R)
		}

		if c.Kind == cabinet.Button {
			ctl.Pressed = c.ID + pressedTag
			pressed[ctl.Pressed] = pressedSprite(master, c.CX, c.CY, c.R)
		}

		ctls = append(ctls, ctl)
	}

	return ctls
}

// put scales an insert's art into the cabinet master.
func put(master, art *image.NRGBA, in insert) error {
	from, to := in.From.rect(), in.To.rect()

	switch {
	case from.Empty() || to.Empty():
		return fmt.Errorf("%w: %s: an empty rectangle", errRecipe, in.Source)
	case !from.In(art.Bounds()):
		return fmt.Errorf("%w: %s: from %v outside the %v image", errRecipe, in.Source, from, art.Bounds())
	case !to.In(master.Bounds()):
		return fmt.Errorf("%w: %s: to %v outside the cabinet", errRecipe, in.Source, to)
	case math.Abs(aspect(from)/aspect(to)-1) > aspectSlack:
		return fmt.Errorf("%w: %s: from %v and to %v differ in shape", errRecipe, in.Source, from, to)
	}

	scaleInto(master, to, art.SubImage(from))

	return nil
}

func aspect(r image.Rectangle) float64 { return float64(r.Dx()) / float64(r.Dy()) }

// backdropOf cuts the part of the glass art that lies under the screen, at
// the art's own resolution.
func backdropOf(art *image.NRGBA, glass insert, screen image.Rectangle) *image.NRGBA {
	from, to := glass.From.rect(), glass.To.rect()
	sx, sy := float64(from.Dx())/float64(to.Dx()), float64(from.Dy())/float64(to.Dy())
	at := func(p image.Point) image.Point {
		return image.Pt(from.Min.X+int(math.Round(float64(p.X-to.Min.X)*sx)),
			from.Min.Y+int(math.Round(float64(p.Y-to.Min.Y)*sy)))
	}

	return toNRGBA(art.SubImage(image.Rectangle{Min: at(screen.Min), Max: at(screen.Max)}))
}

func relative(r image.Rectangle, origin image.Point) cabinet.Rect {
	r = r.Sub(origin)

	return cabinet.Rect{X: float64(r.Min.X), Y: float64(r.Min.Y), W: float64(r.Dx()), H: float64(r.Dy())}
}

// writeOutputs writes the cabinet and the pressed buttons at @1x, @2x and
// @3x, and the backdrop.
func writeOutputs(dir string, out *outputs) error {
	if err := writeJPEG(filepath.Join(dir, backdropFile), out.backdrop); err != nil {
		return err
	}

	w, h := out.cabinet.Bounds().Dx(), out.cabinet.Bounds().Dy()

	for n := 1; n <= cabinet.MaxScale; n++ {
		size := func(v int) int { return int(math.Round(float64(v*n) / cabinet.MaxScale)) }

		if err := writeJPEG(filepath.Join(dir, cabinet.FileName(imageBase, n, "jpg")),
			scaleTo(out.cabinet, size(w), size(h))); err != nil {
			return err
		}

		for name, s := range out.pressed {
			b := s.Bounds()
			if err := writePNG(filepath.Join(dir, cabinet.FileName(name, n, "png")),
				scaleTo(s, size(b.Dx()), size(b.Dy()))); err != nil {
				return err
			}
		}
	}

	return nil
}

func parseColour(s string) (color.NRGBA, error) {
	if len(s) != len("#rrggbb") || s[0] != '#' {
		return color.NRGBA{}, fmt.Errorf("%w: colour %q", errRecipe, s)
	}

	v, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return color.NRGBA{}, fmt.Errorf("%w: colour %q", errRecipe, s)
	}

	return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xFF}, nil //nolint:gosec // G115: 24-bit value
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read recipe: %w", err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("%w: %s: %w", errRecipe, path, err)
	}

	return nil
}

func readPNG(path string) (*image.NRGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open master image: %w", err)
	}
	defer f.Close()

	m, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	return toNRGBA(m), nil
}

func writePNG(path string, m image.Image) error {
	return writeFile(path, func(f *os.File) error {
		enc := png.Encoder{CompressionLevel: png.BestCompression}

		return enc.Encode(f, m)
	})
}

func writeJPEG(path string, m image.Image) error {
	return writeFile(path, func(f *os.File) error {
		return jpeg.Encode(f, m, &jpeg.Options{Quality: jpegQuality})
	})
}

func writeCabinet(path string, c *cabinet.Cabinet) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("encode cabinet: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write cabinet: %w", err)
	}

	return nil
}

func writeFile(path string, encode func(*os.File) error) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}

	if err := encode(f); err != nil {
		_ = f.Close() // the encode error is the one to report

		return fmt.Errorf("encode %s: %w", path, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}
