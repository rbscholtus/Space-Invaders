// Package platform holds everything that differs between the browser and
// the desktop (DESIGN §16), adapted from Chippie. It is the only package
// that imports syscall/js; every _js.go file has a desktop counterpart in an
// _other.go file, so native builds and tests work.
//
// Calls that may wait (storage, fetching, the file picker) are for
// goroutines or return channels: never call them from Update or Draw.
package platform

import (
	"errors"
	"regexp"
)

// Errors from the platform layer.
var (
	ErrBadKey      = errors.New("platform: bad storage key")
	ErrUnavailable = errors.New("platform: storage unavailable")
	ErrUnsupported = errors.New("platform: not supported here")
	ErrTooLarge    = errors.New("platform: too large")
	ErrHTTPStatus  = errors.New("platform: download failed")
)

// appName names the app's storage: the IndexedDB database, the localStorage
// prefix, the user cache and config folders.
const appName = "space-invaders"

// validKey keeps keys safe as file names and IndexedDB keys.
var validKey = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)
