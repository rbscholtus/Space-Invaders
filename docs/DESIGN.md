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
19. [Credits and third-party licences](#19-credits-and-third-party-licences)

---

## 1. Goals

- **Play Space Invaders in the browser with nothing to install.** Hosted on
  GitHub Pages; works on phones, tablets and desktops.
- **The same program on the desktop**, from one Go code base (Ebitengine).
- **Correct emulation.** The Intel 8080 passes the standard CPU exercisers
  (§15), and the board behaves as the Midway hardware does: shift register,
  ports, interrupts on the right scanlines, coin and DIP switches.
- **Sound without samples.** The analog sound circuits are emulated from
  their component values (a port of MAME's circuit models, §14), so no
  recordings need to be hosted.
- **No ROMs hosted.** Players bring their own ROM set; the site never serves
  or downloads it (§9). Without ROMs, the screen says how to load them and
  lists the controls.
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
| No-ROM mode | The ROM screen: what's needed, how to load it, and the controls (§9.4). The CPU exercisers stay test-only. |
| Sound | Ported from MAME's emulation of the sound board's circuits (`mw8080bw_a.cpp`, `disc_*.hxx`, `sn76477.cpp`; BSD-3-Clause, notice in `internal/synth/LICENSE-MAME`), tested against MAME's own output (§14). No samples. |
| Third-party work | Everything used from others, how, and under which licence, is listed in §19. |
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
| Coin bit clears when read, and is active high | The coin switch is active low (§7.2): hold the bit at 0 for a few frames, as a real coin switch does (§13). |
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
│  │  cpm/        minimal CP/M for the exercisers (tests)                 │  │
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
taken). Source: Intel 8080 Microcomputer Systems User's Manual (98-153B,
September 1975), p. 4-13: "The interrupt system is enabled following the
execution of the next instruction." Tested by `TestEIDelay`.

## 7. The Space Invaders board

Midway's 8080 board (1978), also used by other Midway/Taito games. CPU clock
19.968 MHz / 10 = **1.9968 MHz**.

### 7.1 Memory map

Source: MAME `src/mame/midw8080/mw8080bw.cpp`, `main_map` (§18).

| Address | Size | Contents |
|---|---|---|
| `0x0000–0x07FF` | 2 KB | ROM `invaders.h` |
| `0x0800–0x0FFF` | 2 KB | ROM `invaders.g` |
| `0x1000–0x17FF` | 2 KB | ROM `invaders.f` |
| `0x1800–0x1FFF` | 2 KB | ROM `invaders.e` |
| `0x2000–0x23FF` | 1 KB | Work RAM (stack at the top, near `0x2400`) |
| `0x2400–0x3FFF` | 7 KB | Video RAM, 1 bit per pixel (§10) |
| `0x4000–0x5FFF` | 8 KB | Unused ROM space: reads 0, writes ignored |
| `0x6000–0x7FFF` | 8 KB | Mirror of the RAM at `0x2000–0x3FFF` |
| `0x8000–0xFFFF` | | Mirror of `0x0000–0x7FFF` (address line 15 is not decoded) |

Writes to ROM are ignored. `Read`/`Write` mask the address with `0x7FFF`,
then map it as above, so all 8 KB of RAM (work + video) is mirrored as a
whole, not per 1 KB as in the Java version (§3).

### 7.2 Input ports

Source: MAME `mw8080bw.cpp`, `INPUT_PORTS_START( invaders )` and
`invaders_state::io_map`. Port numbers are decoded with 3 bits, so port
`n + 4` reads as port `n` (only ports 0–3 are read).

