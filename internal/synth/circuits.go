package synth

// The six discrete sound circuits of the Midway sound board (schematics
// M051-00739-A005 and B005), wired as in MAME's invaders_discrete
// (mw8080bw_a.cpp; D. Renaud, 2007). Each sample, the blocks are stepped in
// MAME's node order, each reading what the blocks before it produced in the
// same sample. Component names (R72, C23, ...) and IC pins are the
// schematic's, as MAME gives them.

// Enable-input levels: MAME's DISCRETE_INPUTX_LOGIC gains, the voltage a
// set bit puts on each circuit's input.
const (
	ufoHitInputV  = 12
	triggerInputV = 5 // shot, player dies, invader dies
	bonusToneHz   = 480
	outputGain    = 2500 // DISCRETE_OUTPUT gain
	streamScale   = 1.0 / 32768
	routeGain     = 0.5 // each of the discrete board and the SN76477 to the speaker
	missileDiodeV = 0.5 // D1's drop, in the shot's modulation path
	vcc           = 5
	opAmpSupply   = 12
)

// discrete is the six circuits and their mixer.
type discrete struct {
	// Inputs, as voltages or data, set by the port writes.
	ufoHitEN, invaderEN, explosionEN, missileEN float64
	bonusEN                                     bool
	fleetData                                   int

	noise *noise

	// UFO hit (saucer hit).
	ufoHitRamp *opAmp    // IC B3 pin 9
	ufoHitLFO  *opAmpOsc // IC A4 pin 5
	ufoHitVCO  *opAmpOsc // IC A4 pin 9
	ufoHitAmp  *opAmp    // IC B3 pin 10

	// Fleet ("thump").
	fleetR       [16]float64 // R1 of the 555 for each note value (COMP_ADDER)
	fleet555     *astable555 // IC F3 pin 6
	fleetFilter1 rcFilter    // R132, C31
	fleetFilter2 rcFilter    // R132 + R133, C32

	// Extended play (bonus missile base).
	bonus555  *astable555 // IC F4 pin 9
	bonusTone *squareWave // the 480 Hz square wave

	// Invader dies.
	invaderOneShot *oneShot  // IC D3 pin 9
	invaderRamp    *opAmp    // IC D3 pin 10
	invaderLFO     *opAmpOsc // IC B4 pin 5
	invaderVCO     *opAmpOsc // IC B4 pin 4
	invaderAmp     *opAmp    // IC D3 pin 4

	// Player dies (explosion).
	explosionOneShot *oneShot // IC D2 pin 10
	explosionVCA     *tvca    // IC D2 pin 4
	explosionFilter1 rcFilter // R84, C27
	explosionFilter2 rcFilter // R84 + R85, C28

	// Shot (missile).
	missileOneShot *oneShot  // IC B3 pin 4
	missileRamp    *opAmp    // IC B3 pin 5
	missileNoiseHP crFilter  // R29 + R11, C57
	missileVCO     *opAmpOsc // IC C1 pin 4
	missileAmp     *opAmp    // IC A3 pin 9
	missileVCA     *tvca     // IC A3 pin 10

	mixer *opAmpMixer
	mixIn []float64 // reused for the mixer's inputs
}

func newDiscrete(dt float64) *discrete {
	d := &discrete{noise: newNoise(dt), mixIn: make([]float64, 6)}
	d.initUFOHit(dt)
	d.initFleet(dt)
	d.initBonus(dt)
	d.initInvader(dt)
	d.initExplosion(dt)
	d.initMissile(dt)

	// The mixer (invaders_mixer): R78, R134 + R133 + R132, R136, R59,
	// R86 + R85 + R84, R28; C11 on the shot; rF R105; C45 to the amplifier.
	d.mixer = newOpAmpMixer(
		[]float64{200 * kilo, 10*kilo + 100 + 100, 150 * kilo, 200 * kilo, 2*kilo + 6.8*kilo + 5.6*kilo, 150 * kilo},
		[]float64{0, 0, 0, 0, 0, 0.001 * micro},
		100*kilo, 0.1*micro, 1, dt)

	return d
}

// initUFOHit builds the UFO hit (invaders_saucer_hit_*).
func (d *discrete) initUFOHit(dt float64) {
	// R72, R71, C23.
	d.ufoHitRamp = newOpAmp(opAmpParams{r2: 100 * kilo, r3: 1 * mega, c: 1 * micro, vP: opAmpSupply}, dt)
	// R70, R64, R61, R63, R62, C21.
	d.ufoHitLFO = newOpAmpOsc(osc1Norton, oscOutCap, oscParams{
		r1: 1 * mega, r2: 470 * kilo, r3: 100 * kilo, r4: 120 * kilo, r5: 1 * mega, c: 0.1 * micro, vP: opAmpSupply,
	}, dt)
	// R65, R66, R67, R69, R68, C22.
	d.ufoHitVCO = newOpAmpOsc(vco1Norton, oscOutSquare, oscParams{
		r1: 1 * mega, r2: 470 * kilo, r3: 680 * kilo, r4: 1 * mega, r5: 1 * mega, c: 470 * pico, vP: opAmpSupply,
	}, dt)
	// R73, R77, R74, R75.
	d.ufoHitAmp = newOpAmp(opAmpParams{
		r1: 680 * kilo, r2: 680 * kilo, r3: 2.7 * mega, r4: 680 * kilo, vP: opAmpSupply,
	}, dt)
}

