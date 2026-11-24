// Command soundref measures the reference recordings made by
// scripts/mame-reference.sh and writes the measurements that are committed
// (DESIGN §14.5): one .json per sequence in
// internal/synth/testdata/reference/. The recordings stay local.
//
//	scripts/mame-reference.sh && go run ./cmd/soundref
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rbscholtus/Space-Invaders/internal/soundref"
)

// Exit codes.
const (
	exitOK    = 0
	exitFail  = 1
	exitUsage = 2
)

func main() {
	os.Exit(cli(os.Args[1:], os.Stdout, os.Stderr))
}

// cli parses args and measures the recordings; it returns the exit code.
func cli(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("soundref", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", "internal/synth/testdata/reference", "folder with sequences.txt, wav/ and the .json files")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: soundref [-dir folder]")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}

		return exitUsage
	}

	if fs.NArg() != 0 {
		fs.Usage()

		return exitUsage
	}

	if err := run(*dir, stdout); err != nil {
		fmt.Fprintln(stderr, "soundref:", err)

		return exitFail
	}

	return exitOK
}

func run(dir string, stdout io.Writer) error {
	spec, err := os.Open(filepath.Join(dir, "sequences.txt"))
	if err != nil {
		return fmt.Errorf("opening the sequences: %w", err)
	}
	defer spec.Close()

	seqs, err := soundref.ReadSequences(spec)
	if err != nil {
		return fmt.Errorf("reading the sequences: %w", err)
	}

	source := "MAME"

	if v, err := os.ReadFile(filepath.Join(dir, "wav", "mame-version.txt")); err == nil {
		version, _, _ := strings.Cut(strings.TrimSpace(string(v)), " ")
		source += " " + version
	}

	for _, seq := range seqs {
		if err := measure(dir, source, seq, stdout); err != nil {
			return err
		}
	}

	return nil
}

func measure(dir, source string, seq soundref.Sequence, stdout io.Writer) error {
	data, err := os.ReadFile(filepath.Join(dir, "wav", seq.Name+".wav"))
	if err != nil {
		return fmt.Errorf("%s: %w (record it with scripts/mame-reference.sh)", seq.Name, err)
	}

	samples, rate, err := soundref.ReadWAV(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%s: %w", seq.Name, err)
	}

	start := 0
	if len(seq.Writes) > 0 {
		start = seq.Writes[0].Frame
	}

	from := min(soundref.SampleAt(start, rate), len(samples))
	sum := sha256.Sum256(data)
	ref := soundref.Reference{
		Sound: seq.Name, Source: source, WAVSHA256: hex.EncodeToString(sum[:]), StartFrame: start,
		Features: soundref.Analyze(samples[from:], rate),
	}

	var buf bytes.Buffer
	if err := soundref.WriteReference(&buf, ref); err != nil {
		return fmt.Errorf("%s: %w", seq.Name, err)
	}

	path := filepath.Join(dir, seq.Name+".json")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	fmt.Fprintf(stdout, "%-14s %5.2f s measured, sound lasts %.2f s\n",
		seq.Name, float64(len(samples)-from)/float64(rate), ref.Duration())

	return nil
}
