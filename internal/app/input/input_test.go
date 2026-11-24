package input_test

import (
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/app/input"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
)

// TestCoinPulse: a press closes the coin switch for CoinFrames frames, even
// when the key is held longer or released sooner.
func TestCoinPulse(t *testing.T) {
	t.Parallel()

	for _, held := range []int{1, input.CoinFrames, 60} {
		var c input.Controls

		coinFrames := 0

		for frame := range 100 {
			if c.Update(input.Keys{Coin: frame < held}).Coin {
				coinFrames++
			}
		}

		if coinFrames != input.CoinFrames {
			t.Errorf("key held %d frames: coin switch closed %d frames, want %d", held, coinFrames, input.CoinFrames)
		}
	}
}

func TestEachPressIsACoin(t *testing.T) {
	t.Parallel()

	var c input.Controls

	pulses := 0
	wasCoin := false

	for frame := range 100 {
		down := frame%20 < 2 // press for 2 frames every 20
		coin := c.Update(input.Keys{Coin: down}).Coin

		if coin && !wasCoin {
			pulses++
		}

		wasCoin = coin
	}

	if pulses != 5 {
		t.Errorf("%d coin pulses for 5 presses", pulses)
	}
}

func TestKeysPassThrough(t *testing.T) {
	t.Parallel()

	var c input.Controls

	got := c.Update(input.Keys{Start1: true, Start2: true, Left: true, Right: true, Fire: true, Tilt: true})
	want := invaders.Inputs{Start1: true, Start2: true, Left: true, Right: true, Fire: true, Tilt: true}

	if got != want {
		t.Errorf("inputs %+v, want %+v", got, want)
	}
}
