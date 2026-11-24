package soundref

import (
	"math"
	"math/cmplx"
	"slices"
)

// Analysis parameters.
const (
	// WindowsPerSecond sets the hop between measurements: 10 ms.
	WindowsPerSecond = 100

	// Pitch is looked for first in a short frame (21 ms at 48 kHz), which
	// follows sweeping tones, then in a long one (50 ms), which holds two
	// periods of the lowest fleet note (46 Hz).
	shortPitchFrame = 1024
	longPitchFrame  = 2400
	centroidFrame   = 2048 // samples analysed for the spectral centroid
	maxPitchHz      = 5000
	// clarityVoiced is the normalised autocorrelation peak above which a
	// window counts as pitched (McLeod's "clarity").
	clarityVoiced = 0.8
	// keyPeakRatio picks the first autocorrelation peak within this ratio of
	// the highest (McLeod's k).
	keyPeakRatio = 0.9
	// silentRMS is the level below which a window is silence (-100 dBFS).
	silentRMS = 1e-5
	// durationFloor is the level, relative to the loudest window, that
	// counts as the end of a sound: -40 dB.
	durationFloor = 0.01
)

// Features are measurements of a sound, one value per 10 ms window.
type Features struct {
	SampleRate int       `json:"sampleRate"`
	Hop        int       `json:"hop"`      // samples per window
	RMS        []float64 `json:"rms"`      // loudness, full scale = 1
	Pitch      []float64 `json:"pitch"`    // Hz; 0 when unpitched (noise, silence)
	Centroid   []float64 `json:"centroid"` // spectral centroid in Hz; 0 when silent
}

// Analyze measures x, sampled at rate.
func Analyze(x []float32, rate int) Features {
	hop := rate / WindowsPerSecond
	n := len(x) / hop
	f := Features{
		SampleRate: rate, Hop: hop,
		RMS: make([]float64, n), Pitch: make([]float64, n), Centroid: make([]float64, n),
	}

	hann := make([]float64, centroidFrame)
	for i := range hann {
		hann[i] = 0.5 - 0.5*math.Cos(2*math.Pi*float64(i)/centroidFrame)
	}

	for w := range n {
		f.RMS[w] = rms(x[w*hop : (w+1)*hop])
		if f.RMS[w] < silentRMS {
			continue
		}

		center := w*hop + hop/2

		f.Pitch[w] = pitch(frame(x, center, shortPitchFrame), rate)
		if f.Pitch[w] == 0 {
			f.Pitch[w] = pitch(frame(x, center, longPitchFrame), rate)
		}

		f.Centroid[w] = centroid(frame(x, center, centroidFrame), hann, rate)
	}

	return f
}

func rms(x []float32) float64 {
	var sum float64
	for _, v := range x {
		sum += float64(v) * float64(v)
	}

	return math.Sqrt(sum / float64(len(x)))
}

// frame returns size samples centred on center, zero beyond the ends.
func frame(x []float32, center, size int) []float64 {
	out := make([]float64, size)

	for i := range out {
		if j := center - size/2 + i; j >= 0 && j < len(x) {
			out[i] = float64(x[j])
		}
	}

	return out
}

// pitch estimates the fundamental frequency of x with the McLeod pitch
// method: peaks of the normalised square difference function. It returns 0
// when x has no clear pitch.
func pitch(x []float64, rate int) float64 {
	nsdf := squareDifference(x)
	minLag := rate / maxPitchHz
	maxLag := len(nsdf) - 1

	// Peaks: the highest point of each positive stretch after a negative one.
	var peaks []int

	for tau := 1; tau <= maxLag; tau++ {
		if nsdf[tau] <= 0 || nsdf[tau-1] > 0 {
			continue
		}

		best := tau
		for ; tau <= maxLag && nsdf[tau] > 0; tau++ {
			if nsdf[tau] > nsdf[best] {
				best = tau
			}
		}

		if best >= minLag && best < maxLag {
			peaks = append(peaks, best)
		}
	}

	if len(peaks) == 0 {
		return 0
	}

	highest := slices.MaxFunc(peaks, func(a, b int) int { return cmpFloat(nsdf[a], nsdf[b]) })
	if nsdf[highest] < clarityVoiced {
		return 0
	}

	for _, p := range peaks {
		if nsdf[p] >= keyPeakRatio*nsdf[highest] {
			// Parabolic interpolation around the peak.
			a, b, c := nsdf[p-1], nsdf[p], nsdf[p+1]
			shift := 0.0

			if d := a - 2*b + c; d != 0 {
				shift = 0.5 * (a - c) / d
			}

			return float64(rate) / (float64(p) + shift)
		}
	}

	return 0
}

