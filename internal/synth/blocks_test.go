package synth

import (
	"math"
	"testing"
)

const (
	rate = 48000
	dt   = 1.0 / rate
)

// frequency measures out's frequency in Hz from the times of its first and
// last rising crossings of the midpoint, interpolated between samples. The
// first 100 ms are skipped, while capacitors charge from 0 V.
func frequency(out []float64) float64 {
	out = out[rate/10:]

	lo, hi := math.Inf(1), math.Inf(-1)
	for _, v := range out {
		lo, hi = min(lo, v), max(hi, v)
	}

	mid := (lo + hi) / 2
	first, last, n := -1.0, -1.0, 0

	for i := 1; i < len(out); i++ {
		if out[i-1] < mid && out[i] >= mid {
			at := float64(i-1) + (mid-out[i-1])/(out[i]-out[i-1])
			if first < 0 {
				first = at
			}

			last = at
			n++
		}
	}

	if n < 2 {
		return 0
	}

	return float64(n-1) / ((last - first) / rate)
}

// risingEdges counts the rising crossings of the midpoint of out's range.
func risingEdges(out []float64) int {
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, v := range out {
		lo, hi = min(lo, v), max(hi, v)
	}

	mid := (lo + hi) / 2
	n := 0

	for i := 1; i < len(out); i++ {
		if out[i-1] < mid && out[i] >= mid {
			n++
		}
	}

	return n
}

func within(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()

	if math.Abs(got/want-1) > tolerance {
		t.Errorf("%s = %.4g, want %.4g (±%.1f%%)", name, got, want, tolerance*100)
	}
}

// TestRCFilterStepResponse: after one time constant, an RC low-pass has
// covered 1 − 1/e of a step; a CR high-pass has 1/e of it left.
func TestRCFilterStepResponse(t *testing.T) {
	t.Parallel()

	const r, c = 10 * kilo, 1 * micro // τ = 10 ms = 480 samples

	lp, hp := newRCFilter(r, c, dt), newCRFilter(r, c, dt)

	var low, high float64
	for range 480 {
		low, high = lp.step(1), hp.step(1)
	}

	within(t, "low-pass after τ", low, 1-1/math.E, 0.005)
	within(t, "high-pass after τ", high, 1/math.E, 0.005)
}

// TestNoise: the generator is a 17-bit maximal-length-style LFSR stepped
// at 7515 Hz: output 0 or 12 V, about half the time each, changing at most
// at the clock rate, and repeating exactly.
func TestNoise(t *testing.T) {
	t.Parallel()

	a, b := newNoise(dt), newNoise(dt)
	changes, high := 0, 0
	last := a.step()
	b.step()

	for range rate - 1 {
		v := a.step()
		if v != b.step() {
			t.Fatal("two noise generators differ")
		}

		if v != 0 && v != 12 {
			t.Fatalf("noise level %v, want 0 or 12", v)
		}

		if v != last {
			changes++
		}

		if v == 12 {
			high++
		}

		last = v
	}

	// A random bit changes on half the clock ticks.
	if changes < noiseClockHz*4/10 || changes > noiseClockHz*6/10 {
		t.Errorf("%d changes in 1 s, want about half the clock's %d", changes, noiseClockHz)
	}

	if high < rate*45/100 || high > rate*55/100 {
		t.Errorf("high %d of %d samples, want about half", high, rate)
	}
}

// period555 is the textbook period of a 555 astable whose capacitor
// charges towards vCharge (not vPos) between ⅓ and ⅔ of vPos.
func period555(r1, r2, c, vPos, vCharge float64) float64 {
	tHigh := (r1 + r2) * c * math.Log((vCharge-vPos/3)/(vCharge-vPos*2/3))
	tLow := r2 * c * math.Ln2

	return tHigh + tLow
}

// TestAstable555Fleet: the four fleet notes, from the resistors the note
// bits select (R126+R127, R128, R129, R130), with R131 and C29; the
// capacitor charges to 5 V − a diode drop (MAME's invaders_thump_555).
func TestAstable555Fleet(t *testing.T) {
	t.Parallel()

	for i, r1 := range []float64{40 * kilo, 68 * kilo, 82 * kilo, 100 * kilo} {
		a := newAstable555(r1, 75*kilo, 0.1*micro, 5, 5-0.6, ttlLogic1, true, false, dt)
		out := make([]float64, 2*rate)

		for k := range out {
			out[k] = a.step(false, r1)
		}

		within(t, "fleet note "+string(rune('1'+i)), frequency(out), 1/period555(r1, 75*kilo, 0.1*micro, 5, 4.4), 0.005)
	}
}

