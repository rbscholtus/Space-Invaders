package sound_test

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/app/sound"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

// TestEventsLandAtTheirSample: a write halfway through the frame's cycles
// takes effect halfway through its samples.
func TestEventsLandAtTheirSample(t *testing.T) {
	t.Parallel()

	f := sound.NewFrames()
	f.Render(nil) // a quiet frame first

	half := invaders.CyclesPerFrame / 2
	out := f.Render([]invaders.SoundEvent{{Cycle: half, Port: 3, Value: 0x24}}) // amplifier on, player dies

	if len(out) != sound.SamplesPerFrame {
		t.Fatalf("%d samples, want %d", len(out), sound.SamplesPerFrame)
	}

	for i, v := range out[:sound.SamplesPerFrame/2] {
		if v != 0 {
			t.Fatalf("sample %d = %v before the write, want silence", i, v)
		}
	}

	loud := false
	for _, v := range out[sound.SamplesPerFrame/2:] {
		loud = loud || v != 0
	}

	if !loud {
		t.Error("no sound after the write")
	}
}

// TestLateEventsAreClamped: an event past the frame's last cycle (the
// last instruction's overshoot) lands on the last sample, not beyond.
func TestLateEventsAreClamped(t *testing.T) {
	t.Parallel()

	f := sound.NewFrames()
	out := f.Render([]invaders.SoundEvent{{Cycle: invaders.CyclesPerFrame + 10, Port: 5, Value: 1}})

	if len(out) != sound.SamplesPerFrame {
		t.Errorf("%d samples, want %d", len(out), sound.SamplesPerFrame)
	}
}

// read pulls n stereo frames and returns the left channel, checking the
// right one matches.
func read(t *testing.T, s *sound.Stream, n int) []float32 {
	t.Helper()

	p := make([]byte, n*8)
	if got, err := s.Read(p); err != nil || got != len(p) {
		t.Fatalf("Read = %d, %v; want %d, nil", got, err, len(p))
	}

	out := make([]float32, n)
	for i := range out {
		l := math.Float32frombits(binary.LittleEndian.Uint32(p[i*8:]))
		r := math.Float32frombits(binary.LittleEndian.Uint32(p[i*8+4:]))

		if l != r {
			t.Fatalf("frame %d: left %v, right %v", i, l, r)
		}

		out[i] = l
	}

	return out
}

func equal(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func ramp(n int, from float32) []float32 {
	x := make([]float32, n)
	for i := range x {
		x[i] = from + float32(i)/10000
	}

	return x
}

func TestStreamOrderAndSilence(t *testing.T) {
	t.Parallel()

	s := sound.NewStream(1)
	frame := ramp(sound.SamplesPerFrame, 0.1)
	s.Push(frame)
	s.Push(frame)

	got := read(t, s, 2*sound.SamplesPerFrame+100)
	if got[0] != frame[0] || got[sound.SamplesPerFrame-1] != frame[sound.SamplesPerFrame-1] {
		t.Errorf("first frame came out as %v ... %v", got[0], got[sound.SamplesPerFrame-1])
	}

	for i, v := range got[2*sound.SamplesPerFrame:] {
		if v != 0 {
			t.Fatalf("sample %d after the buffer ran out = %v, want silence", i, v)
		}
	}
}

// TestStreamLevelControl: below the low level a sample is repeated, above
// the high level one is dropped; in between, samples pass unchanged.
func TestStreamLevelControl(t *testing.T) {
	t.Parallel()

	s := sound.NewStream(1)
	s.Push(ramp(100, 0))

	if n := s.Level(); n != 101 {
		t.Errorf("level %d after 100 samples into an empty stream, want 101 (one repeated)", n)
	}

	s = sound.NewStream(1)
	for range 2 {
		s.Push(ramp(sound.SamplesPerFrame, 0))
	}

	if n := s.Level(); n != 2*sound.SamplesPerFrame {
		t.Errorf("level %d after two frames, want %d (unchanged)", n, 2*sound.SamplesPerFrame)
	}

	for range 3 {
		s.Push(ramp(sound.SamplesPerFrame, 0))
	}

	if n := s.Level(); n != 5*sound.SamplesPerFrame-1 {
		t.Errorf("level %d after five frames, want %d (one dropped)", n, 5*sound.SamplesPerFrame-1)
	}
}

// TestStreamOverflow: a full ring makes room by dropping the oldest.
func TestStreamOverflow(t *testing.T) {
	t.Parallel()

	s := sound.NewStream(1)
	for range 40 {
		s.Push(ramp(sound.SamplesPerFrame, 0))
	}

	if n := s.Level(); n != 16*sound.SamplesPerFrame {
		t.Errorf("level %d, want the ring's %d", n, 16*sound.SamplesPerFrame)
	}
}

func TestStreamVolumeAndMute(t *testing.T) {
	t.Parallel()

	s := sound.NewStream(0.5)
	s.Push([]float32{0.4, 0.4, 3, -3}) // past full scale: clipped after the volume

	got := read(t, s, 6) // the last sample is repeated (low level), then silence
	if want := []float32{0.2, 0.2, 1, -1, -1, 0}; !equal(got, want) {
		t.Errorf("samples at half volume %v, want %v", got, want)
	}

	s.SetVolume(7)

	if v := s.Volume(); v != 1 {
		t.Errorf("volume %v after SetVolume(7), want 1", v)
	}

	s.SetMuted(true)
	s.Push(ramp(sound.SamplesPerFrame, 0.5))

	if !s.Muted() {
		t.Error("not muted")
	}

	for _, v := range read(t, s, 10) {
		if v != 0 {
			t.Fatalf("muted sample %v", v)
		}
	}
}
