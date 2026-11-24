package i8080

import "testing"

// regs is the register state before or after a test instruction. F includes
// bit 1 (flagOne).
type regs struct {
	A, B, C, D, E, H, L, F byte
	SP                     uint16
}

func (c *CPU) regs() regs {
	return regs{A: c.A, B: c.B, C: c.C, D: c.D, E: c.E, H: c.H, L: c.L, F: c.F, SP: c.SP}
}

func (c *CPU) setRegs(r regs) {
	c.A, c.B, c.C, c.D, c.E, c.H, c.L, c.F, c.SP = r.A, r.B, r.C, r.D, r.E, r.H, r.L, r.F, r.SP
}

// opCase runs code at 0x0100 from the state before, and checks the registers
// after it, memory, the PC, the last OUT and the cycles of the first
// instruction.
type opCase struct {
	name    string
	code    []byte
	before  regs
	mem     map[uint16]byte // memory before
	in      byte            // what every input port returns
	after   regs
	wantMem map[uint16]byte
	wantPC  uint16 // 0: right after the instruction
	wantOut *[2]byte
}

const f1 = flagOne // every F value includes bit 1

// opCases covers the result and flags of every instruction. The cycle
// counts and control flow of all 256 opcodes are covered by
// TestEveryOpcodeFlowAndCycles, the flags of the ALU exhaustively in
// flags_test.go.
var opCases = []opCase{
	// Data transfer.
	{name: "MOV B,C", code: []byte{0x41}, before: regs{C: 0x5A, F: f1, SP: stackAt}, after: regs{B: 0x5A, C: 0x5A, F: f1, SP: stackAt}},
	{name: "MOV A,M", code: []byte{0x7E}, before: regs{H: 0x30, L: 0x01, F: f1, SP: stackAt}, mem: map[uint16]byte{0x3001: 0x99}, after: regs{A: 0x99, H: 0x30, L: 0x01, F: f1, SP: stackAt}},
	{name: "MOV M,E", code: []byte{0x73}, before: regs{E: 0x77, H: 0x30, F: f1, SP: stackAt}, after: regs{E: 0x77, H: 0x30, F: f1, SP: stackAt}, wantMem: map[uint16]byte{0x3000: 0x77}},
	{name: "MVI L", code: []byte{0x2E, 0xAB}, before: regs{F: f1, SP: stackAt}, after: regs{L: 0xAB, F: f1, SP: stackAt}},
	{name: "MVI M", code: []byte{0x36, 0xCD}, before: regs{H: 0x30, F: f1, SP: stackAt}, after: regs{H: 0x30, F: f1, SP: stackAt}, wantMem: map[uint16]byte{0x3000: 0xCD}},
	{name: "LXI B", code: []byte{0x01, 0x34, 0x12}, before: regs{F: f1, SP: stackAt}, after: regs{B: 0x12, C: 0x34, F: f1, SP: stackAt}},
	{name: "LXI D", code: []byte{0x11, 0x34, 0x12}, before: regs{F: f1, SP: stackAt}, after: regs{D: 0x12, E: 0x34, F: f1, SP: stackAt}},
	{name: "LXI H", code: []byte{0x21, 0x34, 0x12}, before: regs{F: f1, SP: stackAt}, after: regs{H: 0x12, L: 0x34, F: f1, SP: stackAt}},
	{name: "LXI SP", code: []byte{0x31, 0x34, 0x12}, before: regs{F: f1}, after: regs{F: f1, SP: 0x1234}},
	{name: "LDA", code: []byte{0x3A, 0x00, 0x30}, before: regs{F: f1, SP: stackAt}, mem: map[uint16]byte{0x3000: 0x42}, after: regs{A: 0x42, F: f1, SP: stackAt}},
	{name: "STA", code: []byte{0x32, 0x00, 0x30}, before: regs{A: 0x42, F: f1, SP: stackAt}, after: regs{A: 0x42, F: f1, SP: stackAt}, wantMem: map[uint16]byte{0x3000: 0x42}},
	{name: "LHLD", code: []byte{0x2A, 0x00, 0x30}, before: regs{F: f1, SP: stackAt}, mem: map[uint16]byte{0x3000: 0x34, 0x3001: 0x12}, after: regs{H: 0x12, L: 0x34, F: f1, SP: stackAt}},
	{name: "SHLD", code: []byte{0x22, 0x00, 0x30}, before: regs{H: 0x12, L: 0x34, F: f1, SP: stackAt}, after: regs{H: 0x12, L: 0x34, F: f1, SP: stackAt}, wantMem: map[uint16]byte{0x3000: 0x34, 0x3001: 0x12}},
	{name: "LDAX B", code: []byte{0x0A}, before: regs{B: 0x30, C: 0x05, F: f1, SP: stackAt}, mem: map[uint16]byte{0x3005: 0x66}, after: regs{A: 0x66, B: 0x30, C: 0x05, F: f1, SP: stackAt}},
	{name: "LDAX D", code: []byte{0x1A}, before: regs{D: 0x30, E: 0x06, F: f1, SP: stackAt}, mem: map[uint16]byte{0x3006: 0x67}, after: regs{A: 0x67, D: 0x30, E: 0x06, F: f1, SP: stackAt}},
	{name: "STAX B", code: []byte{0x02}, before: regs{A: 0x11, B: 0x30, C: 0x07, F: f1, SP: stackAt}, after: regs{A: 0x11, B: 0x30, C: 0x07, F: f1, SP: stackAt}, wantMem: map[uint16]byte{0x3007: 0x11}},
	{name: "STAX D", code: []byte{0x12}, before: regs{A: 0x22, D: 0x30, E: 0x08, F: f1, SP: stackAt}, after: regs{A: 0x22, D: 0x30, E: 0x08, F: f1, SP: stackAt}, wantMem: map[uint16]byte{0x3008: 0x22}},
	{name: "XCHG", code: []byte{0xEB}, before: regs{D: 1, E: 2, H: 3, L: 4, F: f1, SP: stackAt}, after: regs{D: 3, E: 4, H: 1, L: 2, F: f1, SP: stackAt}},

	// Arithmetic.
	{name: "ADD B", code: []byte{0x80}, before: regs{A: 0x2E, B: 0x74, F: f1, SP: stackAt}, after: regs{A: 0xA2, B: 0x74, F: f1 | FlagS | FlagAC, SP: stackAt}},
	{name: "ADD A carry", code: []byte{0x87}, before: regs{A: 0x80, F: f1, SP: stackAt}, after: regs{F: f1 | FlagZ | FlagP | FlagCY, SP: stackAt}},
	{name: "ADC C", code: []byte{0x89}, before: regs{A: 0x3D, C: 0x42, F: f1 | FlagCY, SP: stackAt}, after: regs{A: 0x80, C: 0x42, F: f1 | FlagS | FlagAC, SP: stackAt}},
	{name: "ADD M", code: []byte{0x86}, before: regs{A: 1, H: 0x30, F: f1, SP: stackAt}, mem: map[uint16]byte{0x3000: 2}, after: regs{A: 3, H: 0x30, F: f1 | FlagP, SP: stackAt}},
	{name: "ADI", code: []byte{0xC6, 0x01}, before: regs{A: 0xFF, F: f1, SP: stackAt}, after: regs{F: f1 | FlagZ | FlagP | FlagAC | FlagCY, SP: stackAt}},
	{name: "ACI", code: []byte{0xCE, 0x01}, before: regs{A: 0x01, F: f1 | FlagCY, SP: stackAt}, after: regs{A: 0x03, F: f1 | FlagP, SP: stackAt}},
	{name: "SUB D", code: []byte{0x92}, before: regs{A: 0x3E, D: 0x3E, F: f1, SP: stackAt}, after: regs{D: 0x3E, F: f1 | FlagZ | FlagP | FlagAC, SP: stackAt}},
	{name: "SUB borrow", code: []byte{0x92}, before: regs{A: 0x01, D: 0x02, F: f1, SP: stackAt}, after: regs{A: 0xFF, D: 0x02, F: f1 | FlagS | FlagP | FlagCY, SP: stackAt}},
	{name: "SBB E", code: []byte{0x9B}, before: regs{A: 0x04, E: 0x02, F: f1 | FlagCY, SP: stackAt}, after: regs{A: 0x01, E: 0x02, F: f1 | FlagAC, SP: stackAt}},
	{name: "SUI", code: []byte{0xD6, 0x01}, before: regs{A: 0x00, F: f1, SP: stackAt}, after: regs{A: 0xFF, F: f1 | FlagS | FlagP | FlagCY, SP: stackAt}},
	{name: "SBI", code: []byte{0xDE, 0x01}, before: regs{A: 0x03, F: f1 | FlagCY, SP: stackAt}, after: regs{A: 0x01, F: f1 | FlagAC, SP: stackAt}},
	{name: "INR A", code: []byte{0x3C}, before: regs{A: 0xFF, F: f1 | FlagCY, SP: stackAt}, after: regs{F: f1 | FlagZ | FlagP | FlagAC | FlagCY, SP: stackAt}},
	{name: "INR M", code: []byte{0x34}, before: regs{H: 0x30, F: f1, SP: stackAt}, mem: map[uint16]byte{0x3000: 0x0F}, after: regs{H: 0x30, F: f1 | FlagAC, SP: stackAt}, wantMem: map[uint16]byte{0x3000: 0x10}},
	{name: "DCR C", code: []byte{0x0D}, before: regs{C: 0x01, F: f1, SP: stackAt}, after: regs{F: f1 | FlagZ | FlagP | FlagAC, SP: stackAt}},
	{name: "DCR M", code: []byte{0x35}, before: regs{H: 0x30, F: f1 | FlagCY, SP: stackAt}, mem: map[uint16]byte{0x3000: 0x00}, after: regs{H: 0x30, F: f1 | FlagS | FlagP | FlagCY, SP: stackAt}, wantMem: map[uint16]byte{0x3000: 0xFF}},
	{name: "INX B", code: []byte{0x03}, before: regs{B: 0x00, C: 0xFF, F: f1, SP: stackAt}, after: regs{B: 0x01, C: 0x00, F: f1, SP: stackAt}},
	{name: "INX SP wraps", code: []byte{0x33}, before: regs{F: f1, SP: 0xFFFF}, after: regs{F: f1, SP: 0x0000}},
	{name: "DCX D", code: []byte{0x1B}, before: regs{F: f1, SP: stackAt}, after: regs{D: 0xFF, E: 0xFF, F: f1, SP: stackAt}},
	{name: "DCX H", code: []byte{0x2B}, before: regs{H: 0x10, F: f1, SP: stackAt}, after: regs{H: 0x0F, L: 0xFF, F: f1, SP: stackAt}},
	{name: "DCX SP", code: []byte{0x3B}, before: regs{F: f1, SP: stackAt}, after: regs{F: f1, SP: stackAt - 1}},
	{name: "DAD B", code: []byte{0x09}, before: regs{B: 0x33, C: 0x9F, H: 0xA1, L: 0x7B, F: f1 | FlagZ, SP: stackAt}, after: regs{B: 0x33, C: 0x9F, H: 0xD5, L: 0x1A, F: f1 | FlagZ, SP: stackAt}},
	{name: "DAD H carry", code: []byte{0x29}, before: regs{H: 0x80, L: 0x01, F: f1, SP: stackAt}, after: regs{H: 0x00, L: 0x02, F: f1 | FlagCY, SP: stackAt}},
	{name: "DAD D", code: []byte{0x19}, before: regs{D: 0x00, E: 0x01, H: 0x00, L: 0x01, F: f1 | FlagCY, SP: stackAt}, after: regs{D: 0x00, E: 0x01, L: 0x02, F: f1, SP: stackAt}},
	{name: "DAD SP", code: []byte{0x39}, before: regs{H: 0x00, L: 0x10, F: f1, SP: stackAt}, after: regs{H: 0x20, L: 0x10, F: f1, SP: stackAt}},
	{name: "DAA", code: []byte{0x27}, before: regs{A: 0x9B, F: f1, SP: stackAt}, after: regs{A: 0x01, F: f1 | FlagAC | FlagCY, SP: stackAt}},

	// Logic.
	{name: "ANA H", code: []byte{0xA4}, before: regs{A: 0xFC, H: 0x0F, F: f1 | FlagCY, SP: stackAt}, after: regs{A: 0x0C, H: 0x0F, F: f1 | FlagAC | FlagP, SP: stackAt}},
	{name: "ANI", code: []byte{0xE6, 0x0F}, before: regs{A: 0x3A, F: f1, SP: stackAt}, after: regs{A: 0x0A, F: f1 | FlagAC | FlagP, SP: stackAt}},
	{name: "XRA A", code: []byte{0xAF}, before: regs{A: 0x5C, F: f1 | FlagCY | FlagAC, SP: stackAt}, after: regs{F: f1 | FlagZ | FlagP, SP: stackAt}},
	{name: "XRI", code: []byte{0xEE, 0x81}, before: regs{A: 0x3B, F: f1, SP: stackAt}, after: regs{A: 0xBA, F: f1 | FlagS, SP: stackAt}},
	{name: "ORA L", code: []byte{0xB5}, before: regs{A: 0x33, L: 0x0F, F: f1 | FlagCY, SP: stackAt}, after: regs{A: 0x3F, L: 0x0F, F: f1 | FlagP, SP: stackAt}},
	{name: "ORI", code: []byte{0xF6, 0x80}, before: regs{A: 0x01, F: f1, SP: stackAt}, after: regs{A: 0x81, F: f1 | FlagS | FlagP, SP: stackAt}},
	{name: "CMP E equal", code: []byte{0xBB}, before: regs{A: 0x0A, E: 0x0A, F: f1, SP: stackAt}, after: regs{A: 0x0A, E: 0x0A, F: f1 | FlagZ | FlagP | FlagAC, SP: stackAt}},
	{name: "CMP E less", code: []byte{0xBB}, before: regs{A: 0x02, E: 0x05, F: f1, SP: stackAt}, after: regs{A: 0x02, E: 0x05, F: f1 | FlagS | FlagCY, SP: stackAt}},
	{name: "CPI", code: []byte{0xFE, 0x40}, before: regs{A: 0x4A, F: f1, SP: stackAt}, after: regs{A: 0x4A, F: f1 | FlagP | FlagAC, SP: stackAt}},
	{name: "RLC", code: []byte{0x07}, before: regs{A: 0xF2, F: f1, SP: stackAt}, after: regs{A: 0xE5, F: f1 | FlagCY, SP: stackAt}},
	{name: "RRC", code: []byte{0x0F}, before: regs{A: 0xF2, F: f1 | FlagCY, SP: stackAt}, after: regs{A: 0x79, F: f1, SP: stackAt}},
	{name: "RAL", code: []byte{0x17}, before: regs{A: 0xB5, F: f1, SP: stackAt}, after: regs{A: 0x6A, F: f1 | FlagCY, SP: stackAt}},
	{name: "RAR", code: []byte{0x1F}, before: regs{A: 0x6A, F: f1 | FlagCY, SP: stackAt}, after: regs{A: 0xB5, F: f1, SP: stackAt}},
	{name: "CMA", code: []byte{0x2F}, before: regs{A: 0x51, F: f1, SP: stackAt}, after: regs{A: 0xAE, F: f1, SP: stackAt}},
	{name: "STC", code: []byte{0x37}, before: regs{F: f1, SP: stackAt}, after: regs{F: f1 | FlagCY, SP: stackAt}},
	{name: "CMC", code: []byte{0x3F}, before: regs{F: f1 | FlagCY | FlagZ, SP: stackAt}, after: regs{F: f1 | FlagZ, SP: stackAt}},

	// Stack, I/O and machine control.
	{name: "PUSH B", code: []byte{0xC5}, before: regs{B: 0x12, C: 0x34, F: f1, SP: stackAt}, after: regs{B: 0x12, C: 0x34, F: f1, SP: stackAt - 2}, wantMem: map[uint16]byte{stackAt - 1: 0x12, stackAt - 2: 0x34}},
	{name: "PUSH D", code: []byte{0xD5}, before: regs{D: 0x56, E: 0x78, F: f1, SP: stackAt}, after: regs{D: 0x56, E: 0x78, F: f1, SP: stackAt - 2}, wantMem: map[uint16]byte{stackAt - 1: 0x56, stackAt - 2: 0x78}},
	{name: "PUSH H", code: []byte{0xE5}, before: regs{H: 0x9A, L: 0xBC, F: f1, SP: stackAt}, after: regs{H: 0x9A, L: 0xBC, F: f1, SP: stackAt - 2}, wantMem: map[uint16]byte{stackAt - 1: 0x9A, stackAt - 2: 0xBC}},
	{name: "PUSH PSW", code: []byte{0xF5}, before: regs{A: 0x1F, F: f1 | FlagS | FlagCY, SP: stackAt}, after: regs{A: 0x1F, F: f1 | FlagS | FlagCY, SP: stackAt - 2}, wantMem: map[uint16]byte{stackAt - 1: 0x1F, stackAt - 2: 0x83}},
	{name: "POP B", code: []byte{0xC1}, before: regs{F: f1, SP: stackAt}, mem: map[uint16]byte{stackAt: 0x34, stackAt + 1: 0x12}, after: regs{B: 0x12, C: 0x34, F: f1, SP: stackAt + 2}},
	{name: "POP D", code: []byte{0xD1}, before: regs{F: f1, SP: stackAt}, mem: map[uint16]byte{stackAt: 0x78, stackAt + 1: 0x56}, after: regs{D: 0x56, E: 0x78, F: f1, SP: stackAt + 2}},
	{name: "POP H", code: []byte{0xE1}, before: regs{F: f1, SP: stackAt}, mem: map[uint16]byte{stackAt: 0xBC, stackAt + 1: 0x9A}, after: regs{H: 0x9A, L: 0xBC, F: f1, SP: stackAt + 2}},
	// POP PSW: bits 3 and 5 of F read 0 and bit 1 reads 1, whatever was pushed.
	{name: "POP PSW fixed bits", code: []byte{0xF1}, before: regs{F: f1, SP: stackAt}, mem: map[uint16]byte{stackAt: 0xFF, stackAt + 1: 0x42}, after: regs{A: 0x42, F: 0xD7, SP: stackAt + 2}},
	{name: "POP PSW zero", code: []byte{0xF1}, before: regs{F: f1, SP: stackAt}, after: regs{F: f1, SP: stackAt + 2}},
	{name: "XTHL", code: []byte{0xE3}, before: regs{H: 0x0B, L: 0x3C, F: f1, SP: stackAt}, mem: map[uint16]byte{stackAt: 0xF0, stackAt + 1: 0x0D}, after: regs{H: 0x0D, L: 0xF0, F: f1, SP: stackAt}, wantMem: map[uint16]byte{stackAt: 0x3C, stackAt + 1: 0x0B}},
	{name: "SPHL", code: []byte{0xF9}, before: regs{H: 0x50, L: 0x6C, F: f1}, after: regs{H: 0x50, L: 0x6C, F: f1, SP: 0x506C}},
	{name: "IN", code: []byte{0xDB, 0x03}, in: 0x5A, before: regs{F: f1, SP: stackAt}, after: regs{A: 0x5A, F: f1, SP: stackAt}},
	{name: "OUT", code: []byte{0xD3, 0x04}, before: regs{A: 0xA5, F: f1, SP: stackAt}, after: regs{A: 0xA5, F: f1, SP: stackAt}, wantOut: &[2]byte{0x04, 0xA5}},
	{name: "NOP", code: []byte{0x00}, before: regs{A: 1, F: f1, SP: stackAt}, after: regs{A: 1, F: f1, SP: stackAt}},
	{name: "undocumented NOP", code: []byte{0x38}, before: regs{A: 1, F: f1, SP: stackAt}, after: regs{A: 1, F: f1, SP: stackAt}},

	// Branches: where they go is checked for every opcode elsewhere; here,
	// what they push and pop.
	{name: "CALL", code: []byte{0xCD, 0x00, 0x30}, before: regs{F: f1, SP: stackAt}, after: regs{F: f1, SP: stackAt - 2}, wantPC: 0x3000, wantMem: map[uint16]byte{stackAt - 1: 0x01, stackAt - 2: 0x03}},
	{name: "undocumented CALL", code: []byte{0xFD, 0x00, 0x30}, before: regs{F: f1, SP: stackAt}, after: regs{F: f1, SP: stackAt - 2}, wantPC: 0x3000, wantMem: map[uint16]byte{stackAt - 1: 0x01, stackAt - 2: 0x03}},
	{name: "CZ taken", code: []byte{0xCC, 0x00, 0x30}, before: regs{F: f1 | FlagZ, SP: stackAt}, after: regs{F: f1 | FlagZ, SP: stackAt - 2}, wantPC: 0x3000, wantMem: map[uint16]byte{stackAt - 1: 0x01, stackAt - 2: 0x03}},
	{name: "CZ not taken", code: []byte{0xCC, 0x00, 0x30}, before: regs{F: f1, SP: stackAt}, after: regs{F: f1, SP: stackAt}},
	{name: "RET", code: []byte{0xC9}, before: regs{F: f1, SP: stackAt}, mem: map[uint16]byte{stackAt: 0x00, stackAt + 1: 0x30}, after: regs{F: f1, SP: stackAt + 2}, wantPC: 0x3000},
	{name: "RNC not taken", code: []byte{0xD0}, before: regs{F: f1 | FlagCY, SP: stackAt}, after: regs{F: f1 | FlagCY, SP: stackAt}},
	{name: "RST 7", code: []byte{0xFF}, before: regs{F: f1, SP: stackAt}, after: regs{F: f1, SP: stackAt - 2}, wantPC: 0x0038, wantMem: map[uint16]byte{stackAt - 1: 0x01, stackAt - 2: 0x01}},
	{name: "PCHL", code: []byte{0xE9}, before: regs{H: 0x41, L: 0x3E, F: f1, SP: stackAt}, after: regs{H: 0x41, L: 0x3E, F: f1, SP: stackAt}, wantPC: 0x413E},
}

