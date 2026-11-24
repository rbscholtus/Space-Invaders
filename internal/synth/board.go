package synth

// SampleRate is the rate the board is designed and tested at: MAME's
// default, which the reference recordings use (DESIGN §14.4).
const SampleRate = 48000

// p3Amp is port 3 bit 5, the amplifier enable (DESIGN §7.3).
const p3Amp = 0x20

// Board is the Space Invaders sound board: write the sound ports, render
// samples. It starts as at power-on, muted (port 3 = 0).
type Board struct {
	discrete *discrete
	ufo      *ufo
	muted    bool
}

// New returns a sound board rendering at sampleRate.
func New(sampleRate int) *Board {
	dt := 1 / float64(sampleRate)

	return &Board{discrete: newDiscrete(dt), ufo: newUFO(dt), muted: true}
}

// WritePort3 applies a write to port 3: the UFO, shot, player dies,
// invader dies, extended play and the amplifier enable (DESIGN §7.3).
func (b *Board) WritePort3(v byte) {
	b.discrete.writePort3(v)
	b.ufo.enable(v&0x01 != 0)
	b.muted = v&p3Amp == 0
}

// WritePort5 applies a write to port 5: the fleet notes and UFO hit.
func (b *Board) WritePort5(v byte) {
	b.discrete.writePort5(v)
}

// Render fills dst with the next samples, full scale ±1, as MAME's
// speaker output: the discrete board through its output gain, and the
// SN76477, each at half level. Loud moments can exceed ±1, as in MAME;
// the caller clips after its volume control.
func (b *Board) Render(dst []float32) {
	for i := range dst {
		v := b.discrete.step()*outputGain*streamScale*routeGain + b.ufo.step()*routeGain
		if b.muted {
			v = 0
		}

		dst[i] = float32(v)
	}
}
