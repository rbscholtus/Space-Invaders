package app

import (
	"strconv"

	"github.com/rbscholtus/Space-Invaders/internal/app/sound"
)

// Preference keys (platform.Prefs).
const (
	prefVolume = "volume" // 0–100
	prefMuted  = "muted"  // "1" when muted
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
