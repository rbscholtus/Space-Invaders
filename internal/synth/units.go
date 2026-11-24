package synth

import "math"

// Component values, as in MAME's rescap.h.
const (
	kilo  = 1e3
	mega  = 1e6
	micro = 1e-6
	pico  = 1e-12
)

// Voltages used by MAME's discrete models (discrete.h).
const (
	// nortonVBE is a Norton op-amp's input junction voltage
	// (OP_AMP_NORTON_VBE).
	nortonVBE = 0.5
	// ttlLogic1 is a TTL output's high voltage (DEFAULT_TTL_V_LOGIC_1).
	ttlLogic1 = 3.4
	// diodeDrop is a silicon diode's forward voltage.
	diodeDrop = 0.6
)

// chargeExp is the fraction of the remaining distance an RC network
// charges in dt (RC_CHARGE_EXP_DT).
func chargeExp(rc, dt float64) float64 { return 1 - math.Exp(-dt/rc) }

// parallel is two resistors in parallel (RES_2_PARALLEL).
func parallel(r1, r2 float64) float64 { return r1 * r2 / (r1 + r2) }

// divider is the ratio of a voltage divider's output, r2 to ground
// (RES_VOLTAGE_DIVIDER).
func divider(r1, r2 float64) float64 { return r2 / (r1 + r2) }
