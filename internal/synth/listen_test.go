package synth_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/soundref"
	"github.com/rbscholtus/Space-Invaders/internal/synth"
)

var listen = flag.Bool("listen", false, "write our render of each reference sequence to testdata/listen/ (git-ignored)")

// TestListen writes our render of each reference sequence as a WAV file,
// to compare by ear with MAME's in testdata/reference/wav/:
//
//	go test ./internal/synth -run TestListen -listen
func TestListen(t *testing.T) {
	t.Parallel()

	if !*listen {
		t.Skip("run with -listen to write the files")
	}

	dir := filepath.Join("testdata", "listen")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	for _, seq := range sequences(t) {
		var buf bytes.Buffer
		if err := soundref.WriteWAV(&buf, render(seq), synth.SampleRate); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(dir, seq.Name+".wav"), buf.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
