package i8080

import (
	"strings"
	"testing"
)

// testBus is 64 KB of RAM and 256 ports; it records the last OUT.
type testBus struct {
	mem     [0x10000]byte
	in      [256]byte
	outPort byte
	outVal  byte
	outs    int
}

func (b *testBus) Read(addr uint16) byte     { return b.mem[addr] }
func (b *testBus) Write(addr uint16, v byte) { b.mem[addr] = v }
func (b *testBus) In(port byte) byte         { return b.in[port] }
func (b *testBus) Out(port, v byte) {
	b.outPort, b.outVal = port, v
	b.outs++
}

const (
	codeAt  = 0x0100 // where test code is loaded
	stackAt = 0x2000 // SP in tests
)

// newTestCPU returns a CPU with code at 0x0100, PC there and SP at 0x2000.
func newTestCPU(code ...byte) (*CPU, *testBus) {
	bus := &testBus{}
	copy(bus.mem[codeAt:], code)
	c := New(bus)
	c.PC, c.SP = codeAt, stackAt

	return c, bus
}

// condTaken evaluates a condition suffix (NZ, Z, NC, C, PO, PE, P, M) the way
// the Intel manual defines it, independently of CPU.cond.
func condTaken(cc string, f byte) bool {
	switch cc {
	case "NZ":
		return f&FlagZ == 0
	case "Z":
		return f&FlagZ != 0
	case "NC":
		return f&FlagCY == 0
	case "C":
		return f&FlagCY != 0
	case "PO":
		return f&FlagP == 0
	case "PE":
		return f&FlagP != 0
	case "P":
		return f&FlagS == 0
	case "M":
		return f&FlagS != 0
	}

	panic("bad condition " + cc)
}

// expectFlow returns the PC and cycles expected after running op with flags
// f, from the opcode table's mnemonic: where it jumps, calls or returns to,
// and whether a conditional call or return is taken.
func expectFlow(op, f byte) (uint16, bool) {
	const (
		target = 0x1234 // the operand bytes after the opcode
		retTo  = 0x4321 // the word on the stack
		hl     = 0x3000
	)

	info := Op(op)
	next := uint16(codeAt + info.Length)

	name, _, _ := strings.Cut(info.Mnemonic, " ")
	switch {
	case name == "JMP" || name == "CALL":
		return target, false
	case name == "RET":
		return retTo, false
	case name == "PCHL":
		return hl, false
	case name == "RST":
		return uint16(op & 0x38), false
	case op&0xC7 == 0xC2: // Jcc
		if condTaken(name[1:], f) {
			return target, false
		}
	case op&0xC7 == 0xC4: // Ccc
		if condTaken(name[1:], f) {
			return target, true
		}
	case op&0xC7 == 0xC0: // Rcc
		if condTaken(name[1:], f) {
			return retTo, true
		}
	}

	return next, false
}

// TestEveryOpcodeFlowAndCycles runs each of the 256 opcodes with all flags
// clear and all set, and checks the PC after it and its cycle count.
func TestEveryOpcodeFlowAndCycles(t *testing.T) {
	t.Parallel()

	for op := range 256 {
		for _, f := range []byte{flagOne, flagsUsed | flagOne} {
			c, bus := newTestCPU(byte(op), 0x34, 0x12)
			c.F = f
			c.setHL(0x3000)

			bus.mem[stackAt], bus.mem[stackAt+1] = 0x21, 0x43
			cycles := c.Step()

			wantPC, taken := expectFlow(byte(op), f)

			wantCycles := Op(byte(op)).Cycles
			if taken {
				wantCycles = Op(byte(op)).CyclesTaken
			}

			if c.PC != wantPC {
				t.Errorf("%02X %s (F=%02X): PC = %04X, want %04X", op, Op(byte(op)).Mnemonic, f, c.PC, wantPC)
			}

			if cycles != wantCycles || c.Cycles != uint64(wantCycles) {
				t.Errorf("%02X %s (F=%02X): %d cycles (total %d), want %d",
					op, Op(byte(op)).Mnemonic, f, cycles, c.Cycles, wantCycles)
			}
		}
	}
}

