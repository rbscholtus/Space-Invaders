package app

import (
	"math"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/app/sound"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

// near compares volumes: Stream keeps them as float32.
func near(a, b float64) bool { return math.Abs(a-b) < 1e-3 }

type memPrefs map[string]string

func (m memPrefs) Get(k string) (string, bool) { v, ok := m[k]; return v, ok }

func (m memPrefs) Set(k, v string) {
	if v == "" {
		delete(m, k)
	} else {
		m[k] = v
	}
}

func TestSoundSettingsRoundTrip(t *testing.T) {
	t.Parallel()

	p := memPrefs{}
	s := sound.NewStream(0.5)
	s.SetVolume(0.3)
	s.SetMuted(true)
	saveSound(s, p)

	if p[prefVolume] != "30" || p[prefMuted] != "1" {
		t.Fatalf("saved %v", p)
	}

	again := sound.NewStream(0.5)
	applySaved(again, p, false)

	if !near(again.Volume(), 0.3) || !again.Muted() {
		t.Errorf("restored volume %v, muted %v", again.Volume(), again.Muted())
	}

	s.SetMuted(false)
	saveSound(s, p)

	if _, ok := p[prefMuted]; ok {
		t.Error("unmuting should clear the preference")
	}
}

// TestGivenVolumeWins: -volume or ?volume= overrides the remembered one.
func TestGivenVolumeWins(t *testing.T) {
	t.Parallel()

	s := sound.NewStream(0.8)
	applySaved(s, memPrefs{prefVolume: "30"}, true)

	if !near(s.Volume(), 0.8) {
		t.Errorf("volume %v, want the given 0.8", s.Volume())
	}

	for _, bad := range []string{"loud", "-5", "300"} {
		s := sound.NewStream(0.8)
		applySaved(s, memPrefs{prefVolume: bad}, false)

		if !near(s.Volume(), 0.8) {
			t.Errorf("preference %q: volume %v, want it ignored", bad, s.Volume())
		}
	}
}

// TestHiScoreRemembered: the remembered HI-SCORE goes on a new board, a
// higher one is remembered, and unusable values are ignored.
func TestHiScoreRemembered(t *testing.T) {
	t.Parallel()

	newBoard := func() *invaders.Board {
		b, err := invaders.New(make([]byte, 0x2000), invaders.DIPs{})
		if err != nil {
			t.Fatal(err)
		}

		return b
	}

	p := memPrefs{prefHiScore: "1230"}
	b := newBoard()

	saved := restoreHiScore(b, p)
	if saved != 1230 || b.HiScore() != 1230 {
		t.Fatalf("restored %d, board has %d, want 1230", saved, b.HiScore())
	}

	if saved = saveHiScore(b, p, saved); saved != 1230 || p[prefHiScore] != "1230" {
		t.Errorf("unchanged: saved %d, prefs %v", saved, p)
	}

	if err := b.SetHiScore(1500); err != nil {
		t.Fatal(err)
	}

	if saved = saveHiScore(b, p, saved); saved != 1500 || p[prefHiScore] != "1500" {
		t.Errorf("after a higher score: saved %d, prefs %v", saved, p)
	}

	// A board whose RAM is cleared (before the game copies the HI-SCORE
	// in) doesn't wipe the remembered one.
	if saved = saveHiScore(newBoard(), p, saved); saved != 1500 || p[prefHiScore] != "1500" {
		t.Errorf("after a lower score: saved %d, prefs %v", saved, p)
	}

	for _, bad := range []string{"", "abc", "-10", "10000"} {
		b := newBoard()
		if n := restoreHiScore(b, memPrefs{prefHiScore: bad}); n != 0 || b.HiScore() != 0 {
			t.Errorf("restoring %q gave %d, board %d, want 0", bad, n, b.HiScore())
		}
	}

	if n := restoreHiScore(newBoard(), memPrefs{}); n != 0 {
		t.Errorf("with nothing remembered, restored %d", n)
	}
}
