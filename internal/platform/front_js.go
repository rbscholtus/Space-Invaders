//go:build js && wasm

package platform

import "syscall/js"

// InFront reports whether the app is in front, so the game should run: in
// the browser, whether the page is visible. windowFocused (Ebitengine's
// IsFocused) is ignored, because it also asks document.hasFocus(), which
// iOS Safari can report as false while the page is in front (as found in
// Chippie).
func InFront(_ bool) bool {
	return !js.Global().Get("document").Get("hidden").Truthy()
}