func TestOpcodes(t *testing.T) {
	t.Parallel()

	for _, tt := range opCases {
		c, bus := newTestCPU(tt.code...)
		for addr, v := range tt.mem {
			bus.mem[addr] = v
		}

		for i := range bus.in {
			bus.in[i] = tt.in
		}

		c.setRegs(tt.before)
		c.Step()

		if got := c.regs(); got != tt.after {
			t.Errorf("%s: registers\n got %+v\nwant %+v", tt.name, got, tt.after)
		}

		wantPC := tt.wantPC
		if wantPC == 0 {
			wantPC = codeAt + uint16(len(tt.code))
		}

		if c.PC != wantPC {
			t.Errorf("%s: PC = %04X, want %04X", tt.name, c.PC, wantPC)
		}

		for addr, want := range tt.wantMem {
			if got := bus.mem[addr]; got != want {
				t.Errorf("%s: mem[%04X] = %02X, want %02X", tt.name, addr, got, want)
			}
		}

		if tt.wantOut != nil && (bus.outs != 1 || bus.outPort != tt.wantOut[0] || bus.outVal != tt.wantOut[1]) {
			t.Errorf("%s: %d OUTs, last port %02X value %02X; want one, %02X %02X",
				tt.name, bus.outs, bus.outPort, bus.outVal, tt.wantOut[0], tt.wantOut[1])
		}
	}
}

// TestMemoryWraps: a word read or written at 0xFFFF wraps to 0x0000.
func TestMemoryWraps(t *testing.T) {
	t.Parallel()

	c, bus := newTestCPU(0x22, 0xFF, 0xFF) // SHLD 0FFFFH
	c.setHL(0x1234)
	c.Step()

	if bus.mem[0xFFFF] != 0x34 || bus.mem[0x0000] != 0x12 {
		t.Errorf("SHLD 0FFFFH wrote %02X at FFFF and %02X at 0000, want 34 and 12", bus.mem[0xFFFF], bus.mem[0])
	}
}
