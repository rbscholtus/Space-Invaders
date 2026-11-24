// Package invaders emulates Midway's Space Invaders board (1978): memory map,
// shift register, ports, DIP switches and the two interrupts per frame
// (DESIGN §7, §8).
//
// It uses only the standard library and never reads the clock: it advances
// one video frame per RunFrame.
package invaders

import (
	"errors"
	"fmt"

	"github.com/rbscholtus/Space-Invaders/internal/i8080"
)

// Video timing (DESIGN §8; MAME mw8080bw.h): 320 pixels per line at
// 19.968 MHz / 4, so 128 CPU cycles per line, and 262 lines per frame.
const (
	CyclesPerLine  = 128
	LinesPerFrame  = 262
	CyclesPerFrame = CyclesPerLine * LinesPerFrame

	// The two interrupts per frame (DESIGN §8; MAME mw8080bw.cpp).
	midScreenLine = 96  // RST 1, in the middle of the picture
	vblankLine    = 224 // RST 2, at the start of vertical blank
	opRST1        = 0xCF
	opRST2        = 0xD7
)

// ErrROMSize is returned when the ROM image is not 8 KB.
var ErrROMSize = errors.New("invaders: the ROM image must be 8 KB")

// Board is the Space Invaders board: CPU, memory, shift register and ports.
type Board struct {
	// CPU is the board's 8080, exported for tracing and inspection.
	CPU *i8080.CPU

	mem     [memSize]byte // ROM, then work and video RAM (DESIGN §7.1)
	shifter shifter
	inputs  Inputs
	dips    DIPs

	out3, out5 byte // the last values written to the sound ports
	watchdog   int  // writes to port 6

	// intVector is the RST the INT line requests, 0 when it isn't raised.
	// It stays raised until the CPU accepts it (DESIGN §8).
	intVector byte
	// overshoot is how many cycles the last line ran past its end; the next
	// line is that much shorter.
	overshoot int
	frame     uint64

	frameStart uint64       // CPU.Cycles when the frame began
	events     []SoundEvent // this frame's sound-port changes
}

// New returns a board running rom, the 8 KB program image from
// romset.Invaders, with the DIP switches set to dips, after power-on.
func New(rom []byte, dips DIPs) (*Board, error) {
	if len(rom) != romSize {
		return nil, fmt.Errorf("%w, not %d bytes", ErrROMSize, len(rom))
	}

	b := &Board{dips: dips, events: make([]SoundEvent, 0, maxSoundEvents)}
	copy(b.mem[:romSize], rom)
	b.CPU = i8080.New(b)
	b.Reset()

	return b, nil
}

// Reset is a power-on reset: the CPU starts at 0, RAM is cleared (real RAM
// powers up with noise; clearing it keeps runs repeatable), and the shift
// register and outputs are cleared.
func (b *Board) Reset() {
	clear(b.mem[romSize:])
	b.shifter = shifter{}
	b.out3, b.out5, b.watchdog = 0, 0, 0
	b.intVector, b.overshoot, b.frame = 0, 0, 0
	b.CPU.Reset()
	b.frameStart = 0
	b.events = b.events[:0]
}

// SetInputs sets the controls, as held during the next frames.
func (b *Board) SetInputs(in Inputs) { b.inputs = in }

// Frame returns the number of frames run since Reset.
func (b *Board) Frame() uint64 { return b.frame }

// RunFrame runs one video frame: 262 lines of 128 cycles, raising RST 1 at
// line 96 and RST 2 at line 224 (DESIGN §8). Each line carries the
// overshoot of its last instruction into the next one, so emulated time
// doesn't drift.
func (b *Board) RunFrame() {
	b.frameStart = b.CPU.Cycles
	b.events = b.events[:0]

	for line := range LinesPerFrame {
		switch line {
		case midScreenLine:
			b.raiseInterrupt(opRST1)
		case vblankLine:
			b.raiseInterrupt(opRST2)
		}

		b.runLine()
	}

	b.frame++
}

// raiseInterrupt sets the INT flip-flop. As in MAME, a request made while
// the CPU has interrupts disabled is lost.
func (b *Board) raiseInterrupt(vector byte) {
	if b.CPU.IntEnabled {
		b.intVector = vector
	} else {
		b.intVector = 0
	}
}

// runLine runs one line's worth of cycles, offering a raised interrupt
// before each instruction until the CPU accepts it.
func (b *Board) runLine() {
	c := b.CPU
	start := c.Cycles
	target := uint64(max(CyclesPerLine-b.overshoot, 0))

	for c.Cycles-start < target {
		if b.intVector != 0 && c.Interrupt(b.intVector) {
			b.intVector = 0

			continue
		}

		c.Step()
	}

	b.overshoot = int(c.Cycles-start) - int(target)
}
