package i8080

import "testing"

// The reference functions below follow the Intel 8080 Assembly Language
// Programming Manual step by step, on ints, and differently from the
// implementation in flags.go, so that both can't share a mistake.

func refParity(v int) bool {
	ones := 0
	for i := range 8 {
		ones += v >> i & 1
	}

	return ones%2 == 0
}

func refFlags(r int, ac, cy bool) byte {
	f := flagOne
	if r&0x80 != 0 {
		f |= FlagS
	}

	if r&0xFF == 0 {
		f |= FlagZ
	}

	if ac {
		f |= FlagAC
	}

	if refParity(r & 0xFF) {
		f |= FlagP
	}

	if cy {
		f |= FlagCY
	}

	return f
}

// refAdd is ADD/ADC: the sum of A, the operand and the carry.
func refAdd(a, b, carry int) (byte, byte) {
	r := a + b + carry
	ac := a&0xF+b&0xF+carry > 0xF

	return byte(r), refFlags(r, ac, r > 0xFF)
}

// refSub is SUB/SBB/CMP. The manual describes subtraction as adding the
// two's complement of the operand (and of the borrow): CY is the inverse of
// the carry out of that addition, AC the plain carry out of bit 3.
func refSub(a, b, borrow int) (byte, byte) {
	nb := ^b & 0xFF
	in := 1 - borrow
	sum := a + nb + in
	ac := a&0xF+nb&0xF+in > 0xF

	return byte(sum), refFlags(sum, ac, sum <= 0xFF)
}

// TestALUExhaustive compares ADD, ADC, SUB, SBB and CMP with the reference
// for all 256 × 256 × 2 inputs (DESIGN §15.1).
func TestALUExhaustive(t *testing.T) {
	t.Parallel()

	c, bus := newTestCPU()

	for a := range 256 {
		for b := range 256 {
			for cy := range 2 {
				for _, op := range []byte{0x80, 0x88, 0x90, 0x98, 0xB8} { // ADD ADC SUB SBB CMP B
					bus.mem[codeAt] = op
					c.PC, c.A, c.B, c.F = codeAt, byte(a), byte(b), flagOne|byte(cy)
					c.Step()

					var wantA, wantF byte

					switch op {
					case 0x80:
						wantA, wantF = refAdd(a, b, 0)
					case 0x88:
						wantA, wantF = refAdd(a, b, cy)
					case 0x90:
						wantA, wantF = refSub(a, b, 0)
					case 0x98:
						wantA, wantF = refSub(a, b, cy)
					default:
						_, wantF = refSub(a, b, 0)
						wantA = byte(a)
					}

					if c.A != wantA || c.F != wantF {
						t.Fatalf("%s A=%02X B=%02X CY=%d: A=%02X F=%02X, want A=%02X F=%02X",
							Op(op).Mnemonic, a, b, cy, c.A, c.F, wantA, wantF)
					}
				}
			}
		}
	}
}

// TestLogicExhaustive compares ANA, XRA and ORA with the reference for all
// 256 × 256 inputs.
func TestLogicExhaustive(t *testing.T) {
	t.Parallel()

	c, bus := newTestCPU()

	for a := range 256 {
		for b := range 256 {
			for _, op := range []byte{0xA0, 0xA8, 0xB0} { // ANA XRA ORA B
				bus.mem[codeAt] = op
				c.PC, c.A, c.B, c.F = codeAt, byte(a), byte(b), flagsUsed|flagOne
				c.Step()

				var r int

				ac := false

				switch op {
				case 0xA0:
					r = a & b
					ac = (a|b)&0x08 != 0
				case 0xA8:
					r = a ^ b
				default:
					r = a | b
				}

				if wantF := refFlags(r, ac, false); c.A != byte(r) || c.F != wantF {
					t.Fatalf("%s A=%02X B=%02X: A=%02X F=%02X, want A=%02X F=%02X",
						Op(op).Mnemonic, a, b, c.A, c.F, r, wantF)
				}
			}
		}
	}
}

// TestINRDCRExhaustive checks INR and DCR for every value, with CY both ways.
func TestINRDCRExhaustive(t *testing.T) {
	t.Parallel()

	c, bus := newTestCPU()

	for v := range 256 {
		for cy := range 2 {
			bus.mem[codeAt] = 0x04 // INR B
			c.PC, c.B, c.F = codeAt, byte(v), flagOne|byte(cy)
			c.Step()

			r := v + 1
			if want := refFlags(r, v&0xF+1 > 0xF, cy == 1); c.B != byte(r) || c.F != want {
				t.Fatalf("INR B=%02X CY=%d: B=%02X F=%02X, want %02X %02X", v, cy, c.B, c.F, byte(r), want)
			}

			bus.mem[codeAt] = 0x05 // DCR B: adds 0xFF, AC is the carry out of bit 3
			c.PC, c.B, c.F = codeAt, byte(v), flagOne|byte(cy)
			c.Step()

			r = v + 0xFF
			if want := refFlags(r, v&0xF+0xF > 0xF, cy == 1); c.B != byte(r) || c.F != want {
				t.Fatalf("DCR B=%02X CY=%d: B=%02X F=%02X, want %02X %02X", v, cy, c.B, c.F, byte(r), want)
			}
		}
	}
}

// refDAA follows the Intel manual's two steps literally: adjust the low
// digit, then look at the high digit of the result.
func refDAA(a int, ac, cy bool) (byte, byte) {
	newAC := false
	if a&0xF > 9 || ac {
		newAC = a&0xF+6 > 0xF
		a += 6
	}

	if a>>4&0xF > 9 || cy || a > 0xFF {
		a += 0x60
		cy = true
	}

	return byte(a), refFlags(a, newAC, cy)
}

// TestDAAExhaustive checks DAA for every A, AC and CY (DESIGN §15.1).
func TestDAAExhaustive(t *testing.T) {
	t.Parallel()

	c, bus := newTestCPU()
	bus.mem[codeAt] = 0x27

	for a := range 256 {
		for _, f := range []byte{0, FlagAC, FlagCY, FlagAC | FlagCY} {
			c.PC, c.A, c.F = codeAt, byte(a), flagOne|f
			c.Step()

			wantA, wantF := refDAA(a, f&FlagAC != 0, f&FlagCY != 0)
			if c.A != wantA || c.F != wantF {
				t.Fatalf("DAA A=%02X F=%02X: A=%02X F=%02X, want A=%02X F=%02X", a, flagOne|f, c.A, c.F, wantA, wantF)
			}
		}
	}
}

// TestConditions checks every condition against every flag combination.
func TestConditions(t *testing.T) {
	t.Parallel()

	names := []string{"NZ", "Z", "NC", "C", "PO", "PE", "P", "M"}

	c, _ := newTestCPU()
	for f := range 256 {
		c.F = byte(f)
		for cc, name := range names {
			if got, want := c.cond(byte(cc)), condTaken(name, byte(f)); got != want {
				t.Errorf("cond %s with F=%02X = %v, want %v", name, f, got, want)
			}
		}
	}
}
