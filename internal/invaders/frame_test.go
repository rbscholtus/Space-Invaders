package invaders_test

import (
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

// A test program: a counting loop, with interrupt handlers that save the
// count, so the tests can see when each interrupt came. The loop counts in
// HL, 15 cycles per pass; RST 1 saves HL at 2110H, RST 2 at 2112H, and both
// count interrupts at 2114H.
var countingProgram = map[uint16][]byte{
	0x0000: {0xC3, 0x30, 0x00},                                           // JMP 0030H (past the RST vectors)
	0x0008: {0xCD, 0x20, 0x00, 0x22, 0x10, 0x21, 0xFB, 0xC9},             // RST 1: CALL 0020H; SHLD 2110H; EI; RET
	0x0010: {0xCD, 0x20, 0x00, 0x22, 0x12, 0x21, 0xFB, 0xC9},             // RST 2: CALL 0020H; SHLD 2112H; EI; RET
	0x0020: {0xE5, 0x2A, 0x14, 0x21, 0x23, 0x22, 0x14, 0x21, 0xE1, 0xC9}, // count: PUSH H; LHLD 2114H; INX H; SHLD 2114H; POP H; RET
	0x0030: {0x31, 0x00, 0x24, 0x21, 0x00, 0x00, 0xFB, 0xC3, 0x40, 0x00}, // LXI SP,2400H; LXI H,0; EI; JMP 0040H
	0x0040: {0x23, 0xC3, 0x40, 0x00},                                     // loop: INX H; JMP 0040H (5 + 10 cycles)
}

func word(b *invaders.Board, addr uint16) int {
	return int(b.Read(addr)) | int(b.Read(addr+1))<<8
}

// TestInterruptsPerFrame: each frame raises RST 1 and RST 2 once.
func TestInterruptsPerFrame(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, countingProgram)
	for range 10 {
		b.RunFrame()
	}

	if got := word(b, 0x2114); got != 20 {
		t.Errorf("%d interrupts in 10 frames, want 20", got)
	}

	if b.Frame() != 10 {
		t.Errorf("Frame() = %d, want 10", b.Frame())
	}
}

// TestMidScreenInterruptLine: RST 1 comes at line 96 and RST 2 at line 224,
// not at half the frame's cycles as in the Java version (DESIGN §3, §8).
// The loop counts in 15-cycle passes; both handlers take the same time, so
// the count between them measures the distance between the lines.
func TestMidScreenInterruptLine(t *testing.T) {
	t.Parallel()

	const (
		loopCycles    = 15
		handlerCycles = 11 + 17 + 11 + 16 + 5 + 16 + 10 + 10 + 16 + 4 + 10 // RST, CALL, the counter, SHLD, EI, RET
		rst1ToRst2    = (224 - 96) * invaders.CyclesPerLine
		rst2ToRst1    = (262 - 224 + 96) * invaders.CyclesPerLine
	)

	b := newBoard(t, invaders.DIPs{}, countingProgram)
	b.RunFrame() // RST 1 and RST 2 of frame 0
	rst1, rst2 := word(b, 0x2110), word(b, 0x2112)
	b.RunFrame()
	nextRst1 := word(b, 0x2110)

	check := func(name string, got, cycles int) {
		t.Helper()

		want := (cycles - handlerCycles) / loopCycles
		if got < want-2 || got > want+2 {
			t.Errorf("%s: %d loop passes, want %d ± 2", name, got, want)
		}
	}
	check("RST 1 → RST 2", (rst2-rst1)&0xFFFF, rst1ToRst2)
	check("RST 2 → RST 1", (nextRst1-rst2)&0xFFFF, rst2ToRst1)

	// Line 96 is 96 lines into the frame: the first count is the passes
	// from the start (after the 10+10+10+4+10 cycles of set-up) to there.
	check("start → RST 1", rst1, 96*invaders.CyclesPerLine-44+handlerCycles)
}

// TestFrameCyclesCarryOver: the cycles run per frame average exactly
// 33,536; the overshoot of each line is carried, not dropped (DESIGN §3).
func TestFrameCyclesCarryOver(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, countingProgram)

	const frames = 100
	for range frames {
		b.RunFrame()
	}

	got := int(b.CPU.Cycles)
	if want := frames * invaders.CyclesPerFrame; got < want || got > want+20 {
		t.Errorf("%d cycles in %d frames, want %d (+ at most one instruction)", got, frames, want)
	}
}

// TestNoInterruptsWhenDisabled: with interrupts disabled, the board's
// requests are lost, and not taken later when the program enables them (as
// in MAME, DESIGN §8).
func TestNoInterruptsWhenDisabled(t *testing.T) {
	t.Parallel()

	program := map[uint16][]byte{
		0x0000: {0x31, 0x00, 0x24, 0xC3, 0x40, 0x00}, // LXI SP,2400H; JMP 0040H (interrupts stay off)
		0x0008: {0x3E, 0x01, 0x32, 0x00, 0x21, 0xC9}, // RST 1: MVI A,1; STA 2100H; RET
		0x0010: {0x3E, 0x02, 0x32, 0x00, 0x21, 0xC9}, // RST 2: MVI A,2; STA 2100H; RET
		0x0040: {0xC3, 0x40, 0x00},                   // loop: JMP 0040H
	}
	b := newBoard(t, invaders.DIPs{}, program)
	b.RunFrame()
	b.CPU.IntEnabled = true // as if EI ran: no old request may be pending
	b.CPU.Step()

	if got := b.Read(0x2100); got != 0 {
		t.Errorf("handler %d ran with interrupts disabled", got)
	}
}

// TestHaltWakesForInterrupt: the game's idle loop halts; each interrupt
// wakes it.
func TestHaltWakesForInterrupt(t *testing.T) {
	t.Parallel()

	program := map[uint16][]byte{
		0x0000: {0x31, 0x00, 0x24, 0xFB, 0x76, 0xC3, 0x03, 0x00}, // LXI SP; EI; HLT; JMP 0003H
		0x0008: {0xCD, 0x20, 0x00, 0xFB, 0xC9},                   // RST 1: CALL count; EI; RET
		0x0010: {0xCD, 0x20, 0x00, 0xFB, 0xC9},                   // RST 2: same
		0x0020: {0xE5, 0x2A, 0x14, 0x21, 0x23, 0x22, 0x14, 0x21, 0xE1, 0xC9},
	}
	b := newBoard(t, invaders.DIPs{}, program)
	b.RunFrame()
	b.RunFrame()

	if got := word(b, 0x2114); got != 4 {
		t.Errorf("%d interrupts in 2 frames while halting, want 4", got)
	}
}

// TestReset: a reset clears RAM and restarts the program.
func TestReset(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, countingProgram)
	b.RunFrame()
	b.Reset()

	if b.CPU.PC != 0 || b.CPU.Cycles != 0 || b.Frame() != 0 || word(b, 0x2114) != 0 {
		t.Errorf("after Reset: PC=%04X cycles=%d frame=%d count=%d", b.CPU.PC, b.CPU.Cycles, b.Frame(), word(b, 0x2114))
	}
}

func BenchmarkFrame(b *testing.B) {
	rom := make([]byte, 0x2000)
	for addr, code := range countingProgram {
		copy(rom[addr:], code)
	}

	board, err := invaders.New(rom, invaders.DIPs{})
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		board.RunFrame()
	}
}