// initFleet builds the fleet (invaders_thump_resistors, invaders_thump_555):
// the note bits switch R126 + R127, R128, R129 and R130 in parallel as the
// 555's R1; R131, C29; the capacitor charges to 5 V less a diode drop.
func (d *discrete) initFleet(dt float64) {
	notes := [4]float64{20*kilo + 20*kilo, 68 * kilo, 82 * kilo, 100 * kilo}

	for v := range d.fleetR {
		var g float64

		for bit, r := range notes {
			if v&(1<<bit) != 0 {
				g += 1 / r
			}
		}

		if g != 0 {
			d.fleetR[v] = 1 / g
		}
	}

	d.fleet555 = newAstable555(ne555Params{
		r2: 75 * kilo, c: 0.1 * micro, vPos: vcc, vCharge: vcc - diodeDrop, vOutHigh: ttlLogic1, energy: true,
	}, false, dt)
	d.fleetFilter1 = newRCFilter(100, 4.7*micro, dt)    // R132, C31
	d.fleetFilter2 = newRCFilter(100+100, 10*micro, dt) // R132 + R133, C32
}

// initBonus builds extended play (invaders_bonus_555): R94, R95, C34, the
// 555's default high output (1.2 V below its supply), gating 480 Hz.
func (d *discrete) initBonus(dt float64) {
	d.bonus555 = newAstable555(ne555Params{
		r1: 100 * kilo, r2: 47 * kilo, c: 1 * micro, vPos: vcc, vCharge: vcc, vOutHigh: vcc - 1.2,
	}, true, dt)
	d.bonusTone = newSquareWave(bonusToneHz, 1, 50, 0.5, dt)
}

// oneShotWith returns the board's one-shot with timing capacitor c1. The
// three one-shots share their resistors: R1 to the supply, R2 from the
// trigger capacitor (C2), R3 ∥ R4 to c1, R5 the latch (the explosion's R90,
// R88, R91, R89, R92 and C25; the others' match).
func oneShotWith(c1 float64) oneShotParams {
	return oneShotParams{
		r1: 4.7 * mega, r2: 100 * kilo, r3: 1 * mega, r4: 1 * mega, r5: 2.2 * mega,
		c1: c1, c2: 470 * pico, vP: opAmpSupply,
	}
}

// initInvader builds invader dies (invaders_invader_hit_*).
func (d *discrete) initInvader(dt float64) {
	d.invaderOneShot = newOneShot(oneShotWith(0.1*micro), dt) // R49, R51, R48, R50, R52, C18, C20
	// R53, R137, C19.
	d.invaderRamp = newOpAmp(opAmpParams{r2: 10 * kilo, r3: 1 * mega, c: 0.47 * micro, vP: opAmpSupply}, dt)
	// R37, R41, R38, R40, R39, C16.
	d.invaderLFO = newOpAmpOsc(osc1Norton, oscOutCap, oscParams{
		r1: 1 * mega, r2: 10 * kilo, r3: 100 * kilo, r4: 120 * kilo, r5: 1 * mega, c: 0.1 * micro, vP: opAmpSupply,
	}, dt)
	// R42, R43, R44, R46, R45, C16.
	d.invaderVCO = newOpAmpOsc(vco1Norton, oscOutCap, oscParams{
		r1: 1 * mega, r2: 470 * kilo, r3: 680 * kilo, r4: 1 * mega, r5: 1 * mega, c: 330 * pico, vP: opAmpSupply,
	}, dt)
	// R55, R54, R56, R57.
	d.invaderAmp = newOpAmp(opAmpParams{
		r1: 470 * kilo, r2: 680 * kilo, r3: 2.7 * mega, r4: 680 * kilo, vP: opAmpSupply,
	}, dt)
}

// initExplosion builds player dies (invaders_explosion_*).
func (d *discrete) initExplosion(dt float64) {
	d.explosionOneShot = newOneShot(oneShotWith(2.2*micro), dt) // C24, C25
	// R80, R79, R82, R93, R83, C26.
	d.explosionVCA = newTVCA(tvcaParams{
		r1: 2.7 * mega, r2: 680 * kilo, r4: 680 * kilo, r5: 10 * kilo, r7: 680 * kilo, c1: 1 * micro,
		v1: opAmpSupply - nortonVBE, vP: opAmpSupply, f2: trigTrg0,
	}, dt)
	d.explosionFilter1 = newRCFilter(5.6*kilo, 0.1*micro, dt)          // R84, C27
	d.explosionFilter2 = newRCFilter(5.6*kilo+6.8*kilo, 0.1*micro, dt) // R84 + R85, C28
}

