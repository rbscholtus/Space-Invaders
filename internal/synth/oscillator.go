package synth

// The op-amp oscillators, ported from MAME's dss_op_amp_osc (disc_wav.hxx;
// D. Renaud, 2004). A Schmitt trigger flips when the capacitor reaches its
// upper or lower threshold; the capacitor charges and discharges with
// constant currents (linear charge), and a flip within a sample is handled
// by carrying the overshoot time, so the frequency is exact.

// oscKind is the circuit, MAME's DISC_OP_AMP_OSCILLATOR_* | DISC_OP_AMP_IS_NORTON.
type oscKind int

const (
	osc1Norton oscKind = iota // OSCILLATOR_1: fixed frequency
	vco1Norton                // VCO_1: frequency set by vMod1
	vco3Norton                // VCO_3: frequency set by vMod1 and vMod2
)

// oscOutput is what the oscillator outputs (DISC_OP_AMP_OSCILLATOR_OUT_*).
type oscOutput int

const (
	oscOutCap    oscOutput = iota // the capacitor voltage (a triangle)
	oscOutSquare                  // the Schmitt output: 0 or vOutHigh
)

// oscParams are the components (MAME's discrete_op_amp_osc_info).
type oscParams struct {
	r1, r2, r3, r4, r5, r6, r7, r8 float64
	c, vP                          float64
}

type opAmpOsc struct {
	kind   oscKind
	output oscOutput
	p      oscParams
	dt     float64

	vOutHigh                    float64
	thresholdLow, thresholdHigh float64
	flipFlopXOR                 bool // the cap charges while flipFlop != this
	flipFlop                    bool
	iFixed, rTotal              float64
	temp1                       float64
	vCap                        float64
}

func newOpAmpOsc(kind oscKind, output oscOutput, p oscParams, dt float64) *opAmpOsc {
	o := &opAmpOsc{kind: kind, output: output, p: p, dt: dt, vOutHigh: p.vP - nortonVBE}

	switch kind {
	case osc1Norton:
		// Charges while the flip-flop is high; with no charge on the cap,
		// the Schmitt inverter starts high. (No enable resistor r6 here.)
		o.flipFlop = true
		// The thresholds; MAME works them out each step because the
		// resistors may be nodes. Here they are constants.
		i1 := o.vOutHigh / p.r5
		o.thresholdLow = (i1+(0-nortonVBE)/p.r4)*p.r3 + nortonVBE
		o.thresholdHigh = (i1+(o.vOutHigh-nortonVBE)/p.r4)*p.r3 + nortonVBE
	case vco1Norton:
		// Charges while the flip-flop is low; it starts low.
		o.flipFlopXOR = true
		if p.r6 != 0 {
			o.iFixed += p.vP / p.r6
		}

		o.iFixed += nortonVBE/p.r1 + nortonVBE/p.r2

		if p.r7 != 0 { // the Millman input resistance, used only with r7
			total := 1/p.r1 + 1/p.r2 + 1/p.r7
			if p.r6 != 0 {
				total += 1 / p.r6
			}

			if p.r8 != 0 {
				total += 1 / p.r8
			}

			o.rTotal = 1 / total
		}

		i1 := (p.vP - nortonVBE) / p.r5
		o.thresholdLow = (i1-(p.vP-2*nortonVBE)/p.r4)*p.r3 + nortonVBE
		o.thresholdHigh = (i1-(0-nortonVBE)/p.r4)*p.r3 + nortonVBE
	case vco3Norton:
		o.flipFlop = true // charges while high; starts high (no enable r8 here)
		if p.r7 != 0 {
			o.iFixed = (p.vP - nortonVBE) / p.r7
		}

		o.temp1 = (p.vP - 2*nortonVBE) / p.r2
		i1 := (p.vP - nortonVBE) / p.r5
		o.thresholdLow = (i1+(0-nortonVBE)/p.r4)*p.r3 + nortonVBE
		o.thresholdHigh = (i1+(p.vP-2*nortonVBE)/p.r4)*p.r3 + nortonVBE
	}

	o.step(0, 0) // MAME's reset steps once

	return o
}

