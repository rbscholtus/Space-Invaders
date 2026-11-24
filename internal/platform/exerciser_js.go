//go:build js && wasm

package platform

import (
	"context"
	"errors"
	"fmt"
	"syscall/js"
)

// errFetch reports a failed fetch() call (network error, abort).
var errFetch = errors.New("platform: fetch failed")

// ReadExerciser fetches exercisers/<name> next to the page with the
// browser's fetch(), not net/http, which would add about 1.7 MB (gzip) to
// the WebAssembly. See exerciser.go.
func ReadExerciser(ctx context.Context, name string) ([]byte, error) {
	url := "exercisers/" + name
	abort := js.Global().Get("AbortController").New()
	opts := js.Global().Get("Object").New()
	opts.Set("signal", abort.Get("signal"))

	resp, err := await(ctx, abort, js.Global().Call("fetch", url, opts))
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}

	if !resp.Get("ok").Bool() {
		return nil, fmt.Errorf("fetch %s: %w: %d", url, ErrHTTPStatus, resp.Get("status").Int())
	}

	buf, err := await(ctx, abort, resp.Call("arrayBuffer"))
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}

	n := buf.Get("byteLength").Int()
	if n > maxExerciserSize {
		return nil, fmt.Errorf("fetch %s: %w", url, ErrTooLarge)
	}

	body := make([]byte, n)
	js.CopyBytesToGo(body, js.Global().Get("Uint8Array").New(buf))

	return body, nil
}

// await waits for a JavaScript promise, or for ctx to end (then it aborts
// the request).
func await(ctx context.Context, abort, promise js.Value) (js.Value, error) {
	type result struct {
		v   js.Value
		err error
	}

	ch := make(chan result, 1)

	onOK := js.FuncOf(func(_ js.Value, args []js.Value) any {
		ch <- result{v: args[0]}

		return nil
	})
	defer onOK.Release()

	onErr := js.FuncOf(func(_ js.Value, args []js.Value) any {
		ch <- result{err: fmt.Errorf("%w: %s", errFetch, args[0].Call("toString").String())}

		return nil
	})
	defer onErr.Release()

	promise.Call("then", onOK, onErr)

	select {
	case r := <-ch:
		return r.v, r.err
	case <-ctx.Done():
		abort.Call("abort")

		return js.Value{}, fmt.Errorf("%w: %w", errFetch, ctx.Err())
	}
}
