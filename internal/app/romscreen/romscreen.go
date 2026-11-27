// Package romscreen is the flow from no ROMs to a running game (DESIGN
// §9.2): it takes the remembered image, dropped or picked files, and turns
// them into a ROM image or a message for the player, and it draws the ROM
// screen's text. It is plain Go, without Ebitengine, so tests drive it.
package romscreen

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/rbscholtus/Space-Invaders/internal/app/font"
	"github.com/rbscholtus/Space-Invaders/internal/romset"
)

// Mode is where the flow is.
type Mode int

// The modes, in order.
const (
	Loading Mode = iota // looking for remembered ROMs
	NoROMs              // the ROM screen: waiting for the player's files
	Loaded              // the game runs
)

// String names the mode, as the browser tests see it.
func (m Mode) String() string {
	switch m {
	case Loading:
		return "loading"
	case NoROMs:
		return "roms"
	case Loaded:
		return "game"
	}

	return "unknown"
}

// Screen layout, in character cells (28 × 32 on the monitor).
const (
	columns     = 28
	firstRow    = 1
	maxMsgLines = 6
	keyColumn   = 12 // where the controls' descriptions start
)

// Flow is the ROM flow. The zero value is not ready: use New.
type Flow struct {
	mode    Mode
	message string
	canPick bool
	set     romset.Set
}

// New returns the flow, looking for remembered ROMs. canPick says whether
// the file picker is offered (the browser on a computer).
func New(canPick bool) *Flow {
	return &Flow{mode: Loading, canPick: canPick, set: romset.Invaders()}
}

// Mode returns where the flow is.
func (f *Flow) Mode() Mode { return f.mode }

// Message returns what the ROM screen tells the player, "" for nothing.
func (f *Flow) Message() string { return f.message }

// Start starts the game with an image that is already checked (the
// -roms flag).
func (f *Flow) Start() {
	f.mode, f.message = Loaded, ""
}

// Restored handles the remembered image, if any: it returns the image to
// run, or nil and shows the ROM screen. A remembered image that no longer
// checks out is ignored with a message.
func (f *Flow) Restored(image []byte, found bool, err error) []byte {
	if f.mode != Loading {
		return nil
	}

	f.mode = NoROMs

	switch {
	case err != nil:
		f.message = fmt.Sprintf("Can't read remembered ROMs: %v", err)
	case !found:
	case romset.CheckImage(f.set, image) != nil:
		f.message = "The remembered ROMs were damaged: please load them again."
	default:
		f.mode = Loaded

		return image
	}

	return nil
}

// Offer tries the files the player dropped or picked: it returns the image
// to run and remember, or nil with a message saying what's wrong.
func (f *Flow) Offer(fsys fs.FS) []byte {
	image, err := romset.Load(f.set, fsys)
	if err != nil {
		f.Fail(err)

		return nil
	}

	f.mode, f.message = Loaded, ""

	return image
}

// Fail shows a problem on the ROM screen (a failed pick, for example).
func (f *Flow) Fail(err error) {
	f.mode = NoROMs
	f.message = explain(err)
}

// Forget goes back to the ROM screen (the remembered image is deleted by
// the caller).
func (f *Flow) Forget() {
	f.mode = NoROMs
	f.message = "ROMs and HI-SCORE forgotten."
}

// explain words romset's errors for the player.
func explain(err error) string {
	switch {
	case errors.Is(err, romset.ErrNotAROMSet):
		if _, found, ok := strings.Cut(err.Error(), "found: "); ok {
			return "These aren't the invaders ROMs. Found: " + found
		}

		return "These aren't the invaders ROMs."
	case errors.Is(err, romset.ErrMissingROM):
		return "Not all four ROMs are there. " + err.Error()
	case errors.Is(err, romset.ErrCRCMismatch):
		return "These ROMs differ from the MAME invaders set. " + err.Error()
	}

	return err.Error()
}

// Lines returns the ROM screen's text, at most 28 columns each.
func (f *Flow) Lines() []string {
	if f.mode == Loading {
		return []string{"", "   Looking for your ROMs"}
	}

	lines := []string{
		"      NO ROMS LOADED",
		"",
		"This emulator needs your",
		"own copy of the MAME",
		"invaders ROM set: drop",
		"invaders.zip (or its four",
		"files) on this window.",
	}

	if f.canPick {
		lines = append(lines, "", "Or press O to open them.")
	}

	if f.message != "" {
		msg := font.Wrap(f.message, columns)
		if len(msg) > maxMsgLines {
			msg = append(msg[:maxMsgLines-1], "...")
		}

		lines = append(append(lines, ""), msg...)
	}

	return lines
}

// Controls returns the keys a new player needs (DESIGN §13), shown at the
// bottom of the ROM screen, at most 28 columns each. The keys fill rows
// 23–29 exactly: the overlay's green band (the strip below it is only
// partly green). O is listed only where the file picker is offered.
func (f *Flow) Controls() []string {
	roms := fmt.Sprintf("%-*s%s", keyColumn, "F", "forget the ROMs")
	if f.canPick {
		roms = fmt.Sprintf("%-*s%s", keyColumn, "O / F", "open/forget ROMs")
	}

	return []string{
		"          CONTROLS",
		fmt.Sprintf("%-*s%s", keyColumn, "5 or C", "insert a coin"),
		fmt.Sprintf("%-*s%s", keyColumn, "1 or 2", "1 or 2 players"),
		fmt.Sprintf("%-*s%s", keyColumn, "Left Right", "move"),
		fmt.Sprintf("%-*s%s", keyColumn, "Space", "fire"),
		fmt.Sprintf("%-*s%s", keyColumn, "P / V", "pause / view"),
		fmt.Sprintf("%-*s%s", keyColumn, "M, - and =", "mute, volume"),
		roms,
	}
}

// controlsBottom is how many rows stay empty below the controls.
const controlsBottom = 2

// Draw draws the ROM screen into a picture width pixels wide, one byte per
// pixel: the text from the top, the controls at the bottom, leaving the
// last two rows empty.
func (f *Flow) Draw(dst []byte, width int) {
	for i, line := range f.Lines() {
		font.Draw(dst, width, 0, (firstRow+i)*font.CellHeight, line)
	}

	controls := f.Controls()
	rows := len(dst) / width / font.CellHeight

	for i, line := range controls {
		font.Draw(dst, width, 0, (rows-controlsBottom-len(controls)+i)*font.CellHeight, line)
	}
}
