// Command invaders runs the emulator in a window (DESIGN §4), with your own
// ROM set (DESIGN §9):
//
//	go run ./cmd/invaders -roms roms/invaders.zip
//
// Keys: 5 or C coin, 1 and 2 start, ←/→ (or A/D) move, Space (or ↑/W)
// fire, P pause, M mute, - and = volume.
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
	"github.com/rbscholtus/Space-Invaders/internal/platform"
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

// The lives DIP switch's range (DESIGN §7.2), and the volume's.
const (
	minLives      = 3
	maxLives      = 6
	maxVolume     = 100
	defaultVolume = 50
)

func main() {
	os.Exit(cli(os.Args[1:], os.Stderr, runWindow))
}

// defaultROMs is where the ROM set is looked for without -roms.
const defaultROMs = "roms/invaders.zip"

// cli parses args, loads the ROMs and runs the app with run; it returns the
// exit code. Usage and errors go to stderr. Without ROMs the app shows its
// ROM screen.
func cli(args []string, stderr io.Writer, run func(app.Options) error) int {
	fs := flag.NewFlagSet("invaders", flag.ContinueOnError)
	fs.SetOutput(stderr)
	roms := fs.String("roms", "", "your ROM set: a MAME `invaders.zip` or a folder with its files "+
		"(default: "+defaultROMs+" if there, else the remembered set or the ROM screen)")
	lives := fs.Int("lives", minLives, "DIP switch: lives per game, 3–6")
	bonus1000 := fs.Bool("bonus1000", false, "DIP switch: bonus life at 1000 points instead of 1500")
	volume := fs.Int("volume", defaultVolume, "sound volume, 0–100")
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

	if fs.NArg() != 0 || *lives < minLives || *lives > maxLives || *volume < 0 || *volume > maxVolume {
		fs.Usage()

		return exitUsage
	}

	image, err := initialROMs(*roms)
	if err != nil {
		fmt.Fprintln(stderr, "invaders:", err)

		return exitFail
	}

	// In the browser, the page's URL can set the volume and the demo's
	// speed (platform.URLParams); on the desktop it gives nothing.
	params := platform.URLParams()
	if params.Volume >= 0 {
		*volume = params.Volume
	}

	volumeGiven := params.Volume >= 0

	fs.Visit(func(f *flag.Flag) { volumeGiven = volumeGiven || f.Name == "volume" })

	opts := app.Options{
		ROMs:        image,
		DemoFast:    params.Demo == platform.DemoFast,
		VolumeGiven: volumeGiven,
		DIPs:        invaders.DIPs{Lives: *lives, BonusAt1000: *bonus1000},
		Volume:      float64(*volume) / maxVolume,
		OnError:     func(err error) { fmt.Fprintln(stderr, "invaders:", err) },
	}

	if err := run(opts); err != nil {
		fmt.Fprintln(stderr, "invaders:", err)

		return exitFail
	}

	return exitOK
}

// initialROMs loads the -roms set (an error if it can't), or the default
// set if it's there (nil if not: then the app finds the remembered set or
// shows its ROM screen).
func initialROMs(path string) ([]byte, error) {
	if path != "" {
		return loadROMs(path)
	}

	if !exists(defaultROMs) {
		return nil, nil
	}

	return loadROMs(defaultROMs)
}

func exists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
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
func runWindow(opts app.Options) error {
	ebiten.SetWindowTitle("Invaders: a Midway 8080 emulator")
	ebiten.SetWindowSize(invaders.ScreenWidth*windowScale, invaders.ScreenHeight*windowScale)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(app.New(opts)); err != nil {
		return fmt.Errorf("running the window: %w", err)
	}

	return nil
}
