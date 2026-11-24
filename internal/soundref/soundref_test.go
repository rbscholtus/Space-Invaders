package soundref_test

import (
	"bytes"
	"errors"
	"math"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/soundref"
)

const rate = 48000

func sine(hz, seconds, amp float64) []float32 {
	x := make([]float32, int(seconds*rate))
	for i := range x {
		x[i] = float32(amp * math.Sin(2*math.Pi*hz*float64(i)/rate))
	}

	return x
}

func square(hz, seconds, amp float64) []float32 {
	x := make([]float32, int(seconds*rate))
	for i := range x {
		if math.Mod(float64(i)*hz/rate, 1) < 0.5 {
			x[i] = float32(amp)
		} else {
			x[i] = float32(-amp)
		}
	}

	return x
}

func noise(seconds, amp float64) []float32 {
	r := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // G404: deterministic test noise
	x := make([]float32, int(seconds*rate))

	for i := range x {
		x[i] = float32(amp * (2*r.Float64() - 1))
	}

	return x
}

func median(x []float64) float64 {
	var nz []float64

	for _, v := range x {
		if v > 0 {
			nz = append(nz, v)
		}
	}

	if len(nz) == 0 {
		return 0
	}

	for i := range nz { // insertion sort; short slices
		for j := i; j > 0 && nz[j] < nz[j-1]; j-- {
			nz[j], nz[j-1] = nz[j-1], nz[j]
		}
	}

	return nz[len(nz)/2]
}

func TestPitchOfTones(t *testing.T) {
	t.Parallel()

	for _, hz := range []float64{45, 55, 110, 440, 1000, 3000} {
		for name, x := range map[string][]float32{"sine": sine(hz, 0.5, 0.5), "square": square(hz, 0.5, 0.5)} {
			got := median(soundref.Analyze(x, rate).Pitch)
			if math.Abs(got/hz-1) > 0.01 {
				t.Errorf("%s %.0f Hz: pitch %.2f Hz", name, hz, got)
			}
		}
	}
}

func TestNoiseHasNoPitch(t *testing.T) {
	t.Parallel()

	f := soundref.Analyze(noise(0.5, 0.5), rate)

	pitched := 0

	for _, p := range f.Pitch {
		if p > 0 {
			pitched++
		}
	}

	if pitched > len(f.Pitch)/10 {
		t.Errorf("%d of %d noise windows have a pitch", pitched, len(f.Pitch))
	}

	// White noise's centroid is near a quarter of the sample rate.
	if c := median(f.Centroid); c < 9000 || c > 15000 {
		t.Errorf("noise centroid %.0f Hz, want about 12 kHz", c)
	}
}

func TestRMSAndDuration(t *testing.T) {
	t.Parallel()

	x := append(make([]float32, rate/2), sine(440, 1, 0.5)...)
	x = append(x, make([]float32, rate/2)...)
	f := soundref.Analyze(x, rate)

	if len(f.RMS) != 200 || f.Hop != 480 {
		t.Fatalf("%d windows of %d samples, want 200 of 480", len(f.RMS), f.Hop)
	}

	if got := f.RMS[100]; math.Abs(got-0.5/math.Sqrt2) > 0.01 {
		t.Errorf("RMS of a 0.5 sine = %.3f, want 0.354", got)
	}

	if f.RMS[10] != 0 || f.Pitch[10] != 0 || f.Centroid[10] != 0 {
		t.Errorf("silence measured as %v %v %v", f.RMS[10], f.Pitch[10], f.Centroid[10])
	}

	if d := f.Duration(); math.Abs(d-1) > 0.02 {
		t.Errorf("duration %.2f s, want 1", d)
	}

	if d := soundref.Analyze(make([]float32, rate), rate).Duration(); d != 0 {
		t.Errorf("duration of silence %.2f", d)
	}
}

