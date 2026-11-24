// Package synth emulates the Space Invaders sound board (DESIGN §14): six
// discrete analog circuits and an SN76477 for the UFO, stepped once per
// audio sample from the schematic's component values.
//
// It is a port of MAME's discrete-circuit models (src/devices/sound/disc_*
// and sn76477.cpp) and its Space Invaders sound description
// (src/mame/midw8080/mw8080bw_a.cpp), all BSD-3-Clause; see LICENSE-MAME.
// The ports keep MAME's arithmetic and order of operations, so the output
// matches MAME's closely enough to test against it (DESIGN §14.5).
//
// The package uses only the standard library, never reads the clock and
// imports nothing from this module.
package synth
