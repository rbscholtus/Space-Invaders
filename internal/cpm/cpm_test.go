package cpm_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/cpm"
)

// run loads program, runs it to its warm boot and returns its output.
func run(t *testing.T, program []byte) (string, error) {
	t.Helper()

	var out bytes.Buffer

	m, err := cpm.New(program, &out)
	if err != nil {
		t.Fatal(err)
	}

	done, err := m.Run(1_000_000)
	if err == nil && !done {
		t.Fatal("no warm boot")
	}

	if done != m.Done() {
		t.Errorf("Run reported done=%v, Done()=%v", done, m.Done())
	}

	return out.String(), err
}

func TestConsoleOutput(t *testing.T) {
	t.Parallel()

	got, err := run(t, []byte{
		0x0E, 0x02, // MVI C,2
		0x1E, 'h', //  MVI E,'h'
		0xCD, 0x05, 0x00, // CALL 5
		0x1E, 'i', //  MVI E,'i'
		0xCD, 0x05, 0x00, // CALL 5 (C is preserved by the BDOS here)
		0xC3, 0x00, 0x00, // JMP 0
	})
	if err != nil {
		t.Fatal(err)
	}

	if got != "hi" {
		t.Errorf("output %q, want %q", got, "hi")
	}
}

func TestPrintString(t *testing.T) {
	t.Parallel()

	got, err := run(t, []byte{
		0x0E, 0x09, // MVI C,9
		0x11, 0x09, 0x01, // LXI D,0109H (the string below)
		0xCD, 0x05, 0x00, // CALL 5
		0xC7, //             RST 0 (warm boot)
		'o', 'k', '\r', '\n', '$', 'x',
	})
	if err != nil {
		t.Fatal(err)
	}

	if got != "ok\r\n" {
		t.Errorf("output %q, want %q", got, "ok\r\n")
	}
}

// TestTopOfTPA checks the word at 0x0006, which the exercisers load into SP.
func TestTopOfTPA(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	m, err := cpm.New([]byte{0x2A, 0x06, 0x00, 0xF9}, &out) // LHLD 6; SPHL
	if err != nil {
		t.Fatal(err)
	}

	m.Step()
	m.Step()

	if m.CPU.SP != 0xFE00 {
		t.Errorf("SP = %04X, want FE00", m.CPU.SP)
	}
}

func TestUnsupportedBDOS(t *testing.T) {
	t.Parallel()

	_, err := run(t, []byte{0x0E, 0x01, 0xCD, 0x05, 0x00, 0xC7}) // MVI C,1; CALL 5
	if !errors.Is(err, cpm.ErrUnsupported) {
		t.Errorf("err = %v, want ErrUnsupported", err)
	}
}

func TestProgramTooBig(t *testing.T) {
	t.Parallel()

	_, err := cpm.New(make([]byte, 0xFE00), &bytes.Buffer{})
	if !errors.Is(err, cpm.ErrTooBig) {
		t.Errorf("err = %v, want ErrTooBig", err)
	}
}

type failingWriter struct{}

var errWrite = errors.New("write failed")

func (failingWriter) Write([]byte) (int, error) { return 0, errWrite }

func TestOutputError(t *testing.T) {
	t.Parallel()

	m, err := cpm.New([]byte{0x0E, 0x02, 0xCD, 0x05, 0x00, 0xC7}, failingWriter{})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := m.Run(1000); !errors.Is(err, errWrite) {
		t.Errorf("err = %v, want the writer's error", err)
	}

	if m.Step() != 0 {
		t.Error("Step after an error should do nothing")
	}
}

func TestPortsAreInert(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	m, err := cpm.New([]byte{0xD3, 0x01, 0xDB, 0x01}, &out) // OUT 1; IN 1
	if err != nil {
		t.Fatal(err)
	}

	m.Step()
	m.Step()

	if m.CPU.A != 0xFF {
		t.Errorf("IN returned %02X, want FF", m.CPU.A)
	}
}
