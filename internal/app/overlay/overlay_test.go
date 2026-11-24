package overlay_test

import (
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/app/overlay"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

const w, h = invaders.ScreenWidth, invaders.ScreenHeight

// TestMidwayColours checks the strips' edges (MAME layout/invaders.lay).
func TestMidwayColours(t *testing.T) {
	t.Parallel()

	screen := make([]byte, w*h)
	for i := range screen {
		screen[i] = 1
	}

	dst := make([]byte, w*h*4)
	overlay.NewPainter(overlay.Midway()).Paint(screen, dst)

	const (
		white = "white"
		red   = "red"
		green = "green"
	)

	colour := func(x, y int) string {
		i := (y*w + x) * 4
		switch [3]byte{dst[i], dst[i+1], dst[i+2]} {
		case [3]byte{255, 255, 255}:
			return white
		case [3]byte{255, 32, 32}:
			return red
		case [3]byte{32, 255, 32}:
			return green
		}

		return "?"
	}

	tests := []struct {
		x, y int
		want string
	}{
		{100, 31, white},
		{100, 32, red},
		{100, 61, red},
		{100, 62, white},
		{0, 183, white},
		{0, 184, green},
		{223, 239, green},
		{15, 240, white},
		{16, 240, green},
		{133, 255, green},
		{134, 250, white},
		{223, 255, white},
	}
	for _, tt := range tests {
		if got := colour(tt.x, tt.y); got != tt.want {
			t.Errorf("(%d, %d) is %s, want %s", tt.x, tt.y, got, tt.want)
		}
	}

	if dst[3] != 255 {
		t.Errorf("alpha %d, want 255", dst[3])
	}
}

func TestDarkPixelsAreBlack(t *testing.T) {
	t.Parallel()

	screen := make([]byte, w*h)
	dst := make([]byte, w*h*4)

	for i := range dst {
		dst[i] = 0x77
	}

	overlay.NewPainter(overlay.Midway()).Paint(screen, dst)

	for i := range len(dst) / 4 {
		if px := [4]byte(dst[i*4 : i*4+4]); px != [4]byte{0, 0, 0, 255} {
			t.Fatalf("pixel %d = % X, want opaque black", i, px)
		}
	}
}

func TestRectsAreClipped(t *testing.T) {
	t.Parallel()

	p := overlay.NewPainter([]overlay.Rect{{-10, -10, 1000, 1000, 1, 2, 3}})
	screen := []byte{1}
	dst := make([]byte, 4)
	p.Paint(screen, dst)

	if dst[0] != 1 || dst[1] != 2 || dst[2] != 3 {
		t.Errorf("pixel 0 = % X, want 01 02 03", dst)
	}
}

func BenchmarkPaint(b *testing.B) {
	screen := make([]byte, w*h)
	for i := range screen {
		screen[i] = byte(i/7) & 1
	}

	dst := make([]byte, w*h*4)
	p := overlay.NewPainter(overlay.Midway())

	for b.Loop() {
		p.Paint(screen, dst)
	}
}