// TestAstable555Bonus: the extended-play gate (R94, R95, C34) runs at about
// 7.4 Hz, and stops in reset.
func TestAstable555Bonus(t *testing.T) {
	t.Parallel()

	a := newAstable555(100*kilo, 47*kilo, 1*micro, 5, 5, 5-1.2, false, true, dt)
	out := make([]float64, 4*rate)

	for k := range out {
		out[k] = a.step(false, 100*kilo)
	}

	within(t, "bonus gate", frequency(out), 1/period555(100*kilo, 47*kilo, 1*micro, 5, 5), 0.005)

	for range 100 {
		if v := a.step(true, 100*kilo); v != 0 {
			t.Fatalf("555 in reset outputs %v", v)
		}
	}
}

// TestAstable555Stops: with no charge resistor (no note bit), the fleet's
// 555 stops and its capacitor bleeds away.
func TestAstable555Stops(t *testing.T) {
	t.Parallel()

	a := newAstable555(40*kilo, 75*kilo, 0.1*micro, 5, 4.4, ttlLogic1, true, false, dt)
	for range rate / 10 {
		a.step(false, 40*kilo)
	}

	out := make([]float64, rate)
	for k := range out {
		out[k] = a.step(false, 0)
	}

	if n := risingEdges(out[rate/2:]); n > 0 {
		t.Errorf("%d edges with no charge resistor", n)
	}
}

// oscPeriod is the period of a linear-charge oscillator: the capacitor
// swings between the thresholds at the discharge and charge currents.
func oscPeriod(o *opAmpOsc, vMod1, vMod2 float64) float64 {
	c0, c1 := o.chargeRates(vMod1, vMod2)
	swing := o.thresholdHigh - o.thresholdLow

	return o.p.c*swing/c0 + o.p.c*swing/c1
}

// TestOscillatorFrequencies: each oscillator type runs at the frequency
// its thresholds and charge currents give, with the board's components.
func TestOscillatorFrequencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		kind         oscKind
		p            oscParams
		vMod1, vMod2 float64
	}{
		// UFO hit, IC A4 pin 5: R70, R64, R61, R63, R62, C21.
		{"UFO hit LFO", osc1Norton, oscParams{r1: 1 * mega, r2: 470 * kilo, r3: 100 * kilo, r4: 120 * kilo, r5: 1 * mega, c: 0.1 * micro, vP: 12}, 0, 0},
		// UFO hit, IC A4 pin 9: R65–R69, C22, at two modulation voltages.
		{"UFO hit VCO at 6 V", vco1Norton, oscParams{r1: 1 * mega, r2: 470 * kilo, r3: 680 * kilo, r4: 1 * mega, r5: 1 * mega, c: 470 * pico, vP: 12}, 6, 0},
		{"UFO hit VCO at 10 V", vco1Norton, oscParams{r1: 1 * mega, r2: 470 * kilo, r3: 680 * kilo, r4: 1 * mega, r5: 1 * mega, c: 470 * pico, vP: 12}, 10, 0},
		// Shot, IC C1 pin 4: vMod1 is filtered noise (about ±1.5 V), vMod2
		// the one-shot's ramp (0–11.5 V).
		{"shot VCO", vco3Norton, oscParams{r1: parallel(1*mega, 330*kilo) + 1.5*mega, r2: 1 * mega, r3: 560 * kilo, r4: 2.2 * mega, r5: 1 * mega, r6: 4.7 * mega, r7: 3.3 * mega, c: 330 * pico, vP: 12}, 1, 6},
	}
	for _, tt := range tests {
		o := newOpAmpOsc(tt.kind, oscOutCap, tt.p, dt)
		out := make([]float64, 2*rate)

		for k := range out {
			out[k] = o.step(tt.vMod1, tt.vMod2)
		}

		within(t, tt.name, frequency(out), 1/oscPeriod(o, tt.vMod1, tt.vMod2), 0.005)
	}
}

// TestOscillatorSquare: the square output is 0 or vOutHigh, high about as
// long as the capacitor charges.
func TestOscillatorSquare(t *testing.T) {
	t.Parallel()

	p := oscParams{r1: 1 * mega, r2: 470 * kilo, r3: 680 * kilo, r4: 1 * mega, r5: 1 * mega, c: 470 * pico, vP: 12}
	o := newOpAmpOsc(vco1Norton, oscOutSquare, p, dt)

	for range rate {
		if v := o.step(6, 0); v != 0 && v != 11.5 {
			t.Fatalf("square output %v, want 0 or 11.5", v)
		}
	}
}

func TestSquareWave(t *testing.T) {
	t.Parallel()

	s := newSquareWave(480, 1, 50, 0.5, dt)
	out := make([]float64, rate)
	high := 0

	for k := range out {
		out[k] = s.step()
		if out[k] == 1 {
			high++
		}
	}

	if n := risingEdges(out); n < 479 || n > 481 {
		t.Errorf("%d cycles in 1 s, want 480", n)
	}

	within(t, "duty", float64(high)/rate, 0.5, 0.01)
}

