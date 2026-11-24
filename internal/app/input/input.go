// Package input turns the keys a player holds into the board's inputs
// (DESIGN §13). It is plain Go, so tests need no Ebitengine.
package input

import "github.com/rbscholtus/Space-Invaders/internal/invaders"

// Keys is what the player holds down this frame, after key mapping.
type Keys struct {
	Coin, Start1, Start2 bool
	Left, Right, Fire    bool
	Tilt                 bool
}

// CoinFrames is how long a coin holds the coin switch: a few frames, like a
// coin rolling past it. The game adds the credit on the switch's first frame
// (DESIGN §7.2, §13).
const CoinFrames = 5

// Controls keeps the state between frames: the coin pulse.
type Controls struct {
	coinWasDown bool
	coinLeft    int // frames the coin switch stays closed
}

// Update returns the board's inputs for one frame. Each new press of the
// coin key inserts one coin; holding it doesn't insert more.
func (c *Controls) Update(k Keys) invaders.Inputs {
	if k.Coin && !c.coinWasDown {
		c.coinLeft = CoinFrames
	}

	c.coinWasDown = k.Coin

	coin := c.coinLeft > 0
	if coin {
		c.coinLeft--
	}

	return invaders.Inputs{
		Coin:   coin,
		Start1: k.Start1,
		Start2: k.Start2,
		Fire:   k.Fire,
		Left:   k.Left,
		Right:  k.Right,
		Tilt:   k.Tilt,
	}
}
