package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/app"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
	"github.com/rbscholtus/Space-Invaders/internal/testrig"
)

var errWindow = errors.New("window failed")

func TestCLI(t *testing.T) {
	t.Parallel()

	notAZip := filepath.Join(t.TempDir(), "x.zip")
	if err := os.WriteFile(notAZip, []byte("nope"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		args       []string
		runErr     error
		wantCode   int
		wantStderr string
		wantRun    bool
	}{
		{name: "help", args: []string{"-h"}, wantCode: exitOK, wantStderr: "-roms"},
		{name: "stray argument", args: []string{"x"}, wantCode: exitUsage, wantStderr: "usage: invaders"},
		{name: "unknown flag", args: []string{"-x"}, wantCode: exitUsage, wantStderr: "not defined"},
		{name: "too many lives", args: []string{"-lives", "7"}, wantCode: exitUsage, wantStderr: "usage: invaders"},
		{name: "too few lives", args: []string{"-lives", "2"}, wantCode: exitUsage, wantStderr: "usage: invaders"},
		{name: "too loud", args: []string{"-volume", "101"}, wantCode: exitUsage, wantStderr: "usage: invaders"},
		{name: "negative volume", args: []string{"-volume", "-1"}, wantCode: exitUsage, wantStderr: "usage: invaders"},
		{name: "no ROMs", args: []string{"-roms", "/nonexistent.zip"}, wantCode: exitFail, wantStderr: "no ROM set at /nonexistent.zip"},
		{name: "not a zip", args: []string{"-roms", notAZip}, wantCode: exitFail, wantStderr: "reading the zip"},
		{name: "empty folder", args: []string{"-roms", t.TempDir()}, wantCode: exitFail, wantStderr: "no ROMs of this set found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stderr bytes.Buffer

			ran := false
			run := func(*invaders.Board, app.Options) error {
				ran = true

				return tt.runErr
			}

			if code := cli(tt.args, &stderr, run); code != tt.wantCode {
				t.Errorf("exit code %d, want %d\nstderr: %s", code, tt.wantCode, stderr.String())
			}

			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr %q, want it to contain %q", stderr.String(), tt.wantStderr)
			}

			if ran != tt.wantRun {
				t.Errorf("window opened: %v, want %v", ran, tt.wantRun)
			}
		})
	}
}

// TestCLIWithROMs runs the command with your own ROM set (skipped without
// one): the board is built with the DIP switches from the flags, and a
// window error is reported.
func TestCLIWithROMs(t *testing.T) {
	t.Parallel()

	testrig.ROMs(t) // skip without ROMs

	path := testrig.ROMPath()

	var got *invaders.Board

	var stderr bytes.Buffer

	var gotOpts app.Options

	code := cli([]string{"-roms", path, "-lives", "5", "-bonus1000", "-volume", "30"}, &stderr, func(b *invaders.Board, o app.Options) error {
		got, gotOpts = b, o

		return nil
	})
	if code != exitOK || got == nil {
		t.Fatalf("exit code %d, board %v\nstderr: %s", code, got, stderr.String())
	}

	if v := got.In(2); v&0x0B != 0x0A { // lives 5 = 10b, bonus at 1000 = bit 3
		t.Errorf("port 2 = %02X, want the DIP switches for 5 lives and bonus at 1000", v)
	}

	if gotOpts.Volume != 0.3 || gotOpts.OnError == nil {
		t.Errorf("options %+v, want volume 0.3 and an error reporter", gotOpts)
	}

	gotOpts.OnError(errWindow)

	if !strings.Contains(stderr.String(), "window failed") {
		t.Errorf("OnError wrote %q", stderr.String())
	}

	stderr.Reset()

	if code := cli([]string{"-roms", path}, &stderr, func(*invaders.Board, app.Options) error { return errWindow }); code != exitFail ||
		!strings.Contains(stderr.String(), "window failed") {
		t.Errorf("exit code %d, stderr %q; want the window's error", code, stderr.String())
	}
}
