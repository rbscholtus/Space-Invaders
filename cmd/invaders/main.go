// Command invaders runs the emulator in a window (DESIGN §4), with your own
// ROM set (DESIGN §9):
//
//	go run ./cmd/invaders -roms roms/invaders.zip
//
// Keys: 5 or C coin, 1 and 2 start, ←/→ (or A/D) move, Space (or ↑/W)
// fire, P pause.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/rbscholtus/Space-Invaders/internal/app"
	"github.com/rbscholtus/Space-Invaders/internal/invaders"
	"github.com/rbscholtus/Space-Invaders/internal/romset"
)

// Exit codes.
const (
	exitOK    = 0
	exitFail  = 1
	exitUsage = 2
)

// windowScale is the window's size in native pixels at start-up.
const windowScale = 3

// The lives DIP switch's range (DESIGN §7.2).
const (
	minLives = 3
	maxLives = 6
)

func main() {
	os.Exit(cli(os.Args[1:], os.Stderr, runWindow))
}

// cli parses args, loads the ROMs and runs the board with run; it returns
// the exit code. Usage and errors go to stderr.
func cli(args []string, stderr io.Writer, run func(*invaders.Board) error) int {
	fs := flag.NewFlagSet("invaders", flag.ContinueOnError)
	fs.SetOutput(stderr)
	roms := fs.String("roms", "roms/invaders.zip", "your ROM set: a MAME `invaders.zip` or a folder with its files")
	lives := fs.Int("lives", minLives, "DIP switch: lives per game, 3–6")
	bonus1000 := fs.Bool("bonus1000", false, "DIP switch: bonus life at 1000 points instead of 1500")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: invaders [flags]")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}

		return exitUsage
	}

	if fs.NArg() != 0 || *lives < minLives || *lives > maxLives {
		fs.Usage()

		return exitUsage
	}

	image, err := loadROMs(*roms)
	if err != nil {
		fmt.Fprintln(stderr, "invaders:", err)

		return exitFail
	}

	board, err := invaders.New(image, invaders.DIPs{Lives: *lives, BonusAt1000: *bonus1000})
	if err != nil {
		fmt.Fprintln(stderr, "invaders:", err)

		return exitFail
	}

	if err := run(board); err != nil {
		fmt.Fprintln(stderr, "invaders:", err)

		return exitFail
	}

	return exitOK
}

// loadROMs reads the ROM set from a zip file or a folder.
func loadROMs(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("no ROM set at %s (bring your own, see the README): %w", path, err)
	}

	var image []byte

	if info.IsDir() {
		image, err = romset.Load(romset.Invaders(), os.DirFS(path))
	} else {
		var data []byte

		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading the ROM set: %w", err)
		}

		image, err = romset.LoadZip(romset.Invaders(), data)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return image, nil
}

// runWindow opens the window and runs the game until it's closed.
func runWindow(board *invaders.Board) error {
	ebiten.SetWindowTitle("Invaders: a Midway 8080 emulator")
	ebiten.SetWindowSize(invaders.ScreenWidth*windowScale, invaders.ScreenHeight*windowScale)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(app.New(board)); err != nil {
		return fmt.Errorf("running the window: %w", err)
	}

	return nil
}
