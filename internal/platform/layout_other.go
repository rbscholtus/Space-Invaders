//go:build !(js && wasm)

package platform

// PublishLayout does nothing on the desktop. See layout_js.go.
func PublishLayout(string) {}
