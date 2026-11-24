package i8080

import "math/bits"

// Flag bits in F, the low byte of the PSW (Intel 8080 Assembly Language
// Programming Manual, "Program Status Word"): S Z 0 AC 0 P 1 CY. F is kept as
// that byte, so PUSH PSW and POP PSW need no conversion (DESIGN §6.1).
const (
	FlagS  byte = 0x80 // sign: bit 7 of the result
	FlagZ  byte = 0x40 // zero
	FlagAC byte = 0x10 // auxiliary carry: carry out of bit 3
	FlagP  byte = 0x04 // parity: set when the result has an even number of 1 bits
	FlagCY byte = 0x01 // carry (or borrow, for subtraction)

	flagOne   byte = 0x02                                    // bit 1 always reads 1
	flagsUsed byte = FlagS | FlagZ | FlagAC | FlagP | FlagCY // bits 3 and 5 always read 0
)

// szp returns the S, Z and P flags for a result, with bit 1 set.
func szp(v byte) byte {
	f := v&FlagS | flagOne
	if v == 0 {
		f |= FlagZ
	}

	if bits.OnesCount8(v)&1 == 0 {
		f |= FlagP
	}

	return f
}

// add sets A to A + v + carry, with all flags. AC is the carry out of bit 3.
func (c *CPU) add(v, carry byte) {
	sum := uint16(c.A) + uint16(v) + uint16(carry)
	r := byte(sum)

	f := szp(r) | (c.A^v^r)&FlagAC
	if sum > 0xFF {
		f |= FlagCY
	}

	c.A, c.F = r, f
}

// sub returns A - v - borrow and sets all flags; A is unchanged (CMP uses it
// as is). The 8080 subtracts by adding the complement of v (and of borrow), so
// AC is the carry out of bit 3 of that addition, and CY is set on a borrow
// (DESIGN §3).
func (c *CPU) sub(v, borrow byte) byte {
	diff := uint16(c.A) - uint16(v) - uint16(borrow)
	r := byte(diff)

	f := szp(r) | ^(c.A^v^r)&FlagAC
	if diff > 0xFF {
		f |= FlagCY
	}

	c.F = f

	return r
}

// ana sets A to A & v. CY is cleared; AC is bit 3 of A | v (DESIGN §3).
func (c *CPU) ana(v byte) {
	ac := (c.A | v) << 1 & FlagAC
	c.A &= v
	c.F = szp(c.A) | ac
}

// xra sets A to A ^ v, clearing CY and AC (DESIGN §3).
func (c *CPU) xra(v byte) {
	c.A ^= v
	c.F = szp(c.A)
}

// ora sets A to A | v, clearing CY and AC (DESIGN §3).
func (c *CPU) ora(v byte) {
	c.A |= v
	c.F = szp(c.A)
}

// inr returns v + 1 and sets S, Z, P and AC; CY is unchanged. AC is set when
// the low nibble carries, i.e. when the result's low nibble is 0.
func (c *CPU) inr(v byte) byte {
	r := v + 1

	f := szp(r) | c.F&FlagCY
	if r&0x0F == 0 {
		f |= FlagAC
	}

	c.F = f

	return r
}

// dcr returns v - 1 and sets S, Z, P and AC; CY is unchanged. Like SUB, the
// 8080 adds 0xFF, so AC is set when the low nibble does not borrow: when the
// result's low nibble is not 0xF (DESIGN §3).
func (c *CPU) dcr(v byte) byte {
	r := v - 1

	f := szp(r) | c.F&FlagCY
	if r&0x0F != 0x0F {
		f |= FlagAC
	}

	c.F = f

	return r
}

// daa adjusts A to two BCD digits after an addition (DESIGN §3). The low digit
// gets +6 when it is over 9 or AC is set; the high digit gets +0x60 when A is
// over 0x99 or CY is set. CY is then set if the high digit was adjusted and is
// otherwise unchanged; AC is the carry out of bit 3 of the adjustment.
func (c *CPU) daa() {
	var corr byte

	cy := c.F & FlagCY
	if c.A&0x0F > 9 || c.F&FlagAC != 0 {
		corr = 0x06
	}

	if c.A > 0x99 || cy != 0 {
		corr |= 0x60
		cy = FlagCY
	}

	r := c.A + corr
	c.F = szp(r) | (c.A^corr^r)&FlagAC | cy
	c.A = r
}

// cond reports whether condition cc (bits 3–5 of a Jcc, Ccc or Rcc opcode)
// holds: NZ, Z, NC, C, PO, PE, P, M.
func (c *CPU) cond(cc byte) bool {
	var set bool

	switch cc >> 1 {
	case 0:
		set = c.F&FlagZ != 0
	case 1:
		set = c.F&FlagCY != 0
	case 2:
		set = c.F&FlagP != 0
	default:
		set = c.F&FlagS != 0
	}

	return set == (cc&1 != 0)
}
