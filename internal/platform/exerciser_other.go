//go:build !(js && wasm)

package platform

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// exercisersDir is where the exercisers are in a repository checkout.
const exercisersDir = "third_party/8080-exercisers"

// ReadExerciser reads the file from the repository. See exerciser.go.
func ReadExerciser(_ context.Context, name string) ([]byte, error) {
	return readExerciser(exercisersDir, name)
}

func readExerciser(dir, name string) ([]byte, error) {
	f, err := os.Open(filepath.Join(dir, filepath.Base(name)))
	if err != nil {
		return nil, fmt.Errorf("reading the exerciser: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxExerciserSize+1))
	if err != nil {
		return nil, fmt.Errorf("reading the exerciser: %w", err)
	}

	if len(data) > maxExerciserSize {
		return nil, fmt.Errorf("%s: %w", name, ErrTooLarge)
	}

	return data, nil
}