// squareDifference returns the normalised square difference function of x
// for lags up to half its length (two periods in the frame):
// nsdf(τ) = 2·r(τ) / m(τ), with the autocorrelation r computed by FFT and
// m(τ) = Σ x[j]² + x[j+τ]² updated as τ grows.
func squareDifference(x []float64) []float64 {
	n := len(x)
	size := nextPow2(2 * n)

	buf := make([]complex128, size)
	for i, v := range x {
		buf[i] = complex(v, 0)
	}

	fft(buf, false)

	for i := range buf {
		buf[i] = complex(real(buf[i])*real(buf[i])+imag(buf[i])*imag(buf[i]), 0)
	}

	fft(buf, true) // not scaled: r(τ) = real(buf[τ]) / size

	nsdf := make([]float64, n/2+1)
	m := 2 * real(buf[0]) / float64(size)

	for tau := 1; tau < len(nsdf); tau++ {
		m -= x[tau-1]*x[tau-1] + x[n-tau]*x[n-tau]
		if m > 0 {
			nsdf[tau] = 2 * real(buf[tau]) / float64(size) / m
		}
	}

	return nsdf
}

func cmpFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}

	return 0
}

// centroid returns the magnitude-weighted mean frequency of x.
func centroid(x, window []float64, rate int) float64 {
	buf := make([]complex128, len(x))
	for i, v := range x {
		buf[i] = complex(v*window[i], 0)
	}

	fft(buf, false)

	var weighted, total float64

	for k := 1; k < len(buf)/2; k++ {
		mag := cmplx.Abs(buf[k])
		weighted += mag * float64(k) * float64(rate) / float64(len(buf))
		total += mag
	}

	if total == 0 {
		return 0
	}

	return weighted / total
}

func nextPow2(n int) int {
	p := 1
	for p < n {
		p <<= 1
	}

	return p
}

// fft is an in-place radix-2 FFT; len(a) must be a power of two. The
// inverse is not scaled.
func fft(a []complex128, inverse bool) {
	n := len(a)

	for i, j := 1, 0; i < n; i++ {
		bit := n >> 1
		for ; j&bit != 0; bit >>= 1 {
			j ^= bit
		}

		j ^= bit
		if i < j {
			a[i], a[j] = a[j], a[i]
		}
	}

	sign := -1.0
	if inverse {
		sign = 1
	}

	for size := 2; size <= n; size <<= 1 {
		step := cmplx.Rect(1, sign*2*math.Pi/float64(size))

		for start := 0; start < n; start += size {
			w := complex(1, 0)

			for k := range size / 2 {
				u, v := a[start+k], a[start+k+size/2]*w
				a[start+k], a[start+k+size/2] = u+v, u-v
				w *= step
			}
		}
	}
}

// Duration returns the time in seconds from the first to the last window
// louder than -40 dB below the loudest one, or 0 for silence.
func (f Features) Duration() float64 {
	loudest := slices.Max(append([]float64{0}, f.RMS...))
	if loudest < silentRMS {
		return 0
	}

	first, last := -1, -1

	for i, v := range f.RMS {
		if v >= loudest*durationFloor {
			if first < 0 {
				first = i
			}

			last = i
		}
	}

	return float64(last-first+1) / WindowsPerSecond
}

// Comparison is how a rendering compares with a reference.
type Comparison struct {
	// Lag is how many windows the rendering is late (negative: early),
	// found by aligning the loudness envelopes; the other numbers are
	// measured after the shift.
	Lag int
	// EnvelopeCorrelation is the Pearson correlation of the two loudness
	// envelopes, 1 for the same shape.
	EnvelopeCorrelation float64
	// LoudnessRatio is the rendering's RMS level over the reference's, over
	// the whole of both.
	LoudnessRatio float64
	// PitchRatio is the median of the rendering's pitch over the reference's,
	// in windows where both are pitched; 0 when there are none.
	PitchRatio float64
	// PitchedWindows counts those windows.
	PitchedWindows int
	// CentroidRatio is the median ratio of spectral centroids, in windows
	// where both are louder than -40 dB below their loudest.
	CentroidRatio float64
	// DurationRatio is the rendering's Duration over the reference's.
	DurationRatio float64
}

