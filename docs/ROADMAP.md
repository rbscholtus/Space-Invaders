# Roadmap

Milestones for the Go version. The design is in [DESIGN.md](DESIGN.md).

## M0: Groundwork ✅ (this PR)

- Java version moved to `java/`, tagged `java-2010`.
- Licence (BSD-3), `.gitignore`, `.golangci.yml`, CLAUDE.md, CONTRIBUTING.md,
  DESIGN.md, this roadmap.

## M1: The 8080

- `internal/i8080`: all 256 opcodes, flags, cycles, interrupts, `HLT`.
- Opcode table and disassembler; `cmd/trace`.
- `internal/cpm` and `cmd/cputest`; `8080PRE` and `8080EXM` pass.
- Verify the licences of the exercisers; commit them in `third_party/`.
- `go.mod`, CI (lint, tests, build), Dependabot.

## M2: The board, on the desktop

- `internal/invaders`: memory map, shift register, ports, interrupts per
  scanline, DIP switches, sound-latch events.
- `internal/romset`: zip/files, CRC check (verified CRCs from MAME).
- `cmd/invaders`: a plain Ebitengine window, keyboard, coins, overlay
  colours. The game is playable.
- Golden-frame tests (local, with ROMs).
- Resolve every **⚠ verify** in DESIGN §7–§10.

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
- Shaders: phosphor, overlay + backdrop (additive), CRT.
- Cabinet and screen views, auto by size, `-view` / `?view=` / `V` / setting.
- Clickable coin slot and buttons; touch controls in screen view; gamepads.
- DIP switch settings.

## Ideas

- Keep the high score between sessions (read it from work RAM ⚠ find the
  address in the disassembly).
- Desktop release binaries (`release.yml`).
- Cocktail cabinet view, using the flip bit.
- Other games on the Midway 8080 board.
- Codeberg mirror (DESIGN §17.3).
- A TinyGo build with a minimal web frontend, if download size ever matters.
