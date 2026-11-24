package invaders_test

import (
	"bytes"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/invaders"
	"github.com/rbscholtus/Space-Invaders/internal/testrig"
)

// The game's credit count, in BCD. The coin handler in the VBLANK interrupt
// (0x002D–0x0038 in the ROM) increments it with ADI 1; DAA; STA 20EBH.
const creditsAddr = 0x20EB

// Port 3 bits (DESIGN §7.3).
const (
	out3Shot = 0x02
	out3Amp  = 0x20
)

// game drives a board with the real ROM a number of frames at a time.
type game struct {
	t      *testing.T
	b      *invaders.Board
	events []invaders.SoundEvent // every sound event since clearEvents
}

func newGame(t *testing.T) *game {
	t.Helper()

	b, err := invaders.New(testrig.ROMs(t), invaders.DIPs{})
	if err != nil {
		t.Fatal(err)
	}

	return &game{t: t, b: b}
}

// hold runs frames with in held down, collecting the sound events.
func (g *game) hold(in invaders.Inputs, frames int) {
	g.b.SetInputs(in)

	for range frames {
		g.b.RunFrame()
		g.events = append(g.events, g.b.SoundEvents()...)
	}
}

// clearEvents forgets the sound events collected so far.
func (g *game) clearEvents() { g.events = g.events[:0] }

// port3Set reports whether any event since clearEvents set bit in port 3.
func (g *game) port3Set(bit byte) bool {
	for _, e := range g.events {
		if e.Port == 3 && e.Value&bit != 0 {
			return true
		}
	}

	return false
}

func litPixels(video []byte) int {
	n := 0

	for _, v := range video {
		for ; v != 0; v &= v - 1 {
			n++
		}
	}

	return n
}

// TestROMSmoke plays the real game for a while (DESIGN §15.3; local only,
// skipped without your own ROMs): the attract mode draws and keeps the
// watchdog alive, a coin adds a credit, 1P start begins a game with the
// amplifier on, and fire plays the shot sound.
func TestROMSmoke(t *testing.T) {
	t.Parallel()

	const second = 60 // frames

	g := newGame(t)
	g.hold(invaders.Inputs{}, 10*second)

	if n := litPixels(g.b.Video()); n < 500 {
		t.Errorf("attract mode: %d pixels lit after 10 s, want a picture", n)
	}

	if w := g.b.Watchdog(); w < 50 {
		t.Errorf("attract mode: %d watchdog resets in 10 s, want the program to keep resetting it", w)
	}

	if c := g.b.Read(creditsAddr); c != 0 {
		t.Fatalf("credits = %02X before a coin, want 00", c)
	}

	g.hold(invaders.Inputs{Coin: true}, 5)
	g.hold(invaders.Inputs{}, second)

	if c := g.b.Read(creditsAddr); c != 0x01 {
		t.Fatalf("credits = %02X after a coin, want 01", c)
	}

	g.hold(invaders.Inputs{Start1: true}, 5)
	g.hold(invaders.Inputs{}, 5*second)

	if c := g.b.Read(creditsAddr); c != 0 {
		t.Errorf("credits = %02X after 1P start, want 00", c)
	}

	if p3, _ := g.b.Sound(); p3&out3Amp == 0 {
		t.Errorf("port 3 = %02X 5 s into the game, want the amplifier on", p3)
	}

	// The cannon can fire 5–6 s after start (measured), once the "PLAY
	// PLAYER 1" screen is over and the fleet is drawn.
	g.hold(invaders.Inputs{}, 2*second)

	if g.port3Set(out3Shot) {
		t.Error("shot sound before fire")
	}

	g.clearEvents()
	g.hold(invaders.Inputs{Fire: true}, 3)
	g.hold(invaders.Inputs{}, second/2)

	if !g.port3Set(out3Shot) {
		t.Error("no shot sound after fire")
	}
}

// TestROMCoinNeedsAnEdge: holding the coin switch adds one credit, not one
// per frame (the game adds a credit when the bit drops, DESIGN §7.2).
func TestROMCoinNeedsAnEdge(t *testing.T) {
	t.Parallel()

	g := newGame(t)
	g.hold(invaders.Inputs{}, 120)
	g.hold(invaders.Inputs{Coin: true}, 60)

	if c := g.b.Read(creditsAddr); c != 0x01 {
		t.Errorf("credits = %02X after holding the coin for 1 s, want 01", c)
	}

	g.hold(invaders.Inputs{}, 5)
	g.hold(invaders.Inputs{Coin: true}, 5)
	g.hold(invaders.Inputs{}, 5)

	if c := g.b.Read(creditsAddr); c != 0x02 {
		t.Errorf("credits = %02X after a second coin, want 02", c)
	}
}

// TestROMDeterministic: two runs of the same inputs give the same picture.
func TestROMDeterministic(t *testing.T) {
	t.Parallel()

	run := func() []byte {
		g := newGame(t)
		g.hold(invaders.Inputs{}, 300)
		g.hold(invaders.Inputs{Coin: true}, 5)
		g.hold(invaders.Inputs{}, 300)

		return bytes.Clone(g.b.Video())
	}

	if !bytes.Equal(run(), run()) {
		t.Error("two identical runs ended with different pictures")
	}
}
