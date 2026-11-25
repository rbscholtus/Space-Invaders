# Space Invaders emulator: notes for Claude

A Space Invaders (Midway 8080 board, 1978) emulator written entirely in Go
with Ebitengine (own widgets), for the desktop and compiled to WebAssembly,
presented as an upright arcade cabinet. A port of the 2010 Java version in
`java/` (tag `java-2010`), which is kept for reference only.

- **Design:** [docs/DESIGN.md](docs/DESIGN.md) (cited as "DESIGN §N"). Read the
  relevant section before changing architecture, the memory map, timing,
  sound, ROM handling or the tests. §3 lists the Java version's bugs; don't
  port them. Facts marked **⚠ verify** must be checked against the
  references in DESIGN §18 before code relies on them; remove the mark and
  cite the source once checked.
- **README.md is for users:** what it is, how to play, ROMs, platforms,
  roadmap summary. Technical detail goes in DESIGN.md, milestones and ideas
  in [docs/ROADMAP.md](docs/ROADMAP.md).
- **Coding standards:** [CONTRIBUTING.md](CONTRIBUTING.md). Follow them
  strictly, especially the package rules and the Ebitengine rules. Keep
  per-frame allocations few, but don't make code awkward for it; optimise
  only when a benchmark or profile shows the need.
- **Go 1.27**, golangci-lint v2 with all linters on except those disabled in
  `.golangci.yml`. Never add a bare `//nolint`.
- **Before finishing a change:** `golangci-lint run` (also with
  `GOOS=js GOARCH=wasm`), `go test ./...`, `go build ./cmd/invaders`, and
  `./scripts/build.sh && (cd e2e && npx playwright test)` for anything the
  browser sees.
- **The core packages (`internal/i8080`, `internal/invaders`, `internal/cpm`,
  `internal/synth`) stay dependency-free and clock-free:** standard library
  only, no `time`, no I/O, no `reflect`.
- **`syscall/js` only in `internal/platform/`**, always with a native stub.
  **No `net/http` outside `internal/platform/`.**
- **Never block `Update` or `Draw`.** Async work goes in goroutines with
  channels.
- **ROMs are never committed, hosted or downloaded by the app.** Players
  bring their own (DESIGN §9). Don't add ROM files, ROM URLs or screenshots
  of the game to the repository. Golden-frame tests that need the ROM run
  locally only and skip without it.
- **Original art only.** Nothing copied from the real cabinet art, marquee
  or backdrop; "Space Invaders" is a Taito trademark (DESIGN §12).
- **Test the CPU against the 8080 exercisers** (`8080PRE`, `8080EXM`,
  DESIGN §15.2). Never weaken a golden (text, frame or audio) to make a test
  pass. Fix the emulator, or explain in the PR why the golden was wrong.
- **The user prefers clarifying questions** before big or ambiguous changes.
- **Lessons so far** are in CONTRIBUTING (Tests, Ebitengine rules, Git),
  DESIGN §7.6 (what the game does) and §14.5 (sound references and MAME's
  timing). In short: `go test ./pkg -update` (the
  package before the flag); no `ebitenutil`; check facts in the game code
  or by running it, not from memory; local reference PDFs in `docs/` are
  git-ignored. Browser: the e2e tests read `window.invadersState`
  (`platform.PublishState`); keep it up to date when the app's modes
  change. Sound: `internal/synth` is a port of MAME's circuits, so
  read the MAME source before changing a block, and keep
  `TestAgainstMAME` passing without loosening its tolerances.
- **Name issues and PRs in full** when writing to the user, in commits and
  in PRs: say whether it's an issue or a PR, and name the repository when
  it isn't this one, e.g. "Chippie PR #20", "Space-Invaders issue #3". A
  bare "#3" is ambiguous, because Chippie and this repository have
  overlapping numbers.
