//go:build !(js && wasm)

package assets

import (
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/app/cabinet"
)

// Every image cabinet.json names is embedded, at every scale.
func TestCabinetArt(t *testing.T) {
	t.Parallel()

	c, err := cabinet.Parse(CabinetJSON)
	if err != nil {
		t.Fatal(err)
	}

	names := []string{c.Backdrop}

	for n := 1; n <= cabinet.MaxScale; n++ {
		names = append(names, cabinet.FileName(c.Image, n, "jpg"))

		for _, ctl := range c.Controls {
			if ctl.Pressed != "" {
				names = append(names, cabinet.FileName(ctl.Pressed, n, "png"))
			}
		}
	}

	for _, name := range names {
		if data, err := ReadArt(name); err != nil || len(data) == 0 {
			t.Errorf("%s: %v", name, err)
		}
	}
}
