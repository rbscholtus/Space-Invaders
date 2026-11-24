package synth

// ufo is the UFO sound: an SN76477 sound chip, ported from MAME's
// sn76477_device (sn76477.cpp; Z. Vasvari, BSD-3-Clause) for the
// configuration of invaders_audio_device (mw8080bw_a.cpp). Only the parts
// that reach the output with that configuration are ported:
//
//   - VCO mode 1: the SLF (super-low-frequency oscillator, C 1 µF, R 120 k)
//     sets the VCO's (C 0.1 µF, R 8.2 k) upper voltage, so the tone sweeps;
//   - mixer = VCO only (the chip's noise doesn't reach the output);
//   - envelope = one-shot, with no one-shot parts fitted: once enabled, the
//     envelope opens and stays open; attack and decay are instant (no
//     attack/decay capacitor);
//   - amplitude R 56 k, feedback R 10 k.
//
// MAME's voltages and rate formulas were measured on a real chip.
type ufo struct {
	enabled bool // the chip's /ENABLE pin is low: port 3 bit 0 is set

	slfCap, vcoCap   float64
	slfHigh, vcoHigh bool // the SLF and VCO flip-flops
	slfUp, slfDown   float64
	vcoStep          float64
	envelopeOpen     bool // the one-shot is running (it never stops here)
	adCap            float64
	centerToPeak     float64
}

// SN76477 constants (MAME's measurements).
const (
	snSLFMin       = 0.33 // SLF triangle, bottom and top
	snSLFMax       = 2.37
	snVCOToSLF     = 0.35 // the VCO's top is the SLF's voltage plus this
	snVCOMin       = snSLFMin
	snVCOMax       = snSLFMax + snVCOToSLF
	snADMax        = 4.44 // attack/decay capacitor's top
	snOutCenter    = 2.57 // OUT with the volume at 0
	snOutHighClip  = 3.51
	snOutLowClip   = 0.715
	snSLFCap       = 1 * micro
	snSLFRes       = 120 * kilo
	snVCOCap       = 0.1 * micro
	snVCORes       = 8.2 * kilo
	snAmplitudeRes = 56 * kilo
	snFeedbackRes  = 10 * kilo
)

// The OUT gain at each 0.1 V of the attack/decay capacitor (MAME's
// out_pos_gain and out_neg_gain, measured). With instant attack the
// capacitor is at snADMax, index 44, whenever the envelope is open.
const (
	snPosGainFull = 1.00
	snNegGainFull = -0.85
)

func newUFO(dt float64) *ufo {
	rate := 1 / dt
	slfRange := snSLFMax - snSLFMin

	return &ufo{
		slfCap: snSLFMin, vcoCap: snVCOMin,
		// compute_slf_cap_charging_rate / discharging_rate, per sample.
		slfUp:   slfRange / (0.5885*snSLFRes*snSLFCap + 0.001300) / rate,
		slfDown: slfRange / (0.5413*snSLFRes*snSLFCap + 0.001343) / rate,
		// compute_vco_cap_charging_discharging_rate, 50% duty cycle (no
		// external VCO voltage).
		vcoStep:      0.64 * 2 * (snVCOMax - snVCOMin) / (snVCORes * snVCOCap) / rate,
		centerToPeak: 3.818*(snFeedbackRes/snAmplitudeRes) + 0.03,
	}
}

// enable sets the UFO bit. Its rising edge (the chip's /ENABLE falling)
// starts the envelope.
func (u *ufo) enable(on bool) {
	if on && !u.enabled {
		u.adCap = 0
		u.envelopeOpen = true
	}

	u.enabled = on
}

// step advances one sample and returns the chip's output as a stream
// sample: 0 at the centre level, −1 at the low clip.
func (u *ufo) step() float64 {
	// The SLF: a triangle between its limits.
	if u.slfHigh {
		u.slfCap = max(u.slfCap-u.slfDown, snSLFMin)
	} else {
		u.slfCap = min(u.slfCap+u.slfUp, snSLFMax)
	}

	if u.slfCap >= snSLFMax {
		u.slfHigh = true
	} else if u.slfCap <= snSLFMin {
		u.slfHigh = false
	}

	// The VCO: charges up to the SLF's voltage plus 0.35 V.
	vcoTop := u.slfCap + snVCOToSLF
	if u.vcoHigh {
		u.vcoCap = max(u.vcoCap-u.vcoStep, snVCOMin)
	} else {
		u.vcoCap = min(u.vcoCap+u.vcoStep, vcoTop)
	}

	if u.vcoCap >= vcoTop {
		u.vcoHigh = true
	} else if u.vcoCap <= snVCOMin {
		u.vcoHigh = false
	}

	// The envelope: instant attack while the one-shot runs, else instant
	// decay.
	if u.envelopeOpen {
		u.adCap = snADMax
	} else {
		u.adCap = 0
	}

	voltage := snOutCenter

	if u.enabled && u.vcoCap <= snVCOMax {
		posGain, negGain := 0.0, 0.0
		if u.adCap >= snADMax {
			posGain, negGain = snPosGainFull, snNegGainFull
		}

		if u.vcoHigh {
			voltage = min(snOutCenter+u.centerToPeak*posGain, snOutHighClip)
		} else {
			voltage = max(snOutCenter+u.centerToPeak*negGain, snOutLowClip)
		}
	}

	return (voltage-snOutLowClip)/(snOutCenter-snOutLowClip) - 1
}
