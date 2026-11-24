// Command cputest runs a CP/M .COM program, such as the 8080 exercisers in
// third_party/8080-exercisers, and prints its output (DESIGN §15.2):
//
//	go run ./cmd/cputest third_party/8080-exercisers/8080EXM.COM
//
// The exit code is 0 when the program ends with a warm boot, 1 when it fails
// and 2 for a usage error.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rbscholtus/Space-Invaders/internal/cpm"
)

// i8080Hz is the Space Invaders board's CPU clock (DESIGN §7), used to show
// how long a run would take on the real hardware.
const i8080Hz = 1_996_800

// Exit codes.
const (
	exitOK    = 0
	exitFail  = 1
	exitUsage = 2
)

var errNoWarmBoot = errors.New("no warm boot")

func main() {
	os.Exit(cli(os.Args[1:], os.Stdout, os.Stderr))
}

// cli runs the command with args (without the program name) and returns the
// exit code. The program's output goes to stdout; usage, errors and the
// cycle count go to stderr.
func cli(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("cputest", flag.ContinueOnError)
	fs.SetOutput(stderr)
	maxCycles := fs.Uint64("max", 100_000_000_000, "stop after this many cycles")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: cputest [-max cycles] FILE.COM")
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

	if err := run(fs.Arg(0), *maxCycles, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, "cputest:", err)

		return exitFail
	}

	return exitOK
}

func run(path string, maxCycles uint64, stdout, stderr io.Writer) error {
	program, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading the program: %w", err)
	}

	m, err := cpm.New(program, stdout)
	if err != nil {
		return fmt.Errorf("loading %s: %w", path, err)
	}

	start := time.Now()
	done, err := m.Run(maxCycles)
	elapsed := time.Since(start)

	fmt.Fprintln(stdout)

	cycles := m.CPU.Cycles
	fmt.Fprintf(stderr, "%d cycles in %v (%.0f MHz; %v at 1.9968 MHz)\n",
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
