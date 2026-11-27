//go:build !(js && wasm)

package platform

import "fmt"

// Fetch is only for the browser: the desktop app embeds what the browser
// fetches (assets.ReadArt), and needs no net/http. See fetch_js.go.
func Fetch(url string, _ int) ([]byte, error) {
	return nil, fmt.Errorf("fetch %s: %w", url, ErrUnsupported)
}
