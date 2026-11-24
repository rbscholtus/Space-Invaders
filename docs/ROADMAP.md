# Roadmap

Milestones for the Go version. The design is in [DESIGN.md](DESIGN.md).

## M0: Groundwork ✅

- Java version moved to `java/`, tagged `java-2010`.
- Licence (BSD-3), `.gitignore`, `.golangci.yml`, CLAUDE.md, CONTRIBUTING.md,
  DESIGN.md, this roadmap.

## Project setup ✅

- `go.mod` (Go 1.27.4).
- CI: lint (native + WASM), tests with the race detector (`-short`) and
  coverage, the CPU exercisers without `-race`, native and WASM builds;
  `govulncheck` for both builds. Ebitengine's Linux packages since M2.
- Dependabot for Go modules and GitHub Actions.

## M1: The 8080 ✅

- `internal/i8080`: all 256 opcodes, flags, cycles, interrupts, `HLT`;
  the `EI` delay verified in the Intel manual.
- Opcode table and disassembler; `cmd/trace`.
- `internal/cpm` and `cmd/cputest`; `8080PRE` and `8080EXM` pass.
- Tests for the commands too (`cli` functions).
- Licences of the exercisers verified (GPL v2 or later); committed in
  `third_party/`.

## M2: The board, on the desktop ✅

- Every **⚠ verify** in DESIGN §7–§10 resolved, against MAME's driver
  and the game code.
- `internal/invaders`: memory map, shift register, ports, interrupts per
  scanline, DIP switches, video, sound-port events.
- `internal/romset`: zip or folder, matched by CRC-32, checked by SHA-1
  (MAME's table).
- `cmd/invaders`: a plain Ebitengine window, keyboard, coins, pause, the
  overlay colours from MAME's artwork. The game is playable.
- `internal/testrig` and smoke tests with your own ROMs (local only).

## After M2: golden frames

- Golden-frame tests (local, with ROMs); decide hashes only or nothing in
  git (DESIGN §15.3).

## M3: Sound

- `internal/synth`: the seven sounds (§14), tuned against references.
- Audio output, amplifier enable, volume.

## M4: Browser and GitHub Pages

- `internal/platform` (from Chippie): IndexedDB, file picker, drop, prefs,
  URL params, `InFront`.
- ROM screen: drop/pick, errors, forget; the optional external link.
- CPU-test demo when no ROMs are loaded.
- Loader page, `scripts/build.sh`, `pages.yml`, Playwright smoke test.
- Measure the WASM size and record it in DESIGN §16.

## M5: The cabinet

- Original cabinet art and moon backdrop; `cabinet.json`.
- Shaders: phosphor, overlay + backdrop (additive), CRT. The overlay moves
  from `internal/app/overlay` (CPU) into the shader.
- Cabinet and screen views, auto by size, `-view` / `?view=` / `V` / setting.
- Clickable coin slot and buttons; touch controls in screen view; gamepads.
- DIP switch settings in the app (M2 has `-lives` and `-bonus1000` flags).

## Ideas

- Keep the high score between sessions (read it from work RAM ⚠ find the
  address in the disassembly).
- Desktop release binaries (`release.yml`).
- Cocktail cabinet view, using the flip bit.
- Other games on the Midway 8080 board.
- Codeberg mirror (DESIGN §17.3).
- A TinyGo build with a minimal web frontend, if download size ever matters.
