// Package font is the app's pixel font (DESIGN §10): 5 × 7 glyphs in 8 × 8
// cells, ASCII from space to '~', drawn into the same 1-byte-per-pixel
// pictures as the game, so the overlay colours text too. It is plain Go,
// without Ebitengine, so tests can check every pixel.
package font

import (
	"strings"
	"sync"
)

// Cell size in pixels: 28 columns by 32 rows on the 224 × 256 monitor.
const (
	CellWidth  = 8
	CellHeight = 8

	glyphWidth  = 5
	glyphHeight = 7
	glyphLeft   = 1 // the glyph's column within its cell (centred)

	first     = ' '
	last      = '~'
	numGlyphs = last - first + 1
	unknown   = '?' // drawn for characters outside the font
)

// table is the parsed font: one row of 5 bits per glyph row, bit 4
// leftmost. It is read-only after parse.
var table struct { //nolint:gochecknoglobals // the font, parsed once from glyphArt
	once   sync.Once
	glyphs [numGlyphs][glyphHeight]uint8
}

// glyph returns c's rows, parsing the font on first use.
func glyph(c rune) *[glyphHeight]uint8 {
	table.once.Do(parse)

	if c < first || c > last {
		c = unknown
	}

	return &table.glyphs[c-first]
}

// parse reads glyphArt. The art is checked by the tests; a malformed
// glyph would draw blank.
func parse() {
	lines := strings.Split(strings.TrimSpace(glyphArt), "\n")

	for i := 0; i+glyphHeight < len(lines); i += glyphHeight + 1 {
		label := lines[i]

		c := rune(label[0])

		switch label {
		case "space":
			c = ' '
		case "backtick":
			c = '`'
		}

		if c < first || c > last {
			continue
		}

		for r, row := range lines[i+1 : i+1+glyphHeight] {
			var bits uint8

			for x := range min(len(row), glyphWidth) {
				if row[x] == '#' {
					bits |= 1 << (glyphWidth - 1 - x)
				}
			}

			table.glyphs[c-first][r] = bits
		}
	}
}

// Width returns the width of text in pixels.
func Width(text string) int { return len([]rune(text)) * CellWidth }

// Draw draws text into dst, a picture width pixels wide with one byte per
// pixel, with the first cell's top-left corner at (x, y). Lit pixels are
// set to 1, the rest of each cell to 0. Characters outside the font draw as
// '?'; pixels outside the picture are left out.
func Draw(dst []byte, width, x, y int, text string) {
	height := len(dst) / width

	for i, c := range []rune(text) {
		g := glyph(c)
		cx := x + i*CellWidth

		for r := range CellHeight {
			py := y + r
			if py < 0 || py >= height {
				continue
			}

			var bits uint8
			if r < glyphHeight {
				bits = g[r]
			}

			for col := range CellWidth {
				px := cx + col
				if px < 0 || px >= width {
					continue
				}

				var lit byte

				if gx := col - glyphLeft; gx >= 0 && gx < glyphWidth && bits&(1<<(glyphWidth-1-gx)) != 0 {
					lit = 1
				}

				dst[py*width+px] = lit
			}
		}
	}
}
