package synth

import "math"

// astable555 is a 555 timer in astable mode, ported from MAME's
// dsd_555_astbl (disc_dev.hxx; D. Renaud, 2004): the capacitor charges
// through r1 + r2 towards vCharge and discharges through r2, between ⅓ and
// ⅔ of vPos. R1 may change (the fleet's resistors are switched by the note
// bits); with r1 = 0 the oscillator stops and the capacitor bleeds away.
type astable555 struct {
	r1, r2, c float64
	vCharge   float64
	vOutHigh  float64
	threshold float64
	trigger   float64
	energy    bool // DISC_555_OUT_ENERGY rather than DISC_555_OUT_SQW
	dt        float64

	tRCBleed, tRCCharge, tRCDischarge float64
	expBleed, expCharge, expDischarge float64

	flipFlop bool
	vCap     float64
}

// bleedR is the 555's leakage resistance (DEFAULT_555_BLEED_R).
const bleedR = 10 * mega

// newAstable555 returns a 555; inReset says whether its reset input is
// active at power-on, as MAME's reset steps once with it.
func newAstable555(r1, r2, c, vPos, vCharge, vOutHigh float64, energy, inReset bool, dt float64) *astable555 {
	a := &astable555{
		r2: r2, c: c, vCharge: vCharge, vOutHigh: vOutHigh, energy: energy, dt: dt,
		threshold: vPos * 2 / 3, trigger: vPos / 3, flipFlop: true,
		r1: math.NaN(), // forces setR1 to work out the constants
	}
	a.setR1(r1)
	a.step(inReset, r1)

	return a
}

// setR1 updates the charge constants when r1 changes.
func (a *astable555) setR1(r1 float64) {
	if r1 == a.r1 {
		return
	}

	a.r1 = r1
	a.tRCBleed = bleedR * a.c
	a.tRCCharge = (r1 + a.r2) * a.c
	a.tRCDischarge = a.r2 * a.c
	a.expBleed = chargeExp(a.tRCBleed, a.dt)
	a.expCharge = chargeExp(a.tRCCharge, a.dt)
	a.expDischarge = chargeExp(a.tRCDischarge, a.dt)
}

// step advances one sample. In reset, the output is 0 and the capacitor
// empty; r1 is the current charge resistor.
func (a *astable555) step(reset bool, r1 float64) float64 {
	if reset {
		a.flipFlop = true
		a.vCap = 0

		return 0
	}

	a.setR1(r1)

	dt := a.dt
	vCap := a.vCap
	flipFlop := a.flipFlop
	updateExp := false
	flips := 0
	xTime := 0.0

	// Use up the whole sample; dt becomes the time left after a flip.
	for dt > 0 {
		var flipped bool

		if flipFlop {
			vCap, dt, flipped = a.charge(vCap, dt, updateExp)
		} else {
			vCap, dt, flipped = a.discharge(vCap, dt, updateExp)
		}

		if flipped {
			flipFlop = !flipFlop
			flips++
			xTime = dt
			updateExp = true
		}
	}

	a.vCap = vCap
	out := a.output(flipFlop, flips, xTime/a.dt)
	a.flipFlop = flipFlop

	return out
}

// charge charges the capacitor for dt (bleeding it without r1) and returns
// its voltage, the time left if it reached the threshold, and whether it did.
// updateExp is set after a flip, when dt is not a whole sample.
func (a *astable555) charge(vCap, dt float64, updateExp bool) (float64, float64, bool) {
	if a.r1 == 0 { // no charge resistor: the cap bleeds
		exp := a.expBleed
		if updateExp {
			exp = chargeExp(a.tRCBleed, dt)
		}

		return vCap - vCap*exp, 0, false
	}

	exp := a.expCharge
	if updateExp {
		exp = chargeExp(a.tRCCharge, dt)
	}

	next := vCap + (a.vCharge-vCap)*exp
	if next < a.threshold {
		return next, 0, false
	}

	left := a.tRCCharge * math.Log(1/(1-(next-a.threshold)/(a.vCharge-vCap)))

	return a.threshold, left, true
}

// discharge discharges the capacitor through r2, like charge.
func (a *astable555) discharge(vCap, dt float64, updateExp bool) (float64, float64, bool) {
	exp := a.expDischarge
	if updateExp {
		exp = chargeExp(a.tRCDischarge, dt)
	}

	next := vCap - vCap*exp
	if next > a.trigger {
		return next, 0, false
	}

	left := 0.0
	if next < a.trigger {
		left = a.tRCDischarge * math.Log(1/(1-(a.trigger-next)/vCap))
	}

	return a.trigger, left, true
}

// output is the 555's output after a step: with energy output, the fraction
// of the sample spent high (against aliasing); otherwise a square wave in
// which at least one toggle shows.
func (a *astable555) output(flipFlop bool, flips int, xTime float64) float64 {
	if a.energy {
		if xTime == 0 {
			xTime = 1
		}

		if flipFlop {
			return a.vOutHigh * xTime
		}

		return a.vOutHigh * (1 - xTime)
	}

	if flips >= 2 {
		if !a.flipFlop {
			return a.vOutHigh
		}

		return 0
	}

	if flipFlop {
		return a.vOutHigh
	}

	return 0
}
