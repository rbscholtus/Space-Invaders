# Space Invaders emulator: design

How the Go port is built, and why. For what it is and how to play it, see the
[README](../README.md); for coding standards,
[CONTRIBUTING.md](../CONTRIBUTING.md); for what's done and what's next,
[ROADMAP.md](ROADMAP.md).

Section numbers are stable: code comments cite them as "DESIGN §N".

Facts marked **⚠ verify** come from memory or secondary sources and must be
checked against the references in §18 (MAME's `mw8080bw` driver, the Midway
schematics, the Computer Archeology disassembly) before code relies on them.
Remove the mark once checked, and cite the source.

---

## Contents

1. [Goals](#1-goals)
2. [Key decisions](#2-key-decisions)
3. [Lessons from the Java version](#3-lessons-from-the-java-version)
4. [Architecture](#4-architecture)
5. [Repository layout](#5-repository-layout)
6. [The Intel 8080 core](#6-the-intel-8080-core)
7. [The Space Invaders board](#7-the-space-invaders-board)
8. [Timing](#8-timing)
9. [ROMs](#9-roms)
10. [Video](#10-video)
11. [Presentation: the arcade cabinet](#11-presentation-the-arcade-cabinet)
12. [Cabinet artwork](#12-cabinet-artwork)
13. [Input](#13-input)
14. [Sound](#14-sound)
15. [Testing](#15-testing)
16. [Browser integration and performance](#16-browser-integration-and-performance)
17. [Build, CI and deployment](#17-build-ci-and-deployment)
18. [References](#18-references)

---

## 1. Goals

- **Play Space Invaders in the browser with nothing to install.** Hosted on
  GitHub Pages; works on phones, tablets and desktops.
- **The same program on the desktop**, from one Go code base (Ebitengine).
- **Correct emulation.** The Intel 8080 passes the standard CPU exercisers
  (§15), and the board behaves as the Midway hardware does: shift register,
  ports, interrupts on the right scanlines, coin and DIP switches.
- **Sound without samples.** The analog sound circuits are imitated with
  synthesis (§14), so no recordings need to be hosted.
- **No ROMs hosted.** Players bring their own ROM set; the site never serves
  or downloads it (§9). Without ROMs, a CPU-test demo runs instead.
- **Retro presentation.** A 1978 upright arcade cabinet, with the colour
  overlay and the moon backdrop seen through the half-mirror. On small screens,
  only the screen, overlay and backdrop (§11).
- **One language.** Everything is Go, including the UI and the shaders (Kage).

Out of scope for now: other games on the same Midway 8080 board (Lunar
Rescue, Space Invaders Part II, ...), save states, a debugger UI (§3), netplay.

## 2. Key decisions

| Topic | Decision |
|---|---|
| Language | Go 1.27 (latest patch), everything including UI. Shaders in Kage. |
| Engine | Ebitengine v2.10+ (game loop, rendering, input, audio), as in Chippie. Our own widgets; no UI toolkit. |
| Targets | Browser (`GOOS=js GOARCH=wasm`) and desktop (macOS, Windows, Linux), both first-class. |
| TinyGo | Not used: Ebitengine doesn't support it. The core packages stay TinyGo-friendly (standard library only, no `reflect`), so a tiny web frontend stays possible later. |
| CPU core | A fresh Go 8080 in `internal/i8080`, written against the Intel manual and the exercisers, not translated line by line from the Java (§3). |
| ROMs | Bring your own: drop a MAME `invaders.zip` or the four files; checked by CRC-32; kept in IndexedDB (browser) or a user folder (desktop). An optional link to an external page the player opens themselves. Never hosted or fetched by us. |
| No-ROM mode | A CPU-test demo: the 8080 exerciser (GPL) running on the cabinet screen (§9.4). |
| Sound | Synthesised from models of each analog circuit (§14). No samples. |
| Look | Upright cabinet with all decorations; "screen" view (screen + overlay + moon) on small screens, or when chosen (§11). |
| Art | Original artwork only. Nothing copied from Taito/Midway cabinet art, marquee or backdrop (§12). |
| Debugger | No GUI. A headless trace tool and an on-screen stats overlay instead (§3). |
| Tests | CPU exercisers in CI; golden frames from the real ROM only locally, because CI has no ROMs (§15). |
| Hosting | GitHub (code, Actions, Pages). Codeberg is possible later (§17.3). |
| Memory | Few allocations in the per-frame path, without awkward code; optimise when profiles show the need. |

**Why not keep Java?** Java Web Start was removed in Java 11 and applets are
gone, so today the Java version needs a JDK, a build and the ROMs and samples
copied into the source tree. Running Java in the browser (CheerpJ) would work
but adds a large runtime and a third-party dependency for a 3,000-line program.
A Go rewrite reuses everything Chippie already solved: WASM loading, IndexedDB,
touch input, audio unlock, Kage CRT shaders and GitHub Pages deployment.

## 3. Lessons from the Java version

The Java emulator (`java/`, tag `java-2010`) runs the game, but has these bugs.
Don't port them; each one gets a test in the Go version (§15).

**CPU (`I8080.java`):**

| Bug | Correct behaviour |
|---|---|
| `interrupt()` ignores `inte` | An interrupt is accepted only when interrupts are enabled (`EI`); accepting one disables them. |
| `HLT` never wakes up | An accepted interrupt resumes a halted CPU. |
| `reset()` sets `inte = true` | The 8080 starts with interrupts disabled. |
| `DAA` uses `A &= 0x0f` for the high digit | It adds `0x60`. The low-digit `+6` can carry into the high digit; `A` stays 8-bit; carry is set if the high adjustment carries, otherwise unchanged. |
| `DCR` sets AC when the result's low nibble is `0xF` | The 8080 sets AC when there was *no* borrow from bit 4, i.e. when the low nibble is not `0xF`. |
| `SUB`/`SBB`/`CMP` compute AC as for addition | The 8080 subtracts by adding the complement; AC is the carry out of bit 3 of that addition. |
| `ANA` leaves AC unchanged | AC = bit 3 of `(A \| operand)`. |
| `XRA`/`ORA` leave AC unchanged | They clear AC (and carry). |
| Extra cycles at the end of each half-frame are dropped | Carry them over, or emulated time slowly drifts. |

**Board (`InvadersFrame.java`):**

| Bug | Correct behaviour |
|---|---|
| Reads at `0x4000` and above go out of bounds | Reads and writes both follow the memory map, with RAM mirrors (§7.1). |
| RAM mirror uses `0x2000 \| (addr & 0x3FF)` | Mirror per the address decoding (§7.1). |
| ROM (`0x0000–0x1FFF`) is writable | Writes to ROM are ignored. |
| `writeWord` skips the pixel update | All writes to video RAM update the screen (or, better, the screen is built from video RAM each frame, §10). |
| Port 1 bit 3 is 0 | It always reads 1. |
| Coin bit clears when read | Hold the coin bit for a few frames, as a real coin switch does (§13). |
| Mid-screen interrupt at half the frame's cycles | At the scanline where the hardware fires it (§8). |
| Sounds trigger on a level, march sounds restart on every write | Trigger on the 0→1 edge of each bit; the UFO loops while its bit is held (§14). |
| Amplifier enable (port 3 bit 5) ignored | All sound is muted while it's 0. |
| Port 5 bit 5 (flip screen) and port 6 (watchdog) ignored | Flip is implemented for cocktail mode; watchdog writes are accepted and ignored (or counted for tests). |

**The debugger GUI** (`i8080/gui`) was needed when the 8080 was written by
hand with few references. The 8080 is now thoroughly documented and the
exercisers point at failing instructions directly, so the Go version has no
debugger UI. What remains useful:

- `cmd/trace`: runs a ROM or exerciser headless and writes one line per
  instruction (PC, opcode, disassembly, registers, flags, cycle count), to
  compare with a known-good emulator's log using `diff`.
- An on-screen stats overlay (FPS, emulated frame rate, audio buffer level),
  toggled with a key.
- The opcode table from `I8080OpInfo.java` (mnemonic, length, cycles), as
  data for the disassembler and the cycle counts.

## 4. Architecture

```
┌─────────────────────────────── Browser / desktop ──────────────────────────┐
│  web/index.html   loader: download + unpack + start invaders.wasm          │
│                                                                            │
│  ┌──────────────────────── invaders(.wasm) ─────────────────────────────┐  │
│  │                                                                      │  │
│  │  app/        Ebitengine Game: Update (60 TPS) / Draw (per refresh)   │  │
│  │   ├─ cabinet/  cabinet skin, layout per view, coin slot, buttons     │  │
│  │   ├─ crt/      Kage shaders: phosphor, overlay + backdrop, CRT       │  │
│  │   ├─ input/    keyboard, touch, gamepad → board inputs               │  │
│  │   └─ sound/    synth → Ebitengine audio player                       │  │
│  │                                                                      │  │
│  │  romset/     find, unzip, CRC-check, cache the ROM set               │  │
│  │  platform/   browser-only bits: IndexedDB, file picker,              │  │
│  │              localStorage, URL params (stubs on desktop)             │  │
│  │  synth/      sound circuits as synthesis models (no clock, no I/O)   │  │
│  │  invaders/   the board: memory map, shifter, ports, interrupts,      │  │
│  │              DIP switches, sound latches (no clock, no I/O)          │  │
│  │  cpm/        minimal CP/M for the exercisers (tests + demo)          │  │
│  │  i8080/      the CPU, disassembler, opcode table (no dependencies)   │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────────────┘
```

**Dependency direction:** `app` → `romset`, `platform`, `synth`, `invaders`,
`cpm`. `invaders` and `cpm` → `i8080`. `i8080` and `synth` depend on nothing.
Only `platform` imports `syscall/js`.

**The frame:**

1. Ebitengine calls `Update` 60 times per second (fixed TPS).
2. `Update` reads input, sets the board's input bits, and calls
   `board.RunFrame()` once, unless paused.
3. `RunFrame` runs one video frame of CPU time (§8), firing the two
   interrupts, and records sound-latch changes with their cycle times.
4. `app/sound` hands those events to `synth`, which renders this frame's
   samples into the audio stream.
5. Ebitengine calls `Draw` once per display refresh. `Draw` converts video
   RAM into a preallocated texture (when it changed), runs the phosphor,
   overlay/backdrop and CRT shaders, then draws the cabinet around it.

## 5. Repository layout

The same shape as Chippie: programs in `cmd/`, packages in `internal/`, files
compiled into the app in `assets/`.

```
Space-Invaders/
├── go.mod                      # module github.com/rbscholtus/Space-Invaders; go 1.27.x
├── cmd/
│   ├── invaders/               # main: ebiten.RunGame (browser + desktop)
│   ├── trace/                  # headless run with an instruction trace (§3)
│   └── cputest/                # runs the CP/M exercisers in the terminal
├── internal/
│   ├── i8080/                  # CPU: stdlib only, no clock, no I/O
│   │   ├── cpu.go  ops.go  flags.go  opcodes.go  disasm.go
│   │   ├── testdata/           # disassembly goldens
│   │   └── *_test.go
│   ├── invaders/               # the board
│   │   ├── board.go  memory.go  ports.go  shifter.go  video.go  dips.go  sound.go
│   │   ├── testdata/golden/    # golden frames (made locally from the real ROM, §15.3)
│   │   └── *_test.go
│   ├── cpm/                    # BDOS calls 2 and 9, warm boot = done
│   ├── synth/                  # one model per sound (§14); renders float32 PCM
│   ├── romset/                 # zip/files → verified 8 KB image (§9)
│   ├── platform/               # syscall/js lives only here (_js.go + _other.go pairs)
│   ├── testrig/                # test helpers (golden frames, ROM lookup)
│   └── app/
│       ├── cabinet/  crt/  input/  sound/
│       └── game.go  layout.go  settings.go  osd.go
├── assets/                     # cabinet art, backdrop, fonts, assets.go (go:embed)
│   └── LICENSES.md             # origin and licence of every asset
├── third_party/
│   └── 8080-exercisers/        # 8080PRE/8080EXM + sources + GPL licence (§15.2)
├── web/                        # loader page; build output in web/dist/ (ignored)
├── e2e/                        # Playwright browser tests
├── scripts/build.sh
├── docs/                       # DESIGN.md, ROADMAP.md, references, investigations
└── java/                       # the Java version (2010), for reference only
```

**Test files:** Go requires `_test.go` files next to the code they test.
Unit tests use the package itself (`package i8080`), because opcode tests
need registers and flags. Conformance tests (exercisers, golden frames,
board-level behaviour) use the black-box package (`package invaders_test`)
and only the exported API. Both run with `go test ./...`; coverage across
packages needs `-coverpkg=./...` (set in `scripts/` and CI).

## 6. The Intel 8080 core

### 6.1 Principles

- **No dependencies, no clock, no I/O.** The CPU advances only through
  `Step` and `Run(cycles)`, so every run is repeatable.
- **The bus is an interface.** The board and the CP/M harness implement it.
  At 2 MHz, an interface call per memory access costs far less than a frame's
  budget; change it only if a profile says so (§16).
- **Cycle counts are exact per instruction**, including the different counts
  for taken and not-taken conditional calls and returns (11/5 for `Rcc`,
  17/11 for `Ccc`). Jumps take 10 cycles either way.
- **Flags as a byte.** Keep `F` as the real PSW byte (bit 1 always 1, bits 3
  and 5 always 0), so `PUSH PSW`/`POP PSW` and the exercisers' CRCs come out
  right without conversion. Parity from `math/bits.OnesCount8` (itself a
  256-entry table).
- **Undocumented opcodes** behave as on a real 8080: `0x08 0x10 0x18 0x20
  0x28 0x30 0x38` are `NOP`, `0xCB` is `JMP`, `0xD9` is `RET`, `0xDD 0xED
  0xFD` are `CALL`.

### 6.2 Public API (sketch)

```go
package i8080

// Bus is what the CPU sees: memory and the I/O ports.
type Bus interface {
    Read(addr uint16) byte
    Write(addr uint16, v byte)
    In(port byte) byte
    Out(port byte, v byte)
}

type CPU struct {
    A, B, C, D, E, H, L, F byte
    SP, PC                 uint16
    Halted, IntEnabled     bool
    Cycles                 uint64 // total since Reset
    // bus Bus (unexported)
}

func New(bus Bus) *CPU
func (c *CPU) Reset()                       // PC=0, interrupts off, not halted
func (c *CPU) Step() int                    // one instruction (or 4 idle cycles when halted); returns cycles
func (c *CPU) Run(cycles int) int           // until at least `cycles` have run; returns the overshoot
func (c *CPU) Interrupt(opcode byte) bool   // RST n from the bus; false if interrupts are off
func (c *CPU) EnableTrace(n int)            // ring buffer, allocated once
func (c *CPU) Trace() []State               // the last n states, oldest first
func (c *CPU) Snapshot() State              // registers, flags and the bytes at PC
func Op(op byte) OpInfo                     // mnemonic, length, cycles (taken/not)
func Disassemble(mem []byte, pc uint16) (string, int)
```

### 6.3 Interrupts

The board places an `RST n` opcode on the data bus. When interrupts are
enabled, the CPU: disables interrupts, leaves the halt state, pushes `PC`
and jumps to `n*8` (11 cycles). When they're disabled, the request is
ignored; Space Invaders keeps them enabled except in short critical sections,
and the board re-raises its request each frame anyway. `EI` takes effect
after the *next* instruction (so `EI; RET` returns before an interrupt is
taken). ⚠ verify the `EI` delay against the Intel manual: implemented and
tested (`TestEIDelay`), but the exercisers don't cover it, and the scanned
manual found online has no text layer to check it against.

## 7. The Space Invaders board

Midway's 8080 board (1978), also used by other Midway/Taito games. CPU clock
19.968 MHz / 10 = **1.9968 MHz**.

### 7.1 Memory map

| Address | Size | Contents |
|---|---|---|
| `0x0000–0x07FF` | 2 KB | ROM `invaders.h` |
| `0x0800–0x0FFF` | 2 KB | ROM `invaders.g` |
| `0x1000–0x17FF` | 2 KB | ROM `invaders.f` |
| `0x1800–0x1FFF` | 2 KB | ROM `invaders.e` |
| `0x2000–0x23FF` | 1 KB | Work RAM (stack at the top, near `0x2400`) |
| `0x2400–0x3FFF` | 7 KB | Video RAM, 1 bit per pixel (§10) |
| `0x4000–` | | Mirrors. ⚠ verify the decoding in MAME's `mw8080bw` memory map; the game is known to read/write slightly beyond `0x3FFF` in some situations. |

Writes to ROM are ignored. Memory is one flat 16 KB array; `Read`/`Write`
mask the address according to the map.

### 7.2 Input ports

| Port | Bit | Meaning |
|---|---|---|
| 1 | 0 | Coin (credit), 1 = coin inserted |
| 1 | 1 | 2P start |
| 1 | 2 | 1P start |
| 1 | 3 | Always 1 |
| 1 | 4 | P1 fire |
| 1 | 5 | P1 left |
| 1 | 6 | P1 right |
| 1 | 7 | Unused |
| 2 | 0–1 | DIP: lives, 00 = 3, 01 = 4, 10 = 5, 11 = 6 |
| 2 | 2 | Tilt |
| 2 | 3 | DIP: bonus life at 1000 (1) or 1500 (0) |
| 2 | 4 | P2 fire |
| 2 | 5 | P2 left |
| 2 | 6 | P2 right |
| 2 | 7 | DIP: coin info shown in attract mode, 0 = on |
| 3 | 0–7 | Shift register result (§7.4) |

Port 0 is read by the game but has no function on Space Invaders. ⚠ verify
which bits port 0 returns (MAME documents it).

### 7.3 Output ports

| Port | Bits | Meaning |
|---|---|---|
| 2 | 0–2 | Shift amount (§7.4) |
| 3 | 0 | UFO (repeats while 1) |
| 3 | 1 | Shot |
| 3 | 2 | Player dies ("flash") |
| 3 | 3 | Invader dies |
| 3 | 4 | Extended play (bonus life) |
| 3 | 5 | Amplifier enable |
| 4 | 0–7 | Shift register data (§7.4) |
| 5 | 0–3 | Fleet movement, notes 1–4 |
| 5 | 4 | UFO hit |
| 5 | 5 | Flip screen (cocktail cabinet, player 2's turn) |
| 6 | – | Watchdog reset |

### 7.4 Shift register

The 8080 has no barrel shifter, so the board has one, used to draw sprites at
any x position.

- **Write port 4:** the 16-bit register shifts right by 8 and the new byte
  goes into the high byte: `reg = uint16(v)<<8 | reg>>8`.
- **Write port 2:** `amount = v & 7`.
- **Read port 3:** `byte(reg >> (8 - amount))`.

### 7.5 DIP switches and settings

The board's DIP switches (lives, bonus life, coin info) are settings in the
app, stored in preferences, and applied on the next power-on. Defaults: 3
lives, bonus at 1500, coin info on (the factory setting, ⚠ verify).

## 8. Timing

**The video frame** (⚠ verify against MAME's `mw8080bw` constants):

- Pixel clock 19.968 MHz / 4 = 4.992 MHz; 320 pixels per line; 262 lines.
- Frame rate 4,992,000 / (320 × 262) = **59.54 Hz**.
- CPU cycles per line: 1,996,800 / (59.54 × 262) = **128**; per frame
  **33,536**. (One CPU cycle is 2.5 pixels.)
- 224 visible lines; the rest is vertical blank.

**Interrupts:** two per frame.

- `RST 1` (`0xCF`) when the beam is in the middle of the visible picture. The
  game then updates the objects in the top half, which the beam has already
  drawn.
- `RST 2` (`0xD7`) at the start of vertical blank. The game updates the
  bottom half.

⚠ verify the exact line numbers (MAME triggers on vertical counter values; the
common description is "line 96" and "line 224").

**In the emulator:**

- `board.RunFrame()` runs the frame line by line: 262 × 128 cycles, carrying
  each instruction's overshoot into the next line, and raises each interrupt
  at its line. This replaces the Java version's two half-frames.
- **One emulated frame per `Update`.** Ebitengine runs `Update` at 60 Hz, the
  game at 59.54 Hz, so the game runs 0.8% fast. Nobody can hear or see 0.8%,
  and it avoids the judder of skipping one `Update` in 130. Audio follows the
  emulated frame: each frame renders 1/60 s of samples (§14.4), so the audio
  buffer never drifts.
- **Pausing:** when the tab is hidden (`platform.InFront`, as in Chippie), when
  a menu is open, or with a pause key. Natively, when the window loses focus.
- **Speed is not adjustable** in normal play (the game's speed is its timing),
  except a fast-forward key for the exerciser demo.

## 9. ROMs

### 9.1 What's needed

The MAME set **`invaders`** (Space Invaders, Midway, 1978): four 2 KB files.

| File | Address | CRC-32 |
|---|---|---|
| `invaders.h` | `0x0000` | `734f5ad8` ⚠ verify |
| `invaders.g` | `0x0800` | `6bfaca4a` ⚠ verify |
| `invaders.f` | `0x1000` | `0ccead96` ⚠ verify |
| `invaders.e` | `0x1800` | `14e538b0` ⚠ verify |

Verify the CRCs (and add SHA-1s) from MAME's driver source before shipping;
the table becomes `internal/romset/sets.go`. Other versions (Taito's `sitv`,
`sicv`, ...) may be added later, each with its own table.

### 9.2 Bringing your own

The ROMs are copyrighted by Taito. They are never committed, hosted, or
downloaded by the app.

- **Browser:** drop a `.zip` or the four files on the page, or pick them with
  the file picker. `romset` reads the zip (`archive/zip`, in memory), matches
  files by CRC-32 (so renamed files still work), checks all four are present,
  and builds the 8 KB image. The image is stored in IndexedDB, so the next
  visit starts at once. A "forget ROMs" setting removes it.
- **Desktop:** the same drop and picker, and a `-roms <path>` flag (zip or
  folder). The image is cached in the user config folder.
- **Errors** say exactly what's missing or wrong: "invaders.f is missing",
  "invaders.zip has no Space Invaders ROMs (found: ...)", "CRC mismatch: this
  looks like a bootleg set".

### 9.3 The external link

The ROM screen may also show a link to an external page about obtaining the
ROMs. The player opens it in a new tab themselves; the app never fetches
from it. Which page (if any) is the owner's decision, documented in the
README; it must not be a direct download link to the ROM files.

### 9.4 Without ROMs: the CPU-test demo

When no ROMs are loaded, the cabinet still does something: it runs the 8080
exerciser `8080EXM` (GPL, §15.2) under the `cpm` harness, at full speed, and
shows its output on the cabinet screen in a 1978-style font: each instruction
group and "OK" or "ERROR". This shows the emulator works, gives the browser
tests something to render without ROMs, and makes the "load your ROMs" message
less of a dead end.

The exerciser files are served next to the WASM (not compiled into it), with
their licence and source, like Chippie's test-suite ROMs.

## 10. Video

- **Resolution:** 256 × 224, 1 bit per pixel, 32 bytes per row, 224 rows.
- **Rotated monitor:** the monitor is mounted turned 90° anticlockwise, so the
  player sees **224 wide × 256 tall**. Byte `i` of video RAM (`0x2400 + i`)
  holds pixels of memory row `i / 32`, x = `(i % 32) * 8 + bit`, with bit 0 the
  leftmost in memory. On screen: column `x_screen = i / 32`, row
  `y_screen = 255 - ((i % 32) * 8 + bit)`.
- **Build the picture once per frame from video RAM,** not on every write.
  It's 7 KB, and converting per frame keeps the board free of display code
  and makes golden frames trivial (§15.3). `invaders.Board.Video()` returns a
  read-only view; `app` turns it into an `ebiten.Image` with `WritePixels` on
  a preallocated buffer.
- **Monochrome output.** The board outputs white only. Colour comes from the
  overlay (§11.2), applied by the shader, never by the board.
- **Flip:** in cocktail mode, port 5 bit 5 rotates the picture 180° for
  player 2. Upright cabinets ignore it.
- **Mid-frame tearing:** the real game relies on updating each half after the
  beam has passed it. The emulator shows the whole frame at once; this is
  standard and invisible. If artefacts show, consider rendering the top half
  at `RST 1` and the bottom at `RST 2`.

## 11. Presentation: the arcade cabinet

### 11.1 Views

| View | What's drawn | When |
|---|---|---|
| **Cabinet** | The whole upright cabinet: marquee, bezel, screen, control panel (joystick or buttons, fire, 1P/2P start), coin door with a coin slot. | Default when the window is big enough. |
| **Screen** | Only the picture: phosphor + colour overlay + moon backdrop, with a thin bezel. Touch controls under or beside it on touch screens. | Default on small screens (phones), or when chosen. |

- **Auto:** "Screen" when the picture in cabinet view would be smaller than
  about 2× native height (512 px of real pixels) ⚠ tune on devices.
- **Overrides:** `-view=auto|cabinet|screen` (desktop flag), `?view=...`
  (URL parameter, through `platform` params as in Chippie), a key (`V`) and a
  setting. A choice made by the player is remembered.

### 11.2 The colour overlay

The original monitor is black and white. Coloured cellophane strips on the
screen tint parts of it: a **red** strip near the top (where the UFO flies)
and a **green** strip at the bottom (shields and player), including the
lives area at the bottom left. In the Java version: red for screen rows
32–63, green for rows 184–239, and green for row 240+ when x is 16–133.
⚠ verify against photos of real cabinets and MAME's artwork.

The overlay is a small texture (or a list of rectangles) sampled by the
shader, so a different overlay (e.g. the Taito version's) is data, not code.

### 11.3 The moon backdrop and the half-mirror

In the upright cabinet the monitor faces up or back and is seen through a
half-silvered mirror, with a lit backdrop painting (a moon landscape) behind
it. Where the screen is black, the backdrop shows; where it's lit, the
phosphor is brighter than the backdrop. The shader therefore composites
**additively**: `out = backdrop * backdropLevel + phosphor * overlay`, not a
plain blend. The backdrop is our own art (§12).

### 11.4 CRT

A Kage CRT shader like Chippie's: phosphor persistence (Space Invaders'
flickering sprites look wrong without some afterglow), slight bloom,
scanlines and curvature. Reuse Chippie's `crt.kage` and `phosphor.kage` as a
starting point, adjusted for a rotated, vertical picture.

## 12. Cabinet artwork

- **Original art only.** The real cabinet's side art, marquee, bezel and
  backdrop are copyrighted (Taito/Midway), and "Space Invaders" is a Taito
  trademark. The cabinet is a generic 1978 upright in that style, with our
  own marquee text and our own moon backdrop. No screenshots or scans of real
  artwork.
- **Pipeline:** as in Chippie: AI-generated masters (outside the repository,
  `assets/source/`), a recipe, and a generator that writes pre-scaled `@1x`,
  `@2x`, `@3x` files into `assets/cabinet/`. Layout positions (screen, buttons,
  coin slot) come from a `cabinet.json`, not code.
- **Every asset** has its origin and licence recorded in `assets/LICENSES.md`.

## 13. Input

**Keyboard** (defaults; remappable later):

| Key | Input |
|---|---|
| `5` or `C` | Insert coin |
| `1` | 1P start |
| `2` | 2P start |
| `←` `→` (or `A` `D`) | Move |
| `Space` (or `↑`, `W`) | Fire |
| `T` | Tilt |
| `P` | Pause |
| `V` | Switch view |
| `F1` | Stats overlay |

**Coins:**

- A coin is a pulse: `Update` sets port 1 bit 0 for a few frames (⚠ tune;
  the game counts the switch's edge and debounces it), then clears it. Holding
  the key inserts one coin; each new press inserts another.
- In cabinet view, the coin slot on the coin door is clickable/tappable, with
  a short coin-drop animation and sound.
- In screen view, a coin button sits with the touch controls.

**Two players:** on an upright cabinet both players use the same controls,
taking turns; the game reads player 2's bits (port 2) only in cocktail mode
⚠ verify in the disassembly. By default, the same keys drive both players'
bits, so it doesn't matter.

**Touch:** in screen view, on-screen left/right/fire, coin, 1P and 2P.
In cabinet view, the cabinet's own buttons are the controls. Multi-touch,
so moving and firing work together (as in Chippie's remote).

**Gamepads:** Ebitengine's standard gamepad layout: d-pad/stick to move, A to
fire, Select = coin, Start = 1P start.

**Tilt:** the tilt switch ends the current game, as on the real cabinet.

## 14. Sound

### 14.1 Why synthesis, not samples

The original sounds are made by analog circuits on the sound board (Midway's
discrete circuits and an SN76477 sound chip ⚠ verify which sounds use the
SN76477). MAME historically used recorded samples, then modelled the circuits.

| Approach | Pros | Cons |
|---|---|---|
| Samples (Java version) | Exact recordings | Recordings' rights are unclear and can't be hosted; samples need files the player must find; looping and retriggering artefacts; fixed pitch and length |
| Circuit simulation (SPICE-like) | Most faithful | Large, slow in WASM, hard to get right |
| **Synthesis models** (chosen) | Small, no files, tunable, deterministic and testable | Must be tuned by ear against references |

Each sound gets a small model built from what its circuit does: oscillators,
noise, filters and envelopes. Pure Go in `internal/synth`, no dependencies.

### 14.2 The sounds

| Sound | Trigger | Circuit behaviour to imitate |
|---|---|---|
| UFO | Port 3 bit 0, while held | A tone whose pitch is swept up and down by a slow oscillator, repeating (the SN76477's VCO driven by its SLF) |
| Shot | Port 3 bit 1, rising edge | A short noise-based "pew": filtered noise with a falling pitch/brightness |
| Player dies | Port 3 bit 2, rising edge | A longer noise explosion with a decaying envelope |
| Invader dies | Port 3 bit 3, rising edge | A short noise burst with a fast decay |
| Extended play | Port 3 bit 4, rising edge | A short tone (bonus life) |
| Fleet notes 1–4 | Port 5 bits 0–3, rising edge | Four low "thump" tones, falling in pitch, each with a quick decay; the game plays them in a loop, faster as invaders die |
| UFO hit | Port 5 bit 4, rising edge | A tone sweep for the score display |

All of this is ⚠ verify: take frequencies, time constants and filter
values from the schematics (component values give RC times and oscillator
rates) and MAME's audio code, then compare against reference recordings.

### 14.3 Rules

- **Edges, not levels.** A one-shot sound starts on the bit's 0→1 change
  and plays its envelope to the end; the game holding the bit doesn't restart
  it. The UFO tone plays while its bit is 1.
- **Amplifier enable** (port 3 bit 5): 0 mutes the output (the game turns it
  off in attract mode).
- **Sample-accurate timing.** The board records every port 3/5 write that
  changes a bit as `(cycle, port, value)` in a preallocated per-frame list;
  the synth starts each sound at the matching sample.
- **Deterministic:** the noise generator is a seeded LFSR (like the real noise
  source), so tests can compare rendered audio exactly.

### 14.4 Audio output

- Ebitengine's `audio` package, 48 kHz, a `io.Reader` stream fed by
  `app/sound` from the synth's output (float32, mono → stereo).
- One emulated frame renders 800 samples (48,000 / 60), matching `Update`'s
  rate (§8), so the buffer stays level. Small corrections (±1 sample) keep
  it from slowly filling or emptying.
- The audio context is created on the first user gesture (coin, start,
  click), as browsers require. Volume and mute are settings.

## 15. Testing

### 15.1 Unit tests (`internal/i8080`)

- **Every opcode has table-driven tests:** registers and flags before,
  expected registers, flags, memory and cycles after. Every bug in §3 has a
  test named after it.
- **Flags tested exhaustively** where cheap: `ADD`/`ADC`/`SUB`/`SBB`/`CMP`
  for all 256 × 256 × 2 inputs against a reference function written from the
  manual; `DAA` for all `A`, carry and AC combinations.
- **Disassembler** goldens for all 256 opcodes.

### 15.2 CPU exercisers (CI, no ROMs needed)

The classic 8080 test programs run as CP/M `.COM` files: loaded at `0x0100`,
printing through `CALL 0x0005` (BDOS function 2: character in `E`; function
9: `$`-terminated string at `DE`), ending with a jump to `0x0000`. The `cpm`
package provides exactly that.

| Program | What it checks | Licence | Use |
|---|---|---|---|
| `8080PRE.COM` (Ian Bartholomew, 2009) | Basic instructions, as a precondition | GPL v2 or later (Frank Cringle's `prelim.z80`; stated in the `.MAC` header) | Committed in `third_party/`, CI |
| `8080EXM.COM` (same) | Every instruction group, all flags, against CRCs from a real 8080 | GPL v2 or later (Frank Cringle's `zexlax.z80`; stated in the `.MAC` header) | Committed, CI (`-short` skips it); also the demo (§9.4) |
| `TST8080.COM` (Microcosm Associates, 1980) | Diagnostic | Unclear | Only if the licence is verified; otherwise downloaded by a dev tool, not committed |
| `CPUTEST.COM` (SuperSoft, 1981) | Diagnostic | Commercial | Not committed; optional local run |

The tests compare the full printed output with a golden text file. `8080EXM`
runs 23.8 billion cycles (about 45 minutes on a real 8080 at 2 MHz): the test
is part of the normal suite but skipped with `-short`. Natively it takes
about 27 s (Apple M-series, Go 1.27.4, about 870 MHz emulated).

### 15.3 Board tests and golden frames

- **Board tests (CI):** small 8080 programs, written as byte slices with
  comments, exercising the shift register, port bits, the memory map and
  mirrors, interrupt timing (count cycles between `RST 1` and `RST 2`),
  `EI`/`DI`/`HLT` and the sound-latch events.
- **Golden frames (local only):** the real ROM isn't available in CI. With a
  ROM set in `roms/invaders.zip` (git-ignored), or `INVADERS_ROMS=<path>`,
  the tests run the attract mode and a scripted game (coin, start, fire, move
  at fixed frames) and compare frames at fixed frame numbers with goldens in
  `internal/invaders/testdata/golden/`. Without ROMs they're skipped with a
  clear message. The goldens are 1-bit pictures of copyrighted screens, so
  they are stored as hashes plus a local `.png` that's never committed
  ⚠ decide: hashes only in git, or no goldens in git at all.
- **Never weaken a golden to make a test pass.** Fix the emulator, or explain
  in the PR why the golden was wrong.

### 15.4 Sound tests

The synth renders each sound for a fixed seed; tests check duration, peak
level, silence after the envelope, and that the amplifier bit mutes. Exact
waveforms are compared against committed goldens (our own synthesis, so no
rights issue), regenerated with `-update` and checked by ear in the PR.

### 15.5 Browser tests (Playwright)

As in Chippie: build, serve `web/dist/`, load the page in headless Chromium,
check the loader goes away, and that the CPU-test demo prints its first "OK"
line (no ROM needed). Allow for software rendering.

## 16. Browser integration and performance

**Browser-specific code** lives in `internal/platform/`, with desktop stubs,
as in Chippie: IndexedDB (the ROM image), `localStorage` (settings, view,
DIP switches), the file picker, URL parameters, `platform.InFront`, and
`platform.OpenURL` for the external link. Nothing there blocks `Update` or
`Draw`; async work runs in goroutines and reports on channels.

**Performance:** the 8080 at 2 MHz is about 0.5 million instructions per
second; a plain Go interpreter does far more, even in WASM. The cost is in
drawing: the CRT and overlay shaders at high pixel densities, as Chippie
found. Benchmarks: `BenchmarkFrame` (one emulated frame), `BenchmarkEXM`
(exerciser throughput), `BenchmarkVideo` (video RAM → pixels). PRs that touch
`i8080`, `invaders` or `crt` include `benchstat` output if they could affect
speed.

**Size:** expected a few MB gzip (Ebitengine + Go runtime; no UI toolkit, one
font). Measure it in M1 and record it here.

## 17. Build, CI and deployment

### 17.1 Local

```sh
go run ./cmd/invaders -roms ~/roms/invaders.zip   # desktop
go run ./cmd/cputest third_party/8080-exercisers/8080PRE.COM
go test ./...                                      # -short skips 8080EXM
golangci-lint run && GOOS=js GOARCH=wasm golangci-lint run
./scripts/build.sh && (cd web/dist && python3 -m http.server 8080)
```

`scripts/build.sh`, as in Chippie: build `invaders.wasm` with
`-trimpath -ldflags="-s -w"`, gzip it as `invaders-gz.wasm`, inline
`wasm_exec.js` into the loader page, copy the exercisers (with licence and
source) and the cabinet art into `web/dist/`.

### 17.2 GitHub Actions

Copied from Chippie and trimmed:

| Workflow | Trigger | Does |
|---|---|---|
| `ci.yml` | push, PR | lint (native + WASM), unit + exerciser tests, WASM build, native build, Playwright smoke test |
| `pages.yml` | push to `main` | build and deploy to GitHub Pages |
| `vulncheck.yml` | push, PR, weekly | `govulncheck` for both builds |
| `release.yml` | tag `v*` | desktop binaries (macOS, Windows, Linux) as release assets ⚠ decide; Ebitengine needs cgo on Linux and macOS, so build on each OS's runner |

Dependabot for Go modules, npm (Playwright) and Actions.

### 17.3 Codeberg (later)

- **Hosting:** Codeberg Pages (the git-pages based service since December
  2025; the old v2 service stops on 2026-12-31) can serve `web/dist/`.
- **CI:** Forgejo Actions runners on Codeberg are offered in limited form,
  with queues; Woodpecker CI needs an application. A self-hosted runner is the
  reliable option. The Playwright job is the heavy one.
- **Plan:** start on GitHub. Mirror to Codeberg when wanted; move only when
  its CI can run the full `ci.yml` in reasonable time.

## 18. References

- **Intel 8080 Microcomputer Systems User's Manual** (1975) and the **8080
  Assembly Language Programming Manual**: instruction behaviour, flags,
  cycle counts.
- **MAME source**, `src/mame/midw8080/mw8080bw.cpp` and `mw8080bw_a.cpp`:
  memory map, ports, video timing, interrupt lines, DIP switches, sound
  circuits, ROM CRCs. ⚠ verify the paths.
- **Computer Archeology, Space Invaders**: a fully commented disassembly
  and hardware notes.
- **Emulator 101** (emulator101.com): a step-by-step Space Invaders emulator
  tutorial; useful for the board, not authoritative for the 8080's flags.
- **Midway Space Invaders schematics and manual** (sound board component
  values for §14).
- **8080 exercisers:** Ian Bartholomew's `8080EXM`/`8080PRE` with sources;
  Frank Cringle's `zexall`.
- **Chippie** (github.com/rbscholtus/chippie): the loader, platform layer,
  shaders, CI and Pages setup this project reuses.
