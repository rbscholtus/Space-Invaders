package testrig

import "github.com/rbscholtus/Space-Invaders/internal/invaders"

// The scripted session played by the golden-frame test (DESIGN §15.3) and
// by the session sound check (§14.5), in frames (60 per second):
//
//	0–5399       attract mode, 90 s: the whole attract cycle with its demo
//	5400–5604    three coins
//	5700         1P start; from 6000 the player sweeps left and right and
//	             fires, until all ships are lost (about frame 9100)
//	11400        2P start with the two remaining credits; both players
//	             play the same way until game over (about frame 20500)
//	to 23999     attract mode again
//
// scripts/mame/session.lua plays the same script in MAME: keep them in step.
const (
	ScriptFrames = 24000

	firstCoin   = 5400
	coinGap     = 100
	coins       = 3
	start1Frame = 5700
	playFrom    = 6000
	start2Frame = 11400
	pressFrames = 5  // how long a coin or start button is held
	sweepFrames = 45 // frames moving one way
	fireEvery   = 16 // frames between presses of fire
	fireFrames  = 2
)

// Script returns the inputs held at frame f of the session.
func Script(f int) invaders.Inputs {
	var in invaders.Inputs

	for c := range coins {
		if start := firstCoin + c*coinGap; f >= start && f < start+pressFrames {
			in.Coin = true
		}
	}

	in.Start1 = f >= start1Frame && f < start1Frame+pressFrames
	in.Start2 = f >= start2Frame && f < start2Frame+pressFrames

	if f >= playFrom {
		in.Right = (f/sweepFrames)%2 == 0
		in.Left = !in.Right
		in.Fire = f%fireEvery < fireFrames
	}

	return in
}
