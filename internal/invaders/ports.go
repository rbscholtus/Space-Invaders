package invaders

// Inputs are the cabinet's controls, as held down.
type Inputs struct {
	Coin   bool // a coin going through the coin switch (hold it a few frames, DESIGN §13)
	Start1 bool
	Start2 bool
	Fire   bool
	Left   bool
	Right  bool
	// Tilt sets the bit the software reads as tilt. The Midway board doesn't
	// connect it (DESIGN §7.2); it's here as a debugging aid.
	Tilt bool
}

// DIPs are the board's DIP switches (DESIGN §7.5). The zero value is MAME's
// default: 3 lives, bonus life at 1500, coin info shown.
type DIPs struct {
	Lives        int  // 3–6; anything else means 3
	BonusAt1000  bool // bonus life at 1000 points instead of 1500
	HideCoinInfo bool // don't show the coin info in attract mode
}

// Port bits (DESIGN §7.2, §7.3; MAME mw8080bw.cpp).
const (
	in1Coin   = 0x01 // active low
	in1Start2 = 0x02
	in1Start1 = 0x04
	inAlways1 = 0x08 // ports 0 and 1, bit 3
	inFire    = 0x10 // ports 0, 1 and 2 (upright: P2's bits are P1's controls)
	inLeft    = 0x20
	inRight   = 0x40
	in2Tilt   = 0x04
	in2Bonus  = 0x08 // 1: bonus life at 1000
	in2NoCoin = 0x80 // 1: coin info off

	minLives = 3
	maxLives = 6

	// Ports are decoded with 3 address bits; reads also ignore bit 2.
	outPortMask = 0x07
	inPortMask  = 0x03

	portShiftCount = 2
	portSound1     = 3
	portShiftData  = 4
	portSound2     = 5
	portWatchdog   = 6
)

// controls returns the fire, left and right bits.
func (in Inputs) controls() byte {
	var v byte
	if in.Fire {
		v |= inFire
	}

	if in.Left {
		v |= inLeft
	}

	if in.Right {
		v |= inRight
	}

	return v
}

// In implements i8080.Bus.
func (b *Board) In(port byte) byte {
	in := b.inputs

	switch port & inPortMask {
	case 0:
		return inAlways1 | in.controls()
	case 1:
		v := inAlways1 | in.controls()
		if !in.Coin {
			v |= in1Coin
		}

		if in.Start1 {
			v |= in1Start1
		}

		if in.Start2 {
			v |= in1Start2
		}

		return v
	case 2:
		return b.dips.bits() | b.tiltBit() | in.controls()
	default:
		return b.shifter.result()
	}
}

func (b *Board) tiltBit() byte {
	if b.inputs.Tilt {
		return in2Tilt
	}

	return 0
}

// bits returns the DIP switches as port 2 reads them.
func (d DIPs) bits() byte {
	var v byte
	if d.Lives >= minLives && d.Lives <= maxLives {
		v = byte(d.Lives - minLives)
	}

	if d.BonusAt1000 {
		v |= in2Bonus
	}

	if d.HideCoinInfo {
		v |= in2NoCoin
	}

	return v
}

// Out implements i8080.Bus.
func (b *Board) Out(port, v byte) {
	switch port & outPortMask {
	case portShiftCount:
		b.shifter.count(v)
	case portSound1:
		b.out3 = v
	case portShiftData:
		b.shifter.data(v)
	case portSound2:
		b.out5 = v
	case portWatchdog:
		b.watchdog++
	}
}

// Watchdog returns how many times the program has reset the watchdog. The
// real board resets the CPU after 255 frames without one; the emulator only
// counts them (DESIGN §3).
func (b *Board) Watchdog() int { return b.watchdog }

// Sound returns the last values written to sound ports 3 and 5 (DESIGN §7.3).
func (b *Board) Sound() (byte, byte) { return b.out3, b.out5 }
