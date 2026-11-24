//go:build !(js && wasm)

package platform

// InFront reports whether the app is in front, so the game should run: on
// the desktop, whether the window has focus.
func InFront(windowFocused bool) bool { return windowFocused }