| Port | Bit | Meaning |
|---|---|---|
| 0 | 0 | DIP switch 8 (unused), 0 |
| 0 | 1–2 | DIP switches 6 and 7 (unused), 0 |
| 0 | 3 | Always 1 |
| 0 | 4 | P1 fire |
| 0 | 5 | P1 left |
| 0 | 6 | P1 right |
| 0 | 7 | DIP switch 5 (unused), 0 |
| 1 | 0 | Coin, **active low**: 1 at rest, 0 while a coin goes in |
| 1 | 1 | 2P start |
| 1 | 2 | 1P start |
| 1 | 3 | Always 1 |
| 1 | 4 | P1 fire |
| 1 | 5 | P1 left |
| 1 | 6 | P1 right |
| 1 | 7 | Unused, 0 |
| 2 | 0–1 | DIP: lives, 00 = 3, 01 = 4, 10 = 5, 11 = 6 |
| 2 | 2 | Tilt in the software, but not connected on the Midway board (MAME); reads 0 |
| 2 | 3 | DIP: bonus life at 1000 (1) or 1500 (0) |
| 2 | 4 | P2 fire (upright board: wired to P1's controls) |
| 2 | 5 | P2 left (same) |
| 2 | 6 | P2 right (same) |
| 2 | 7 | DIP: coin info shown in attract mode, 0 = on |
| 3 | 0–7 | Shift register result (§7.4) |

The game never reads port 0 (there is no `IN 0` in the ROM). The coin polarity
is confirmed by the game code: the VBLANK handler (`RST 2`, `0x0010`) reads
port 1 and adds a credit when bit 0 drops from 1 to 0.

### 7.3 Output ports

Source: MAME `invaders_state::io_map` and `invaders_audio_device::p1_w`,
`p2_w` (`mw8080bw_a.cpp`).

| Port | Bits | Meaning |
|---|---|---|
| 2 | 0–2 | Shift amount (§7.4) |
| 3 | 0 | UFO (repeats while 1) |
| 3 | 1 | Shot |
| 3 | 2 | Player dies ("flash") |
| 3 | 3 | Invader dies |
| 3 | 4 | Extended play (bonus life) |
| 3 | 5 | Amplifier enable |
| 3 | 6–7 | Not connected |
| 4 | 0–7 | Shift register data (§7.4) |
| 5 | 0–3 | Fleet movement, notes 1–4 |
| 5 | 4 | UFO hit |
| 5 | 5 | Flip screen; only wired on the cocktail board |
| 5 | 6–7 | Not connected |
| 6 | – | Watchdog reset: the board resets the CPU if 255 frames pass without one |

### 7.4 Shift register

The 8080 has no barrel shifter, so the board has one, used to draw sprites at
any x position.

- **Write port 4:** the 16-bit register shifts right by 8 and the new byte
  goes into the high byte: `reg = uint16(v)<<8 | reg>>8`.
- **Write port 2:** `amount = v & 7`.
- **Read port 3:** `byte(reg >> (8 - amount))`.

### 7.5 DIP switches and settings

The board's DIP switches (lives, bonus life, coin info) are settings in the
app, stored in preferences, and applied on the next power-on. Defaults, as
in MAME: 3 lives, bonus at 1500, coin info on.

### 7.6 What the game does (from the ROM and from running it)

Facts the tests rely on, found by disassembling the player's ROM and by
running it; addresses are work RAM.

- **Reset:** the 8080 starts at `0x0000`, which jumps to `0x18D4`: stack
  at `0x2400`, RAM set-up, then the attract mode at `0x0AEA`. There is no
  self-test or RAM/ROM check: the Midway board's diagnostics were separate
  test ROMs.
- **Ports read:** 1, 2 and 3 only; never 0.
- **Interrupt handlers:** `RST 1` (`0x0008`) and `RST 2` (`0x0010`). The
  coin switch is read in the `RST 2` handler (§7.2).

| Address | Meaning |
|---|---|
| `0x20EB` | Credits, BCD |
| `0x20EF` | Game mode: 1 while a game runs |
| `0x21FF` | Player 1's ships in reserve |
| `0x20F8`, `0x20F9` | Player 1's score, BCD, low byte first |
| `0x20FC`, `0x20FD` | Player 2's score |

**Timing, measured** (frames at 60 per second, from the start button):
the amplifier turns on (port 3 bit 5) after about 200 frames (3.3 s), when
"PLAY PLAYER 1" ends; the cannon can fire after 300–350 frames (5–6 s).
Tests that drive the game must wait at least that long; guessing shorter
times made the first smoke test fail.

## 8. Timing

**The video frame** (MAME `mw8080bw.h`, `MW8080BW_*` constants, and the
timing notes at the top of `mw8080bw.cpp`):

- Pixel clock 19.968 MHz / 4 = 4.992 MHz; 320 pixels per line; 262 lines.
- Frame rate 4,992,000 / (320 × 262) = **59.54 Hz**.
- CPU cycles per line: 1,996,800 / (59.54 × 262) = **128**; per frame
  **33,536**. (One CPU cycle is 2.5 pixels.)
- 224 visible lines (0–223); vertical blank from line 224.

**Interrupts:** two per frame.

- `RST 1` (`0xCF`) at **line 96**, in the middle of the visible picture. The
  game then updates the objects in the top half, which the beam has already
  drawn.
- `RST 2` (`0xD7`) at **line 224**, the start of vertical blank. The game
  updates the bottom half.

Source: MAME `mw8080bw.cpp`, "Interrupts" and `interrupt_trigger`: the
vertical counter values `0x80` and `0xDA` correspond to lines 96 and 224,
and the vector is `0xC7 | (64V << 4) | (!64V << 3)`.

The INT line is a flip-flop that stays set until the CPU accepts the
interrupt. MAME raises it only when interrupts are enabled at that moment
(otherwise the request is lost), and keeps it raised until accepted, so a
request that arrives during the instruction after `EI` is taken one
instruction later. The board does the same: when its line comes, it raises
the request if `IntEnabled`, and offers it before every instruction until the
CPU accepts it.

**In the emulator:**

- `board.RunFrame()` runs the frame line by line: 262 × 128 cycles, carrying
  each instruction's overshoot into the next line, and raises each interrupt
  at its line. A frame starts at line 0, the top of the picture. This replaces the Java version's two half-frames.
- **One emulated frame per `Update`.** Ebitengine runs `Update` at 60 Hz, the
  game at 59.54 Hz, so the game runs 0.8% fast. Nobody can hear or see 0.8%,
  and it avoids the judder of skipping one `Update` in 130. Audio follows the
  emulated frame: each frame renders 1/60 s of samples (§14.4), so the audio
  buffer never drifts.
- **Pausing:** when the tab is hidden (`platform.InFront`, as in Chippie), when
  a menu is open, or with a pause key. Natively, when the window loses focus.
- **Speed is not adjustable** in normal play (the game's speed is its timing).

## 9. ROMs

### 9.1 What's needed

The MAME set **`invaders`** (Space Invaders, Midway, 1978): four 2 KB files.

| File | MAME name | Address | CRC-32 | SHA-1 |
|---|---|---|---|---|
| `invaders.h` | `9316b-0869_m739h.h1` | `0x0000` | `734f5ad8` | `ff6200af4c9110d8181249cbcef1a8a40fa40b7f` |
| `invaders.g` | `9316b-0856_m739g.g1` | `0x0800` | `6bfaca4a` | `16f48649b531bdef8c2d1446c429b5f414524350` |
| `invaders.f` | `9316b-0855_m739f.f1` | `0x1000` | `0ccead96` | `537aef03468f63c5b9e11dd61e253f7ae17d9743` |
| `invaders.e` | `9316b-0854_m739e.e1` | `0x1800` | `14e538b0` | `1d6ca0c99f9df71e2990b610deb9d7da0125e2d8` |

Source: MAME `mw8080bw.cpp`, `ROM_START( invaders )`. Older MAME sets name
the files `invaders.h` ... `invaders.e`; current MAME uses the chip names.
`romset` matches by CRC-32, so both work. The table becomes
`internal/romset/sets.go`. Other versions (Taito's `sitv`, `sicv`, ...) may
be added later, each with its own table.

### 9.2 Bringing your own

The ROMs are copyrighted by Taito. They are never committed, hosted, or
downloaded by the app.

- **Browser:** drop a `.zip` or the four files on the page, or press `O` to
  pick them with the file picker. `romset` reads the zip (`archive/zip`, in
  memory), matches files by CRC-32 (so renamed files still work), checks all
  four are present, and builds the 8 KB image. The image is stored in
  IndexedDB, so the next visit starts at once. `F` forgets it.
- **Desktop:** the same drop, and a `-roms <path>` flag (zip or folder),
  now optional. The image is cached in the user cache folder.
- **The flow** (`internal/app/romscreen`, plain Go with tests): at start, the
  cached image is loaded in a goroutine (never in `Update`) and re-checked
  (`romset.CheckImage`); while there is none, the ROM screen shows, with the
  controls below it (§9.4); a good set starts the game and is cached.
  Dropped files come from `ebiten.DroppedFiles()`, read in a goroutine (in
  the browser a read can wait on JavaScript) with limits (64 files, 16 MB
  each, 32 MB in all); picked files arrive as `platform.Picked`. Both go
  through `romset.Files`, an in-memory `fs.FS`, and a zip among them is
  opened one level deep, so a dropped folder holding `invaders.zip` works.
- **Errors** say what's wrong, then `romset`'s detail: "These aren't the
  invaders ROMs. Found: holiday.jpg", "Not all four ROMs are there. …"
  (which file is missing), "These ROMs differ from the MAME invaders set. …"
  (which CRC differs), wrapped to the monitor's 28 columns.

### 9.3 The external link

**Decided (M4): no link.** The ROM screen and the README explain that the
MAME `invaders` set is needed and how to load it. If a link is ever added,
the player opens it themselves, the app never fetches from it, and it must
not be a direct download of ROM files.

### 9.4 Without ROMs: the controls

Below the ROM screen's text, at the bottom of the monitor, the screen lists
the controls a new player needs (`Flow.Controls`: coin, start, move,
fire, pause, sound, and open (where the picker is offered) and forget the
ROMs), the keys in the overlay's green band. In the
browser there is nowhere else to show them until the cabinet (M5).

**Decided (M4): no CPU-test demo.** M4 first ran the 8080 exerciser on the
ROM screen, to show the emulator works without ROMs. It was taken out before
release: it meant nothing to players (instruction-group names for 9.5
minutes), and it shipped GPL test programs with the site. The exercisers are
test data only (§15.2), and the tests also run compiled to WebAssembly
(`scripts/test-wasm.sh`), which checks the CPU as the browser runs it more
directly than the demo did.

## 10. Video

- **Resolution:** 256 × 224, 1 bit per pixel, 32 bytes per row, 224 rows.
- **Rotated monitor:** the monitor is mounted turned 90° anticlockwise (MAME
  `ROT270`), so the
  player sees **224 wide × 256 tall**. Byte `i` of video RAM (`0x2400 + i`)
  holds pixels of memory row `i / 32`, x = `(i % 32) * 8 + bit`, with bit 0 the
  leftmost in memory. On screen: column `x_screen = i / 32`, row
  `y_screen = 255 - ((i % 32) * 8 + bit)`.
- **A 4-pixel offset:** the video hardware reads each byte 4 pixels late,
  so MAME's picture is 260 pixels wide, with the 256 pixels of video RAM
  after 4 black ones (`screen_update_invaders`, `mw8080bw_v.cpp`). The
  emulator shows the 256 pixels; the offset is only a border.
- **Build the picture once per frame from video RAM,** not on every write.
  It's 7 KB, and converting per frame keeps the board free of display code
  and makes golden frames trivial (§15.3). `invaders.Board.Video()` returns a
  read-only view; `app` turns it into an `ebiten.Image` with `WritePixels` on
  a preallocated buffer.
- **Text** (the ROM screen) uses our own 8×8 pixel font
  (`internal/app/font`, ASCII 32–126, original glyphs in a 1978 terminal
  style): 28 columns by 32 rows on the 224 × 256 monitor, drawn into the same
  1-byte-per-pixel picture as the game, so the overlay colours it too. No
  font file, nothing from the game's own font.
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
lives area at the bottom left. From MAME's `layout/invaders.lay` (CC0), in
screen pixels:

| Strip | Columns | Rows | Lit pixels multiplied by (R, G, B) |
|---|---|---|---|
| Red | 0–223 | 32–61 | 1, 0.125, 0.125 |
| Green | 0–223 | 184–239 | 0.125, 1, 0.125 |
| Green (lives) | 16–133 | 240–255 | 0.125, 1, 0.125 |

The strips are tints, not pure colours: they let a little of the other
colours through. The Java version had the red strip at rows 32–63. Until
the shaders (M5), `internal/app/overlay` paints the picture on the CPU.

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
| `T` | Tilt (not wired on the Midway board, §7.2; for testing) |
| `P` | Pause |
| `M` | Mute |
| `-` / `=` | Volume down / up (10% steps; `-volume 0–100` at start, default 50) |
| `O` | Open the ROM files (browser file picker), on the ROM screen |
| `F` | Forget the remembered ROMs and go back to the ROM screen |
| `V` | Switch view |
| `F1` | Stats overlay |

**Coins:**

- A coin is a pulse: each new press of the coin key closes the coin switch
  (port 1 bit 0 reads 0, §7.2) for 5 frames (`input.CoinFrames`); the game
  adds the credit on the first. Holding the key inserts one coin
  (`TestROMCoinNeedsAnEdge`). Holding
  the key inserts one coin; each new press inserts another.
- In cabinet view, the coin slot on the coin door is clickable/tappable, with
  a short coin-drop animation and sound.
- In screen view, a coin button sits with the touch controls.

**Two players:** on an upright cabinet both players use the same controls,
taking turns: the upright board wires port 2's player 2 bits to player 1's
controls (MAME `invaders_in2_control_r`). The emulator does the same.

**Touch:** in screen view, on-screen left/right/fire, coin, 1P and 2P.
In cabinet view, the cabinet's own buttons are the controls. Multi-touch,
so moving and firing work together (as in Chippie's remote).

**Gamepads:** Ebitengine's standard gamepad layout: d-pad/stick to move, A to
fire, Select = coin, Start = 1P start.

**Tilt:** the game software handles a tilt bit (port 2 bit 2), but the
Midway board doesn't connect it (MAME). The `T` key sets it anyway, as a
debugging aid; the real cabinet had no tilt.

## 14. Sound

### 14.1 Approach: port MAME's circuit models

The sounds are made by analog circuits on the sound board: six discrete
circuits (op-amps, 555 timers, RC filters, a noise generator) and an SN76477
sound chip for the UFO (MAME `mw8080bw_a.cpp`, `invaders_audio_device::p1_w`).
There are no recordings in the machine; the circuits *are* the sounds.

| Approach | Ease | Faithful | Testable | Performance | Hosting |
|---|---|---|---|---|---|
| Samples (the Java version, most hobby emulators, MAME before 2007) | Easiest | One board's recording, but fixed: loops and retriggers audibly, can't follow the circuit | Hard | Trivial | **Can't host the recordings**; players would need a second file set |
| Hand-tuned synthesis | Medium | As good as the tuning | Nothing objective to compare with | Trivial | Ours |
| **Port of MAME's circuit models** (chosen) | Medium–hard: ~2,000 lines | High: the schematic's component values and MAME's models | Against MAME's output, nearly sample for sample | Small (§16) | Ours (BSD-3, credited) |
| Circuit simulation (SPICE-like, MAME's netlist, MiSTer's analog work) | Hard | Highest in theory | Fragile numerics | Heavy, risky in WASM | Ours |

**Why:** MAME has modelled this board's circuits since 2007 (D. Renaud, from
the Midway schematics M051-00739-A005/B005; the noise clock was breadboarded
and measured), and its code is BSD-3-Clause like ours
(`src/mame/midw8080/mw8080bw_a.cpp`, `src/devices/sound/disc_*.hxx`,
`disc_dev.hxx`, `sn76477.cpp`). Porting it gives faithful sound without
sample files, and MAME itself becomes a repeatable reference (§14.5). The
ported files keep MAME's copyright lines.

**How others do it:** most hobby emulators and tutorials play recorded WAV
samples, usually supplied by the player for copyright reasons; MAME did too
until 2007, and old forks (MAME 2003) still do. MiSTer's FPGA core has
simplified built-in sound and a "samples" variant. Samples are the easy
route when the CPU and video are the focus.

### 14.2 The circuits

From MAME's `mw8080bw_a.cpp` (invaders section, Midway board values):

| Sound | Trigger | Circuit (MAME blocks) |
|---|---|---|
| Noise | always running | 17-bit LFSR (bit 4 XOR bit 16) clocked at 7515 Hz (`LFSR_NOISE`, `midway_lfsr`), used by the shot and the player's explosion |
| UFO | port 3 bit 0, while 1 | SN76477: VCO swept by its SLF (super-low-frequency oscillator), VCO mode 1, enabled while the bit is 1 |
| Shot | port 3 bit 1 | Norton op-amp one-shot → integrator → op-amp VCO (falling pitch) mixed with filtered noise → triggered VCA |
| Player dies | port 3 bit 2 | One-shot → triggered VCA on the noise → two RC low-pass filters |
| Invader dies | port 3 bit 3 | One-shot → integrator → op-amp oscillator → VCO → op-amp |
| Extended play | port 3 bit 4 | 555 astable (about 7 Hz, reset by the bit) AND a 480 Hz square wave |
| Fleet notes 1–4 | port 5 bits 0–3 | The note bits pick resistors (`COMP_ADDER`) for a 555 astable → two RC filters: the "thump" |
| UFO hit | port 5 bit 4 | Integrator → op-amp oscillator → VCO → op-amp |
| Mix | — | Op-amp mixer of the six discrete circuits; MAME's output gain 2500, then discrete and SN76477 at 0.5 each |
| Mute | port 3 bit 5 = 0 | All sound off (MAME `system_mute`) |

### 14.3 Rules

- **How long a sound plays comes from its circuit**, not from rules in our
  code. The one-shots (shot, invader dies, player dies) start on a bit's
  rising edge and end after their capacitor's time or when the bit falls,
  whichever is first: their trigger is capacitor-coupled, so a falling bit
  pulls it low. So the shot sounds while the game holds its bit, a held
  "invader dies" lasts its 40 ms pulse plus the decay, and a held "player
  dies" lasts about 0.85 s plus the decay. The UFO plays while its bit is 1,
  extended play while bit 4 is 1, each fleet note while its bit is 1 (the
  game holds them 4 frames).
- **Amplifier enable** (port 3 bit 5): 0 mutes the output (the game turns it
  off in attract mode). The synth starts muted (port 3 = 0), so there is no
  power-on thump; MAME, which starts unmuted, plays one while its circuits
  charge.
- **Sample-accurate timing.** The board records every port 3/5 write that
  changes a bit as `(cycle, port, value)` in a preallocated per-frame list
  (64 entries; a frame with more keeps the latest value in the last one).
  `app/sound` turns the cycle into a sample offset, `cycle × 800 / 33,536`,
  and the synth applies the write at that sample (±1 sample, 21 µs). The
  cycle is when the `OUT` instruction begins, at most 10 cycles (5 µs) before
  the write.
- **Real time, not game time.** The circuits run at the audio sample rate,
  so pitches are right even though the game runs 0.8% fast (§8); only the
  fleet's march is 0.8% quicker.
- **Deterministic:** the noise is an LFSR, as on the board, so renders can be
  compared exactly.

### 14.4 Audio output

- Ebitengine's `audio` package, 48 kHz, an `io.Reader` stream fed by
  `app/sound` from the synth's output (float32, mono → stereo).
- One emulated frame renders 800 samples (48,000 / 60), matching `Update`'s
  rate (§8). A small buffer between the game loop and the audio player is
  kept at about 3 frames (50 ms): when it drifts, one sample is dropped or
  repeated. `Update` never blocks.
- Volume and mute: `-volume` or `?volume=`, and the keys (§13). Both are
  remembered (`platform.Prefs`); a volume given at start wins over the
  remembered one. In the browser, sound starts on the first key press,
  click, touch or drop (browsers allow audio only after a user gesture); on
  the desktop at once.

### 14.5 Testing against references

1. **The schematic's maths:** each building block is tested against its
   textbook behaviour: 555 frequency `1.44 / ((R1 + 2·R2)·C)`, RC step
   response, LFSR sequence, oscillator frequencies, one-shot lengths.
2. **MAME's output, sound by sound:** `scripts/mame-reference.sh` runs MAME
   with a Lua script that writes the sound ports during attract mode, when
   the game writes none (checked: the golden session has no sound events
   before the first game), and records each sound to a WAV file with
   `-wavwrite`. The WAVs stay local (git-ignored): they are renderings of
   Taito's sound design. `cmd/soundref` measures them (loudness envelope,
   pitch and spectral centroid per 10 ms, duration) and the measurements
   are committed; CI compares our synth with them. Locally, with the WAVs
   present, the waveforms themselves are compared (correlation).
3. **MAME's output for the whole golden session** (local): the same script
   in Lua, compared with our session by per-window features.
4. **Real cabinets** (recordings, videos): for listening only. Components
   vary 5–20% between boards, and speakers and rooms colour the sound.
5. **Our own renders:** `go test ./internal/synth -run TestListen -listen`
   writes them as WAVs to `testdata/listen/` (git-ignored), to compare by
   ear with MAME's. No exact hashes: floating point may differ in its last
   bits between arm64 and amd64 (FMA, assembly math), and the oscillators
   amplify that.

**Results (M3):** every sound matches MAME's recording: waveform
correlation 0.9989–1.0000, all measurements within 0.7% (the fleet's
brightness 1.7%). Over the whole golden session, in the 112 s where both
make sound: envelope correlation 0.9998, waveform 0.9996, loudness 0.9998.
`Render` takes 70 µs per frame with every sound playing (§16).

**What we learned about MAME's timing** (measured, MAME 0.289):

- The Lua frame callback for frame f runs at the start of that frame's
  vertical blank, (f + 1) frame periods after power-on: MAME's screen starts
  at line 224 at power-on, ours at line 0, so MAME's interrupts (and the
  game's sound writes) come 38 lines later than ours relative to power-on.
- MAME latches inputs once per frame, at the start of vertical blank: values
  set in callback N take effect at line 224 of frame N + 1.
  `Board.SetInputsAt(224, …)` reproduces it, and the game then runs exactly
  as in MAME: work and video RAM match at every frame of the 24,000-frame
  session.
- MAME applies the amplifier mute (`system_mute`) to whole buffers of audio
  as its sound thread processes them, about 0.8 s late here; ours mutes at
  the sample, as the board does. The session check compares only windows
  where both make sound.
- MAME's discrete noise is a normal LFSR: its feedback helper masks both
  operands (a first port that missed this produced lopsided noise).

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
| `8080EXM.COM` (same) | Every instruction group, all flags, against CRCs from a real 8080 | GPL v2 or later (Frank Cringle's `zexlax.z80`; stated in the `.MAC` header) | Committed, CI (`-short` skips it) |
| `TST8080.COM` (Microcosm Associates, 1980) | Diagnostic | Unclear | Only if the licence is verified; otherwise downloaded by a dev tool, not committed |
| `CPUTEST.COM` (SuperSoft, 1981) | Diagnostic | Commercial | Not committed; optional local run |

The tests compare the full printed output with a golden text file. `8080EXM`
runs 23.8 billion cycles (about 45 minutes on a real 8080 at 2 MHz): the test
is part of the normal suite but skipped with `-short`. Natively it takes
about 27 s (Apple M3, Go 1.27.4, about 870 MHz emulated). With `-race` it
takes many minutes, so CI runs the race detector with `-short` and the
exercisers in a separate step without it.

**As WebAssembly:** `scripts/test-wasm.sh` compiles every package without
Ebitengine or browser storage to WASM and runs its tests under Node
(`go_js_wasm_exec`), 8080EXM included: the CPU, board, sound and ROM code as
the browser runs it. 8080EXM takes about 165 s that way on the M3; CI runs
it as its own job, next to the native tests. The exercisers are test data
only: never part of the app or the site (§9.4).

### 15.3 Board tests and golden frames

- **Board tests (CI):** small 8080 programs, written as byte slices with
  comments, exercising the shift register, port bits, the memory map and
  mirrors, interrupt timing (count cycles between `RST 1` and `RST 2`),
  `EI`/`DI`/`HLT` and the sound-latch events.
- **Golden frames (local only):** the real ROM isn't available in CI. With a
  ROM set in `roms/invaders.zip` (git-ignored), or `INVADERS_ROMS=<path>`,
  `TestGoldenSession` plays a scripted session of 24,000 frames (6 min 40 s
  of game time, under a second to run): 90 s of attract mode, three coins, a
  1-player game until all ships are lost, a 2-player game until game over,
  and attract mode again. Every 30 frames it compares a SHA-256 of video RAM
  and a SHA-256 of the sound events since the last checkpoint with
  `internal/invaders/testdata/golden/session.txt` (800 checkpoints). The test
  also checks that the session still reaches every phase (2 games, 2 game
  overs, ships lost, both players scoring), so the hashes can't silently
  cover less. Without ROMs it's skipped with a clear message.
- **Only hashes are committed:** the pictures are Taito's. `-update`
  rewrites the hashes and writes every checkpoint's picture to
  `testdata/golden/png/` (git-ignored) to check by eye; on a mismatch the
  test writes the differing pictures to `testdata/failures/` (git-ignored).
- **Never weaken a golden to make a test pass.** Fix the emulator, or explain
  in the PR why the golden was wrong.

### 15.4 Sound tests

See §14.5: block tests against the schematic's maths (`blocks_test.go`),
`TestAgainstMAME` against the committed measurements of MAME's output (and
the waveforms, locally, when the recordings exist), `TestSessionAgainstMAME`
(local: the golden session against MAME's), the amplifier mute, determinism,
and `Render` not allocating. `internal/app/sound` tests where each event
lands in the frame and the stream's buffering; `TestROMSound` (local) plays
the real game through the sound path.

### 15.5 Browser tests (Playwright)

`e2e/`, adapted from Chippie: `scripts/build.sh`, then Playwright serves
`web/dist/` and loads the page in headless Chromium with software WebGL (CI
has no GPU). The tests read the state the app publishes
(`window.invadersState`: the mode and the ROM screen's message;
`platform.PublishState`) rather than comparing pixels, and poll it instead
of waiting fixed times:

- the loader goes away and the app asks for ROMs;
- picking a file that isn't the ROMs (`O`, the file chooser) shows the
  explanation;
- with your own ROMs (`roms/invaders.zip` or `INVADERS_ROMS`; local only,
  skipped without them): picking them starts the game, a reload starts it
  again from IndexedDB, and `F` forgets them.

## 16. Browser integration and performance

**Browser-specific code** lives in `internal/platform/`, adapted from
Chippie, each `_js.go` with an `_other.go` for the desktop: `Cache`
(IndexedDB; the user cache folder), `Prefs` (`localStorage`; a JSON file:
volume, mute, later the view and DIP switches), `Params` (URL query; flags),
`PickFiles`, `InFront` (page visibility; iOS Safari can misreport focus)
and `PublishState` (the app's state as `window.invadersState`, for the
browser tests). The app downloads nothing, so there is no `fetch()` and no
`net/http`. Nothing there blocks
`Update` or `Draw`; async work runs in goroutines and reports on channels.

**Audio in the browser** starts on the first key press, click, touch or
drop, as browsers require a user gesture (§14.4); on the desktop at start.

**Performance:** the 8080 at 2 MHz is about 0.5 million instructions per
second; a plain Go interpreter does far more, even in WASM. Measured
natively (Apple M3, Go 1.27.4, M2):

| Benchmark | Time | Notes |
|---|---|---|
| `BenchmarkStep` | 4.9 ns per instruction | |
| `BenchmarkEXM` | 864 MHz emulated | about 430 × real time |
| `BenchmarkFrame` | 33 µs per frame | 0.2% of a 16.7 ms frame |
| `BenchmarkScreen` | 31 µs | video RAM → rotated picture |
| `BenchmarkPaint` | 50 µs | overlay colours on the CPU, until M5's shader |
| `BenchmarkRender` (M3) | 70 µs per frame | all seven sounds playing (800 samples) |

The whole per-frame path is about 0.19 ms natively with sound; WASM is
typically 2–5 × slower, which still leaves the frame to drawing. Measured
in M4 on the same M3: 8080EXM's 23.8 G cycles took 134 s in headless
Chromium (with the CPU-test demo, since removed, §9.4) while still drawing
every frame, and take 165 s under Node (`scripts/test-wasm.sh`): 140–180 M
cycles per second, 70–90 × real time, about 5 × slower than natively. The
cost is in
drawing: the CRT and overlay shaders at high pixel densities, as Chippie
found. Benchmarks: `BenchmarkFrame` (one emulated frame), `BenchmarkEXM`
(exerciser throughput), `BenchmarkVideo` (video RAM → pixels). PRs that touch
`i8080`, `invaders` or `crt` include `benchstat` output if they could affect
speed.

**Size** (M4, Go 1.27.4, Ebitengine 2.10.6, `-trimpath -ldflags="-s -w"`):
`invaders-gz.wasm` is 3,429 KB (14,540 KB unpacked). Nearly all of it is
the Go runtime and Ebitengine; there is no `net/http` (CI checks) and no font
file. CI fails the build above 4 MB gzip and writes the size to the job
summary. Served from `localhost`, the page downloads, unpacks and compiles
it in well under a second; over the network the 3.4 MB download dominates.
The loader shows its progress.

## 17. Build, CI and deployment

### 17.1 Local

```sh
go run ./cmd/invaders -roms ~/roms/invaders.zip   # desktop
go run ./cmd/cputest third_party/8080-exercisers/8080PRE.COM
go test ./...                                      # -short skips 8080EXM
golangci-lint run && GOOS=js GOARCH=wasm golangci-lint run
./scripts/build.sh && (cd web/dist && python3 -m http.server 8080)
./scripts/build.sh && (cd e2e && npm ci && npx playwright test)
./scripts/test-wasm.sh                             # the tests as WASM (Node)
```

Every test command (ROM tests, the sound against MAME, listening,
regenerating goldens and references) is listed in CONTRIBUTING, "Running
tests".

`scripts/build.sh`, as in Chippie, fills `web/dist/`: it builds
`invaders.wasm` with `-trimpath -ldflags="-s -w"`, gzips it as
`invaders-gz.wasm`, inlines `wasm_exec.js` into the loader page
(`web/page.html` → `index.html`), and copies the licence notices
(`LICENSE`, `LICENSE-MAME`, and a short `NOTICE.txt`, §19). The cabinet art joins them
in M5.

### 17.2 GitHub Actions

Copied from Chippie and trimmed:

| Workflow | Trigger | Does |
|---|---|---|
| `ci.yml` | push, PR | lint (native + WASM), unit + exerciser tests natively and as WASM under Node (§15.2), native and WASM builds, no `net/http` in WASM, `scripts/build.sh` with the 4 MB size budget (§16); then a browser-test job runs Playwright on the built site (§15.5) |
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

- **Intel 8080 Microcomputer Systems User's Manual** (98-153B, September
  1975), chapter 4 (instruction set, with states per instruction), and the
  **8080 Assembly Language Programming Manual**: instruction behaviour,
  flags, cycle counts. Keep local copies in `docs/` (git-ignored).
- **MAME source**, `src/mame/midw8080/`: `mw8080bw.cpp` and `mw8080bw.h`
  (memory map, ports, video timing, interrupt lines, DIP switches, ROM CRCs),
  `mw8080bw_v.cpp` (video), `mw8080bw_a.cpp` (sound circuits). Checked at
  commit `11f90a3`, the latest to touch `mw8080bw.cpp` at the time.
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

## 19. Credits and third-party licences

Our own code is BSD-3-Clause (`LICENSE`). What we use from others:

| What | From | Licence | How it's used | Notice |
|---|---|---|---|---|
| Sound emulation (`internal/synth`) | MAME: the Space Invaders sound board (`mw8080bw_a.cpp`) and the discrete-circuit and SN76477 emulation (`disc_*.hxx`, `discrete.h`, `sn76477.cpp`, `rescap.h`); K. Wilkins, D. Renaud, Z. Vasvari and others | BSD-3-Clause | Ported to Go: source code derived from theirs | `internal/synth/LICENSE-MAME`, and a "Ported from MAME's …" comment in each file |
| 8080 exercisers (`third_party/8080-exercisers`) | Frank D. Cringle (1994), converted to the 8080 by Ian Bartholomew (2009) | GPL v2 or later | Test data: run by the tests, never linked into our code or shipped with the app or the site | `third_party/8080-exercisers/README.md`, `COPYING` |
| Colour overlay (`internal/app/overlay`) | MAME's `layout/invaders.lay` | CC0 | The strip positions and colours | Cited in the code |
| Hardware facts (memory map, ports, timing, interrupts, DIP switches, ROM CRCs) | MAME's `mw8080bw` driver; the Intel 8080 manual | Facts, not code | Checked and cited per fact (§7–§10) | Cited in DESIGN and the code |
| Reference recordings and measurements (`internal/synth/testdata/reference`) | Made locally with MAME from the player's own ROMs | Our measurements (numbers only) | Tests (§14.5); the recordings themselves are never committed | — |
| Loader, platform layer, CI, browser tests | Chippie (our own project) | BSD-3-Clause | Copied and adapted | — |
| Browser tests (`e2e/`) | Playwright (Microsoft) | Apache-2.0 | A development tool, never shipped | — |

**The pixel font** (`internal/app/font`) is our own drawing: nothing from
the game's font or another font.

**Binary releases** (the WASM site, desktop binaries) must carry the BSD-3
notices too (clause 2): `scripts/build.sh` puts `LICENSE`, `LICENSE-MAME`
and a `NOTICE.txt` (what each covers, the trademark)
in the site (M4); `release.yml` will do the same for desktop binaries, and
an "About" screen lists them (M5).

**The game itself** (its program, graphics and sound design, the name
"Space Invaders") is Taito's. We ship none of it: players bring their own
ROMs (§9), the art is our own (§12), and the sound is emulated, not
recorded.
