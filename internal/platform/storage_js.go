//go:build js && wasm

package platform

import (
	"fmt"
	"sync"
	"syscall/js"
)

// IndexedDB from Go: the callback API is turned into blocking calls with
// channels. Callers run them in goroutines, so the game loop never waits.

const idbVersion = 1

// conn is the page's one IndexedDB connection, opened on first use.
var conn struct { //nolint:gochecknoglobals // one connection per page, like the database itself
	once sync.Once
	db   js.Value
	err  error
}

// openDB opens the database once and keeps it open.
func openDB() (js.Value, error) {
	conn.once.Do(func() {
		factory := js.Global().Get("indexedDB")
		if factory.IsUndefined() || factory.IsNull() {
			conn.err = fmt.Errorf("%w: no indexedDB", ErrUnavailable)

			return
		}

		req := factory.Call("open", appName, idbVersion)

		upgrade := js.FuncOf(func(js.Value, []js.Value) any {
			db := req.Get("result")
			if !db.Get("objectStoreNames").Call("contains", storeROMs).Bool() {
				db.Call("createObjectStore", storeROMs)
			}

			return nil
		})
		defer upgrade.Release()

		req.Set("onupgradeneeded", upgrade)

		conn.db, conn.err = wait(req)
	})

	return conn.db, conn.err
}

// wait blocks until an IDBRequest succeeds or fails.
func wait(req js.Value) (js.Value, error) {
	done := make(chan error, 1)

	ok := js.FuncOf(func(js.Value, []js.Value) any {
		done <- nil

		return nil
	})
	defer ok.Release()

	fail := js.FuncOf(func(js.Value, []js.Value) any {
		done <- fmt.Errorf("%w: %s", ErrUnavailable, req.Get("error").Call("toString").String())

		return nil
	})
	defer fail.Release()

	req.Set("onsuccess", ok)
	req.Set("onerror", fail)

	if err := <-done; err != nil {
		return js.Value{}, err
	}

	return req.Get("result"), nil
}

func objectStore(mode string) (js.Value, error) {
	db, err := openDB()
	if err != nil {
		return js.Value{}, err
	}

	return db.Call("transaction", storeROMs, mode).Call("objectStore", storeROMs), nil
}

// defaultCacheDir is unused in the browser: IndexedDB has no folders.
func defaultCacheDir() string { return "" }

func (c *Cache) get(key string) ([]byte, bool, error) {
	st, err := objectStore("readonly")
	if err != nil {
		return nil, false, err
	}

	v, err := wait(st.Call("get", key))
	if err != nil {
		return nil, false, err
	}

	if v.IsUndefined() || v.IsNull() {
		return nil, false, nil
	}

	data := make([]byte, v.Get("byteLength").Int())
	js.CopyBytesToGo(data, v)

	return data, true, nil
}

func (c *Cache) put(key string, value []byte) error {
	st, err := objectStore("readwrite")
	if err != nil {
		return err
	}

	arr := js.Global().Get("Uint8Array").New(len(value))
	js.CopyBytesToJS(arr, value)

	_, err = wait(st.Call("put", arr, key))

	return err
}

func (c *Cache) delete(key string) error {
	st, err := objectStore("readwrite")
	if err != nil {
		return err
	}

	_, err = wait(st.Call("delete", key))

	return err
}
