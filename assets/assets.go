// Package assets embeds the files compiled into the app: the cabinet's
// layout, cabinet/cabinet.json, in every build, and its images in the
// desktop build only. The browser fetches the images from next to the page
// (scripts/build.sh copies them there), so they don't grow the WebAssembly
// download. Every file's origin and licence is in LICENSES.md. This is the
// one package outside cmd/ and internal/: go:embed only reaches files at or
// below its package's folder.
package assets

import _ "embed"

// CabinetJSON is cabinet/cabinet.json, written by cmd/cabinet-gen: small,
// and it fixes the layout from the first frame, before the images arrive.
//
//go:embed cabinet/cabinet.json
var CabinetJSON []byte
