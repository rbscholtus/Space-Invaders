package invaders_test

import (
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

// TestScreenRotation checks the corners of the rotated picture (DESIGN §10):
// memory row 0 is the screen's left column, and x = 0 in memory is the
// bottom of the screen.
func TestScreenRotation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		addr     uint16
		bit      int
		col, row int
	}{
		{"first byte, bit 0: bottom left", 0x2400, 0, 0, 255},
		{"first byte, bit 7", 0x2400, 7, 0, 248},
		{"end of row 0, bit 7: top left", 0x241F, 7, 0, 0},
		{"row 1, first byte, bit 0", 0x2420, 0, 1, 255},
		{"last byte, bit 7: top right", 0x3FFF, 7, 223, 0},
		{"start of the last row, bit 0: bottom right", 0x3FE0, 0, 223, 255},
	}
	for _, tt := range tests {
		b := newBoard(t, invaders.DIPs{}, nil)
		b.Write(tt.addr, 1<<tt.bit)

		screen := make([]byte, invaders.ScreenWidth*invaders.ScreenHeight)
		invaders.Screen(b.Video(), screen)

		lit := 0

		for i, p := range screen {
			if p == 0 {
				continue
			}

			lit++

			if col, row := i%invaders.ScreenWidth, i/invaders.ScreenWidth; col != tt.col || row != tt.row || p != 1 {
				t.Errorf("%s: pixel %d at (%d, %d), want 1 at (%d, %d)", tt.name, p, col, row, tt.col, tt.row)
			}
		}

		if lit != 1 {
			t.Errorf("%s: %d pixels lit, want 1", tt.name, lit)
		}
	}
}

// TestVideoIsALiveView: Video reflects later writes without copying.
func TestVideoIsALiveView(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, nil)
	video := b.Video()

	if len(video) != 7168 {
		t.Fatalf("len(Video()) = %d, want 7168", len(video))
	}

	b.Write(0x2401, 0x5A)

	if video[1] != 0x5A {
		t.Errorf("Video()[1] = %02X after writing 2401H, want 5A", video[1])
	}
}

func BenchmarkScreen(b *testing.B) {
	video := make([]byte, 7168)
	for i := range video {
		video[i] = byte(i * 37)
	}

	dst := make([]byte, invaders.ScreenWidth*invaders.ScreenHeight)
	for b.Loop() {
		invaders.Screen(video, dst)
	}
}