// Compare measures got against ref, allowing a shift of up to maxLag
// windows between them.
func Compare(ref, got Features, maxLag int) Comparison {
	c := Comparison{Lag: bestLag(ref.RMS, got.RMS, maxLag)}

	a, b := overlap(ref.RMS, got.RMS, c.Lag)
	c.EnvelopeCorrelation = correlation(a, b)

	if ea, eb := sumSquares(ref.RMS), sumSquares(got.RMS); ea > 0 {
		c.LoudnessRatio = math.Sqrt(eb / ea)
	}

	pa, pb := overlap(ref.Pitch, got.Pitch, c.Lag)
	ca, cb := overlap(ref.Centroid, got.Centroid, c.Lag)
	loudA := slices.Max(append([]float64{0}, a...)) * durationFloor
	loudB := slices.Max(append([]float64{0}, b...)) * durationFloor

	var pitchRatios, centroidRatios []float64

	for i := range pa {
		if pa[i] > 0 && pb[i] > 0 {
			pitchRatios = append(pitchRatios, pb[i]/pa[i])
		}

		if a[i] > loudA && b[i] > loudB && ca[i] > 0 && cb[i] > 0 {
			centroidRatios = append(centroidRatios, cb[i]/ca[i])
		}
	}

	c.PitchedWindows = len(pitchRatios)
	c.PitchRatio = median(pitchRatios)
	c.CentroidRatio = median(centroidRatios)

	if d := ref.Duration(); d > 0 {
		c.DurationRatio = got.Duration() / d
	}

	return c
}

// bestLag finds the shift of b against a, within ±maxLag, with the highest
// cross-correlation Σ a[i]·b[i+lag] (zero outside the ends). Unlike a
// correlation of the overlapping parts, this has a single peak even for an
// exponential decay, whose shifted copy is a scaled copy.
func bestLag(a, b []float64, maxLag int) int {
	best, bestDot := 0, math.Inf(-1)

	for lag := -maxLag; lag <= maxLag; lag++ {
		var dot float64

		for i, v := range a {
			if j := i + lag; j >= 0 && j < len(b) {
				dot += v * b[j]
			}
		}

		if dot > bestDot {
			best, bestDot = lag, dot
		}
	}

	return best
}

// overlap returns the parts of a and b that line up when b is shifted lag
// windows later.
func overlap(a, b []float64, lag int) ([]float64, []float64) {
	if lag >= 0 {
		b = b[min(lag, len(b)):]
	} else {
		a = a[min(-lag, len(a)):]
	}

	n := min(len(a), len(b))

	return a[:n], b[:n]
}

func correlation(a, b []float64) float64 {
	n := float64(len(a))
	if n < 2 {
		return 0
	}

	var ma, mb float64
	for i := range a {
		ma += a[i]
		mb += b[i]
	}

	ma /= n
	mb /= n

	var cov, va, vb float64
	for i := range a {
		cov += (a[i] - ma) * (b[i] - mb)
		va += (a[i] - ma) * (a[i] - ma)
		vb += (b[i] - mb) * (b[i] - mb)
	}

	if va == 0 || vb == 0 {
		return 0
	}

	return cov / math.Sqrt(va*vb)
}

func sumSquares(x []float64) float64 {
	var s float64
	for _, v := range x {
		s += v * v
	}

	return s
}

func median(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}

	s := slices.Clone(x)
	slices.Sort(s)

	if len(s)%2 == 1 {
		return s[len(s)/2]
	}

	return (s[len(s)/2-1] + s[len(s)/2]) / 2
}

// WaveformCorrelation returns the highest Pearson correlation of a and b
// over shifts of up to maxLag samples: 1 for identical waveforms.
func WaveformCorrelation(a, b []float32, maxLag int) float64 {
	x := make([]float64, len(a))
	for i, v := range a {
		x[i] = float64(v)
	}

	y := make([]float64, len(b))
	for i, v := range b {
		y[i] = float64(v)
	}

	best := math.Inf(-1)

	for lag := -maxLag; lag <= maxLag; lag++ {
		p, q := overlap(x, y, lag)
		best = max(best, correlation(p, q))
	}

	return best
}
