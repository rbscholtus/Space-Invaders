//go:build !(js && wasm)

package platform

// URLParams returns no options on the desktop: use the flags. See params.go.
func URLParams() Params { return Params{Volume: -1} }
