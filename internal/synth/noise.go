package synth

// noise is the board's noise generator: a 17-bit shift register clocked at
// 7515 Hz (breadboarded and measured, per MAME), output bit 12, swinging
// 0–12 V. Ported from MAME's dss_lfsr_noise (disc_wav.hxx) with the
// midway_lfsr description (mw8080bw_a.cpp): feedback bit 4 XOR bit 16,
// reset value 0x1FFFF (an RC network presets all bits at power-up). Bit 17
// holds the feedback for the next shift, as in MAME, so the noise sequence
// is MAME's.
type noise struct {
	reg         uint32
	tLeft, tClk float64 // time to the next clock, the clock period
	dt          float64
	out         float64
	amp, bias   float64
	bitLength   uint
	bit0, bit1  uint // feedback taps
	outputBit   uint
}

const (
	noiseClockHz = 7515
	noiseAmp     = 12     // peak to peak
	noiseBias    = 12 / 2 // so the output swings 0–12 V
)

func newNoise(dt float64) *noise {
	n := &noise{
		tClk: 1.0 / noiseClockHz, dt: dt, amp: noiseAmp, bias: noiseBias,
		bitLength: 17, bit0: 4, bit1: 16, outputBit: 12,
	}
	n.reset()

	return n
}

func (n *noise) reset() {
	n.tLeft = 0
	n.reg = 0x1FFFF
	n.reg |= n.feedback() << n.bitLength // DISC_LFSR_REPLACE with the feedback bit
	n.out = n.level()
}

// feedback is the XOR of the two taps.
func (n *noise) feedback() uint32 {
	return (n.reg>>n.bit0)&1 ^ (n.reg>>n.bit1)&1
}

func (n *noise) level() float64 {
	if (n.reg>>n.outputBit)&1 != 0 {
		return n.amp/2 + n.bias
	}

	return -n.amp/2 + n.bias
}

// step advances one sample and returns the noise voltage.
func (n *noise) step() float64 {
	cycles := (n.tLeft + n.dt) / n.tClk
	inc := int(cycles)
	n.tLeft = (cycles - float64(inc)) * n.tClk

	mask := uint32(1)<<n.bitLength - 1

	for range inc {
		// MAME masks both operands of each feedback function: the shift
		// drops bit 17, then the new feedback goes into it.
		fb := (n.reg >> n.bitLength) & 1
		n.reg = (n.reg<<1)&mask | fb
		n.reg |= n.feedback() << n.bitLength
	}

	if inc > 0 {
		n.out = n.level()
	}

	return n.out
}
