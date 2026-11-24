package cpm_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/cpm"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata/")

const exerciserDir = "../../third_party/8080-exercisers"

// maxCycles bounds a run, so a CPU bug that loops forever fails the test
// instead of hanging it. 8080EXM needs about 23 billion cycles.
const maxCycles = 100_000_000_000

// runCOM runs a .COM program from third_party/8080-exercisers to its warm
// boot and returns what it printed.
func runCOM(t *testing.T, name string) string {
	t.Helper()

	program, err := os.ReadFile(filepath.Join(exerciserDir, name))
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer

	m, err := cpm.New(program, &out)
	if err != nil {
		t.Fatal(err)
	}

	done, err := m.Run(maxCycles)
	if err != nil {
		t.Fatalf("%s: %v\noutput so far:\n%s", name, err, out.String())
	}

	if !done {
		t.Fatalf("%s: no warm boot after %d cycles\noutput so far:\n%s", name, maxCycles, out.String())
	}

	t.Logf("%s: %d cycles", name, m.CPU.Cycles)

	return out.String()
}

// checkGolden compares got with testdata/<name>.golden, or rewrites it with
// -update.
func checkGolden(t *testing.T, name, got string) {
	t.Helper()

	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}

		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if got != string(want) {
		t.Errorf("%s output differs from %s\ngot:\n%s\nwant:\n%s", name, path, got, want)
	}
}

// Test8080PRE runs the preliminary tests. An early failure jumps to 0
// without printing anything, so the golden's final message is what proves
// the run passed.
func Test8080PRE(t *testing.T) {
	t.Parallel()

	got := runCOM(t, "8080PRE.COM")
	if !strings.Contains(got, "8080 Preliminary tests complete") {
		t.Errorf("8080PRE did not complete:\n%s", got)
	}

	checkGolden(t, "8080PRE", got)
}

// Test8080EXM runs the full instruction exerciser, which checks every
// instruction group against CRCs taken from a real 8080. Skipped with -short.
func Test8080EXM(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("8080EXM runs billions of cycles; skipped with -short")
	}

	got := runCOM(t, "8080EXM.COM")
	if strings.Contains(got, "ERROR") || !strings.Contains(got, "Tests complete") {
		t.Errorf("8080EXM failed:\n%s", got)
	}

	checkGolden(t, "8080EXM", got)
}

// BenchmarkEXM measures exerciser throughput: the first 10 million cycles of
// 8080EXM per iteration (DESIGN §16).
func BenchmarkEXM(b *testing.B) {
	program, err := os.ReadFile(filepath.Join(exerciserDir, "8080EXM.COM"))
	if err != nil {
		b.Fatal(err)
	}

	const cycles = 10_000_000

	for b.Loop() {
		m, err := cpm.New(program, &bytes.Buffer{})
		if err != nil {
			b.Fatal(err)
		}

		if _, err := m.Run(cycles); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportMetric(float64(cycles)*float64(b.N)/b.Elapsed().Seconds()/1e6, "MHz")
}