// TestOpcodeTable checks the table against the Intel manual's counts for a
// few instructions whose counts are easy to get wrong.
func TestOpcodeTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		op                  byte
		mnemonic            string
		length, cyc, cycTkn int
	}{
		{0x00, "NOP", 1, 4, 4},
		{0x34, "INR M", 1, 10, 10},
		{0x36, "MVI M,d8", 2, 10, 10},
		{0x41, "MOV B,C", 1, 5, 5},
		{0x46, "MOV B,M", 1, 7, 7},
		{0x76, "HLT", 1, 7, 7},
		{0xC0, "RNZ", 1, 5, 11},
		{0xC2, "JNZ a16", 3, 10, 10},
		{0xC4, "CNZ a16", 3, 11, 17},
		{0xC5, "PUSH B", 1, 11, 11},
		{0xE3, "XTHL", 1, 18, 18},
		{0xEB, "XCHG", 1, 4, 4},
		{0xF1, "POP PSW", 1, 10, 10},
		{0x22, "SHLD a16", 3, 16, 16},
		{0x32, "STA a16", 3, 13, 13},
	}
	for _, tt := range tests {
		got := Op(tt.op)
		if got.Mnemonic != tt.mnemonic || got.Length != tt.length || got.Cycles != tt.cyc || got.CyclesTaken != tt.cycTkn {
			t.Errorf("Op(%02X) = %+v, want %s %d %d/%d", tt.op, got, tt.mnemonic, tt.length, tt.cyc, tt.cycTkn)
		}
	}

	undocumented := 0

	for op := range 256 {
		if Op(byte(op)).Undocumented {
			undocumented++
		}
	}

	if undocumented != 12 {
		t.Errorf("%d undocumented opcodes, want 12 (DESIGN §6.1)", undocumented)
	}
}

// --- The CPU bugs of the Java version (DESIGN §3) ---

func TestInterruptIgnoredWhenDisabled(t *testing.T) {
	t.Parallel()

	c, _ := newTestCPU(0x00)
	if c.Interrupt(0xCF) {
		t.Fatal("interrupt accepted with interrupts disabled")
	}

	if c.PC != codeAt || c.SP != stackAt || c.Cycles != 0 {
		t.Errorf("rejected interrupt changed state: PC=%04X SP=%04X cycles=%d", c.PC, c.SP, c.Cycles)
	}
}

func TestInterruptAccepted(t *testing.T) {
	t.Parallel()

	c, bus := newTestCPU(0xFB, 0x00) // EI; NOP
	c.Step()
	c.Step()

	if !c.Interrupt(0xD7) { // RST 2
		t.Fatal("interrupt refused with interrupts enabled")
	}

	if c.PC != 0x10 || c.IntEnabled || c.Cycles != 4+4+11 {
		t.Errorf("PC=%04X IntEnabled=%v cycles=%d, want 0010 false 19", c.PC, c.IntEnabled, c.Cycles)
	}

	if ret := uint16(bus.mem[c.SP+1])<<8 | uint16(bus.mem[c.SP]); ret != codeAt+2 {
		t.Errorf("pushed %04X, want %04X", ret, codeAt+2)
	}

	if c.Interrupt(0xD7) {
		t.Error("a second interrupt was accepted; accepting one must disable interrupts")
	}
}

func TestHLTWakesOnInterrupt(t *testing.T) {
	t.Parallel()

	c, _ := newTestCPU(0xFB, 0x76) // EI; HLT
	c.Step()
	c.Step()

	if !c.Halted || c.PC != codeAt+2 {
		t.Fatalf("Halted=%v PC=%04X after HLT", c.Halted, c.PC)
	}

	if n := c.Step(); n != haltCycles || c.PC != codeAt+2 {
		t.Errorf("halted Step: %d cycles, PC=%04X; want %d, %04X", n, c.PC, haltCycles, codeAt+2)
	}

	if !c.Interrupt(0xCF) {
		t.Fatal("interrupt refused while halted")
	}

	if c.Halted || c.PC != 0x08 {
		t.Errorf("after interrupt: Halted=%v PC=%04X, want false 0008", c.Halted, c.PC)
	}
}

func TestResetDisablesInterrupts(t *testing.T) {
	t.Parallel()

	c, _ := newTestCPU(0xFB, 0x76)
	c.Step()
	c.Step()
	c.Reset()

	if c.IntEnabled || c.Halted || c.PC != 0 || c.Cycles != 0 {
		t.Errorf("after Reset: IntEnabled=%v Halted=%v PC=%04X Cycles=%d", c.IntEnabled, c.Halted, c.PC, c.Cycles)
	}

	if c.F != flagOne {
		t.Errorf("after Reset: F=%02X, want %02X", c.F, flagOne)
	}
}

