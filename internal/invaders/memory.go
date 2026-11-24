package invaders

// The memory map (DESIGN §7.1; MAME mw8080bw.cpp, main_map): the CPU sees
// 15 address lines.
const (
	addrMask = 0x7FFF
	romSize  = 0x2000 // 0x0000–0x1FFF
	ramStart = 0x2000 // work RAM 0x2000–0x23FF, video RAM 0x2400–0x3FFF
	ramMask  = 0x1FFF // the 8 KB of RAM repeats at 0x6000–0x7FFF
	gapStart = 0x4000 // 0x4000–0x5FFF: unused ROM space, reads 0
	mirror   = 0x6000
	memSize  = romSize + ramMask + 1

	// VideoRAM is where the 1-bit picture starts (DESIGN §10).
	VideoRAM = 0x2400
)

// Read implements i8080.Bus.
func (b *Board) Read(addr uint16) byte {
	a := addr & addrMask
	switch {
	case a < gapStart:
		return b.mem[a]
	case a < mirror:
		return 0
	default:
		return b.mem[ramStart+a&ramMask]
	}
}

// Write implements i8080.Bus. Writes to ROM and the unused space are ignored.
func (b *Board) Write(addr uint16, v byte) {
	a := addr & addrMask
	if (a >= ramStart && a < gapStart) || a >= mirror {
		b.mem[ramStart+a&ramMask] = v
	}
}
