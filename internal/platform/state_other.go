//go:build !(js && wasm)

package platform

// PublishState does nothing on the desktop. See state_js.go.
func PublishState(string) {}
