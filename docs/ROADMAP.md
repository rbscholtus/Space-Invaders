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

## After M2: golden frames ✅

- `TestGoldenSession` (local, with ROMs): a 6 min 40 s scripted session
  (attract, a 1-player and a 2-player game to game over) with picture and
  sound hashes every half second. Only hashes in git (DESIGN §15.3).

## M3: Sound ✅

- Reference tooling: MAME records each sound (Lua + `-wavwrite`, local);
  `cmd/soundref` commits measurements (DESIGN §14.5).
- `internal/synth`: MAME's circuit building blocks, the six discrete
  circuits, the SN76477 for the UFO, the mixer and mute, ported to Go and
  tested against the schematic's maths and MAME's output: waveform
  correlation 0.9989–1.0000 per sound.
- `internal/app/sound`: audio output at 48 kHz, events at the right sample,
  volume (`-volume`, `-`/`=`) and mute (`M`).
- The golden session against MAME's (local): with MAME's input timing the
  game runs exactly as in MAME, and its sound matches (0.9996).

## M4: Browser and GitHub Pages ✅

- `internal/platform` (from Chippie): IndexedDB, prefs, URL params, file
  picker, `InFront`, a state hook for the browser tests.
- Our own 8×8 pixel font for on-screen text (`internal/app/font`).
- ROM screen (`internal/app/romscreen`): drop (browser and desktop), pick
  (`O`), remembered, forget (`F`), clear errors; a zip inside a dropped
  folder works; no external link (DESIGN §9.3). `-roms` is optional.
- The controls at the bottom of the ROM screen. A CPU-test demo there
  (8080EXM) was built and then dropped: it didn't help players, and the
  exercisers stay test-only (DESIGN §9.4). Instead the tests also run as
  WebAssembly under Node, 8080EXM included (`scripts/test-wasm.sh`, CI).
- Audio starts on the first gesture in the browser; volume and mute
  remembered.
- Loader page, `scripts/build.sh`, `pages.yml`, Playwright tests (no ROMs
  needed in CI; a ROM test locally), a 4 MB size budget in CI.
- The licence notices ship with the site (`LICENSE`, `LICENSE-MAME`,
  `NOTICE.txt`; DESIGN §19).
- WASM size measured: 3.4 MB gzip (DESIGN §16).
- Keyboard only: touch controls come with the cabinet (M5).

## After M4: the HI-SCORE remembered ✅

- The HI-SCORE is kept between sessions (`platform.Prefs`), put back as
  the board's power-on value so the game shows it from the start; `F`
  forgets it with the ROMs (DESIGN §7.6).
- A global leaderboard was considered: a server that trusts the client's
  score can be faked by anyone, and one that replays the game to check it
  would need a copy of the ROM. Not planned.

## M5: The cabinet

- Original cabinet art and moon backdrop; `cabinet.json`.
- Shaders: phosphor, overlay + backdrop (additive), CRT. The overlay moves
  from `internal/app/overlay` (CPU) into the shader.
- Cabinet and screen views, auto by size, `-view` / `?view=` / `V` / setting.
- Clickable coin slot and buttons; touch controls in screen view; gamepads.
- DIP switch settings in the app (M2 has `-lives` and `-bonus1000` flags).

## Ideas

- Desktop release binaries (`release.yml`).
- Cocktail cabinet view, using the flip bit.
- A flag (and setting) to choose the Midway or the Taito sound board. MAME
  notes where the Taito PCB differs (the shot's C12 is 0.22 µF and R13
  2.2 MΩ, and its VCO circuit isn't fully modelled), so the shot would
  change; needs Taito reference recordings to test against.
- Other games on the Midway 8080 board.
- Codeberg mirror (DESIGN §17.3).
- A TinyGo build with a minimal web frontend, if download size ever matters.
