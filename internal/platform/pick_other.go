//go:build !(js && wasm)

package platform

// PickFiles reports ErrUnsupported on the desktop. See pick.go.
func PickFiles() <-chan Picked {
	ch := make(chan Picked, 1)
	ch <- Picked{Err: ErrUnsupported}

	return ch
}

// CanPickFiles reports whether PickFiles works in this build.
func CanPickFiles() bool { return false }
