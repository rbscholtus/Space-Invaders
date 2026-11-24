package synth

// opAmpMixer is an inverting op-amp summer (MAME's dst_mixer with
// DISC_MIXER_IS_OP_AMP and no rI; disc_mth.hxx; D. Renaud, 2004): each
// input passes through its resistor (and capacitor, if any), the sum of
// the currents times rF is the output, which is then high-pass filtered by
// cAmp into an assumed 100 kΩ amplifier input.
type opAmpMixer struct {
	r       []float64
	expIn   []float64 // input high-pass exponents; 0 without a capacitor
	vCapIn  []float64
	rF      float64
	expAmp  float64
	vCapAmp float64
	gain    float64
}

// ampInputR is MAME's assumed final-stage impedance for cAmp.
const ampInputR = 100 * kilo

func newOpAmpMixer(r, c []float64, rF, cAmp, gain, dt float64) *opAmpMixer {
	m := &opAmpMixer{
		r: r, expIn: make([]float64, len(r)), vCapIn: make([]float64, len(r)),
		rF: rF, gain: gain,
	}

	for i := range r {
		if c[i] != 0 {
			m.expIn[i] = chargeExp(r[i]*c[i], dt)
		}
	}

	if cAmp != 0 {
		m.expAmp = chargeExp(ampInputR*cAmp, dt)
	}

	return m
}

// step mixes the inputs (one per resistor).
func (m *opAmpMixer) step(in []float64) float64 {
	var i float64

	for k, v := range in {
		if m.expIn[k] != 0 {
			m.vCapIn[k] += (v - m.vCapIn[k]) * m.expIn[k]
			v -= m.vCapIn[k]
		}

		i += (0 - v) / m.r[k] // vRef = 0: inverting
	}

	v := i * m.rF

	if m.expAmp != 0 {
		m.vCapAmp += (v - m.vCapAmp) * m.expAmp
		v -= m.vCapAmp
	}

	return v * m.gain
}
