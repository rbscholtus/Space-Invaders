// Command trace runs a CP/M .COM program headless and prints one line per
// instruction: address, bytes, disassembly, registers, flags and cycles
// (DESIGN §3). Compare two runs, or a run with another emulator's log, with
// diff. The program's own output goes to stderr.
//
//	go run ./cmd/trace -n 100 third_party/8080-exercisers/8080PRE.COM
//
// Running the Space Invaders ROM comes with the board (ROADMAP M2).
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/rbscholtus/Space-Invaders/internal/cpm"
)

func main() {
	n := flag.Uint64("n", 1000, "stop after this many instructions (0: no limit)")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "usage: trace [-n count] FILE.COM")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(flag.Arg(0), *n); err != nil {
		fmt.Fprintln(os.Stderr, "trace:", err)
		os.Exit(1)
	}
}

func run(path string, limit uint64) error {
	program, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading the program: %w", err)
	}

	m, err := cpm.New(program, os.Stderr)
	if err != nil {
		return fmt.Errorf("loading %s: %w", path, err)
	}

	w := bufio.NewWriter(os.Stdout)

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

	if _, err := m.Run(0); err != nil {
		return fmt.Errorf("running %s: %w", path, err)
	}

	return nil
}
