package main

import (
	"bytes"
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

func TestCLI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		program    []byte // written to a temporary file, appended to args
		wantCode   int
		wantStdout string // substring
		wantStderr string // substring
	}{
		{name: "8080PRE passes", args: []string{pre}, wantCode: exitOK, wantStdout: "8080 Preliminary tests complete", wantStderr: "cycles in"},
		{name: "no arguments", args: nil, wantCode: exitUsage, wantStderr: "usage: cputest"},
		{name: "two files", args: []string{pre, pre}, wantCode: exitUsage, wantStderr: "usage: cputest"},
		{name: "unknown flag", args: []string{"-x", pre}, wantCode: exitUsage, wantStderr: "flag provided but not defined"},
		{name: "help", args: []string{"-h"}, wantCode: exitOK, wantStderr: "-max"},
		{name: "missing file", args: []string{"nonexistent.COM"}, wantCode: exitFail, wantStderr: "cputest: reading the program"},
		{name: "too big", program: make([]byte, 0xFE00), wantCode: exitFail, wantStderr: "does not fit in the TPA"},
		{name: "cycle limit", args: []string{"-max", "100", pre}, wantCode: exitFail, wantStderr: "no warm boot after 100 cycles"},
		{
			name:     "unsupported BDOS call",
			program:  []byte{0x0E, 0x01, 0xCD, 0x05, 0x00, 0xC7}, // MVI C,1; CALL 5; RST 0
			wantCode: exitFail, wantStderr: "unsupported BDOS function 1",
		},
		{
			name:     "prints with BDOS 9",
			program:  []byte{0x0E, 0x09, 0x11, 0x09, 0x01, 0xCD, 0x05, 0x00, 0xC7, 'h', 'i', '$'}, // MVI C,9; LXI D,0109H; CALL 5; RST 0
			wantCode: exitOK, wantStdout: "hi\n", wantStderr: "cycles in",
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

			if !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Errorf("stdout %q, want it to contain %q", stdout.String(), tt.wantStdout)
			}

			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr %q, want it to contain %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}
