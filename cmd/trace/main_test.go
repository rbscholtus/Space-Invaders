package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const pre = "../../third_party/8080-exercisers/8080PRE.COM"

// writeProgram writes a .COM file to a temporary folder and returns its path.
func writeProgram(t *testing.T, program []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "TEST.COM")
	if err := os.WriteFile(path, program, 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

// TestFirstLines checks the trace format on the first instructions of 8080PRE.
func TestFirstLines(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if code := cli([]string{"-n", "3", pre}, &stdout, &stderr); code != exitOK {
		t.Fatalf("exit code %d\nstderr: %s", code, stderr.String())
	}

	want := "" +
		"0100  3E 01     MVI A,01H     A=00 BC=0000 DE=0000 HL=0000 SP=0000 F=..... CYC=0\n" +
		"0102  FE 02     CPI 02H       A=01 BC=0000 DE=0000 HL=0000 SP=0000 F=..... CYC=7\n" +
		"0104  CA 00 00  JZ 0000H      A=01 BC=0000 DE=0000 HL=0000 SP=0000 F=S..PC CYC=14\n"
	if stdout.String() != want {
		t.Errorf("trace:\n%s\nwant:\n%s", stdout.String(), want)
	}
}

// TestWholeRun traces 8080PRE to its warm boot (-n 0): the program's output
// goes to stderr, and the last line is the jump to 0.
func TestWholeRun(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if code := cli([]string{"-n", "0", pre}, &stdout, &stderr); code != exitOK {
		t.Fatalf("exit code %d\nstderr: %s", code, stderr.String())
	}

	if !strings.Contains(stderr.String(), "8080 Preliminary tests complete") {
		t.Errorf("stderr %q, want the program's output", stderr.String())
	}

	lines := strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n")
	if len(lines) < 1000 {
		t.Errorf("%d trace lines, want the whole run", len(lines))
	}

	if last := lines[len(lines)-1]; !strings.Contains(last, "JMP 0000H") {
		t.Errorf("last line %q, want the warm boot's JMP 0000H", last)
	}
}

func TestCLIErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		program    []byte // written to a temporary file, appended to args
		wantCode   int
		wantStderr string // substring
	}{
		{name: "no arguments", wantCode: exitUsage, wantStderr: "usage: trace"},
		{name: "two files", args: []string{pre, pre}, wantCode: exitUsage, wantStderr: "usage: trace"},
		{name: "bad count", args: []string{"-n", "x", pre}, wantCode: exitUsage, wantStderr: "invalid value"},
		{name: "help", args: []string{"-h"}, wantCode: exitOK, wantStderr: "-n"},
		{name: "missing file", args: []string{"nonexistent.COM"}, wantCode: exitFail, wantStderr: "trace: reading the program"},
		{name: "too big", program: make([]byte, 0xFE00), wantCode: exitFail, wantStderr: "does not fit in the TPA"},
		{
			name:     "unsupported BDOS call",
			program:  []byte{0x0E, 0x01, 0xCD, 0x05, 0x00, 0xC7}, // MVI C,1; CALL 5; RST 0
			wantCode: exitFail, wantStderr: "unsupported BDOS function 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			args := tt.args
			if tt.program != nil {
				args = append(args, writeProgram(t, tt.program))
			}

			var stdout, stderr bytes.Buffer

			code := cli(args, &stdout, &stderr)
			if code != tt.wantCode {
				t.Errorf("exit code %d, want %d\nstderr: %s", code, tt.wantCode, stderr.String())
			}

			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr %q, want it to contain %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

var errWrite = errors.New("write failed")

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errWrite }

// TestWriteErrors: a trace that can't be written fails, whether the error
// shows while tracing (a full buffer) or when the buffer is flushed.
func TestWriteErrors(t *testing.T) {
	t.Parallel()

	for _, n := range []string{"0", "1"} {
		var stderr bytes.Buffer
		if code := cli([]string{"-n", n, pre}, failingWriter{}, &stderr); code != exitFail {
			t.Errorf("-n %s: exit code %d, want %d", n, code, exitFail)
		}

		if !strings.Contains(stderr.String(), "writing the trace: write failed") {
			t.Errorf("-n %s: stderr %q", n, stderr.String())
		}
	}
}