func TestCompare(t *testing.T) {
	t.Parallel()

	decay := func(hz float64, delay int) []float32 {
		x := make([]float32, delay, rate+delay)
		for i := range rate {
			x = append(x, float32(math.Exp(-5*float64(i)/rate)*math.Sin(2*math.Pi*hz*float64(i)/rate)))
		}

		return x
	}

	ref := soundref.Analyze(decay(440, 0), rate)

	same := soundref.Compare(ref, soundref.Analyze(decay(440, 0), rate), 10)
	if same.Lag != 0 || same.EnvelopeCorrelation < 0.999 || math.Abs(same.PitchRatio-1) > 0.001 ||
		math.Abs(same.LoudnessRatio-1) > 0.001 || math.Abs(same.DurationRatio-1) > 0.001 {
		t.Errorf("identical sounds compare as %+v", same)
	}

	late := soundref.Compare(ref, soundref.Analyze(decay(440, 5*480), rate), 10)
	if late.Lag != 5 || late.EnvelopeCorrelation < 0.99 {
		t.Errorf("a sound 50 ms late compares as %+v, want lag 5", late)
	}

	sharp := soundref.Compare(ref, soundref.Analyze(decay(466, 0), rate), 10)
	if math.Abs(sharp.PitchRatio-466.0/440) > 0.005 || sharp.PitchedWindows < 50 {
		t.Errorf("a semitone up compares as %+v", sharp)
	}
}

func TestWaveformCorrelation(t *testing.T) {
	t.Parallel()

	a := sine(440, 0.1, 0.5)
	b := append(make([]float32, 7), a...)

	if c := soundref.WaveformCorrelation(a, b, 10); c < 0.999 {
		t.Errorf("shifted copy correlates %.4f", c)
	}

	if c := soundref.WaveformCorrelation(a, noise(0.1, 0.5), 10); c > 0.2 {
		t.Errorf("sine and noise correlate %.4f", c)
	}
}

func TestWAVRoundTrip(t *testing.T) {
	t.Parallel()

	x := []float32{0, 0.5, -0.5, 1, -1, 2, -2}

	var buf bytes.Buffer
	if err := soundref.WriteWAV(&buf, x, rate); err != nil {
		t.Fatal(err)
	}

	got, r, err := soundref.ReadWAV(&buf)
	if err != nil || r != rate || len(got) != len(x) {
		t.Fatalf("ReadWAV: %d samples at %d Hz, %v", len(got), r, err)
	}

	for i, want := range []float32{0, 0.5, -0.5, 1, -1, 1, -1} {
		if math.Abs(float64(got[i]-want)) > 1e-4 {
			t.Errorf("sample %d = %v, want %v", i, got[i], want)
		}
	}

	if _, _, err := soundref.ReadWAV(strings.NewReader("not a wav")); !errors.Is(err, soundref.ErrWAV) {
		t.Errorf("err = %v, want ErrWAV", err)
	}
}

func TestReadSequences(t *testing.T) {
	t.Parallel()

	seqs, err := soundref.ReadSequences(strings.NewReader(`
# comment
sound shot 300
120 3 0x20   # amp on
150 3 0x22
sound fleet 10
1 5 1
`))
	if err != nil {
		t.Fatal(err)
	}

	if len(seqs) != 2 || seqs[0].Name != "shot" || seqs[0].Frames != 300 || len(seqs[0].Writes) != 2 ||
		seqs[0].Writes[1] != (soundref.Write{Frame: 150, Port: 3, Value: 0x22}) || seqs[1].Writes[0].Value != 1 {
		t.Errorf("sequences %+v", seqs)
	}

	for _, bad := range []string{"120 3 0x20", "sound x", "sound x y", "sound x 1\n1 2 0x100", "sound x 1\n1 2 z"} {
		if _, err := soundref.ReadSequences(strings.NewReader(bad)); !errors.Is(err, soundref.ErrSpec) {
			t.Errorf("%q: err = %v, want ErrSpec", bad, err)
		}
	}

	if got := soundref.SampleAt(60, rate); got != 48369 {
		t.Errorf("SampleAt(60) = %d, want 48369 (60 frames at 59.54 Hz)", got)
	}
}
