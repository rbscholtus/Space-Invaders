// Package i8080 emulates the Intel 8080 CPU (DESIGN §6).
//
// It depends only on the standard library and never reads the clock: the CPU
// advances only through Step and Run, so every run is repeatable.
package i8080
