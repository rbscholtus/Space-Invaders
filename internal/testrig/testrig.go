// Package testrig holds helpers for tests: finding the player's own ROM set
// (DESIGN §15.3). Production code never imports it.
package testrig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/romset"
)

// EnvROMs names the environment variable that points at a ROM set: a zip or
// a folder.
const EnvROMs = "INVADERS_ROMS"

// ROMPath returns where the ROM set is looked for: $INVADERS_ROMS, or
// roms/invaders.zip at the repository root (the folder with go.mod, found
// from the test's working directory, which is its package's folder).
func ROMPath() string {
	if path := os.Getenv(EnvROMs); path != "" {
		return path
	}

	dir, err := os.Getwd()
	if err != nil {
		return filepath.Join("roms", "invaders.zip")
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "roms", "invaders.zip")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return filepath.Join("roms", "invaders.zip")
		}

		dir = parent
	}
}

// ROMs returns the 8 KB Space Invaders image from the player's own ROM set,
// or skips the test when there is none. ROMs are never committed, so CI
// skips these tests (DESIGN §9, §15.3). A ROM set that is present but wrong
// fails the test.
func ROMs(tb testing.TB) []byte {
	tb.Helper()

	path := ROMPath()

	info, err := os.Stat(path)
	if err != nil {
		tb.Skipf("no ROM set at %s; put your own there or set %s", path, EnvROMs)
	}

	var image []byte

	if info.IsDir() {
		image, err = romset.Load(romset.Invaders(), os.DirFS(path))
	} else {
		var data []byte

		data, err = os.ReadFile(path)
		if err == nil {
			image, err = romset.LoadZip(romset.Invaders(), data)
		}
	}

	if err != nil {
		tb.Fatalf("ROM set at %s: %v", path, err)
	}

	return image
}
