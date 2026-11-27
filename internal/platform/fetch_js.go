//go:build js && wasm

package platform

import (
	"errors"
	"fmt"
	"syscall/js"
)

// ErrHTTPStatus reports a fetch whose answer wasn't 200 OK.
var ErrHTTPStatus = errors.New("platform: HTTP request failed")

// Fetch downloads url, relative to the page, with the browser's fetch(),
// and returns the body, at most limit bytes. It blocks, so call it from a
// goroutine, never from Update or Draw. It uses fetch() directly because
// Go's net/http would add about 1.7 MB (gzip) to the WebAssembly file.
func Fetch(url string, limit int) ([]byte, error) {
	resp, err := await(js.Global().Call("fetch", url))
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}

	if !resp.Get("ok").Bool() {
		return nil, fmt.Errorf("fetch %s: %w: %d", url, ErrHTTPStatus, resp.Get("status").Int())
	}

	buf, err := await(resp.Call("arrayBuffer"))
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}

	n := buf.Get("byteLength").Int()
	if n > limit {
		return nil, fmt.Errorf("fetch %s: %w", url, ErrTooLarge)
	}

	body := make([]byte, n)
	js.CopyBytesToGo(body, js.Global().Get("Uint8Array").New(buf))

	return body, nil
}
