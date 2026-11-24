package invaders_test

import (
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

// soundProgram writes the sound ports at known cycles:
//
//	0000  MVI A,20H     7   cycles 0–7
//	0002  OUT 3        10   written at cycle 17 (the end of the OUT)
//	0004  OUT 3        10   same value: no event
//	0006  MVI A,01H     7
//	0008  OUT 5        10   written at cycle 44
//	000A  MVI A,22H     7
//	000C  OUT 3        10   written at cycle 61
//	000E  JMP 000EH         spin
var soundProgram = map[uint16][]byte{
	0x0000: {0x3E, 0x20, 0xD3, 0x03, 0xD3, 0x03, 0x3E, 0x01, 0xD3, 0x05, 0x3E, 0x22, 0xD3, 0x03, 0xC3, 0x0E, 0x00},
}

func TestSoundEvents(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, soundProgram)
	b.RunFrame()

	// The board records the cycle count when OUT starts executing its
	// write: the cycles before the instruction (the CPU adds an
	// instruction's cycles after it runs).
	want := []invaders.SoundEvent{
		{Cycle: 7, Port: 3, Value: 0x20},
		{Cycle: 34, Port: 5, Value: 0x01},
		{Cycle: 51, Port: 3, Value: 0x22},
	}

	got := b.SoundEvents()
	if len(got) != len(want) {
		t.Fatalf("events %+v, want %+v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("event %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	if p3, p5 := b.Sound(); p3 != 0x22 || p5 != 0x01 {
		t.Errorf("Sound() = %02X, %02X; want 22, 01", p3, p5)
	}

	b.RunFrame()

	if got := b.SoundEvents(); len(got) != 0 {
		t.Errorf("second frame has events %+v, want none", got)
	}
}

// TestSoundEventsOverflow: beyond the preallocated room, the last event
// keeps the latest value, and nothing allocates. Not parallel:
// testing.AllocsPerRun doesn't allow it.
func TestSoundEventsOverflow(t *testing.T) { //nolint:paralleltest // AllocsPerRun panics in parallel tests
	program := map[uint16][]byte{
		0x0000: {0x3C, 0xD3, 0x03, 0xC3, 0x00, 0x00}, // loop: INR A; OUT 3; JMP 0000H
	}
	b := newBoard(t, invaders.DIPs{}, program)

	allocs := testing.AllocsPerRun(3, b.RunFrame)
	if allocs != 0 {
		t.Errorf("RunFrame allocates %.0f times, want 0", allocs)
	}

	events := b.SoundEvents()
	if len(events) != 64 {
		t.Fatalf("%d events, want 64 (the preallocated room)", len(events))
	}

	if p3, _ := b.Sound(); events[63].Value != p3 {
		t.Errorf("last event %02X, want the port's current value %02X", events[63].Value, p3)
	}
}

func TestResetClearsSoundEvents(t *testing.T) {
	t.Parallel()

	b := newBoard(t, invaders.DIPs{}, soundProgram)
	b.RunFrame()
	b.Reset()

	if got := b.SoundEvents(); len(got) != 0 {
		t.Errorf("events after Reset: %+v", got)
	}

	if p3, p5 := b.Sound(); p3 != 0 || p5 != 0 {
		t.Errorf("Sound() after Reset = %02X, %02X", p3, p5)
	}
}