// initMissile builds the shot (invaders_missile_*, Midway values; the Taito
// PCB differs).
func (d *discrete) initMissile(dt float64) {
	d.missileOneShot = newOneShot(oneShotWith(1*micro), dt) // R32, R30, R31, R33, R34, C12, C15
	// R35, R36, C13.
	d.missileRamp = newOpAmp(opAmpParams{r2: 10 * kilo, r3: 1.5 * mega, c: 0.22 * micro, vP: opAmpSupply}, dt)
	d.missileNoiseHP = newCRFilter(1*mega+330*kilo, 0.1*micro, dt) // R29, R11, C57
	// R29 ∥ R11 + R12, R16, R17, R19, R16, R14, R13, C58.
	d.missileVCO = newOpAmpOsc(vco3Norton, oscOutSquare, oscParams{
		r1: parallel(1*mega, 330*kilo) + 1.5*mega, r2: 1 * mega, r3: 560 * kilo, r4: 2.2 * mega, r5: 1 * mega,
		r6: 4.7 * mega, r7: 3.3 * mega, c: 330 * pico, vP: opAmpSupply,
	}, dt)
	// R22, R15, R20, R21.
	d.missileAmp = newOpAmp(opAmpParams{
		r1: 560 * kilo, r2: 470 * kilo, r3: 2.7 * mega, r4: 560 * kilo, vP: opAmpSupply,
	}, dt)
	// R25, R23, R26, (1 k), R60, C14.
	d.missileVCA = newTVCA(tvcaParams{
		r1: 2.7 * mega, r2: 560 * kilo, r4: 560 * kilo, r5: 1 * kilo, r7: 560 * kilo, c1: 0.1 * micro,
		v1: triggerInputV, vP: opAmpSupply, f2: trigTrg0,
	}, dt)
}

// Port 3 and 5 bits the discrete circuits listen to (DESIGN §7.3).
const (
	p3Shot      = 0x02
	p3Explosion = 0x04
	p3Invader   = 0x08
	p3Bonus     = 0x10
	p5Fleet     = 0x0F
	p5UFOHit    = 0x10
)

func level(set bool, v float64) float64 {
	if set {
		return v
	}

	return 0
}

// writePort3 sets the inputs from port 3 (MAME's invaders_audio_device::p1_w).
func (d *discrete) writePort3(v byte) {
	d.missileEN = level(v&p3Shot != 0, triggerInputV)
	d.explosionEN = level(v&p3Explosion != 0, triggerInputV)
	d.invaderEN = level(v&p3Invader != 0, triggerInputV)
	d.bonusEN = v&p3Bonus != 0
}

// writePort5 sets the inputs from port 5 (p2_w).
func (d *discrete) writePort5(v byte) {
	d.fleetData = int(v & p5Fleet)
	d.ufoHitEN = level(v&p5UFOHit != 0, ufoHitInputV)
}

// step advances one sample and returns the mixer's output voltage.
func (d *discrete) step() float64 {
	nz := d.noise.step()

	// UFO hit.
	ramp := d.ufoHitRamp.step(0, d.ufoHitEN)
	ufoHit := d.ufoHitAmp.step(d.ufoHitVCO.step(d.ufoHitLFO.step(0, 0), 0), ramp)

	// Fleet.
	fleet := d.fleetFilter2.step(d.fleetFilter1.step(d.fleet555.step(false, d.fleetR[d.fleetData])))

	// Extended play: the gate AND the tone AND the enable, at TTL level.
	gate := d.bonus555.step(!d.bonusEN, 100*kilo)
	tone := d.bonusTone.step()
	bonus := level(d.bonusEN && tone != 0 && gate != 0, ttlLogic1)

	// Invader dies.
	ramp = d.invaderRamp.step(0, d.invaderOneShot.step(d.invaderEN))
	invader := d.invaderAmp.step(d.invaderVCO.step(d.invaderLFO.step(0, 0), 0), ramp)

	// Player dies.
	explosion := d.explosionFilter2.step(d.explosionFilter1.step(
		d.explosionVCA.step(d.explosionOneShot.step(d.explosionEN) >= 1, nz)))

	// Shot: the one-shot's ramp, less D1's drop, sweeps the VCO; filtered
	// noise wobbles it.
	sweep := min(max(d.missileRamp.step(0, d.missileOneShot.step(d.missileEN))-missileDiodeV, 0), opAmpSupply)
	wobble := d.missileNoiseHP.step(nz) * (330 * kilo / (1*mega + 330*kilo))
	shot := d.missileVCA.step(d.missileEN >= 1, d.missileAmp.step(d.missileVCO.step(wobble, sweep), sweep))

	d.mixIn[0], d.mixIn[1], d.mixIn[2] = ufoHit, fleet, bonus
	d.mixIn[3], d.mixIn[4], d.mixIn[5] = invader, explosion, shot

	return d.mixer.step(d.mixIn)
}
