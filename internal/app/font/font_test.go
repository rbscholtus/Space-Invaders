package font

import (
	"strings"
	"testing"
)

// TestArtIsWellFormed: 95 glyphs from space to '~', in order, each 7 rows
// of 5 pixels drawn with '#' and '.'.
func TestArtIsWellFormed(t *testing.T) {
	t.Parallel()

	lines := strings.Split(strings.TrimSpace(glyphArt), "\n")
	if want := int(numGlyphs) * (glyphHeight + 1); len(lines) != want {
		t.Fatalf("%d lines, want %d", len(lines), want)
	}

	for i := range int(numGlyphs) {
		label := lines[i*(glyphHeight+1)]

		want := string(rune(int(first) + i))
		switch want {
		case " ":
			want = "space"
		case "`":
			want = "backtick"
		}

		if label != want {
			t.Errorf("glyph %d is labelled %q, want %q", i, label, want)
		}

		for r, row := range lines[i*(glyphHeight+1)+1 : (i+1)*(glyphHeight+1)] {
			if len(row) != glyphWidth || strings.Trim(row, "#.") != "" {
				t.Errorf("%q row %d is %q: want 5 of '#' and '.'", want, r, row)
			}
		}
	}
}

// TestEveryGlyphIsDistinct: every character but space lights pixels, and no
// two look the same.
func TestEveryGlyphIsDistinct(t *testing.T) {
	t.Parallel()

	seen := map[[glyphHeight]uint8]rune{}

	for c := rune(first); c <= last; c++ {
		g := *glyph(c)

		if c != ' ' && g == ([glyphHeight]uint8{}) {
			t.Errorf("%q is blank", c)
		}

		if other, dup := seen[g]; dup {
			t.Errorf("%q looks the same as %q", c, other)
		}

		seen[g] = c
	}
}

// render draws text into a fresh picture and returns it as art rows.
func render(width, height, x, y int, text string) []string {
	pic := make([]byte, width*height)
	for i := range pic {
		pic[i] = 7 // overwritten in drawn cells
	}

	Draw(pic, width, x, y, text)

	rows := make([]string, height)
	for r := range rows {
		var sb strings.Builder

		for c := range width {
			switch pic[r*width+c] {
			case 1:
				sb.WriteByte('#')
			case 0:
				sb.WriteByte('.')
			default:
				sb.WriteByte('-')
			}
		}

		rows[r] = sb.String()
	}

	return rows
}

func TestDrawA(t *testing.T) {
	t.Parallel()

	got := render(10, 9, 1, 1, "A")
	want := []string{
		"----------",
		"-..###...-",
		"-.#...#..-",
		"-.#...#..-",
		"-.#####..-",
		"-.#...#..-",
		"-.#...#..-",
		"-.#...#..-",
		"-........-",
	}

	for r := range want {
		if got[r] != want[r] {
			t.Errorf("row %d: %s, want %s", r, got[r], want[r])
		}
	}
}

// TestDrawClips: text past the picture's edges draws only what fits.
func TestDrawClips(t *testing.T) {
	t.Parallel()

	// "HI" starting 3 pixels left and 4 rows above: the H's bottom right
	// (its rows 4–7, its right column at x = 2) shows; the I's cell starts
	// at x = 5 with its blank column, and its lit pixels fall outside.
	got := render(6, 4, -3, -4, "HI")
	want := []string{
		"..#...",
		"..#...",
		"..#...",
		"......",
	}

	for r := range want {
		if got[r] != want[r] {
			t.Errorf("row %d: %s, want %s", r, got[r], want[r])
		}
	}
}

func TestUnknownCharacters(t *testing.T) {
	t.Parallel()

	if *glyph('é') != *glyph('?') || *glyph('\n') != *glyph('?') {
		t.Error("characters outside the font should draw as '?'")
	}

	if w := Width("PASS é"); w != 6*CellWidth {
		t.Errorf("Width = %d, want %d (6 characters)", w, 6*CellWidth)
	}
}
