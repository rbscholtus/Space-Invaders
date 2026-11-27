//go:build js && wasm

package assets

import (
	"fmt"
	"io/fs"
)

// ArtEmbedded is false in the browser: the cabinet's images are fetched
// from next to the page instead.
const ArtEmbedded = false

// ReadArt finds nothing in the browser. See ArtEmbedded.
func ReadArt(name string) ([]byte, error) {
	return nil, fmt.Errorf("cabinet art %s: %w", name, fs.ErrNotExist)
}
