package app

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/rbscholtus/Space-Invaders/internal/platform"
)

func TestReadDropped(t *testing.T) {
	t.Parallel()

	got := readDropped(fstest.MapFS{
		"invaders.zip":         {Data: []byte("zip")},
		"folder/invaders.h":    {Data: []byte{1}},
		"folder/sub/notes.txt": {Data: []byte("hi")},
	})
	if got.Err != nil {
		t.Fatal(got.Err)
	}

	for _, name := range []string{"invaders.zip", "invaders.h", "notes.txt"} {
		if _, ok := got.Files[name]; !ok {
			t.Errorf("%s missing from %v", name, got.Files)
		}
	}
}

func TestReadDroppedLimits(t *testing.T) {
	t.Parallel()

	many := fstest.MapFS{}
	for i := range maxDroppedFiles + 1 {
		many[string(rune('a'+i%26))+string(rune('a'+i/26))] = &fstest.MapFile{Data: []byte{1}}
	}

	if got := readDropped(many); !errors.Is(got.Err, platform.ErrTooLarge) {
		t.Errorf("too many files: %v, want ErrTooLarge", got.Err)
	}

	big := fstest.MapFS{"big.zip": {Data: make([]byte, maxDroppedFile+1)}}
	if got := readDropped(big); !errors.Is(got.Err, platform.ErrTooLarge) {
		t.Errorf("too large a file: %v, want ErrTooLarge", got.Err)
	}

	total := fstest.MapFS{}
	for _, n := range []string{"a", "b", "c"} {
		total[n] = &fstest.MapFile{Data: make([]byte, maxDroppedFile)}
	}

	if got := readDropped(total); !errors.Is(got.Err, platform.ErrTooLarge) {
		t.Errorf("too much in all: %v, want ErrTooLarge", got.Err)
	}
}
