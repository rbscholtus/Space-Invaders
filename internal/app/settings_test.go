package app

import (
	"math"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/app/sound"
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
