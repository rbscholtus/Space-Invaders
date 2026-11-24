package synth_test

import (
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/synth"
)

func peak(x []float32) float32 {
	var p float32
	for _, v := range x {
		p = max(p, v, -v)
	}

	return p
}

// TestStartsMuted: at power-on port 3 is 0, so the amplifier is off and a
// sound played without it is silent (DESIGN §14.3).
func TestStartsMuted(t *testing.T) {
	t.Parallel()

	b := synth.New(synth.SampleRate)
	b.WritePort3(0x04) // player dies, amplifier off

	out := make([]float32, synth.SampleRate)
	b.Render(out)

	if p := peak(out); p != 0 {
		t.Errorf("peak %v with the amplifier off, want silence", p)
	}
}

// TestAmplifierMutes: switching the amplifier off silences a playing
// sound at once.
func TestAmplifierMutes(t *testing.T) {
	t.Parallel()

	b := synth.New(synth.SampleRate)
	b.WritePort3(0x20)

	settle := make([]float32, synth.SampleRate)
	b.Render(settle)
	b.WritePort3(0x24) // player dies

	on := make([]float32, synth.SampleRate/4)
	b.Render(on)

	if peak(on) == 0 {
		t.Fatal("no sound with the amplifier on")
	}

	b.WritePort3(0x04) // amplifier off, sound still triggered

	off := make([]float32, synth.SampleRate/4)
	b.Render(off)

	if p := peak(off); p != 0 {
		t.Errorf("peak %v after the amplifier went off, want silence", p)
	}
}
