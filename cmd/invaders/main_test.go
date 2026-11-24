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
		// No -roms and no roms/invaders.zip here: the app starts with its
		// ROM screen.
		{name: "no ROMs given", args: nil, wantCode: exitOK, wantRun: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stderr bytes.Buffer

			ran := false
			run := func(app.Options) error {
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

	var stderr bytes.Buffer

	var gotOpts app.Options

	code := cli([]string{"-roms", path, "-lives", "5", "-bonus1000", "-volume", "30"}, &stderr, func(o app.Options) error {
		gotOpts = o

		return nil
	})
	if code != exitOK || len(gotOpts.ROMs) != 0x2000 {
		t.Fatalf("exit code %d, %d bytes of ROM\nstderr: %s", code, len(gotOpts.ROMs), stderr.String())
	}

	if want := (invaders.DIPs{Lives: 5, BonusAt1000: true}); gotOpts.DIPs != want {
		t.Errorf("DIPs %+v, want %+v", gotOpts.DIPs, want)
	}

	if gotOpts.Volume != 0.3 || gotOpts.OnError == nil {
		t.Errorf("options %+v, want volume 0.3 and an error reporter", gotOpts)
	}

	gotOpts.OnError(errWindow)

	if !strings.Contains(stderr.String(), "window failed") {
		t.Errorf("OnError wrote %q", stderr.String())
	}

	stderr.Reset()

	if code := cli([]string{"-roms", path}, &stderr, func(app.Options) error { return errWindow }); code != exitFail ||
		!strings.Contains(stderr.String(), "window failed") {
		t.Errorf("exit code %d, stderr %q; want the window's error", code, stderr.String())
	}
}
