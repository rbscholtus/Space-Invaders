// Command trace runs a CP/M .COM program headless and prints one line per
// instruction: address, bytes, disassembly, registers, flags and cycles
// (DESIGN §3). Compare two runs, or a run with another emulator's log, with
// diff. The program's own output goes to stderr.
//
//	go run ./cmd/trace -n 100 third_party/8080-exercisers/8080PRE.COM
//
// The exit code is 0 on success, 1 when the program fails and 2 for a usage
// error. Running the Space Invaders ROM comes with the board (ROADMAP M2).
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/rbscholtus/Space-Invaders/internal/cpm"
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

// cli runs the command with args (without the program name) and returns the
// exit code. The trace goes to stdout; the program's output, usage and
// errors go to stderr.
func cli(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("trace", flag.ContinueOnError)
	fs.SetOutput(stderr)
	n := fs.Uint64("n", 1000, "stop after this many instructions (0: no limit)")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: trace [-n count] FILE.COM")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}

		return exitUsage
	}

	if fs.NArg() != 1 {
		fs.Usage()

		return exitUsage
	}

	if err := run(fs.Arg(0), *n, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, "trace:", err)

		return exitFail
	}

	return exitOK
}

func run(path string, limit uint64, stdout, stderr io.Writer) error {
	program, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading the program: %w", err)
	}

	m, err := cpm.New(program, stderr)
	if err != nil {
		return fmt.Errorf("loading %s: %w", path, err)
	}

	w := bufio.NewWriter(stdout)

	for i := uint64(0); limit == 0 || i < limit; i++ {
		state := m.CPU.Snapshot()
		if m.Step() == 0 {
			break // warm boot or error
		}

		if _, err := fmt.Fprintln(w, state); err != nil {
			return fmt.Errorf("writing the trace: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("writing the trace: %w", err)
	}

	// Run(0) runs nothing; it reports the error that stopped Step, if any.
	if _, err := m.Run(0); err != nil {
		return fmt.Errorf("running %s: %w", path, err)
	}

	return nil
}
