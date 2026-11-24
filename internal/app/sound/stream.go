package sound

import (
	"encoding/binary"
	"math"
	"sync"
)

// Stream sizes, in samples.
const (
	ringSize = 16 * SamplesPerFrame // the most that is ever buffered
	// The buffer is kept between these levels by dropping or repeating one
	// sample per frame, which nobody hears (DESIGN §14.4).
	lowLevel  = 1 * SamplesPerFrame
	highLevel = 4 * SamplesPerFrame

	bytesPerFrame = 8 // one stereo frame of float32 samples
)

// Stream carries the rendered samples from the game loop to the audio
// player: Push adds a frame's samples (never blocking for long), Read is the
// io.Reader that Ebitengine's float32 player pulls stereo samples from.
// When it runs empty, for example while the game is paused, it plays
// silence.
type Stream struct {
	mu     sync.Mutex
	ring   [ringSize]float32
	head   int // the oldest sample
	n      int // samples buffered
	volume float32
	muted  bool
}

// NewStream returns an empty stream at volume (0–1).
func NewStream(volume float64) *Stream {
	s := &Stream{}
	s.SetVolume(volume)

	return s
}

// SetVolume sets the loudness, 0 to 1.
func (s *Stream) SetVolume(v float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.volume = float32(min(max(v, 0), 1))
}

// Volume returns the loudness, 0 to 1.
func (s *Stream) Volume() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return float64(s.volume)
}

// SetMuted mutes or unmutes the output, keeping the volume.
func (s *Stream) SetMuted(m bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.muted = m
}

// Muted reports whether the output is muted.
func (s *Stream) Muted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.muted
}

// Level returns how many samples are buffered.
func (s *Stream) Level() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.n
}

// Push adds samples. Above the high level, the last one is dropped; below
// the low level, the last one is repeated. If the ring is full, the oldest
// samples make room.
func (s *Stream) Push(samples []float32) {
	if len(samples) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	repeat := false

	switch {
	case s.n+len(samples) > highLevel:
		samples = samples[:len(samples)-1]
	case s.n+len(samples) < lowLevel:
		repeat = true
	}

	for _, v := range samples {
		s.add(v)
	}

	if repeat {
		s.add(samples[len(samples)-1])
	}
}

// Read fills p with little-endian float32 stereo frames, at the volume and
// clipped to ±1; silence where no samples are buffered. It implements
// io.Reader for audio.Context.NewPlayerF32.
func (s *Stream) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	gain := s.volume
	if s.muted {
		gain = 0
	}

	frames := len(p) / bytesPerFrame

	for i := range frames {
		var v float32

		if s.n > 0 {
			v = min(max(s.ring[s.head]*gain, -1), 1)
			s.head = (s.head + 1) % ringSize
			s.n--
		}

		bits := math.Float32bits(v)
		binary.LittleEndian.PutUint32(p[i*bytesPerFrame:], bits)
		binary.LittleEndian.PutUint32(p[i*bytesPerFrame+4:], bits)
	}

	return frames * bytesPerFrame, nil
}

// add appends one sample, dropping the oldest when full. s.mu is held.
func (s *Stream) add(v float32) {
	if s.n == ringSize {
		s.head = (s.head + 1) % ringSize
		s.n--
	}

	s.ring[(s.head+s.n)%ringSize] = v
	s.n++
}
