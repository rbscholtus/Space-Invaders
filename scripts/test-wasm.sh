#!/usr/bin/env bash
# Runs the tests as WebAssembly under Node (DESIGN §15.2): every package
# that doesn't use Ebitengine or the browser's storage (internal/platform),
# so the CPU, the board, the sound and the ROM handling are tested as the
# browser runs them, including the 8080EXM exerciser (about 3 minutes).
# Extra arguments go to go test, e.g. -short to skip 8080EXM.
set -euo pipefail
cd "$(dirname "$0")/.."

export GOOS=js GOARCH=wasm

pkgs=()
for p in $(go list ./...); do
	if [[ $p == */internal/platform ]] || go list -deps "$p" | grep -q '^github.com/hajimehoshi/ebiten'; then
		continue
	fi
	pkgs+=("$p")
done

go test -exec="$(go env GOROOT)/lib/wasm/go_js_wasm_exec" "$@" "${pkgs[@]}"
