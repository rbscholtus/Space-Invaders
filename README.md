# Space Invaders emulator

An emulator of the 1978 Space Invaders arcade machine (the Midway 8080
board), being rewritten in Go so that it runs **in your browser, with nothing
to install**, and on the desktop. It looks like an upright arcade cabinet
(our own art, in the style of Midway's), with the CRT's glow, the coloured
screen overlay and the moon backdrop; on small screens, simply the screen.

> **Status:** playable, with sound, in the browser and on the desktop, in its
> cabinet. Touch and gamepad controls are next, then settings. The plan is in
> [docs/DESIGN.md](docs/DESIGN.md) and the milestones in
> [docs/ROADMAP.md](docs/ROADMAP.md).

## Playing it in the browser

Open **<https://rbscholtus.github.io/Space-Invaders/>** in a current
Chrome, Firefox, Edge or Safari, and drop your `invaders.zip` on the page
(or press `O` to pick it). It's remembered for your next visit; `F` forgets
it. Nothing is uploaded: the ROMs stay in your browser. For now it's played
with the keyboard; touch controls come with the cabinet.

## Playing it on the desktop

You need [Go](https://go.dev/dl/) 1.27 and your own ROMs (below). On Linux,
Ebitengine also needs a few system packages (see its install guide).

```sh
go run ./cmd/invaders -roms path/to/invaders.zip
```

`-roms` also takes a folder with the four files. Without it, the emulator
uses `roms/invaders.zip` if it's there, or the set you loaded last time, or
asks for one: drop it on the window. `-lives 3`…`6` and `-bonus1000` set
the DIP switches; `-volume 0`…`100` the sound (default 50; volume and mute
are remembered). `-shader clean` draws the plain picture instead of the CRT
monitor (`?shader=clean` in the browser). `-view cabinet` or `-view screen`
shows the cabinet or the picture alone (`?view=` in the browser); by default
it depends on the window's size, and `V` switches.

## ROMs

The game's ROMs are copyrighted, so this project doesn't include or download
them. You'll need your own copy of the MAME `invaders` ROM set
(`invaders.zip`, or the files `invaders.h`, `.g`, `.f` and `.e`; renamed
files are recognised by their contents). Without ROMs, the screen says how
to load them and lists the controls.

Your HI-SCORE is remembered too, in the browser and on the desktop, until
you press `F`.

## Controls

| Key | |
|---|---|
| `5` or `C` | Insert coin |
| `1` / `2` | 1 or 2 player start |
| `←` `→` (or `A` `D`) | Move |
| `Space` (or `↑`, `W`) | Fire |
| `P` | Pause (also when the window isn't in front) |
| `M` | Mute |
| `-` / `=` | Volume down / up |
| `O` | Open your ROM files (browser, on the ROM screen) |
| `F` | Forget the remembered ROMs and HI-SCORE |
| `V` | The cabinet, or the picture alone (remembered) |

On touch screens, the cabinet's buttons and coin slot will be the controls.

## The Java version

The original emulator from 2010, written in Java, is in [`java/`](java/) and
tagged `java-2010`. It's kept for reference and no longer maintained.

## Building the browser version

```sh
./scripts/build.sh                          # builds web/dist/
cd web/dist && python3 -m http.server 8080  # then open http://localhost:8080
```

## Credits

- **The sound** is ported from [MAME](https://www.mamedev.org/)'s emulation
  of the Space Invaders sound board, by K. Wilkins, Derrick Renaud, Zsolt
  Vasvari and the MAME team, and checked against MAME's own output.
- **The hardware details** (memory map, ports, timing) and the screen's
  colour overlay come from MAME's Midway 8080 driver and artwork.
- **The CPU tests** are Frank D. Cringle's exercisers, converted to the 8080
  by Ian Bartholomew.

## Licence

BSD 3-Clause, see [LICENSE](LICENSE). The sound code ported from MAME is
also BSD 3-Clause, with MAME's notice in
[internal/synth/LICENSE-MAME](internal/synth/LICENSE-MAME). The 8080
exercisers in `third_party/` are GPL v2 or later; they are test programs the
emulator runs, not part of its code. The Java version in `java/` is Apache
2.0. Details: [DESIGN §19](docs/DESIGN.md#19-credits-and-third-party-licences).

Space Invaders is a trademark of Taito; this project is not affiliated
with Taito or Midway.
