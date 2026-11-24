# Space Invaders emulator

An emulator of the 1978 Space Invaders arcade machine (the Midway 8080
board), being rewritten in Go so that it runs **in your browser, with nothing
to install**, and on the desktop. It will look like an upright arcade cabinet,
with the coloured screen overlay and the moon backdrop, and on phones simply
the screen.

> **Status:** being ported. The plan is in [docs/DESIGN.md](docs/DESIGN.md)
> and the milestones in [docs/ROADMAP.md](docs/ROADMAP.md). The playable
> version will be published on GitHub Pages.

## ROMs

The game's ROMs are copyrighted, so this project doesn't include or download
them. You'll need your own copy of the MAME `invaders` ROM set
(`invaders.zip`, or the files `invaders.h`, `.g`, `.f` and `.e`). Drop it on
the page and it's remembered for your next visit. Without ROMs, the emulator
runs a CPU self-test instead.

## Controls (planned)

| Key | |
|---|---|
| `5` or `C` | Insert coin |
| `1` / `2` | 1 or 2 player start |
| `←` `→` | Move |
| `Space` | Fire |

On touch screens, the cabinet's buttons and coin slot are the controls.

## The Java version

The original emulator from 2010, written in Java, is in [`java/`](java/) and
tagged `java-2010`. It's kept for reference and no longer maintained.

## Licence

BSD 3-Clause, see [LICENSE](LICENSE). The Java version in `java/` is Apache
2.0. Space Invaders is a trademark of Taito; this project is not affiliated
with Taito or Midway.
