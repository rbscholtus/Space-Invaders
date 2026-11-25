//go:build js && wasm

package platform

import (
	"net/url"
	"strconv"
	"syscall/js"
)

// URLParams reads the page's URL query. Bad values are ignored. See
// params.go.
func URLParams() Params {
	p := Params{Volume: -1}

	q, err := url.ParseQuery(trimQuestion(js.Global().Get("location").Get("search").String()))
	if err != nil {
		return p
	}

	if v, err := strconv.Atoi(q.Get("volume")); err == nil && v >= 0 && v <= 100 {
		p.Volume = v
	}

	return p
}

func trimQuestion(s string) string {
	if len(s) > 0 && s[0] == '?' {
		return s[1:]
	}

	return s
}