// TestEIDelay: EI takes effect after the next instruction, so EI; RET returns
// before an interrupt is taken (DESIGN §6.3).
func TestEIDelay(t *testing.T) {
	t.Parallel()

	c, _ := newTestCPU(0xFB, 0xC9) // EI; RET
	c.Step()

	if !c.IntEnabled {
		t.Fatal("EI did not set IntEnabled")
	}

	if c.Interrupt(0xCF) {
		t.Fatal("interrupt accepted directly after EI")
	}

	c.Step()

	if !c.Interrupt(0xCF) {
		t.Fatal("interrupt refused after the instruction after EI")
	}
}

func TestDI(t *testing.T) {
	t.Parallel()

	c, _ := newTestCPU(0xFB, 0x00, 0xF3) // EI; NOP; DI
	c.Step()
	c.Step()
	c.Step()

	if c.IntEnabled || c.Interrupt(0xCF) {
		t.Error("DI did not disable interrupts")
	}
}

func TestDAAHighDigit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, f, wantA, wantF byte
	}{
		// 0x9A: the low digit +6 carries into the high digit, which then gets +0x60.
		{0x9A, flagOne, 0x00, flagOne | FlagZ | FlagP | FlagAC | FlagCY},
		// 0x0F: +6 carries into the high digit: 0x15.
		{0x0F, flagOne, 0x15, flagOne | FlagAC},
		// 0xA0: the high digit is over 9: +0x60 sets CY; A stays 8-bit.
		{0xA0, flagOne, 0x00, flagOne | FlagZ | FlagP | FlagCY},
		// CY in: +0x60, CY stays set.
		{0x12, flagOne | FlagCY, 0x72, flagOne | FlagP | FlagCY},
		// Already BCD, no flags: unchanged, CY stays clear.
		{0x42, flagOne, 0x42, flagOne | FlagP},
	}
	for _, tt := range tests {
		c, _ := newTestCPU(0x27)
		c.A, c.F = tt.a, tt.f
		c.Step()

		if c.A != tt.wantA || c.F != tt.wantF {
			t.Errorf("DAA A=%02X F=%02X: A=%02X F=%02X, want A=%02X F=%02X", tt.a, tt.f, c.A, c.F, tt.wantA, tt.wantF)
		}
	}
}

func TestDCRAuxCarry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		b      byte
		wantAC bool
	}{
		{0x10, false}, // borrow from bit 4: low nibble becomes 0xF
		{0x11, true},
		{0x00, false},
		{0x0F, true},
	}
	for _, tt := range tests {
		c, _ := newTestCPU(0x05) // DCR B
		c.B = tt.b
		c.Step()

		if got := c.F&FlagAC != 0; got != tt.wantAC {
			t.Errorf("DCR B=%02X: AC=%v, want %v", tt.b, got, tt.wantAC)
		}
	}
}

func TestSUBAuxCarry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		op         byte
		a, b       byte
		carry      bool
		wantAC     bool
		wantResult byte
	}{
		{0x90, 0x10, 0x01, false, false, 0x0F}, // SUB B: borrow from bit 4 → no carry out of bit 3
		{0x90, 0x11, 0x01, false, true, 0x10},
		{0x90, 0x00, 0x00, false, true, 0x00}, // 0 + 0xFF + 1: carry out of bit 3
		{0x98, 0x11, 0x00, true, true, 0x10},  // SBB B with borrow
		{0x98, 0x10, 0x00, true, false, 0x0F},
		{0xB8, 0x11, 0x01, false, true, 0x11}, // CMP B: A unchanged
	}
	for _, tt := range tests {
		c, _ := newTestCPU(tt.op)

		c.A, c.B = tt.a, tt.b
		if tt.carry {
			c.F |= FlagCY
		}

		c.Step()

		if got := c.F&FlagAC != 0; got != tt.wantAC || c.A != tt.wantResult {
			t.Errorf("%s A=%02X B=%02X CY=%v: A=%02X AC=%v, want A=%02X AC=%v",
				Op(tt.op).Mnemonic, tt.a, tt.b, tt.carry, c.A, got, tt.wantResult, tt.wantAC)
		}
	}
}

