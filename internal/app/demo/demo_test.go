package demo_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/app/demo"
)

func TestFormat(t *testing.T) {
	t.Parallel()

	tests := []struct{ in, want string }{
		{"dad <b,d,h,sp>................  PASS! crc is:14474ba6", "dad <b,d,h,sp>            OK"},
		{"mvi <b,c,d,e,h,l,m,a>,nn......  PASS! crc is:eaa72044", "mvi <b,c,d,e,h,l,m,a>,nn  OK"},
		{"aluop nn......................  ERROR **** crc expected:9e922f9e found:00000000", "aluop nn               ERROR"},
		{"8080 instruction exerciser", "8080 instruction exerciser"},
		{"a header line that is far too long for the monitor", "a header line that is far to"},
	}
	for _, tt := range tests {
		got := demo.Format(tt.in)
		if got != tt.want {
			t.Errorf("Format(%q) = %q, want %q", tt.in, got, tt.want)
		}

		if len(got) > 28 {
			t.Errorf("Format(%q) is %d columns", tt.in, len(got))
		}
	}
}

// TestRunsTheExerciser: the demo shows the header, then the group being
// tested with "...", then its result.
func TestRunsTheExerciser(t *testing.T) {
	t.Parallel()

	program, err := os.ReadFile("../../../third_party/8080-exercisers/8080EXM.COM")
	if err != nil {
		t.Fatal(err)
	}

	d, err := demo.New(program)
	if err != nil {
		t.Fatal(err)
	}

	d.Run(1_000_000)

	lines := d.Lines()
	if len(lines) != 2 || lines[0] != "8080 instruction exerciser" || !strings.HasSuffix(lines[1], "...") {
		t.Fatalf("after 1 M cycles: %q", lines)
	}

	for !strings.HasSuffix(d.Lines()[1], "OK") {
		d.Run(100_000_000)
	}

	if got := d.Lines()[1]; got != "dad <b,d,h,sp>            OK" {
		t.Errorf("first result %q", got)
	}

	if d.Done() {
		t.Error("done after the first group")
	}

	pic := make([]byte, 224*256)
	d.Draw(pic, 224, 20, 2)

	lit := 0
	for _, p := range pic[20*8*224:] {
		lit += int(p)
	}

	if lit == 0 {
		t.Error("Draw drew nothing")
	}
}

// TestLastLineWithoutNewline: output left without a newline when the
// program ends is a finished line, without "..." (8080EXM ends with
// "Tests complete").
func TestLastLineWithoutNewline(t *testing.T) {
	t.Parallel()

	program := slices.Concat([]byte{
		0x0E, 0x09, // MVI C,9: print a $-terminated string
		0x11, 0x0B, 0x01, // LXI D,0x010B: the string after this code
		0xCD, 0x05, 0x00, // CALL 5
		0xC3, 0x00, 0x00, // JMP 0: warm boot
	}, []byte("Tests complete$"))

	d, err := demo.New(program)
	if err != nil {
		t.Fatal(err)
	}

	d.Run(10_000)

	if got := d.Lines(); !d.Done() || len(got) != 1 || got[0] != "Tests complete" {
		t.Errorf("done %v, lines %q, want [\"Tests complete\"]", d.Done(), got)
	}
}

func TestBadProgram(t *testing.T) {
	t.Parallel()

	if _, err := demo.New(make([]byte, 0x10000)); err == nil {
		t.Error("accepted a program too large for CP/M")
	}

	// A program that calls an unsupported BDOS function stops with the
	// error shown.
	d, err := demo.New([]byte{0x0E, 0x01, 0xCD, 0x05, 0x00, 0xC7})
	if err != nil {
		t.Fatal(err)
	}

	d.Run(1000)

	if !d.Done() || !strings.Contains(strings.Join(d.Lines(), " "), "unsupported") {
		t.Errorf("done %v, lines %q", d.Done(), d.Lines())
	}
}
