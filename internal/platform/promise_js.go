//go:build js && wasm

package platform

import (
	"errors"
	"fmt"
	"syscall/js"
)

// errPromise reports a rejected JavaScript promise.
var errPromise = errors.New("platform: a browser call failed")

// await waits for a JavaScript promise and returns its value. Call it from
// a goroutine, never from Update or Draw.
func await(promise js.Value) (js.Value, error) {
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
		ch <- result{err: fmt.Errorf("%w: %s", errPromise, args[0].Call("toString").String())}

		return nil
	})
	defer onErr.Release()

	promise.Call("then", onOK, onErr)

	r := <-ch

	return r.v, r.err
}