// TestOpAmpGain: a Norton op-amp with only r4 outputs the input current
// difference times r4, clamped to its rails.
func TestOpAmpGain(t *testing.T) {
	t.Parallel()

	a := newOpAmp(680*kilo, 680*kilo, 2.7*mega, 680*kilo, 0, 0, 12, dt)

	// Equal inputs: only r3's fixed current, from 11.5 V, pulls down: clamped at 0.
	if v := a.step(5, 5); v != 0 {
		t.Errorf("equal inputs: %v, want 0", v)
	}

	// 11.5 V into the + input only: (11 V / 680 k − 11.5 V / 2.7 M) × 680 k.
	want := (11/(680*kilo) - 11.5/(2.7*mega)) * 680 * kilo
	if v := a.step(0, 11.5); math.Abs(v-want) > 1e-9 {
		t.Errorf("+ input only: %v, want %v", v, want)
	}
}

// TestOpAmpIntegrator: with only a capacitor, the output ramps linearly at
// i/C, from 0 up to the rail.
func TestOpAmpIntegrator(t *testing.T) {
	t.Parallel()

	a := newOpAmp(0, 1*mega, 0, 0, 1*micro, 0, 12, dt) // 1 M, 1 µF: 10.5 µA → 10.5 V/s

	var v float64

	for range rate / 2 {
		v = a.step(0, 11)
	}

	within(t, "ramp after 0.5 s", v, 10.5/2, 0.001)

	for range rate {
		v = a.step(0, 11)
	}

	if v != 11.5 {
		t.Errorf("after 1.5 s: %v, want the rail, 11.5", v)
	}
}

// TestOneShot: a trigger held high gives one pulse whose length is set by
// c1 charging through r3 ∥ r4; when the trigger falls earlier, the
// capacitor-coupled trigger (c2, r2) pulls the input low and ends the pulse
// at once. (This is why the shot lasts as long as the game holds its bit.)
func TestOneShot(t *testing.T) {
	t.Parallel()

	pulse := func(hold int) int {
		// The shot's one-shot: R32, R30, R31, R33, R34, C12, C15.
		s := newOneShot(4.7*mega, 100*kilo, 1*mega, 1*mega, 2.2*mega, 1*micro, 470*pico, 0, 12, dt)
		for range rate / 10 {
			s.step(0)
		}

		high := 0

		for k := range 2 * rate {
			trig := 0.0
			if k < hold {
				trig = 5
			}

			if s.step(trig) > 0 {
				high++
			}
		}

		return high
	}

	// Held: the pulse ends when the - input's current, (vC1 − 0.5 V) / r3
	// plus the fixed 11.5 V / r1, exceeds the latch current 11.5 V / r5.
	// c1 charges towards 0.5 V + 11 V × r3 / (r3 + r4) with τ = (r3 ∥ r4) c1.
	const r1, r3, r4, r5, c1 = 4.7 * mega, 1 * mega, 1 * mega, 2.2 * mega, 1 * micro

	vEnd := (11.5/r5-11.5/r1)*r3 + nortonVBE
	vTarget := 11*r3/(r3+r4) + nortonVBE
	want := parallel(r3, r4) * c1 * math.Log((vTarget-0)/(vTarget-vEnd)) * rate

	within(t, "held pulse (samples)", float64(pulse(rate)), want, 0.02)

	if got := pulse(1000); got != 1000 {
		t.Errorf("pulse for a 1000-sample trigger: %d samples, want it to end with the trigger", got)
	}
}

// TestTVCA: while triggered, the VCA passes its input; released, it fades
// with c1 discharging through r6 + r7.
func TestTVCA(t *testing.T) {
	t.Parallel()

	// The player-dies VCA: R80, R79, R82, R93, R83, C26.
	v := newTVCA(2.7*mega, 680*kilo, 680*kilo, 10*kilo, 0, 680*kilo, 1*micro, 0, 12-nortonVBE, 12, trigTrg0, dt)

	var on float64
	for range rate / 2 {
		on = v.step(true, 0)
	}

	if on <= 0 {
		t.Fatalf("triggered output %v, want > 0", on)
	}

	for range rate * 3 {
		v.step(false, 0)
	}

	if off := v.step(false, 0); off > on/100 {
		t.Errorf("3 s after release: %v, want below 1%% of %v", off, on)
	}
}

// TestMixer: without capacitors, the op-amp mixer inverts and sums with
// gains rF / r.
func TestMixer(t *testing.T) {
	t.Parallel()

	m := newOpAmpMixer([]float64{100 * kilo, 50 * kilo}, []float64{0, 0}, 100*kilo, 0, 1, dt)
	if v := m.step([]float64{1, 1}); math.Abs(v-(-3)) > 1e-9 {
		t.Errorf("mix of 1 V and 1 V = %v, want -3", v)
	}
}
