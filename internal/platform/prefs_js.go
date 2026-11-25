//go:build js && wasm

package platform

import (
	"strings"
	"syscall/js"
)

// defaultPrefsPath is unused in the browser: preferences are in localStorage.
func defaultPrefsPath() string { return "" }

// prefsPrefix namespaces the app's keys in localStorage.
const prefsPrefix = appName + "."

// tryJS runs f and reports whether it finished: localStorage throws when
// storage is blocked (Safari private mode, disabled storage) or full.
func tryJS(f func()) bool {
	finished := false

	func() {
		defer func() { _ = recover() }()

		f()

		finished = true
	}()

	return finished
}

// localStorage returns window.localStorage, or false if it is blocked.
func localStorage() (js.Value, bool) {
	var st js.Value

	ok := tryJS(func() { st = js.Global().Get("localStorage") })

	return st, ok && !st.IsUndefined() && !st.IsNull()
}

// load reads the app's keys from localStorage.
func (p *Prefs) load() {
	st, ok := localStorage()
	if !ok {
		p.saved = false

		return
	}

	ok = tryJS(func() {
		for i := range st.Get("length").Int() {
			k := st.Call("key", i).String()
			if name, found := strings.CutPrefix(k, prefsPrefix); found {
				p.memory[name] = st.Call("getItem", k).String()
			}
		}
	})
	if !ok {
		p.saved = false
	}
}

// store writes one key. The lock is held.
func (p *Prefs) store(key, value string) bool {
	st, ok := localStorage()
	if !ok {
		return false
	}

	return tryJS(func() {
		if value == "" {
			st.Call("removeItem", prefsPrefix+key)
		} else {
			st.Call("setItem", prefsPrefix+key, value)
		}
	})
}
