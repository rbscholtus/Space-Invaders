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

// everything starts every sound at once: the busiest the board gets.
func everything(b *synth.Board) {
	b.WritePort3(0x3F) // amplifier, UFO, shot, player dies, invader dies, extended play
	b.WritePort5(0x11) // fleet note 1, UFO hit
}

// TestRenderDoesNotAllocate: Render is in the per-frame path.
func TestRenderDoesNotAllocate(t *testing.T) { //nolint:paralleltest // AllocsPerRun panics in parallel tests
	b := synth.New(synth.SampleRate)
	everything(b)

	frame := make([]float32, synth.SampleRate/60)

	if n := testing.AllocsPerRun(10, func() { b.Render(frame) }); n != 0 {
		t.Errorf("Render allocates %.0f times per frame, want 0", n)
	}
}

// TestDeterministic: the same writes give the same samples.
func TestDeterministic(t *testing.T) {
	t.Parallel()

	render := func() []float32 {
		b := synth.New(synth.SampleRate)
		out := make([]float32, synth.SampleRate)
		b.Render(out[:1000])
		everything(b)
		b.Render(out[1000:])

		return out
	}

	a, b := render(), render()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("sample %d differs: %v and %v", i, a[i], b[i])
		}
	}
}

// BenchmarkRender renders one video frame's worth of samples (800) with
// every sound playing (DESIGN §16).
func BenchmarkRender(b *testing.B) {
	board := synth.New(synth.SampleRate)
	everything(board)

	frame := make([]float32, synth.SampleRate/60)
	for b.Loop() {
		board.Render(frame)
	}
}
