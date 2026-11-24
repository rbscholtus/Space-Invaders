package invaders_test

import (
	"errors"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

// newBoard returns a board whose ROM holds the given code at each address.
func newBoard(t *testing.T, dips invaders.DIPs, code map[uint16][]byte) *invaders.Board {
	t.Helper()

	rom := make([]byte, 0x2000)
	for addr, bytes := range code {
		copy(rom[addr:], bytes)
	}

	b, err := invaders.New(rom, dips)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

func TestROMSize(t *testing.T) {
	t.Parallel()

	if _, err := invaders.New(make([]byte, 0x1000), invaders.DIPs{}); !errors.Is(err, invaders.ErrROMSize) {
		t.Errorf("err = %v, want ErrROMSize", err)
	}
}

// --- Memory map (DESIGN §7.1) ---

func TestROMIsReadOnly(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, map[uint16][]byte{0x1234: {0xAA}})
	b.Write(0x1234, 0x55)

	if got := b.Read(0x1234); got != 0xAA {
		t.Errorf("ROM at 1234 = %02X after a write, want AA", got)
	}
}

func TestReadsAbove4000StayInBounds(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, map[uint16][]byte{0x0000: {0x31}})
	b.Write(0x2005, 0x77)

	tests := []struct {
		addr uint16
		want byte
	}{
		{0x4000, 0x00}, // unused ROM space
		{0x5FFF, 0x00},
		{0x6005, 0x77}, // RAM mirror
		{0x8000, 0x31}, // A15 not decoded: ROM again
		{0xA005, 0x77}, // ... and RAM again
		{0xE005, 0x77},
		{0xFFFF, b.Read(0x3FFF)},
	}
	for _, tt := range tests {
		if got := b.Read(tt.addr); got != tt.want {
			t.Errorf("Read(%04X) = %02X, want %02X", tt.addr, got, tt.want)
		}
	}
}

// TestRAMMirror: the whole 8 KB of RAM is mirrored, not 1 KB as in the
// Java version (DESIGN §3).
func TestRAMMirror(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, nil)
	b.Write(0x7FFF, 0x42) // mirror of 0x3FFF, the last byte of video RAM
	b.Write(0x4400, 0x99) // unused space: ignored

	if got := b.Read(0x3FFF); got != 0x42 {
		t.Errorf("Read(3FFF) = %02X after writing 7FFF, want 42", got)
	}

	if got := b.Read(0x2400); got != 0 {
		t.Errorf("Read(2400) = %02X after writing 4400, want 00", got)
	}
}

// --- Ports (DESIGN §7.2–§7.4) ---

func TestPort1Bit3IsAlwaysSet(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, nil)
	if got := b.In(1); got&0x08 == 0 {
		t.Errorf("port 1 = %02X, want bit 3 set", got)
	}
}

// TestCoinIsActiveLow: bit 0 of port 1 is 1 at rest and 0 while a coin
// goes in (DESIGN §7.2). It doesn't change when read (DESIGN §3).
func TestCoinIsActiveLow(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, nil)
	if got := b.In(1); got&0x01 == 0 {
		t.Errorf("port 1 = %02X without a coin, want bit 0 set", got)
	}

	b.SetInputs(invaders.Inputs{Coin: true})

	for range 2 {
		if got := b.In(1); got&0x01 != 0 {
			t.Errorf("port 1 = %02X with a coin, want bit 0 clear", got)
		}
	}
}

func TestInputPorts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   invaders.Inputs
		dips invaders.DIPs
		port byte
		want byte
	}{
		{"port 0 at rest", invaders.Inputs{}, invaders.DIPs{}, 0, 0x08},
		{"port 0 controls", invaders.Inputs{Fire: true, Left: true, Right: true}, invaders.DIPs{}, 0, 0x78},
		{"port 1 at rest", invaders.Inputs{}, invaders.DIPs{}, 1, 0x09},
		{"port 1 starts", invaders.Inputs{Start1: true, Start2: true}, invaders.DIPs{}, 1, 0x0F},
		{"port 1 controls", invaders.Inputs{Fire: true, Right: true}, invaders.DIPs{}, 1, 0x59},
		{"port 2 defaults", invaders.Inputs{}, invaders.DIPs{}, 2, 0x00},
		{"port 2 lives 3", invaders.Inputs{}, invaders.DIPs{Lives: 3}, 2, 0x00},
		{"port 2 lives 6", invaders.Inputs{}, invaders.DIPs{Lives: 6}, 2, 0x03},
		{"port 2 lives out of range", invaders.Inputs{}, invaders.DIPs{Lives: 9}, 2, 0x00},
		{"port 2 bonus and coin info", invaders.Inputs{}, invaders.DIPs{Lives: 4, BonusAt1000: true, HideCoinInfo: true}, 2, 0x89},
		{"port 2 P2 bits are P1's controls", invaders.Inputs{Left: true}, invaders.DIPs{}, 2, 0x20},
		{"port 2 tilt", invaders.Inputs{Tilt: true}, invaders.DIPs{}, 2, 0x04},
		{"port 5 reads as port 1", invaders.Inputs{}, invaders.DIPs{}, 5, 0x09},
		{"port 4 reads as port 0", invaders.Inputs{}, invaders.DIPs{}, 4, 0x08},
	}
	for _, tt := range tests {
		b := newBoard(t, tt.dips, nil)
		b.SetInputs(tt.in)

		if got := b.In(tt.port); got != tt.want {
			t.Errorf("%s: In(%d) = %02X, want %02X", tt.name, tt.port, got, tt.want)
		}
	}
}

func TestShiftRegister(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, nil)
	b.Out(4, 0xAB) // reg = AB00
	b.Out(4, 0xCD) // reg = CDAB

	for amount, want := range []byte{0xCD, 0x9B, 0x36, 0x6D, 0xDA, 0xB5, 0x6A, 0xD5} {
		b.Out(2, byte(amount)|0xF8) // only the low 3 bits count

		if got := b.In(3); got != want {
			t.Errorf("amount %d: In(3) = %02X, want %02X", amount, got, want)
		}

		if got := b.In(7); got != want {
			t.Errorf("amount %d: In(7) = %02X, want port 3's %02X", amount, got, want)
		}
	}
}

func TestOutputPorts(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, nil)
	b.Out(3, 0x21)
	b.Out(5, 0x24)
	b.Out(6, 0)
	b.Out(6+8, 0)  // ports repeat every 8
	b.Out(0, 0xFF) // unconnected: ignored
	b.Out(7, 0xFF)

	if p3, p5 := b.Sound(); p3 != 0x21 || p5 != 0x24 {
		t.Errorf("Sound() = %02X, %02X; want 21, 24", p3, p5)
	}

	if got := b.Watchdog(); got != 2 {
		t.Errorf("Watchdog() = %d, want 2", got)
	}
}
