package app

import (
	"fmt"
	"io"
	"io/fs"
	"path"

	"github.com/rbscholtus/Space-Invaders/internal/platform"
)

// Limits for dropped files: a MAME invaders.zip is a few KB, the four ROMs
// 2 KB each. Larger drops are refused rather than read into memory.
const (
	maxDroppedFiles = 64
	maxDroppedFile  = 16 << 20
	maxDroppedTotal = 32 << 20
)

// errTooMuch reports a drop beyond the limits.
var errTooMuch = fmt.Errorf("%w: drop just invaders.zip or its four files", platform.ErrTooLarge)

// readDropped reads the files dropped on the window (ebiten.DroppedFiles)
// into memory, by base name, so a dropped folder works like its files. It
// may block in the browser: run it in a goroutine.
func readDropped(fsys fs.FS) platform.Picked {
	files := map[string][]byte{}
	total := 0

	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if len(files) == maxDroppedFiles {
			return errTooMuch
		}

		data, err := readLimited(fsys, p)
		if err != nil {
			return err
		}

		total += len(data)
		if total > maxDroppedTotal {
			return errTooMuch
		}

		files[path.Base(p)] = data

		return nil
	})
	if err != nil {
		return platform.Picked{Err: fmt.Errorf("reading the dropped files: %w", err)}
	}

	return platform.Picked{Files: files}
}

func readLimited(fsys fs.FS, name string) ([]byte, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", name, err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxDroppedFile+1))
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", name, err)
	}

	if len(data) > maxDroppedFile {
		return nil, fmt.Errorf("%s: %w", name, errTooMuch)
	}

	return data, nil
}
