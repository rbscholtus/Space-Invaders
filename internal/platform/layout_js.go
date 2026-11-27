//go:build js && wasm

package platform

import "syscall/js"

// PublishLayout puts a JSON description of the app's layout on the page as
// window.invadersLayout, so the browser tests can find the picture and the
// cabinet's controls instead of repeating the layout maths.
func PublishLayout(jsonText string) {
	js.Global().Set("invadersLayout", jsonText)
}
