package app

import (
	"strconv"

	"github.com/rbscholtus/Space-Invaders/internal/app/sound"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

// Preference keys (platform.Prefs).
const (
	prefVolume  = "volume"  // 0–100
	prefMuted   = "muted"   // "1" when muted
	prefHiScore = "hiscore" // the game's HI-SCORE, 0–9999
)

// prefsStore is what the settings need from platform.Prefs.
type prefsStore interface {
	Get(key string) (string, bool)
	Set(key, value string)
}

// applySaved sets the stream's volume and mute from the preferences: the
// volume only when the player didn't give one at start (-volume, ?volume=).
func applySaved(s *sound.Stream, p prefsStore, volumeGiven bool) {
	if v, ok := p.Get(prefVolume); ok && !volumeGiven {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 100 {
			s.SetVolume(float64(n) / 100)
		}
	}

	if v, _ := p.Get(prefMuted); v == "1" {
		s.SetMuted(true)
	}
}

// saveSound remembers the stream's volume and mute.
func saveSound(s *sound.Stream, p prefsStore) {
	p.Set(prefVolume, strconv.Itoa(int(s.Volume()*100+0.5)))

	muted := ""
	if s.Muted() {
		muted = "1"
	}

	p.Set(prefMuted, muted)
}

// restoreHiScore gives a new board the remembered HI-SCORE, before its
// first frame, and returns it (0 when there's none or it's unusable).
func restoreHiScore(b *invaders.Board, p prefsStore) int {
	v, ok := p.Get(prefHiScore)
	if !ok {
		return 0
	}

	n, err := strconv.Atoi(v)
	if err != nil || b.SetHiScore(n) != nil {
		return 0
	}

	return n
}

// saveHiScore remembers the board's HI-SCORE when it has gone above saved,
// the last value remembered, and returns the value now remembered. The
// game only ever raises it (at game over), so a lower value is ignored.
func saveHiScore(b *invaders.Board, p prefsStore, saved int) int {
	n := b.HiScore()
	if n <= saved {
		return saved
	}

	p.Set(prefHiScore, strconv.Itoa(n))

	return n
}
