//go:build js && wasm

package platform

import (
	"fmt"
	"syscall/js"
)

// maxPickSize bounds each picked file: a MAME invaders.zip is a few KB.
const maxPickSize = 16 << 20

// PickFiles opens a hidden <input type="file" multiple> and reads the
// chosen files with arrayBuffer(). See pick.go.
func PickFiles() <-chan Picked {
	ch := make(chan Picked, 1)
	doc := js.Global().Get("document")

	input := doc.Call("createElement", "input")
	input.Set("type", "file")
	input.Set("multiple", true)
	input.Get("style").Set("display", "none")
	doc.Get("body").Call("appendChild", input)

	var onChange, onCancel js.Func

	done := func() {
		onChange.Release()
		onCancel.Release()
		input.Call("remove")
	}

	// Cancelling sends Picked{}, so the app knows the picker closed.
	onCancel = js.FuncOf(func(js.Value, []js.Value) any {
		defer done()

		ch <- Picked{}

		return nil
	})

	onChange = js.FuncOf(func(js.Value, []js.Value) any {
		defer done()

		files := input.Get("files")
		n := files.Get("length").Int()

		if n == 0 {
			ch <- Picked{}

			return nil
		}

		list := make([]js.Value, n)
		for i := range list {
			list[i] = files.Index(i)
		}

		// arrayBuffer() is a promise: wait for it in a goroutine, so this
		// event handler returns at once.
		go func() { ch <- readPicked(list) }()

		return nil
	})

	input.Call("addEventListener", "change", onChange)
	input.Call("addEventListener", "cancel", onCancel)
	input.Call("click")

	return ch
}

// readPicked reads the picked files' contents.
func readPicked(list []js.Value) Picked {
	out := map[string][]byte{}

	for _, file := range list {
		name := file.Get("name").String()
		if file.Get("size").Int() > maxPickSize {
			return Picked{Err: fmt.Errorf("%s: %w", name, ErrTooLarge)}
		}

		buf, err := await(file.Call("arrayBuffer"))
		if err != nil {
			return Picked{Err: fmt.Errorf("reading %s: %w", name, err)}
		}

		data := make([]byte, buf.Get("byteLength").Int())
		js.CopyBytesToGo(data, js.Global().Get("Uint8Array").New(buf))
		out[name] = data
	}

	return Picked{Files: out}
}

// CanPickFiles reports whether the app offers PickFiles: not on phones and
// tablets (a coarse primary pointer), where the browser opens the chooser
// only inside the tap itself, while the app acts on it a frame later.
func CanPickFiles() bool {
	return !js.Global().Call("matchMedia", "(pointer: coarse)").Get("matches").Bool()
}
