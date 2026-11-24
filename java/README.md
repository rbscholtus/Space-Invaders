# Space Invaders emulator: the Java version (2010)

This folder holds the original Java emulator, as it was before the Go port.
The same code is tagged `java-2010`.

It's kept for reference and is no longer maintained. It was launched with
Java Web Start, which was removed in Java 11. Building it needs NetBeans/Ant,
a JDK, the ROMs (in `src/spaceinvaders/roms/`) and the sound samples (in
`src/spaceinvaders/samples/`).

Its known bugs are listed in [DESIGN §3](../docs/DESIGN.md#3-lessons-from-the-java-version).

Licence: Apache 2.0 (`LICENSE`, `NOTICE` in this folder). The Go port in the
rest of the repository is BSD-3-Clause.