// step advances one sample with modulation voltages vMod1 and vMod2 (used
// by the VCOs) and returns the output. (MAME's oscillators can also be
// forced to keep charging, which stops them; none of the board's are.)
func (o *opAmpOsc) step(vMod1, vMod2 float64) float64 {
	charge0, charge1 := o.chargeRates(vMod1, vMod2)

	p := &o.p
	dt := o.dt
	vCap := o.vCap
	flipFlop := o.flipFlop
	flips := 0

	// Use up the whole sample, flipping as often as the thresholds are hit;
	// dt becomes the time left after a flip (the overshoot).
	for dt > 0 {
		if flipFlop != o.flipFlopXOR {
			vCap += charge1 * dt / p.c // charging
			dt = 0

			if vCap > o.thresholdHigh {
				flipFlop = o.flipFlopXOR
				flips++
				dt = p.c * (vCap - o.thresholdHigh) / charge1
				vCap = o.thresholdHigh
			}

			continue
		}

		vCap -= charge0 * dt / p.c // discharging
		dt = 0

		if vCap < o.thresholdLow {
			flipFlop = !o.flipFlopXOR
			flips++
			dt = p.c * (o.thresholdLow - vCap) / charge0
			vCap = o.thresholdLow
		}
	}

	o.vCap = min(max(vCap, 0), o.vOutHigh)

	var out float64

	switch o.output {
	case oscOutCap:
		out = o.vCap
	case oscOutSquare:
		switch {
		case flips >= 2 && !o.flipFlop: // at least one toggle shows
			out = o.vOutHigh
		case flips < 2 && flipFlop:
			out = o.vOutHigh
		}
	}

	o.flipFlop = flipFlop

	return out
}

// chargeRates returns the discharge and charge currents for the
// modulation voltages.
func (o *opAmpOsc) chargeRates(vMod1, vMod2 float64) (float64, float64) {
	p := &o.p

	switch o.kind {
	case osc1Norton:
		c0 := o.vOutHigh / p.r1 // the inputs sit at vOutHigh, as MAME notes

		return c0, (o.vOutHigh-nortonVBE)/p.r2 - c0
	case vco1Norton:
		v := vMod1
		if p.r7 != 0 { // mix the bias and modulation voltages (Millman)
			i := o.iFixed + vMod1/p.r7
			if p.r8 != 0 {
				i += vMod2 / p.r8
			}

			v = i * o.rTotal
		}

		v -= nortonVBE
		c0 := v / p.r1

		return c0, v/p.r2 - c0 // the enable is always on here
	case vco3Norton:
		c0 := o.iFixed + max(vMod1-nortonVBE, 0)/p.r1
		if p.r6 != 0 {
			c0 += (vMod2 - nortonVBE) / p.r6
		}

		return c0, o.temp1 - c0
	}

	return 0, 0
}

// squareWave is a fixed-frequency square wave (MAME's dss_squarewfix,
// disc_wav.hxx): 0 or amp around bias, here 0 or 1 at 50% duty.
type squareWave struct {
	tLeft, tOn, tOff, dt float64
	high                 bool
	amp, bias            float64
}

func newSquareWave(hz, amp, duty, bias, dt float64) *squareWave {
	period := 1 / hz
	s := &squareWave{dt: dt, high: true, amp: amp, bias: bias, tOn: period * duty / 100}
	s.tOff = period - s.tOn

	// MAME's reset: start phase 0, toggle until time is left, then step.
	for s.tLeft <= 0 {
		s.high = !s.high
		if s.high {
			s.tLeft += s.tOn
		} else {
			s.tLeft += s.tOff
		}
	}

	s.step()

	return s
}

func (s *squareWave) step() float64 {
	s.tLeft -= s.dt

	for s.tLeft <= 0 {
		s.high = !s.high
		if s.high {
			s.tLeft += s.tOn
		} else {
			s.tLeft += s.tOff
		}
	}

	if s.high {
		return s.amp/2 + s.bias
	}

	return -s.amp/2 + s.bias
}
