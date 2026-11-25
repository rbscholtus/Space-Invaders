#!/usr/bin/env bash
# Builds the browser version into web/dist/ (DESIGN §17.1):
#   index.html            the loader page, with wasm_exec.js inlined
#   invaders-gz.wasm      the app, gzip-compressed (the loader unpacks it)
#   LICENSE, LICENSE-MAME, NOTICE.txt
#                         the licence notices binary releases must carry
#                         (BSD-3 clause 2, DESIGN §19)
#
# Serve it with:  cd web/dist && python3 -m http.server 8080
set -euo pipefail
cd "$(dirname "$0")/.."

out=web/dist

rm -rf "$out"
mkdir -p "$out"

GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o "$out/invaders.wasm" ./cmd/invaders
gzip -9 -n -c "$out/invaders.wasm" > "$out/invaders-gz.wasm"
raw_size=$(wc -c < "$out/invaders.wasm" | tr -d ' ')
rm "$out/invaders.wasm"

gz_size=$(wc -c < "$out/invaders-gz.wasm" | tr -d ' ')

python3 - "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$gz_size" "$out/index.html" <<'PY'
import sys
wasm_exec, gz_size, dst = sys.argv[1:]
page = open("web/page.html").read()
for marker in ("/*__WASM_EXEC__*/", "__WASM_GZ_SIZE__"):
    if marker not in page:
        sys.exit(f"web/page.html: marker {marker} missing")
page = page.replace("/*__WASM_EXEC__*/", open(wasm_exec).read()).replace("__WASM_GZ_SIZE__", gz_size)
open(dst, "w").write(page)
PY

cp LICENSE "$out/LICENSE"
cp internal/synth/LICENSE-MAME "$out/LICENSE-MAME"
cat > "$out/NOTICE.txt" <<'TXT'
Space Invaders emulator: https://github.com/rbscholtus/Space-Invaders

The emulator is BSD 3-Clause (LICENSE). Its sound emulation is ported from
MAME and keeps MAME's BSD 3-Clause notice (LICENSE-MAME). No ROMs are
included: Space Invaders is a trademark of Taito; this project is not
affiliated with Taito or Midway.
TXT

echo "built $out: invaders-gz.wasm $((gz_size / 1024)) KB ($((raw_size / 1024)) KB unpacked)"
