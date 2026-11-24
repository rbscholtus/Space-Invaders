# Contributing

These are the coding standards for the Go version of the Space Invaders
emulator. The design is described in [docs/DESIGN.md](docs/DESIGN.md); code
comments cite its sections as "DESIGN §N". The Java version in `java/` is
frozen: don't change it.

## Toolchain

- **Go 1.27**, latest patch release. `go.mod` declares that release (e.g.
  `go 1.27.4`) and has no separate `toolchain` line: this is an app that
  nothing imports, so its minimum Go and its build toolchain are the same.
  An older local Go downloads it (`GOTOOLCHAIN=auto`); CI's `setup-go`
  reads it from `go.mod`.
- **Ebitengine v2.10+**. Upgrades are their own PRs. No UI toolkit: menus and
  controls are our own widgets.
- **Dependencies:** as few as possible. Dependabot opens weekly update PRs.
  `govulncheck` runs in CI for the native and the WASM build; after changing
  dependencies, run `govulncheck ./...` and
  `GOOS=js GOARCH=wasm govulncheck ./...` locally.
- **golangci-lint v2.** The config is in `.golangci.yml`.
- **Node LTS** only for the Playwright browser tests in `e2e/`. No JavaScript
  in the app, apart from the small loader page.

## Linting

- **All linters on, minus a short list.** The config enables everything
  (`linters.default: all`) and disables only the very noisy or opinionated
  ones, as listed in `.golangci.yml`.
- **Disabling a linter for the whole project needs a one-line reason**, as a
  comment next to it in `.golangci.yml`.
- **`//nolint` must name the linter and give a reason:**
  `//nolint:gosec // G115: the 8080 bus is 16 bits; truncation is intended`.
  A bare `//nolint` is not allowed.
- **Lint must pass in CI** for both `GOOS=js GOARCH=wasm` and the native
  build. Run `golangci-lint run` before pushing.
- **Formatting:** `gofumpt` + `goimports` (through golangci-lint formatters).
  Kage files (`*.kage`) follow gofmt layout.

## Package rules

Go code lives in `cmd/` (programs) and `internal/` (packages). The only other
Go file is `assets/assets.go`, which embeds the data around it (DESIGN §5).

- **The core: `internal/i8080/`, `internal/invaders/`, `internal/cpm/`,
  `internal/synth/`:**
  - Standard library only, and no `reflect` (keeps them TinyGo-friendly).
  - No `syscall/js`, `time`, `math/rand` global state, `os` or other I/O.
  - They never read the clock. They advance only through `Step`, `Run`,
    `RunFrame` and `Render`.
  - `i8080` and `synth` import nothing from this module.
- **`internal/romset/`:** standard library only (`archive/zip`,
  `hash/crc32`, `crypto/sha1`); no Ebitengine, no `syscall/js`.
- **`internal/platform/`:** the only package with `syscall/js` and
  `net/http`. Every file that uses `syscall/js` is build-tagged
  `//go:build js && wasm` and has a `!(js && wasm)` stub, so native builds
  and tests work.
- **`internal/app/`:** Ebitengine. Keeps emulation out: it calls into
  `invaders`, `synth`, `romset` and `platform`.
- **`internal/testrig/`:** test helpers only, never imported by production
  code.
- **`cmd/`:** entry points and developer tools. Tools may use the file
  system.
- **`assets/`:** the files compiled into the app, and `assets.go`, which only
  embeds them (standard library only).

## Ebitengine rules

- **Never block `Update` or `Draw`.** Storage and file-picker work runs in
  goroutines and reports back on channels that `Update` polls without
  blocking (`select` with `default`).
- **No image or shader creation per frame.** Create `ebiten.Image`s and
  shaders at start-up or on resize, then reuse them.
- **Reuse draw options and uniforms.** Keep `DrawImageOptions`,
  `DrawRectShaderOptions` and their `Uniforms` maps as fields; update values
  in place.
- **Kage:** use `//kage:unit pixels`. Work in source coordinates
  (`srcPos - imageSrc0Origin()`), not `imageDstSize()`, which is the whole
  destination image.
