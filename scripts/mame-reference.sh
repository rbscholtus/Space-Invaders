#!/usr/bin/env bash
# Records MAME's Space Invaders sounds, one WAV per sequence in
# internal/synth/testdata/reference/sequences.txt, as references for the synth
# (DESIGN §14.5). Needs MAME (e.g. `brew install mame`) and your own ROM set.
# The WAVs stay local (git-ignored); then run `go run ./cmd/soundref` to
# measure them into the committed .json files.
#
#   MAME_ROMPATH=~/mame/roms scripts/mame-reference.sh [sound ...]
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
ref="$root/internal/synth/testdata/reference"
rompath="${MAME_ROMPATH:-$HOME/mame/roms}"
mame="${MAME:-mame}"

command -v "$mame" >/dev/null || { echo "mame not found: install it or set MAME" >&2; exit 1; }

# macOS ships bash 3.2: no mapfile.
if [ $# -gt 0 ]; then
  sounds=("$@")
else
  sounds=()
  while read -r s; do sounds+=("$s"); done < <(awk '$1 == "sound" { print $2 }' "$ref/sequences.txt")
fi

mkdir -p "$ref/wav"
"$mame" -version > "$ref/wav/mame-version.txt"

for sound in "${sounds[@]}"; do
  echo "recording $sound"
  # The Lua script ends the run after the sequence; -seconds_to_run is a
  # safety stop (frames at 59.54 Hz, plus a margin).
  frames=$(awk -v s="$sound" '$1 == "sound" && $2 == s { print $3 }' "$ref/sequences.txt")
  [ -n "$frames" ] || { echo "no sequence named $sound" >&2; exit 1; }
  (
    cd "$ref/wav"
    INVADERS_SEQUENCES="$ref/sequences.txt" INVADERS_SOUND="$sound" \
      "$mame" invaders -rompath "$rompath" -samplerate 48000 \
      -video none -sound none -nothrottle -skip_gameinfo \
      -seconds_to_run $((frames / 59 + 5)) \
      -autoboot_script "$root/scripts/mame/sounds.lua" \
      -wavwrite "$sound.wav" >/dev/null
  )
done
