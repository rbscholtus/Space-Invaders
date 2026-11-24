package synth_test

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/invaders"
	"github.com/rbscholtus/Space-Invaders/internal/soundref"
	"github.com/rbscholtus/Space-Invaders/internal/synth"
	"github.com/rbscholtus/Space-Invaders/internal/testrig"
)

// cpuHz is the board's CPU clock (DESIGN §7): the session is rendered on
// the emulated time base, as MAME runs it, not the app's 60 Hz.
const cpuHz = 1996800

// mameScreenStart is where MAME's screen is at power-on: at the start of
// vertical blank (line 224), where ours is at line 0. MAME's interrupts,
// and so the game's sound writes, come 38 lines later after power-on than
// ours, while the sound circuits run from power-on in both (measured: the
// frame callbacks fall at whole frame periods).
const mameScreenStart = (262 - 224) * invaders.CyclesPerLine

// renderSession plays the golden session's script on the board and renders
// its sound with every port write at its emulated time.
func renderSession(t *testing.T) []float32 {
	t.Helper()

	b, err := invaders.New(testrig.ROMs(t), invaders.DIPs{})
	if err != nil {
		t.Fatal(err)
	}

	s := synth.New(synth.SampleRate)
	total := int((int64(testrig.ScriptFrames)*invaders.CyclesPerFrame + mameScreenStart) * synth.SampleRate / cpuHz)
	out := make([]float32, total)
	pos := 0

	// Inputs change as MAME latches them, at the start of vertical blank
	// (scripts/mame/session.lua), so the game runs exactly as in MAME.
	const vblankLine = 224

	for f := range testrig.ScriptFrames {
		b.SetInputsAt(vblankLine, testrig.Script(f+1))
		b.RunFrame()

		for _, e := range b.SoundEvents() {
			at := int((int64(f)*invaders.CyclesPerFrame + int64(e.Cycle) + mameScreenStart) * synth.SampleRate / cpuHz)
			at = min(max(at, pos), total)
			s.Render(out[pos:at])
			pos = at

			if e.Port == 3 {
				s.WritePort3(e.Value)
			} else {
				s.WritePort5(e.Value)
			}
		}
	}

	s.Render(out[pos:])

	for i, v := range out {
		out[i] = max(-1, min(1, v))
	}

	return out
}

// envelope is the RMS per 10 ms.
func envelope(x []float32) []float64 {
	const hop = synth.SampleRate / 100

	env := make([]float64, len(x)/hop)
	for w := range env {
		var sum float64
		for _, v := range x[w*hop : (w+1)*hop] {
			sum += float64(v) * float64(v)
		}

		env[w] = math.Sqrt(sum / hop)
	}

	return env
}

func pearson(a, b []float64) float64 {
	n := float64(min(len(a), len(b)))

	var ma, mb float64
	for i := range int(n) {
		ma += a[i]
		mb += b[i]
	}

	ma /= n
	mb /= n

	var cov, va, vb float64
	for i := range int(n) {
		cov += (a[i] - ma) * (b[i] - mb)
		va += (a[i] - ma) * (a[i] - ma)
		vb += (b[i] - mb) * (b[i] - mb)
	}

	return cov / math.Sqrt(va*vb)
}

// Session check thresholds: over the windows where both make sound.
const (
	audibleRMS          = 1e-4 // -80 dBFS: below this a window is quiet
	minSessionEnvCorr   = 0.99
	minSessionWaveCorr  = 0.99
	sessionLoudnessTol  = 0.01
	sessionWindowSample = synth.SampleRate / 100
)

// TestSessionAgainstMAME compares the sound of the whole golden session
// with MAME playing the same script (local only: your ROMs and
// scripts/mame-reference.sh session). With MAME's input timing the game
// runs exactly as in MAME (work and video RAM match at every frame), so
// the sound should too.
//
// Only windows where both make sound are compared: MAME applies its
// amplifier mute to whole buffers of audio as its sound thread processes
// them, about 0.8 s after the game switches the amplifier off, where ours
// (and the real board) mute at the sample. Those windows are counted, not
// compared.
func TestSessionAgainstMAME(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join(refDir, "wav", "session.wav"))
	if err != nil {
		t.Skip("no MAME session recording: run scripts/mame-reference.sh session")
	}

	mame, _, err := soundref.ReadWAV(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	ours := renderSession(t)
	n := min(len(mame), len(ours))
	envA, envB := envelope(mame[:n]), envelope(ours[:n])

	var (
		a, b         []float64 // envelopes where both are audible
		wavA, wavB   []float32 // and their samples
		onlyA, onlyB int
	)

	for w := range envA {
		switch {
		case envA[w] > audibleRMS && envB[w] > audibleRMS:
			a, b = append(a, envA[w]), append(b, envB[w])
			wavA = append(wavA, mame[w*sessionWindowSample:(w+1)*sessionWindowSample]...)
			wavB = append(wavB, ours[w*sessionWindowSample:(w+1)*sessionWindowSample]...)
		case envA[w] > audibleRMS:
			onlyA++
		case envB[w] > audibleRMS:
			onlyB++
		}
	}

	envCorr := pearson(a, b)
	waveCorr := soundref.WaveformCorrelation(wavA, wavB, 0)

	var ea, eb float64
	for i := range a {
		ea += a[i] * a[i]
		eb += b[i] * b[i]
	}

	loudness := math.Sqrt(eb / ea)

	t.Logf("%d windows audible in both (%.0f s); only MAME %d, only ours %d (MAME's late mute)",
		len(a), float64(len(a))/100, onlyA, onlyB)
	t.Logf("envelope correlation %.4f, waveform correlation %.4f, loudness ratio %.4f", envCorr, waveCorr, loudness)

	if envCorr < minSessionEnvCorr {
		t.Errorf("envelope correlation %.4f, want at least %.2f", envCorr, minSessionEnvCorr)
	}

	if waveCorr < minSessionWaveCorr {
		t.Errorf("waveform correlation %.4f, want at least %.2f", waveCorr, minSessionWaveCorr)
	}

	if loudness < 1-sessionLoudnessTol || loudness > 1+sessionLoudnessTol {
		t.Errorf("loudness ratio %.4f, want 1 ± %.2f", loudness, sessionLoudnessTol)
	}
}
