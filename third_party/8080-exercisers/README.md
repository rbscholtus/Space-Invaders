# 8080 CPU exercisers

The Intel 8080 test programs the emulator's tests run, natively and as
WebAssembly (DESIGN §15.2). They are CP/M `.COM` programs, run by
`internal/cpm`. They are test data only: never part of the app or the site.

| File | What it is |
|---|---|
| `8080PRE.COM`, `8080PRE.MAC` | Preliminary tests: the instructions the exerciser itself needs |
| `8080EXM.COM`, `8080EXM.MAC` | The instruction exerciser, with the CRCs of a real 8080 built in, so it prints "PASS!" or "ERROR" per instruction group |

## Origin

Frank D. Cringle's Z80 exercisers `prelim.z80` and `zexlax.z80` (1994),
converted to the 8080 by Ian Bartholomew (February 2009). `8080EXM` is his
version with the real 8080's CRCs filled in.

Downloaded on 2026-11-25 from <https://altairclone.com/downloads/cpu_tests/>,
unchanged (CRLF line endings kept; see `.gitattributes`).

| File | SHA-256 |
|---|---|
| `8080PRE.COM` | `18eb3c79cba42c0718f160be6a1853cb64cdce7aa47d65780189a57bdd98c4e0` |
| `8080PRE.MAC` | `ca1507444929978038ad4f83d18e13bcce72072df2b850932b3982c31cfc1ccc` |
| `8080EXM.COM` | `6e3286e11bb1a8f47b8ee1280b4a067be813193363e3223c99b0d21912f44aeb` |
| `8080EXM.MAC` | `806d3a069b0021e9925c0b7c26fd74a3c397ca7f618a599ab4a8396ebcd1f3f3` |

## Licence

GNU General Public License, version 2 or (at your option) any later version,
as stated in the header of each `.MAC` source: "Copyright (C) 1994 Frank D.
Cringle". The licence text is in [COPYING](COPYING) (from
<https://www.gnu.org/licenses/old-licenses/gpl-2.0.txt>).

These programs are not linked into the emulator: they are data it loads and
runs, kept here with their source, so the emulator's own code stays under its
BSD-3 licence.
