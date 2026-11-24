# Space Invaders emulator

An emulator of the 1978 Space Invaders arcade machine (the Midway 8080
board), being rewritten in Go so that it runs **in your browser, with nothing
to install**, and on the desktop. It will look like an upright arcade cabinet,
with the coloured screen overlay and the moon backdrop, and on phones simply
the screen.

> **Status:** playable on the desktop, without sound yet. The browser
> version, sound and the cabinet are next. The plan is in
> [docs/DESIGN.md](docs/DESIGN.md) and the milestones in
> [docs/ROADMAP.md](docs/ROADMAP.md). The browser version will be published
> on GitHub Pages.

## Playing it on the desktop

You need [Go](https://go.dev/dl/) 1.27 and your own ROMs (below). On Linux,
Ebitengine also needs a few system packages (see its install guide).

```sh
go run ./cmd/invaders -roms path/to/invaders.zip
```

`-roms` also takes a folder with the four files; by default it looks for
`roms/invaders.zip`. `-lives 3`…`6` and `-bonus1000` set the DIP switches.

## ROMs

The game's ROMs are copyrighted, so this project doesn't include or download
them. You'll need your own copy of the MAME `invaders` ROM set
(`invaders.zip`, or the files `invaders.h`, `.g`, `.f` and `.e`; renamed
files are recognised by their contents). In the browser version, you'll
drop it on the page and it's remembered for your next visit; without ROMs,
the emulator will run a CPU self-test instead.

## Controls

| Key | |
|---|---|
| `5` or `C` | Insert coin |
| `1` / `2` | 1 or 2 player start |
| `←` `→` (or `A` `D`) | Move |
| `Space` (or `↑`, `W`) | Fire |
| `P` | Pause (also when the window isn't in front) |

On touch screens, the cabinet's buttons and coin slot will be the controls.

## The Java version

The original emulator from 2010, written in Java, is in [`java/`](java/) and
tagged `java-2010`. It's kept for reference and no longer maintained.

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
