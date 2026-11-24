package synth_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/soundref"
	"github.com/rbscholtus/Space-Invaders/internal/synth"
)

const refDir = "testdata/reference"

func sequences(t *testing.T) []soundref.Sequence {
	t.Helper()

	f, err := os.Open(filepath.Join(refDir, "sequences.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	seqs, err := soundref.ReadSequences(f)
	if err != nil {
		t.Fatal(err)
	}

	return seqs
}

// render plays a sequence on a fresh board, as MAME played it: each write
// at the sample where it took effect in MAME, from power-on.
func render(seq soundref.Sequence) []float32 {
	b := synth.New(synth.SampleRate)
	out := make([]float32, soundref.SampleAt(seq.Frames, synth.SampleRate))
	pos := 0

	for _, w := range seq.Writes {
		at := min(w.Sample(synth.SampleRate), len(out))
		b.Render(out[pos:at])
		pos = at

		switch w.Port {
		case 3:
			b.WritePort3(w.Value)
		case 5:
			b.WritePort5(w.Value)
		}
	}

	b.Render(out[pos:])

	// Clip as MAME's WAV writer does.
	for i, v := range out {
		out[i] = max(-1, min(1, v))
	}

	return out
}

func reference(t *testing.T, name string) soundref.Reference {
	t.Helper()

	f, err := os.Open(filepath.Join(refDir, name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	ref, err := soundref.ReadReference(f)
	if err != nil {
		t.Fatal(err)
	}

	return ref
}

// Tolerances against MAME's recordings (DESIGN §14.5). The port follows
// MAME's arithmetic, so the measurements agree almost exactly; these leave
// room only for floating-point and rounding differences.
const (
	maxLagWindows    = 1    // 10 ms
	minEnvelopeCorr  = 0.99 // loudness envelope shape
	loudnessTol      = 0.01 // ±1%
	durationTol      = 0.02 // ±2% (two windows of a one-second sound)
	pitchTol         = 0.01 // ±1%, where at least minPitched windows have a pitch
	minPitched       = 10
	centroidTol      = 0.03 // ±3%
	minWaveformCorr  = 0.99 // locally, with the recordings
	maxWaveformShift = 1000 // samples searched for the best waveform alignment
)

// notYetPorted lists the references whose sound is still missing.
var notYetPorted = map[string]bool{"ufo": true}

// TestAgainstMAME plays every reference sequence and compares the output
// with the committed measurements of MAME's recording; with the recording
// present (scripts/mame-reference.sh), also the waveform itself.
func TestAgainstMAME(t *testing.T) {
	t.Parallel()

	for _, seq := range sequences(t) {
		t.Run(seq.Name, func(t *testing.T) {
			t.Parallel()

			if notYetPorted[seq.Name] {
				t.Skip("not ported yet")
			}

			ref := reference(t, seq.Name)
			out := render(seq)
			from := (soundref.Write{Frame: ref.StartFrame}).Sample(synth.SampleRate)
			c := soundref.Compare(ref.Features, soundref.Analyze(out[from:], synth.SampleRate), 10)

			t.Logf("lag %d, envelope %.4f, loudness %.4f, pitch %.4f (%d windows), centroid %.4f, duration %.4f",
				c.Lag, c.EnvelopeCorrelation, c.LoudnessRatio, c.PitchRatio, c.PitchedWindows, c.CentroidRatio, c.DurationRatio)

			if c.Lag < -maxLagWindows || c.Lag > maxLagWindows {
				t.Errorf("shifted %d windows against MAME", c.Lag)
			}

			if c.EnvelopeCorrelation < minEnvelopeCorr {
				t.Errorf("envelope correlation %.4f, want at least %.2f", c.EnvelopeCorrelation, minEnvelopeCorr)
			}

			checkRatio(t, "loudness", c.LoudnessRatio, loudnessTol)
			checkRatio(t, "duration", c.DurationRatio, durationTol)
			checkRatio(t, "centroid", c.CentroidRatio, centroidTol)

			if c.PitchedWindows >= minPitched {
				checkRatio(t, "pitch", c.PitchRatio, pitchTol)
			}

			compareWaveform(t, seq.Name, out, from)
		})
	}
}

func checkRatio(t *testing.T, name string, ratio, tolerance float64) {
	t.Helper()

	if ratio < 1-tolerance || ratio > 1+tolerance {
		t.Errorf("%s ratio %.4f, want 1 ± %.2f", name, ratio, tolerance)
	}
}

// compareWaveform compares with MAME's recording, if it is present.
func compareWaveform(t *testing.T, name string, out []float32, from int) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(refDir, "wav", name+".wav"))
	if err != nil {
		t.Log("no recording: waveform not compared (scripts/mame-reference.sh)")

		return
	}

	mame, _, err := soundref.ReadWAV(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	n := min(len(mame), len(out))
	if corr := soundref.WaveformCorrelation(mame[from:n], out[from:n], maxWaveformShift); corr < minWaveformCorr {
		t.Errorf("waveform correlation %.4f, want at least %.2f", corr, minWaveformCorr)
	} else {
		t.Logf("waveform correlation %.4f", corr)
	}
}
