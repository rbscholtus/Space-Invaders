package romscreen_test

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rbscholtus/Space-Invaders/internal/app/romscreen"
	"github.com/rbscholtus/Space-Invaders/internal/testrig"
)

var (
	errBlocked  = errors.New("blocked")
	errVeryLong = errors.New(strings.Repeat("very long problem ", 20))
)

func TestModeNames(t *testing.T) {
	t.Parallel()

	for m, want := range map[romscreen.Mode]string{
		romscreen.Loading: "loading", romscreen.NoROMs: "roms", romscreen.Loaded: "game", romscreen.Mode(9): "unknown",
	} {
		if m.String() != want {
			t.Errorf("%d.String() = %q, want %q", m, m.String(), want)
		}
	}
}

func TestNothingRemembered(t *testing.T) {
	t.Parallel()

	f := romscreen.New(true)
	if f.Mode() != romscreen.Loading || !strings.Contains(strings.Join(f.Lines(), " "), "Looking") {
		t.Fatalf("start: mode %v, lines %q", f.Mode(), f.Lines())
	}

	if img := f.Restored(nil, false, nil); img != nil || f.Mode() != romscreen.NoROMs {
		t.Fatalf("nothing remembered: image %v, mode %v", img != nil, f.Mode())
	}

	text := strings.Join(f.Lines(), "\n")
	if !strings.Contains(text, "NO ROMS LOADED") || !strings.Contains(text, "press O") {
		t.Errorf("ROM screen:\n%s", text)
	}

	for _, line := range f.Lines() {
		if len(line) > 28 {
			t.Errorf("line %q is longer than 28 columns", line)
		}
	}

	if f.Restored(nil, false, nil) != nil || f.Mode() != romscreen.NoROMs {
		t.Error("a second Restored should do nothing")
	}
}

func TestNoPickerWithoutIt(t *testing.T) {
	t.Parallel()

	f := romscreen.New(false)
	f.Restored(nil, false, nil)

	if strings.Contains(strings.Join(f.Lines(), " "), "press O") {
		t.Error("offers the picker where there is none")
	}
}

func TestStorageProblem(t *testing.T) {
	t.Parallel()

	f := romscreen.New(true)
	f.Restored(nil, false, errBlocked)

	if f.Mode() != romscreen.NoROMs || !strings.Contains(f.Message(), "blocked") {
		t.Errorf("mode %v, message %q", f.Mode(), f.Message())
	}
}

func TestDamagedRemembered(t *testing.T) {
	t.Parallel()

	f := romscreen.New(true)
	if img := f.Restored(make([]byte, 0x2000), true, nil); img != nil || !strings.Contains(f.Message(), "damaged") {
		t.Errorf("damaged image: returned %v, message %q", img != nil, f.Message())
	}
}

// TestWrongFiles: the ROM screen explains what's wrong, wrapped to fit.
func TestWrongFiles(t *testing.T) {
	t.Parallel()

	f := romscreen.New(true)
	f.Restored(nil, false, nil)

	if img := f.Offer(fstest.MapFS{"holiday.jpg": {Data: []byte("not a rom")}}); img != nil {
		t.Fatal("accepted a photo")
	}

	if f.Mode() != romscreen.NoROMs || f.Message() != "These aren't the invaders ROMs. Found: holiday.jpg" {
		t.Errorf("mode %v, message %q", f.Mode(), f.Message())
	}

	lines := f.Lines()
	for _, line := range lines {
		if len(line) > 28 {
			t.Errorf("line %q is longer than 28 columns", line)
		}
	}

	// A long message is cut to six lines, and the text still ends above
	// the controls, with an empty row between (32 rows, from row 1; two
	// empty rows below the controls).
	f.Fail(errVeryLong)

	if n := len(f.Lines()); n > 7+2+1+6 {
		t.Errorf("%d lines with a long message", n)
	}

	if last, controlsAt := 1+len(f.Lines())-1, 32-2-len(romscreen.Controls()); last >= controlsAt-1 {
		t.Errorf("the text ends at row %d, the controls start at row %d", last, controlsAt)
	}

	for _, line := range romscreen.Controls() {
		if len(line) > 28 {
			t.Errorf("control %q is longer than 28 columns", line)
		}
	}

	pic := make([]byte, 224*256)
	f.Draw(pic, 224)

	lit := 0
	for _, p := range pic {
		lit += int(p)
	}

	if lit == 0 {
		t.Error("Draw drew nothing")
	}

	// The controls are at the bottom, in rows 22–29 (the keys in the
	// overlay's green band, 23–29), and the last two rows stay empty.
	bottom := 0
	for _, p := range pic[22*8*224 : 30*8*224] {
		bottom += int(p)
	}

	if bottom == 0 {
		t.Error("no controls at the bottom")
	}

	for _, p := range pic[30*8*224:] {
		if p != 0 {
			t.Fatal("the last two rows aren't empty")
		}
	}
}

func TestForget(t *testing.T) {
	t.Parallel()

	f := romscreen.New(true)
	f.Start()

	if f.Mode() != romscreen.Loaded {
		t.Fatal("Start didn't start")
	}

	f.Forget()

	if f.Mode() != romscreen.NoROMs || f.Message() == "" {
		t.Errorf("after Forget: mode %v, message %q", f.Mode(), f.Message())
	}
}

// TestRealROMs: your own ROMs, dropped, start the game; remembered, they
// start it at once (local only).
func TestRealROMs(t *testing.T) {
	t.Parallel()

	image := testrig.ROMs(t)

	f := romscreen.New(true)
	f.Restored(nil, false, nil)

	if got := f.Offer(fstest.MapFS{
		"invaders.e": {Data: image[0x1800:]}, "invaders.f": {Data: image[0x1000:0x1800]},
		"invaders.g": {Data: image[0x800:0x1000]}, "invaders.h": {Data: image[:0x800]},
	}); got == nil || f.Mode() != romscreen.Loaded {
		t.Fatalf("dropped ROMs: mode %v, message %q", f.Mode(), f.Message())
	}

	again := romscreen.New(true)
	if got := again.Restored(image, true, nil); got == nil || again.Mode() != romscreen.Loaded {
		t.Errorf("remembered ROMs: mode %v, message %q", again.Mode(), again.Message())
	}
}
