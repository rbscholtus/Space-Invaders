package invaders

// The picture (DESIGN §10): video RAM holds 224 rows of 256 pixels, 1 bit
// per pixel, bit 0 leftmost. The monitor is turned 90° anticlockwise, so the
// player sees 224 × 256.
const (
	videoSize    = 32 * 224 // bytes of video RAM
	ScreenWidth  = 224      // as the player sees it
	ScreenHeight = 256
)

// Video returns the video RAM, 7 KB, as a view into the board's memory: it
// changes as the board runs, and must not be written.
func (b *Board) Video() []byte {
	return b.mem[VideoRAM : VideoRAM+videoSize]
}

// Screen converts video to the picture the player sees, one byte per pixel
// (1 = lit, 0 = dark), row by row, into dst, which must hold ScreenWidth ×
// ScreenHeight bytes. Byte i of video RAM holds pixels x = (i%32)*8 + bit of
// memory row i/32; on screen that is column i/32, row 255 - x (DESIGN §10).
func Screen(video, dst []byte) {
	for i, v := range video[:videoSize] {
		col := i / 32
		x := (i % 32) * 8

		for bit := range 8 {
			dst[(ScreenHeight-1-x-bit)*ScreenWidth+col] = v >> bit & 1
		}
	}
}
