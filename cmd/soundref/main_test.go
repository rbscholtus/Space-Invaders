package main

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/soundref"
)

// setup makes a reference folder with one sequence and, if withWAV, its
// recording: a 440 Hz tone from frame 10.
func setup(t *testing.T, withWAV bool) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sequences.txt"), []byte("sound beep 120\n10 3 0x20\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if !withWAV {
		return dir
	}

	const rate = 48000

	x := make([]float32, 2*rate)
	for i := soundref.SampleAt(10, rate); i < len(x); i++ {
		x[i] = float32(0.5 * math.Sin(2*math.Pi*440*float64(i)/rate))
	}

	var buf bytes.Buffer
	if err := soundref.WriteWAV(&buf, x, rate); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "wav"), 0o750); err != nil {
		t.Fatal(err)
	}

	for name, data := range map[string][]byte{"beep.wav": buf.Bytes(), "mame-version.txt": []byte("0.289 (unknown)\n")} {
		if err := os.WriteFile(filepath.Join(dir, "wav", name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func TestMeasures(t *testing.T) {
	t.Parallel()

	dir := setup(t, true)

	var stdout, stderr bytes.Buffer
	if code := cli([]string{"-dir", dir}, &stdout, &stderr); code != exitOK {
		t.Fatalf("exit code %d: %s", code, stderr.String())
	}

	f, err := os.Open(filepath.Join(dir, "beep.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	ref, err := soundref.ReadReference(f)
	if err != nil {
		t.Fatal(err)
	}

	if ref.Sound != "beep" || ref.Source != "MAME 0.289" || ref.StartFrame != 10 || len(ref.WAVSHA256) != 64 {
		t.Errorf("reference header %+v", ref)
	}

	if d := ref.Duration(); math.Abs(d-(2-10/soundref.FrameRate)) > 0.02 {
		t.Errorf("duration %.2f s", d)
	}

	if p := ref.Pitch[50]; math.Abs(p-440) > 2 {
		t.Errorf("pitch %.1f Hz, want 440", p)
	}

	if !strings.Contains(stdout.String(), "beep") {
		t.Errorf("stdout %q", stdout.String())
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantErr  string
	}{
		{"help", []string{"-h"}, exitOK, "-dir"},
		{"stray argument", []string{"x"}, exitUsage, "usage: soundref"},
		{"no folder", []string{"-dir", "/nonexistent"}, exitFail, "opening the sequences"},
		{"no recording", []string{"-dir", setup(t, false)}, exitFail, "record it with scripts/mame-reference.sh"},
	}
	for _, tt := range tests {
		var stdout, stderr bytes.Buffer
		if code := cli(tt.args, &stdout, &stderr); code != tt.wantCode || !strings.Contains(stderr.String(), tt.wantErr) {
			t.Errorf("%s: exit code %d, stderr %q; want %d and %q", tt.name, code, stderr.String(), tt.wantCode, tt.wantErr)
		}
	}
}
