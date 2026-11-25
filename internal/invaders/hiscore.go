package invaders

import (
	"errors"
	"fmt"
)

// The HI-SCORE (DESIGN §7.6): four BCD digits in work RAM, the low two
// digits first. At power-on the game copies 256 bytes of ROM, 0x1B00–0x1BFF,
// to RAM 0x2000–0x20FF (the call at 0x18D9), so the HI-SCORE starts as the
// ROM bytes at 0x1BF4–0x1BF5, which are 0. At game over it copies a higher
// score there (0x166D–0x1695).
const (
	hiScoreAddr    = 0x20F4
	hiScorePowerOn = 0x1BF4 // read only by the power-on copy
	maxHiScore     = 9999
)

// ErrHiScore is returned for a HI-SCORE the game can't show.
var ErrHiScore = errors.New("invaders: the HI-SCORE must be 0–9999")

// HiScore returns the game's HI-SCORE.
func (b *Board) HiScore() int {
	lo, hi := b.mem[hiScoreAddr], b.mem[hiScoreAddr+1]

	return int(hi>>4)*1000 + int(hi&0x0F)*100 + int(lo>>4)*10 + int(lo&0x0F)
}

// SetHiScore sets the HI-SCORE, as if the board had kept it through a
// power cut: in RAM, and as its power-on value, so the game shows it when
// it starts and keeps it after a Reset. Call it before the first frame to
// have the start-up screen show it.
func (b *Board) SetHiScore(score int) error {
	if score < 0 || score > maxHiScore {
		return fmt.Errorf("%w, not %d", ErrHiScore, score)
	}

	lo := byte(score%10 | score/10%10<<4)
	hi := byte(score/100%10 | score/1000<<4)

	b.mem[hiScoreAddr], b.mem[hiScoreAddr+1] = lo, hi
	b.mem[hiScorePowerOn], b.mem[hiScorePowerOn+1] = lo, hi

	return nil
}
