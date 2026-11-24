//go:build js && wasm

package platform

import "syscall/js"

// PublishState puts a JSON description of the app's state on the page as
// window.invadersState, so the browser tests can see what the app shows
// (the ROM screen, the demo's lines) without reading pixels.
func PublishState(jsonText string) {
	js.Global().Set("invadersState", jsonText)
}
