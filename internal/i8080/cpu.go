package i8080

// Bus is what the CPU sees: memory and the I/O ports. The Space Invaders
// board and the CP/M harness implement it (DESIGN §6.1).
type Bus interface {
	Read(addr uint16) byte
	Write(addr uint16, v byte)
	In(port byte) byte
	Out(port, v byte)
}

const (
	// haltCycles is how far a halted CPU advances per Step: it keeps running
	// bus cycles (one 4-state machine cycle) until an interrupt wakes it.
	haltCycles = 4
	// interruptCycles is the cost of accepting an interrupt: the RST
	// instruction the board places on the data bus takes 11 states.
	interruptCycles = 11
)

// CPU is an Intel 8080. Registers and flags are exported so that the board,
// tests and tools can inspect them; F holds the real PSW byte (see FlagS ...).
type CPU struct {
	A, B, C, D, E, H, L, F byte
	SP, PC                 uint16

	// Halted is set by HLT and cleared when an interrupt is accepted.
	Halted bool
	// IntEnabled is the interrupt enable flip-flop (EI sets it, DI and an
	// accepted interrupt clear it).
	IntEnabled bool
	// Cycles counts the states run since Reset, including interrupts.
	Cycles uint64

	bus Bus
	// eiDelay is set by EI and cleared after the next instruction: until
	// then, interrupts are not accepted (DESIGN §6.3).
	eiDelay bool
	trace   traceRing
}

// New returns a CPU on bus, in the state after Reset.
func New(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()

	return c
}

// Reset puts the CPU in its power-on state: PC is 0, interrupts are disabled
// and the CPU is not halted (DESIGN §3). The real 8080 leaves the other
// registers undefined; here they are cleared, so every run is repeatable.
func (c *CPU) Reset() {
	c.A, c.B, c.C, c.D, c.E, c.H, c.L = 0, 0, 0, 0, 0, 0, 0
	c.F = flagOne
	c.SP, c.PC = 0, 0
	c.Halted, c.IntEnabled, c.eiDelay = false, false, false
	c.Cycles = 0
	c.trace.clear()
}

// Step runs one instruction and returns the cycles it took. A halted CPU
// runs no instruction and advances 4 cycles.
func (c *CPU) Step() int {
	if c.Halted {
		c.Cycles += haltCycles
		return haltCycles
	}

	c.eiDelay = false
	if c.trace.enabled() {
		c.trace.add(c.Snapshot())
	}

	op := c.fetch()
	info := &opTable[op]

	cycles := info.Cycles
	if c.execute(op) {
		cycles = info.CyclesTaken
	}

	c.Cycles += uint64(cycles)

	return cycles
}

// Run steps until at least cycles have run, and returns how many it
// ran past that (the overshoot), for the caller to carry into the next run.
// When cycles is zero or negative, nothing runs and -cycles is returned.
func (c *CPU) Run(cycles int) int {
	ran := 0
	for ran < cycles {
		ran += c.Step()
	}

	return ran - cycles
}

// Interrupt offers an interrupt with opcode on the data bus, which must be an
// RST instruction (0xC7 | n<<3). It reports whether the CPU accepted it: not
// while interrupts are disabled, nor in the instruction after EI. Accepting
// one disables interrupts, wakes a halted CPU and calls n*8, taking 11 cycles
// (added to Cycles). DESIGN §6.3.
func (c *CPU) Interrupt(opcode byte) bool {
	if !c.IntEnabled || c.eiDelay {
		return false
	}

	c.IntEnabled = false
	c.Halted = false
	c.push(c.PC)
	c.PC = uint16(opcode & 0x38)
	c.Cycles += interruptCycles

	return true
}

// BC returns the register pair BC.
func (c *CPU) BC() uint16 { return uint16(c.B)<<8 | uint16(c.C) }

// DE returns the register pair DE.
func (c *CPU) DE() uint16 { return uint16(c.D)<<8 | uint16(c.E) }

// HL returns the register pair HL.
func (c *CPU) HL() uint16 { return uint16(c.H)<<8 | uint16(c.L) }

func (c *CPU) setBC(v uint16) { c.B, c.C = byte(v>>8), byte(v) }
func (c *CPU) setDE(v uint16) { c.D, c.E = byte(v>>8), byte(v) }
func (c *CPU) setHL(v uint16) { c.H, c.L = byte(v>>8), byte(v) }

// fetch reads the byte at PC and advances PC.
func (c *CPU) fetch() byte {
	v := c.bus.Read(c.PC)
	c.PC++

	return v
}

// fetch16 reads the little-endian word at PC and advances PC by 2.
func (c *CPU) fetch16() uint16 {
	lo := c.fetch()
	return uint16(c.fetch())<<8 | uint16(lo)
}

func (c *CPU) read16(addr uint16) uint16 {
	return uint16(c.bus.Read(addr+1))<<8 | uint16(c.bus.Read(addr))
}

func (c *CPU) write16(addr, v uint16) {
	c.bus.Write(addr, byte(v))
	c.bus.Write(addr+1, byte(v>>8))
}

func (c *CPU) push(v uint16) {
	c.SP -= 2
	c.write16(c.SP, v)
}

func (c *CPU) pop() uint16 {
	v := c.read16(c.SP)
	c.SP += 2

	return v
}

// reg returns register r as encoded in an opcode: B C D E H L M A, where M is
// the memory byte at HL.
func (c *CPU) reg(r byte) byte {
	switch r {
	case 0:
		return c.B
	case 1:
		return c.C
	case 2:
		return c.D
	case 3:
		return c.E
	case 4:
		return c.H
	case 5:
		return c.L
	case 6:
		return c.bus.Read(c.HL())
	default:
		return c.A
	}
}

// setReg sets register r as encoded in an opcode (see reg).
func (c *CPU) setReg(r, v byte) {
	switch r {
	case 0:
		c.B = v
	case 1:
		c.C = v
	case 2:
		c.D = v
	case 3:
		c.E = v
	case 4:
		c.H = v
	case 5:
		c.L = v
	case 6:
		c.bus.Write(c.HL(), v)
	default:
		c.A = v
	}
}

// pair returns register pair rp as encoded in an opcode: BC, DE, HL, SP.
func (c *CPU) pair(rp byte) uint16 {
	switch rp {
	case 0:
		return c.BC()
	case 1:
		return c.DE()
	case 2:
		return c.HL()
	default:
		return c.SP
	}
}

// setPair sets register pair rp as encoded in an opcode (see pair).
func (c *CPU) setPair(rp byte, v uint16) {
	switch rp {
	case 0:
		c.setBC(v)
	case 1:
		c.setDE(v)
	case 2:
		c.setHL(v)
	default:
		c.SP = v
	}
}