- **HiDPI:** `Layout` returns real pixels (outside size × device scale
  factor). Sizes and fonts are scaled by the same factor, never hard-coded in
  CSS pixels.
- **Widgets:** keep layout and state in plain types that tests can drive
  without Ebitengine; only the drawing touches it.

## Memory and performance

- **Keep allocations in the per-frame path few, without making the code
  awkward.** The per-frame path is `RunFrame`, every instruction, the
  sound-event list, synth rendering, the video RAM → texture copy and shader
  uniform updates. Prefer the plain, readable version. Optimise later, when a
  benchmark or profile shows the need (DESIGN §16).
- **Sensible defaults in hot code:** preallocate buffers (video, audio,
  sound events, trace) once and reuse them; no closures per instruction.
- **`sync.Pool` is not used.** Add one only when a profile shows repeated
  allocations that can't be preallocated, and include benchmark results in
  the PR.
- **Avoid `fmt`, `reflect` and growing strings in hot code.** Disassembly,
  trace formatting and on-screen text are built only when they change.
- **PRs that touch `i8080`, `invaders`, `synth` or `app/crt`** include
  `benchstat` output against `main` if they could affect speed.

## Code style

- **Names:** follow Effective Go and the Go Code Review Comments. Opcode
  handlers are named after the instruction, e.g. `opMOV`, `opDAA`, `opLXI`.
  Port and bit names follow DESIGN §7 (`portIn1`, `in1Coin`, `out3UFO`).
- **Errors:** wrap with `%w` and context. Sentinel errors for conditions
  callers check (`ErrMissingROM`, `ErrCRCMismatch`, `ErrNotAROMSet`). Never
  `panic` on bad ROM input or bad files.
- **Comments:** every exported identifier has a doc comment. Hardware facts
  cite their source (Intel manual page, MAME driver, schematic sheet), or the
  DESIGN section that does.
- **Constants:** no unexplained numbers. Memory addresses, ports, bits, clock
  rates and synth parameters are named. Cabinet layout positions come from
  `cabinet.json`, not code.

## Tests

- **Every opcode has table-driven tests** for its result, flags and cycles.
  Every bug in DESIGN §3 has a test.
- **Unit tests** use the package itself (`package i8080`); conformance tests
  (exercisers, board, golden frames) use the `_test` package and only the
  exported API (DESIGN §5).
- **Exercisers:** `8080PRE` and `8080EXM` output is compared with golden
  text. `-short` skips `8080EXM`.
- **Golden frames need your own ROMs** (`roms/invaders.zip` or
  `INVADERS_ROMS`) and skip without them. CI doesn't run them; run them
  locally before a PR that touches `i8080`, `invaders` or video.
- **Goldens:** `go test ./... -update` regenerates goldens; regenerated ones
  must be checked in the PR (by eye for frames, by ear for audio). Never
  weaken a golden to make a test pass.
- **Deterministic:** seeded noise, frame counts instead of time, no network
  in `go test`.
- **Browser tests** (`e2e/`, Playwright): allow for software rendering in CI.
- **Coverage:** `internal/i8080/` aims for 95%+ line coverage. Coverage is
  reported in CI (`-coverpkg=./...`), not enforced.
- **Before pushing:** `go test ./...` and `go test -race -short ./...`
  (native), the WASM build, and the native build all pass. 8080EXM is
  skipped under `-short`, because it takes many minutes with `-race`.

## Assets

- **Every asset has a recorded origin and licence** in `assets/LICENSES.md`:
  generator, model and prompt for art, licence for fonts, source and licence
  for the exercisers.
- **Original art only** (DESIGN §12). Fonts: OFL or similarly free licences.
- **Never commit ROMs** or pictures made from them.

## Git

- **Branches:** short-lived branches off `main`. PRs need green CI.
- **Commit messages:** Conventional Commits, e.g. `feat(i8080): add DAA`,
  `fix(invaders): mirror RAM above 0x4000`, `test: ...`, `perf: ...`.
