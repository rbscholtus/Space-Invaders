// Package cpm is a minimal CP/M machine for the 8080 exercisers (DESIGN
// §15.2): 64 KB of RAM, a program loaded at 0x0100, BDOS functions 2
// (console output) and 9 (print string), and a warm boot (a jump to 0x0000)
// that ends the run.
//
// It does no I/O of its own: console output goes to the io.Writer it is
// given, and it never reads the clock.
package cpm

import (
	"errors"
	"fmt"
	"io"

	"github.com/rbscholtus/Space-Invaders/internal/i8080"
)

const (
	memSize = 0x10000
	// tpa is where CP/M loads a .COM program (the Transient Program Area).
	tpa = 0x0100
	// bdosEntry is the BDOS entry point that programs CALL, with the
	// function number in C.
	bdosEntry = 0x0005
	// bdosBase is where the JMP at bdosEntry goes. The word at 0x0006 is
	// therefore the top of the TPA, which programs (the exercisers among
	// them) use to set their stack.
	bdosBase = 0xFE00

	opJMP = 0xC3
	opRET = 0xC9

	bdosConOut      = 2 // print the character in E
	bdosPrintString = 9 // print the string at DE, up to '$'
)

var (
	// ErrTooBig is returned when a program doesn't fit below the BDOS.
	ErrTooBig = errors.New("cpm: program does not fit in the TPA")
	// ErrUnsupported is returned when a program calls a BDOS function other
	// than 2 or 9.
	ErrUnsupported = errors.New("cpm: unsupported BDOS function")
)

// Machine is a CP/M machine running one program.
type Machine struct {
	// CPU is the machine's 8080, exported for tracing and inspection.
	CPU *i8080.CPU

	mem  [memSize]byte
	out  io.Writer
	buf  []byte // reused for BDOS output
	done bool
	err  error
}

// New returns a machine with program loaded at 0x0100 and the CPU about to
// run it. Console output is written to out.
func New(program []byte, out io.Writer) (*Machine, error) {
	if len(program) > bdosBase-tpa {
		return nil, fmt.Errorf("%w: %d bytes, at most %d", ErrTooBig, len(program), bdosBase-tpa)
	}

	m := &Machine{out: out}
	copy(m.mem[tpa:], program)
	m.mem[bdosEntry] = opJMP
	m.mem[bdosEntry+1] = byte(bdosBase & 0xFF)
	m.mem[bdosEntry+2] = byte(bdosBase >> 8)
	m.mem[bdosBase] = opRET
	m.CPU = i8080.New(m)
	m.CPU.PC = tpa

	return m, nil
}

// Read implements i8080.Bus.
func (m *Machine) Read(addr uint16) byte { return m.mem[addr] }

// Write implements i8080.Bus.
func (m *Machine) Write(addr uint16, v byte) { m.mem[addr] = v }

// In implements i8080.Bus. CP/M programs use no ports; all read 0xFF.
func (m *Machine) In(byte) byte { return 0xFF }

// Out implements i8080.Bus. Writes to ports are ignored.
func (m *Machine) Out(byte, byte) {}

// Done reports whether the program has ended with a warm boot.
func (m *Machine) Done() bool { return m.done }

// Step runs one instruction, handling a BDOS call or warm boot at PC first,
// and returns the cycles it took (0 once the program has ended or failed).
func (m *Machine) Step() int {
	if m.done || m.err != nil {
		return 0
	}

	switch m.CPU.PC {
	case 0:
		m.done = true
		return 0
	case bdosEntry:
		m.bdos()

		if m.err != nil {
			return 0
		}
	}

	return m.CPU.Step()
}

// Run steps until the program ends, fails, or maxCycles have run. It
// reports whether the program ended, and the first error: a write to the
// output that failed, or an unsupported BDOS call.
func (m *Machine) Run(maxCycles uint64) (bool, error) {
	start := m.CPU.Cycles
	for !m.done && m.err == nil && m.CPU.Cycles-start < maxCycles {
		m.Step()
	}

	return m.done, m.err
}

// bdos handles the BDOS call the CPU is about to make; the JMP and RET at
// the entry point then return to the caller.
func (m *Machine) bdos() {
	c := m.CPU
	m.buf = m.buf[:0]

	switch c.C {
	case bdosConOut:
		m.buf = append(m.buf, c.E)
	case bdosPrintString:
		for addr, n := c.DE(), 0; m.mem[addr] != '$' && n < memSize; addr, n = addr+1, n+1 {
			m.buf = append(m.buf, m.mem[addr])
		}
	default:
		caller := uint16(m.mem[c.SP+1])<<8 | uint16(m.mem[c.SP]) // the CALL's return address
		m.err = fmt.Errorf("%w %d, called before %04XH", ErrUnsupported, c.C, caller)

		return
	}

	if _, err := m.out.Write(m.buf); err != nil {
		m.err = fmt.Errorf("cpm: console output: %w", err)
	}
}
