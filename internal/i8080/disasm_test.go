package i8080

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata/")

// TestDisassembleGolden disassembles all 256 opcodes, each followed by the
// bytes 0xAB 0xCD, and compares with testdata/disasm.golden (DESIGN §15.1).
func TestDisassembleGolden(t *testing.T) {
	t.Parallel()

	var b strings.Builder

	for op := range 256 {
		text, n := Disassemble([]byte{byte(op), 0xAB, 0xCD}, 0)
		fmt.Fprintf(&b, "%02X  %d  %s\n", op, n, text)
	}

	got := b.String()

	path := filepath.Join("testdata", "disasm.golden")
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}

		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if got != string(want) {
		t.Errorf("disassembly differs from %s; run with -update and check the diff", path)
	}
}

func TestDisassemble(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mem  []byte
		pc   uint16
		text string
		n    int
	}{
		{[]byte{0x00, 0x3E, 0x0A}, 1, "MVI A,0AH", 2},
		{[]byte{0x3E, 0xFF}, 0, "MVI A,0FFH", 2},
		{[]byte{0xC3, 0x00, 0x01}, 0, "JMP 0100H", 3},
		{[]byte{0xCD, 0x34, 0xE2}, 0, "CALL 0E234H", 3},
		{[]byte{0x21}, 0, "LXI H,0000H", 3}, // bytes past the end read as 0
		{[]byte{0xD3, 0x06}, 0, "OUT 06H", 2},
		{[]byte{0x7E}, 0, "MOV A,M", 1},
		{nil, 0, "NOP", 1},
	}
	for _, tt := range tests {
		text, n := Disassemble(tt.mem, tt.pc)
		if text != tt.text || n != tt.n {
			t.Errorf("Disassemble(% X, %d) = %q, %d; want %q, %d", tt.mem, tt.pc, text, n, tt.text, tt.n)
		}
	}
}
