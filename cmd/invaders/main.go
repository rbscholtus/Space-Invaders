// Command invaders runs the emulator in a window (DESIGN §4), with your own
// ROM set (DESIGN §9):
//
//	go run ./cmd/invaders -roms roms/invaders.zip
//
// Without -roms it uses roms/invaders.zip if there, else the remembered set,
// else it shows the ROM screen: drop the set on the window. The same
// program runs in the browser (scripts/build.sh).
//
// Keys: 5 or C coin, 1 and 2 start, ←/→ (or A/D) move, Space (or ↑/W)
// fire, P pause, M mute, - and = volume, V cabinet or picture alone, F
// forget the remembered ROMs.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/rbscholtus/Space-Invaders/assets"
	"github.com/rbscholtus/Space-Invaders/internal/app"
	"github.com/rbscholtus/Space-Invaders/internal/app/crt"
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

// The window's size at start-up: landscape, as most screens are, which
// shows the cabinet's upper part (DESIGN §11.1).
const (
	windowWidth  = 1024
	windowHeight = 768
)

// maxArtFile is the largest cabinet image the browser accepts; the largest
// is well under it.
const maxArtFile = 4 << 20

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
	shader := fs.String("shader", "crt", "how the picture is drawn: crt (the monitor) or clean (pixel for pixel)")
	viewName := fs.String("view", "auto", "what the window shows: cabinet, screen (the picture alone) "+
		"or auto (by the window's size)")
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

	mode, modeErr := crt.ParseMode(*shader)
	view, viewErr := app.ParseView(*viewName)

	if fs.NArg() != 0 || *lives < minLives || *lives > maxLives || *volume < 0 || *volume > maxVolume ||
		modeErr != nil || viewErr != nil {
		fs.Usage()

		return exitUsage
	}

	image, err := initialROMs(*roms)
	if err != nil {
		return fail(stderr, err)
	}

	volumeInURL, viewInURL := fromURL(volume, &mode, &view)

	opts := app.Options{
		ROMs:        image,
		VolumeGiven: volumeInURL || given(fs, "volume"),
		DIPs:        invaders.DIPs{Lives: *lives, BonusAt1000: *bonus1000},
		Volume:      float64(*volume) / maxVolume,
		Shader:      mode,
		View:        view,
		ViewGiven:   viewInURL || given(fs, "view"),
		Art:         readArt,
		OnError:     func(err error) { fmt.Fprintln(stderr, "invaders:", err) },
	}

	if err := run(opts); err != nil {
		return fail(stderr, err)
	}

	return exitOK
}

// fail reports err and returns the exit code for it.
func fail(stderr io.Writer, err error) int {
	fmt.Fprintln(stderr, "invaders:", err)

	return exitFail
}

// given reports whether the flag was set on the command line.
func given(fs *flag.FlagSet, name string) bool {
	set := false

	fs.Visit(func(f *flag.Flag) { set = set || f.Name == name })

	return set
}

// fromURL applies the page's URL parameters in the browser
// (platform.URLParams; on the desktop there are none), and reports whether
// they gave the volume and the view.
func fromURL(volume *int, mode *crt.Mode, view *app.View) (bool, bool) {
	params := platform.URLParams()
	if params.Volume >= 0 {
		*volume = params.Volume
	}

	if m, err := crt.ParseMode(params.Shader); err == nil {
		*mode = m
	}

	v, err := app.ParseView(params.View)
	if err == nil {
		*view = v
	}

	return params.Volume >= 0, err == nil
}

// readArt reads one of the cabinet's images: embedded on the desktop,
// fetched from next to the page in the browser (scripts/build.sh puts them
// in cabinet/).
func readArt(name string) ([]byte, error) {
	if assets.ArtEmbedded {
		return assets.ReadArt(name) //nolint:wrapcheck // it names the file
	}

	return platform.Fetch("cabinet/"+name, maxArtFile) //nolint:wrapcheck // it names the URL
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
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(app.New(opts)); err != nil {
		return fmt.Errorf("running the window: %w", err)
	}

	return nil
}
