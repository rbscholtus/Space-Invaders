package invaders_test

import (
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

// TestHiScore: SetHiScore stores four BCD digits, low two first, in RAM
// and as the power-on value, and refuses what the game can't show.
func TestHiScore(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, nil)

	if err := b.SetHiScore(1230); err != nil {
		t.Fatal(err)
	}

	if got := b.HiScore(); got != 1230 {
		t.Errorf("HiScore() = %d, want 1230", got)
	}

	if lo, hi := b.Read(0x20F4), b.Read(0x20F5); lo != 0x30 || hi != 0x12 {
		t.Errorf("RAM 20F4–20F5 = %02X %02X, want 30 12", lo, hi)
	}

	if lo, hi := b.Read(0x1BF4), b.Read(0x1BF5); lo != 0x30 || hi != 0x12 {
		t.Errorf("power-on value at 1BF4–1BF5 = %02X %02X, want 30 12", lo, hi)
	}

	for _, bad := range []int{-10, 10000} {
		if err := b.SetHiScore(bad); !errors.Is(err, invaders.ErrHiScore) {
			t.Errorf("SetHiScore(%d) = %v, want ErrHiScore", bad, err)
		}
	}

	if got := b.HiScore(); got != 1230 {
		t.Errorf("after refused values, HiScore() = %d, want 1230", got)
	}
}

// How the game draws the HI-SCORE (checked in the ROM): RAM 20F6–20F7 hold
// its screen address; each digit is character 0x1A + digit, whose 8 bytes
// are at 0x1E00 + 8 × character and go to video RAM 0x20 apart (one per
// row of the rotated screen), so the next digit starts 0x100 further on.
const (
	hiScoreScreenAddr = 0x20F6
	charSet           = 0x1E00
	charZero          = 0x1A
	charRows          = 8
	rowStride         = 0x20
)

// shownHiScore returns the digits on screen where the game draws the
// HI-SCORE, or -1 where a digit isn't a digit's glyph.
func shownHiScore(b *invaders.Board) []int {
	at := uint16(b.Read(hiScoreScreenAddr)) | uint16(b.Read(hiScoreScreenAddr+1))<<8
	digits := make([]int, 4)

	for i := range digits {
		cell := make([]byte, charRows)
		for r := range cell {
			cell[r] = b.Read(at + uint16(i*charRows*rowStride+r*rowStride))
		}

		digits[i] = -1

		for d := range 10 {
			glyph := make([]byte, charRows)
			for r := range glyph {
				glyph[r] = b.Read(uint16(charSet + (charZero+d)*charRows + r))
			}

			if slices.Equal(cell, glyph) {
				digits[i] = d
			}
		}
	}

	return digits
}

// TestROMHiScoreShown: a HI-SCORE set before the game starts is on screen
// in the attract mode, and still after a Reset (a power cut). With -update
// it writes the screen to testdata/golden/png/hiscore.png to look at.
// Local only: skipped without your own ROMs.
func TestROMHiScoreShown(t *testing.T) {
	t.Parallel()

	const second = 60 // frames

	g := newGame(t)
	g.hold(invaders.Inputs{}, second)

	if got := shownHiScore(g.b); !slices.Equal(got, []int{0, 0, 0, 0}) {
		t.Fatalf("without a remembered HI-SCORE the screen shows %v, want 0000", got)
	}

	g = newGame(t)
	if err := g.b.SetHiScore(1230); err != nil {
		t.Fatal(err)
	}

	g.hold(invaders.Inputs{}, second)

	if got := shownHiScore(g.b); !slices.Equal(got, []int{1, 2, 3, 0}) {
		t.Errorf("the screen shows HI-SCORE %v, want 1230", got)
	}

	if *update {
		writeScreenPNG(t, filepath.Join(pngDir, "hiscore.png"), g.b)
	}

	g.b.Reset()
	g.hold(invaders.Inputs{}, second)

	if got := shownHiScore(g.b); !slices.Equal(got, []int{1, 2, 3, 0}) || g.b.HiScore() != 1230 {
		t.Errorf("after a Reset the screen shows %v and RAM holds %d, want 1230", got, g.b.HiScore())
	}
}

// writeScreenPNG writes the board's screen to path, for a person to look at.
func writeScreenPNG(t *testing.T, path string, b *invaders.Board) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}

	img := image.NewGray(image.Rect(0, 0, invaders.ScreenWidth, invaders.ScreenHeight))
	invaders.Screen(b.Video(), img.Pix)

	for i, lit := range img.Pix {
		img.Pix[i] = lit * 255
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}

	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
