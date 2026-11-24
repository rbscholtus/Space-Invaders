package synth

// The Norton op-amp circuits of the sound board, ported from MAME's
// disc_mth.hxx (dst_op_amp, dst_op_amp_1sht, dst_tvca_op_amp; D. Renaud,
// 2004–2007). A Norton op-amp (an LM3900) works with input currents, not
// voltages: each input sits at nortonVBE, and the output follows the
// difference of the currents flowing into the two inputs.

// opAmp is a Norton op-amp stage (MAME's dst_op_amp, DISC_OP_AMP_IS_NORTON):
// in0 feeds the inverting input through r1, in1 the non-inverting input
// through r2, r3 adds a fixed current from vP; r4 and c set the output (an
// RC charge, a linear integrator, a gain, or a comparator).
type opAmp struct {
	r1, r2, r3, r4, c float64
	vN, vP            float64

	vMax, iFixed, exp float64
	vCap              float64
}

func newOpAmp(r1, r2, r3, r4, c, vN, vP, dt float64) *opAmp {
	a := &opAmp{r1: r1, r2: r2, r3: r3, r4: r4, c: c, vN: vN, vP: vP, vMax: vP - nortonVBE}

	if c > 0 {
		if r4 > 0 {
			a.exp = chargeExp(r4*c, dt) // exponential charge
		} else {
			a.exp = c / dt // linear charge: sample rate × c
		}
	}

	if r3 > 0 {
		a.iFixed = (vP - nortonVBE) / r3
	}

	return a
}

// step returns the output for inputs in0 (through r1, if any) and in1
// (through r2).
func (a *opAmp) step(in0, in1 float64) float64 {
	var iNeg float64

	if a.r1 > 0 {
		iNeg = max((in0-nortonVBE)/a.r1, 0)
	}

	iNeg += a.iFixed
	iPos := max((in1-nortonVBE)/a.r2, 0)
	i := iPos - iNeg

	var out float64

	switch {
	case a.c > 0 && a.r4 > 0:
		a.vCap += (i*a.r4 - a.vCap) * a.exp
		out = a.vCap
	case a.c > 0:
		a.vCap += i / a.exp
		out = a.vCap
	case a.r4 > 0:
		out = i * a.r4
	case i > 0: // no r4: the output swings to a rail
		out = a.vMax
	}

	out = min(max(out, a.vN), a.vMax)
	a.vCap = out

	return out
}

// oneShot is a Norton op-amp one-shot (MAME's dst_op_amp_1sht,
// DISC_OP_AMP_1SHT_1): a rising edge at the trigger, through c2 and r2,
// sets the output high until c1 charges through r3 ∥ r4.
type oneShot struct {
	r2, r3, r5   float64
	vN           float64
	exp1c, exp1d float64
	exp2         float64
	iFixed, vMax float64
	r34ratio     float64
	vCap1, vCap2 float64
	out          float64
}

func newOneShot(r1, r2, r3, r4, r5, c1, c2, vN, vP, dt float64) *oneShot {
	return &oneShot{
		r2: r2, r3: r3, r5: r5, vN: vN,
		exp1c:    chargeExp(parallel(r3, r4)*c1, dt),
		exp1d:    chargeExp(r4*c1, dt),
		exp2:     chargeExp(r2*c2, dt),
		iFixed:   (vP - nortonVBE) / r1,
		vMax:     vP - nortonVBE,
		r34ratio: r3 / (r3 + r4),
	}
}

func (s *oneShot) step(trigger float64) float64 {
	// The trigger circuit.
	iPos := (trigger-s.vCap2)/s.r2 + s.out/s.r5
	s.vCap2 += (trigger - s.vCap2) * s.exp2

	iNeg := max((s.vCap1-nortonVBE)/s.r3, 0) + s.iFixed

	if iPos > iNeg {
		s.out = s.vMax
	} else {
		s.out = s.vN
	}

	// Update c1.
	v := s.out + diodeDrop // the diode's anode, roughly, when discharging

	switch {
	case s.vCap1 > s.out && s.vCap1 > v:
		s.vCap1 = v // immediate discharge through the diode
	case s.vCap1 > s.out:
		s.vCap1 += (s.out - s.vCap1) * s.exp1d // discharge through r4
	default:
		s.vCap1 += ((s.out-nortonVBE)*s.r34ratio + nortonVBE - s.vCap1) * s.exp1c
	}

	return s.out
}

// triggerFunc is how a triggered VCA's trigger input switches part of the
// circuit (MAME's DISC_OP_AMP_TRIGGER_FUNCTION_*). Only the functions the
// board uses are ported.
type triggerFunc int

const (
	trigNone triggerFunc = iota // not used: the circuit is open (always 1)
	trigTrg0                    // grounded when trigger 0 is 0
)

func (f triggerFunc) eval(trig0 bool) bool {
	if f == trigTrg0 {
		return trig0
	}

	return true
}

// tvca is a triggered op-amp VCA (MAME's dst_tvca_op_amp): while the
// trigger holds, c1 charges and lets the input through; when it drops, c1
// discharges and the sound fades. Ported for the configuration the board
// uses: input 0 only, no c2/c3 stages (r9 = r11 = 0), f0, f1 and f3 unused.
type tvca struct {
	r2, r4       float64
	f2           triggerFunc
	r67, vMax    float64
	iFixed       float64
	vTrig        [2]float64
	expC, expD   [2]float64
	exp4         float64
	hasC4        bool
	vCap1, vCap4 float64
}

func newTVCA(r1, r2, r4, r5, r6, r7, c1, c4, v1, vP float64, f2 triggerFunc, dt float64) *tvca {
	v := &tvca{r2: r2, r4: r4, f2: f2, r67: r6 + r7, vMax: vP - nortonVBE}
	v.iFixed = v.vMax / r1
	v.vTrig[0] = (v1 - diodeDrop) * divider(r5, r6)
	v.vTrig[1] = (v1-diodeDrop-nortonVBE)*divider(r5, v.r67) + nortonVBE
	v.expC[0] = chargeExp(parallel(r5, r6)*c1, dt)
	v.expC[1] = chargeExp(parallel(r5, v.r67)*c1, dt)
	v.expD[1] = chargeExp(v.r67*c1, dt)

	if r6 != 0 {
		v.expD[0] = chargeExp(r6*c1, dt)
	}

	if r4 != 0 && c4 != 0 {
		v.hasC4 = true
		v.exp4 = chargeExp(r4*c4, dt)
	}

	v.step(false, 0) // MAME's reset steps once

	return v
}

// step returns the output for trigger 0 and input 0.
func (v *tvca) step(trig0 bool, in0 float64) float64 {
	const f3 = 1 // trigNone: F3 is out of circuit

	var i2 float64
	if v.r2 != 0 { // f0 is trigNone
		i2 = max((in0-nortonVBE)/v.r2, 0)
	}

	iNeg := v.iFixed + i2

	if v.f2.eval(trig0) {
		v.vCap1 += (v.vTrig[f3] - v.vCap1) * v.expC[f3]
	} else {
		v.vCap1 += (nortonVBE - v.vCap1) * v.expD[f3] // to nortonVBE through r6 + r7
	}

	iPos := max((v.vCap1-nortonVBE)/v.r67, 0)
	iOut := max(iPos-iNeg, 0)

	var out float64

	if v.hasC4 {
		v.vCap4 += (iOut*v.r4 - v.vCap4) * v.exp4
		v.vCap4 = max(v.vCap4, 0)
		out = v.vCap4
	} else {
		out = iOut * v.r4
	}

	return min(out, v.vMax)
}
