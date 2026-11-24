package i8080

import (
	"strconv"
	"strings"
)

// Disassemble returns the instruction at mem[pc] in Intel syntax, e.g.
// "MVI B,12H" or "JMP 0100H", and its length in bytes. Bytes past the end of
// mem read as 0.
func Disassemble(mem []byte, pc uint16) (string, int) {
	at := func(i int) byte {
		if i < len(mem) {
			return mem[i]
		}

		return 0
	}
	start := int(pc)
	info := opTable[at(start)]

	text := info.Mnemonic
	switch {
	case strings.HasSuffix(text, "d8"):
		text = strings.TrimSuffix(text, "d8") + intelHex(uint16(at(start+1)), 2)
	case strings.HasSuffix(text, "d16"), strings.HasSuffix(text, "a16"):
		v := uint16(at(start+2))<<8 | uint16(at(start+1))
		text = text[:len(text)-3] + intelHex(v, 4)
	}

	return text, info.Length
}

// intelHex formats v as Intel hex with digits digits: "0FFH", "1234H". A
// number that starts with a letter gets a leading 0, as Intel's assemblers
// require.
func intelHex(v uint16, digits int) string {
	s := strings.ToUpper(strconv.FormatUint(uint64(v), 16))
	if len(s) < digits {
		s = strings.Repeat("0", digits-len(s)) + s
	}

	if s[0] > '9' {
		s = "0" + s
	}

	return s + "H"
}
