// Package sound plays the board's sound (DESIGN §14.4): it renders each
// emulated frame's samples with internal/synth, placing every sound-port
// write at its sample, and streams them to Ebitengine's audio player.
package sound

import (
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
	"github.com/rbscholtus/Space-Invaders/internal/synth"
)

// SampleRate is the audio rate; SamplesPerFrame is one Update's worth: the
// game runs one emulated frame per Update at 60 Hz (DESIGN §8).
const (
	SampleRate      = synth.SampleRate
	SamplesPerFrame = SampleRate / 60
)

// Sound ports (DESIGN §7.3).
const (
	portSound1 = 3
	portSound2 = 5
)

// Frames renders the sound board frame by frame.
type Frames struct {
	board *synth.Board
	buf   []float32
}

// NewFrames returns a sound board at power-on (muted).
func NewFrames() *Frames {
	return &Frames{board: synth.New(SampleRate), buf: make([]float32, SamplesPerFrame)}
}

// Render renders one emulated frame's samples, applying each of the frame's
// sound events at the sample its cycle falls on: cycle × 800 / 33,536
// (DESIGN §14.3). The returned slice is reused by the next call.
func (f *Frames) Render(events []invaders.SoundEvent) []float32 {
	pos := 0

	for _, e := range events {
		at := min(max(e.Cycle*SamplesPerFrame/invaders.CyclesPerFrame, pos), SamplesPerFrame)
		f.board.Render(f.buf[pos:at])
		pos = at

		switch e.Port {
		case portSound1:
			f.board.WritePort3(e.Value)
		case portSound2:
			f.board.WritePort5(e.Value)
		}
	}

	f.board.Render(f.buf[pos:])

	return f.buf
}
