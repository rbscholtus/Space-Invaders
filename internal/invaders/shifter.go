package invaders

// shifter is the board's 16-bit shift register, used to draw sprites at any
// x position (DESIGN §7.4; a Fujitsu MB14241 in MAME).
type shifter struct {
	reg    uint16
	amount byte
}

// data takes a byte written to port 4: the register shifts right by 8 and
// the new byte goes into the high byte.
func (s *shifter) data(v byte) { s.reg = uint16(v)<<8 | s.reg>>8 }

// count takes a byte written to port 2: the shift amount is its low 3 bits.
func (s *shifter) count(v byte) { s.amount = v & 7 }

// result is what port 3 reads: the 8 bits starting amount bits from the top.
func (s *shifter) result() byte { return byte(s.reg >> (8 - s.amount)) }
