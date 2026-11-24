package i8080

// Opcodes are decoded by their fields, as laid out in the Intel manual:
// bits 6–7 select the group, bits 3–5 a destination register, register pair
// or condition (ddd), bits 0–2 a source register or operation (sss).

const opHLT = 0x76

// execute runs opcode op, whose byte has been fetched. It reports whether a
// conditional call or return was taken, which costs extra cycles.
func (c *CPU) execute(op byte) bool {
	switch op >> 6 {
	case 0:
		c.execGroup0(op)
	case 1:
		if op == opHLT {
			c.Halted = true
		} else {
			c.setReg(op>>3&7, c.reg(op&7)) // MOV
		}
	case 2:
		c.alu(op>>3&7, c.reg(op&7)) // ADD ... CMP r
	default:
		return c.execGroup3(op)
	}

	return false
}

// alu runs arithmetic or logic operation n (ADD ADC SUB SBB ANA XRA ORA CMP)
// on A and v.
func (c *CPU) alu(n, v byte) {
	switch n {
	case 0:
		c.add(v, 0)
	case 1:
		c.add(v, c.F&FlagCY)
	case 2:
		c.A = c.sub(v, 0)
	case 3:
		c.A = c.sub(v, c.F&FlagCY)
	case 4:
		c.ana(v)
	case 5:
		c.xra(v)
	case 6:
		c.ora(v)
	default:
		c.sub(v, 0) // CMP
	}
}

// execGroup0 runs opcodes 0x00–0x3F.
func (c *CPU) execGroup0(op byte) {
	d := op >> 3 & 7
	rp := d >> 1

	switch op & 7 {
	case 0: // NOP, and the undocumented NOPs 0x08 ... 0x38 (DESIGN §6.1)
	case 1:
		if d&1 == 0 {
			c.setPair(rp, c.fetch16()) // LXI
		} else {
			c.opDAD(c.pair(rp))
		}
	case 2:
		c.execLoadStore(d)
	case 3:
		if d&1 == 0 {
			c.setPair(rp, c.pair(rp)+1) // INX
		} else {
			c.setPair(rp, c.pair(rp)-1) // DCX
		}
	case 4:
		c.setReg(d, c.inr(c.reg(d))) // INR
	case 5:
		c.setReg(d, c.dcr(c.reg(d))) // DCR
	case 6:
		c.setReg(d, c.fetch()) // MVI
	default:
		c.execAccumulator(d)
	}
}

// execLoadStore runs STAX, LDAX, SHLD, LHLD, STA and LDA (opcodes 0x02 + d<<3).
func (c *CPU) execLoadStore(d byte) {
	switch d {
	case 0:
		c.bus.Write(c.BC(), c.A) // STAX B
	case 1:
		c.A = c.bus.Read(c.BC()) // LDAX B
	case 2:
		c.bus.Write(c.DE(), c.A) // STAX D
	case 3:
		c.A = c.bus.Read(c.DE()) // LDAX D
	case 4:
		c.write16(c.fetch16(), c.HL()) // SHLD
	case 5:
		c.setHL(c.read16(c.fetch16())) // LHLD
	case 6:
		c.bus.Write(c.fetch16(), c.A) // STA
	default:
		c.A = c.bus.Read(c.fetch16()) // LDA
	}
}

// execAccumulator runs RLC, RRC, RAL, RAR, DAA, CMA, STC and CMC
// (opcodes 0x07 + d<<3). The rotates change only CY.
func (c *CPU) execAccumulator(d byte) {
	switch d {
	case 0: // RLC
		cy := c.A >> 7
		c.A = c.A<<1 | cy
		c.F = c.F&^FlagCY | cy
	case 1: // RRC
		cy := c.A & 1
		c.A = c.A>>1 | cy<<7
		c.F = c.F&^FlagCY | cy
	case 2: // RAL
		cy := c.A >> 7
		c.A = c.A<<1 | c.F&FlagCY
		c.F = c.F&^FlagCY | cy
	case 3: // RAR
		cy := c.A & 1
		c.A = c.A>>1 | (c.F&FlagCY)<<7
		c.F = c.F&^FlagCY | cy
	case 4:
		c.daa()
	case 5: // CMA
		c.A = ^c.A
	case 6: // STC
		c.F |= FlagCY
	default: // CMC
		c.F ^= FlagCY
	}
}

// opDAD adds v to HL; only CY changes.
func (c *CPU) opDAD(v uint16) {
	sum := uint32(c.HL()) + uint32(v)
	c.setHL(uint16(sum))

	c.F &^= FlagCY
	if sum > 0xFFFF {
		c.F |= FlagCY
	}
}

// execGroup3 runs opcodes 0xC0–0xFF and reports whether a conditional call
// or return was taken.
func (c *CPU) execGroup3(op byte) bool {
	d := op >> 3 & 7
	switch op & 7 {
	case 0: // Rcc
		if c.cond(d) {
			c.PC = c.pop()
			return true
		}
	case 1:
		c.execPopEtc(d)
	case 2: // Jcc
		addr := c.fetch16()
		if c.cond(d) {
			c.PC = addr
		}
	case 3:
		c.execMisc(d)
	case 4: // Ccc
		addr := c.fetch16()
		if c.cond(d) {
			c.call(addr)
			return true
		}
	case 5:
		if d&1 == 0 {
			c.opPUSH(d >> 1)
		} else {
			c.call(c.fetch16()) // CALL, and the undocumented 0xDD, 0xED, 0xFD
		}
	case 6:
		c.alu(d, c.fetch()) // ADI ... CPI
	default:
		c.call(uint16(d) * 8) // RST
	}

	return false
}

// execPopEtc runs POP, RET (and its undocumented alias 0xD9), PCHL and SPHL
// (opcodes 0xC1 + d<<3).
func (c *CPU) execPopEtc(d byte) {
	switch d {
	case 0, 2, 4:
		c.setPair(d>>1, c.pop()) // POP B, D, H
	case 6: // POP PSW: bits 1, 3 and 5 of F are fixed
		v := c.pop()
		c.A = byte(v >> 8)
		c.F = byte(v)&flagsUsed | flagOne
	case 1, 3:
		c.PC = c.pop() // RET
	case 5:
		c.PC = c.HL() // PCHL
	default:
		c.SP = c.HL() // SPHL
	}
}

// execMisc runs JMP (and its undocumented alias 0xCB), OUT, IN, XTHL, XCHG,
// DI and EI (opcodes 0xC3 + d<<3).
func (c *CPU) execMisc(d byte) {
	switch d {
	case 0, 1:
		c.PC = c.fetch16() // JMP
	case 2:
		c.bus.Out(c.fetch(), c.A) // OUT
	case 3:
		c.A = c.bus.In(c.fetch()) // IN
	case 4: // XTHL
		v := c.read16(c.SP)
		c.write16(c.SP, c.HL())
		c.setHL(v)
	case 5: // XCHG
		c.H, c.L, c.D, c.E = c.D, c.E, c.H, c.L
	case 6: // DI
		c.IntEnabled = false
	default: // EI: takes effect after the next instruction (DESIGN §6.3)
		c.IntEnabled = true
		c.eiDelay = true
	}
}

// opPUSH pushes register pair rp: BC, DE, HL or PSW.
func (c *CPU) opPUSH(rp byte) {
	if rp == 3 {
		c.push(uint16(c.A)<<8 | uint16(c.F))
		return
	}

	c.push(c.pair(rp))
}

func (c *CPU) call(addr uint16) {
	c.push(c.PC)
	c.PC = addr
}
