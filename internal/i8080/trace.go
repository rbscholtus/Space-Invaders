package i8080

import "fmt"

// State is the CPU's state before an instruction: the registers, flags and
// the instruction's bytes. Trace returns the last ones run (DESIGN §3).
type State struct {
	PC, SP                 uint16
	Op                     [3]byte // the opcode and the two bytes after it
	A, F, B, C, D, E, H, L byte
	Cycles                 uint64
}

// Snapshot returns the CPU's current state, with the 3 bytes at PC.
func (c *CPU) Snapshot() State {
	return State{
		PC: c.PC, SP: c.SP,
		Op: [3]byte{c.bus.Read(c.PC), c.bus.Read(c.PC + 1), c.bus.Read(c.PC + 2)},
		A:  c.A, F: c.F, B: c.B, C: c.C, D: c.D, E: c.E, H: c.H, L: c.L,
		Cycles: c.Cycles,
	}
}

// String formats s as one trace line: address, instruction bytes,
// disassembly, registers, flags (a letter when set, "." when clear) and the
// cycle count before the instruction, e.g.
//
//	0100  C3 13 01  JMP 0113H     A=00 BC=0000 DE=0000 HL=0000 SP=0000 F=.Z.P. CYC=0
func (s State) String() string {
	text, n := Disassemble(s.Op[:], 0)

	var bytes string

	switch n {
	case 1:
		bytes = fmt.Sprintf("%02X      ", s.Op[0])
	case 2:
		bytes = fmt.Sprintf("%02X %02X   ", s.Op[0], s.Op[1])
	default:
		bytes = fmt.Sprintf("%02X %02X %02X", s.Op[0], s.Op[1], s.Op[2])
	}

	return fmt.Sprintf("%04X  %s  %-12s  A=%02X BC=%02X%02X DE=%02X%02X HL=%02X%02X SP=%04X F=%s CYC=%d",
		s.PC, bytes, text, s.A, s.B, s.C, s.D, s.E, s.H, s.L, s.SP, flagString(s.F), s.Cycles)
}

// flagString shows S Z AC P CY as letters when set and dots when clear.
func flagString(f byte) string {
	b := []byte(".....")

	for i, flag := range [...]struct {
		bit    byte
		letter byte
	}{{FlagS, 'S'}, {FlagZ, 'Z'}, {FlagAC, 'A'}, {FlagP, 'P'}, {FlagCY, 'C'}} {
		if f&flag.bit != 0 {
			b[i] = flag.letter
		}
	}

	return string(b)
}

// EnableTrace keeps the state before each of the last n instructions, in a
// ring buffer allocated here, once. n <= 0 turns tracing off.
func (c *CPU) EnableTrace(n int) {
	if n <= 0 {
		c.trace = traceRing{}
		return
	}

	c.trace = traceRing{buf: make([]State, n)}
}

// Trace returns the traced states, oldest first, in a new slice.
func (c *CPU) Trace() []State {
	return c.trace.states()
}

// traceRing is a fixed-size ring of the last states.
type traceRing struct {
	buf  []State
	next int // where the next state goes
	full bool
}

func (t *traceRing) enabled() bool { return t.buf != nil }

func (t *traceRing) add(s State) {
	t.buf[t.next] = s

	t.next++
	if t.next == len(t.buf) {
		t.next = 0
		t.full = true
	}
}

func (t *traceRing) clear() {
	t.next, t.full = 0, false
}

func (t *traceRing) states() []State {
	if !t.full {
		return append([]State(nil), t.buf[:t.next]...)
	}

	return append(append([]State(nil), t.buf[t.next:]...), t.buf[:t.next]...)
}
