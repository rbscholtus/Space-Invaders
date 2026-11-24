package testrig_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/testrig"
)

func TestROMPathFindsTheRepositoryRoot(t *testing.T) {
	t.Setenv(testrig.EnvROMs, "")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(wd, "..", "..", "roms", "invaders.zip")
	if got := testrig.ROMPath(); filepath.Clean(got) != filepath.Clean(want) {
		t.Errorf("ROMPath() = %s, want %s", got, want)
	}
}

func TestROMPathFromEnvironment(t *testing.T) {
	t.Setenv(testrig.EnvROMs, "/some/where.zip")

	if got := testrig.ROMPath(); got != "/some/where.zip" {
		t.Errorf("ROMPath() = %s, want the environment's path", got)
	}
}

func TestROMsFromAFolder(t *testing.T) {
	// A folder without the set: ROMs must fail, not skip. A recorder
	// stands in for the test, so the failure doesn't fail this test.
	dir := t.TempDir()
	t.Setenv(testrig.EnvROMs, dir)

	rec := &recorder{TB: t}

	func() {
		defer func() { _ = recover() }() // the recorder panics to end ROMs, as Fatalf would

		testrig.ROMs(rec)
	}()

	if !rec.failed {
		t.Error("ROMs accepted a folder without ROMs")
	}
}

// recorder records Fatalf and Skipf instead of ending the test.
type recorder struct {
	testing.TB

	failed, skipped bool
}

func (r *recorder) Helper() {}

func (r *recorder) Fatalf(string, ...any) {
	r.failed = true

	panic("fatal")
}

func (r *recorder) Skipf(string, ...any) {
	r.skipped = true

	panic("skip")
}

func TestROMsSkipsWithoutASet(t *testing.T) {
	t.Setenv(testrig.EnvROMs, filepath.Join(t.TempDir(), "missing.zip"))

	rec := &recorder{TB: t}

	func() {
		defer func() { _ = recover() }()

		testrig.ROMs(rec)
	}()

	if !rec.skipped || rec.failed {
		t.Errorf("skipped=%v failed=%v, want a skip", rec.skipped, rec.failed)
	}
}
