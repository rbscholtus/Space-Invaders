// Command cputest runs a CP/M .COM program, such as the 8080 exercisers in
// third_party/8080-exercisers, and prints its output (DESIGN §15.2):
//
//	go run ./cmd/cputest third_party/8080-exercisers/8080EXM.COM
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rbscholtus/Space-Invaders/internal/cpm"
)

// i8080Hz is the Space Invaders board's CPU clock (DESIGN §7), used to show
// how long a run would take on the real hardware.
const i8080Hz = 1_996_800

var errNoWarmBoot = errors.New("no warm boot")

func main() {
	maxCycles := flag.Uint64("max", 100_000_000_000, "stop after this many cycles")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "usage: cputest [-max cycles] FILE.COM")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(flag.Arg(0), *maxCycles); err != nil {
		fmt.Fprintln(os.Stderr, "cputest:", err)
		os.Exit(1)
	}
}

func run(path string, maxCycles uint64) error {
	program, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading the program: %w", err)
	}

	m, err := cpm.New(program, os.Stdout)
	if err != nil {
		return fmt.Errorf("loading %s: %w", path, err)
	}

	start := time.Now()
	done, err := m.Run(maxCycles)
	elapsed := time.Since(start)

	fmt.Fprintln(os.Stdout)

	cycles := m.CPU.Cycles
	fmt.Fprintf(os.Stderr, "%d cycles in %v (%.0f MHz; %v at 1.9968 MHz)\n",
		cycles, elapsed.Round(time.Millisecond), float64(cycles)/elapsed.Seconds()/1e6,
		time.Duration(float64(cycles)/i8080Hz*float64(time.Second)).Round(time.Second))

	if err != nil {
		return fmt.Errorf("running %s: %w", path, err)
	}

	if !done {
		return fmt.Errorf("%s: %w after %d cycles", path, errNoWarmBoot, maxCycles)
	}

	return nil
}
