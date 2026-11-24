package synth

// rcFilter is an RC low-pass filter to ground (MAME's dst_rcfilter, fast
// path: constant R and C, reference 0 V; disc_flt.hxx).
type rcFilter struct {
	exp float64
	out float64
}

func newRCFilter(r, c, dt float64) rcFilter { return rcFilter{exp: chargeExp(r*c, dt)} }

func (f *rcFilter) step(in float64) float64 {
	f.out += (in - f.out) * f.exp

	return f.out
}

// crFilter is a CR high-pass filter (MAME's dst_crfilter, reference 0 V;
// disc_flt.hxx).
type crFilter struct {
	exp  float64
	vCap float64
}

func newCRFilter(r, c, dt float64) crFilter { return crFilter{exp: chargeExp(r*c, dt)} }

func (f *crFilter) step(in float64) float64 {
	out := in - f.vCap
	f.vCap += out * f.exp

	return out
}
