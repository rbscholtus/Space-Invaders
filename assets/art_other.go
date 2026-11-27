//go:build !(js && wasm)

package assets

import (
	"embed"
	"fmt"
)

//go:embed cabinet/*.jpg cabinet/*.png
var cabinetArt embed.FS

// ArtEmbedded says whether ReadArt has the cabinet's images: in the desktop
// build, yes.
const ArtEmbedded = true

// ReadArt returns one of the cabinet's images by name (cabinet@2x.jpg,
// backdrop.jpg, fire-pressed@2x.png).
func ReadArt(name string) ([]byte, error) {
	data, err := cabinetArt.ReadFile("cabinet/" + name)
	if err != nil {
		return nil, fmt.Errorf("cabinet art: %w", err)
	}

	return data, nil
}
