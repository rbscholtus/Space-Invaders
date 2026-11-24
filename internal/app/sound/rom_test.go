package sound_test

import (
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/app/sound"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
	"github.com/rbscholtus/Space-Invaders/internal/testrig"
)

// TestROMSound plays the real game through the sound path, as the game
// loop does (local only, with your own ROMs): attract mode is silent, and
// a started game makes sound (the fleet's march). The sounds themselves
// are tested against MAME in internal/synth.
func TestROMSound(t *testing.T) {
	t.Parallel()

	b, err := invaders.New(testrig.ROMs(t), invaders.DIPs{})
	if err != nil {
		t.Fatal(err)
	}

	frames := sound.NewFrames()

	// play runs n frames with in held and returns the loudest sample.
	play := func(in invaders.Inputs, n int) float32 {
		b.SetInputs(in)

		var loudest float32

		for range n {
			b.RunFrame()

			for _, v := range frames.Render(b.SoundEvents()) {
				loudest = max(loudest, v, -v)
			}
		}

		return loudest
	}

	if p := play(invaders.Inputs{}, 600); p != 0 {
		t.Errorf("attract mode: peak %v, want silence", p)
	}

	play(invaders.Inputs{Coin: true}, 5)
	play(invaders.Inputs{}, 60)
	play(invaders.Inputs{Start1: true}, 5)

	if p := play(invaders.Inputs{}, 7*60); p < 0.1 {
		t.Errorf("7 s into a game: peak %v, want the fleet's march", p)
	}
}