func TestANAAuxCarry(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		a, b   byte
		wantAC bool
	}{{0x08, 0x00, true}, {0x00, 0x08, true}, {0xF7, 0xF7, false}} {
		c, _ := newTestCPU(0xA0) // ANA B
		c.A, c.B = tt.a, tt.b
		c.F |= FlagCY
		c.Step()

		if got := c.F&FlagAC != 0; got != tt.wantAC || c.F&FlagCY != 0 {
			t.Errorf("ANA A=%02X B=%02X: AC=%v CY=%v, want AC=%v CY=false", tt.a, tt.b, got, c.F&FlagCY != 0, tt.wantAC)
		}
	}
}

func TestXRAORAClearAuxCarry(t *testing.T) {
	t.Parallel()

	for _, op := range []byte{0xA8, 0xB0, 0xEE, 0xF6} { // XRA B, ORA B, XRI, ORI
		c, _ := newTestCPU(op, 0x0F)
		c.A, c.B = 0x0F, 0x0F
		c.F = flagsUsed | flagOne
		c.Step()

		if c.F&(FlagAC|FlagCY) != 0 {
			t.Errorf("%s: F=%02X, want AC and CY clear", Op(op).Mnemonic, c.F)
		}
	}
}

func TestRunCarriesOvershoot(t *testing.T) {
	t.Parallel()

	c, _ := newTestCPU() // all NOPs: 4 cycles each
	if over := c.Run(10); over != 2 || c.Cycles != 12 {
		t.Errorf("Run(10) overshoot %d, cycles %d; want 2, 12", over, c.Cycles)
	}
	// The caller carries the overshoot: the next run is 2 cycles shorter.
	if over := c.Run(10 - 2); over != 0 || c.Cycles != 20 {
		t.Errorf("Run(8) overshoot %d, cycles %d; want 0, 20", over, c.Cycles)
	}

	if over := c.Run(-3); over != 3 || c.Cycles != 20 {
		t.Errorf("Run(-3) overshoot %d, cycles %d; want 3, 20", over, c.Cycles)
	}
}

func TestTrace(t *testing.T) {
	t.Parallel()

	c, _ := newTestCPU(0x3E, 0x12, 0x06, 0x34, 0x00, 0xC3, 0x00, 0x01) // MVI A; MVI B; NOP; JMP 0100H
	if got := c.Trace(); len(got) != 0 {
		t.Fatalf("trace before EnableTrace: %v", got)
	}

	c.EnableTrace(3)

	for range 4 {
		c.Step()
	}

	got := c.Trace()
	if len(got) != 3 || got[0].PC != 0x0102 || got[2].PC != 0x0105 {
		t.Fatalf("trace = %+v, want the last 3 instructions from 0102", got)
	}

	want := "0105  C3 00 01  JMP 0100H     A=12 BC=3400 DE=0000 HL=0000 SP=2000 F=..... CYC=18"
	if s := got[2].String(); s != want {
		t.Errorf("String() =\n%q, want\n%q", s, want)
	}

	if s := got[0].String(); !strings.HasPrefix(s, "0102  06 34     MVI B,34H") {
		t.Errorf("String() = %q", s)
	}

	if s := got[1].String(); !strings.HasPrefix(s, "0104  00        NOP") {
		t.Errorf("String() = %q", s)
	}

	c.Step()
	c.Reset()

	if got := c.Trace(); len(got) != 0 {
		t.Errorf("trace after Reset has %d entries, want 0", len(got))
	}

	c.EnableTrace(0)
	c.Step()

	if got := c.Trace(); len(got) != 0 {
		t.Errorf("trace after EnableTrace(0) has %d entries, want 0", len(got))
	}
}

func TestFlagString(t *testing.T) {
	t.Parallel()

	if got := flagString(flagsUsed | flagOne); got != "SZAPC" {
		t.Errorf("all set: %q", got)
	}

	if got := flagString(FlagZ | FlagCY); got != ".Z..C" {
		t.Errorf("Z and CY: %q", got)
	}
}

func BenchmarkStep(b *testing.B) {
	// A loop of register moves, arithmetic, memory access and a jump.
	c, _ := newTestCPU(
		0x21, 0x00, 0x30, // LXI H,3000H
		0x7E,             // MOV A,M
		0x80,             // ADD B
		0x77,             // MOV M,A
		0x23,             // INX H
		0x05,             // DCR B
		0xC3, 0x03, 0x01, // JMP 0103H
	)
	for b.Loop() {
		c.Step()
	}
}
