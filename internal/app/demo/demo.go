// Package demo is the CPU-test demo (DESIGN §9.4): without ROMs, the 8080
// exerciser 8080EXM runs under the cpm harness and its results show on the
// monitor, one line per instruction group. It is plain Go, without
// Ebitengine; the app decides how many cycles each frame gets.
package demo

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/rbscholtus/Space-Invaders/internal/app/font"
	"github.com/rbscholtus/Space-Invaders/internal/cpm"
)

// columns is the monitor's width in characters.
const columns = 28

// Demo runs one exerciser.
type Demo struct {
	machine *cpm.Machine
	out     bytes.Buffer // the program's output, not yet turned into lines
	lines   []string
	err     error
}

// New loads an exerciser program (a CP/M .COM file).
func New(program []byte) (*Demo, error) {
	d := &Demo{}

	m, err := cpm.New(program, &d.out)
	if err != nil {
		return nil, fmt.Errorf("the CPU-test demo: %w", err)
	}

	d.machine = m

	return d, nil
}

// Run runs up to cycles more cycles and takes in the output.
func (d *Demo) Run(cycles uint64) {
	if d.Done() {
		return
	}

	if _, err := d.machine.Run(cycles); err != nil {
		d.err = err
	}

	d.takeLines()

	// The exerciser's last line ("Tests complete") has no newline: once the
	// program has ended, it's a finished line, not a group still running.
	if d.machine.Done() {
		if rest := strings.Trim(d.out.String(), "\r\n"); rest != "" {
			d.lines = append(d.lines, Format(rest))
		}

		d.out.Reset()
	}
}

// Done reports whether the program has finished (or failed).
func (d *Demo) Done() bool { return d.machine.Done() || d.err != nil }

// Lines returns the results so far, ending with the group being tested
// (with "...") while one is running, at most 28 columns each.
func (d *Demo) Lines() []string {
	lines := d.lines

	if current := strings.Trim(d.out.String(), "\r\n"); current != "" {
		lines = append(slices.Clip(lines), fit(name(current), "..."))
	}

	if d.err != nil {
		lines = append(slices.Clip(lines), font.Wrap(d.err.Error(), columns)...)
	}

	return lines
}

// Format turns an output line of the exerciser into a monitor line: a
// group's result "dad <b,d,h,sp>....  PASS! crc is:14474ba6" becomes the
// group's name and OK or ERROR; other lines are cut to fit.
func Format(line string) string {
	switch {
	case strings.Contains(line, "PASS"):
		return fit(name(line), "OK")
	case strings.Contains(line, "ERROR"):
		return fit(name(line), "ERROR")
	}

	return fit(line, "")
}

// name is a result line's group name: what comes before the dots.
func name(line string) string {
	if i := strings.Index(line, ".."); i >= 0 {
		line = line[:i]
	}

	return strings.TrimSpace(line)
}

// fit puts text and status on one 28-column line, the status at the right
// edge, cutting the text if needed.
func fit(text, status string) string {
	room := columns
	if status != "" {
		room -= len(status) + 1
	}

	if r := []rune(text); len(r) > room {
		text = string(r[:room])
	}

	if status == "" {
		return text
	}

	return text + strings.Repeat(" ", columns-len([]rune(text))-len(status)) + status
}

// Draw draws the last lines that fit in rows of text, starting at
// character row row, into a picture width pixels wide.
func (d *Demo) Draw(dst []byte, width, row, rows int) {
	lines := d.Lines()
	if len(lines) > rows {
		lines = lines[len(lines)-rows:]
	}

	for i, line := range lines {
		font.Draw(dst, width, 0, (row+i)*font.CellHeight, line)
	}
}

// takeLines moves finished output lines into d.lines.
func (d *Demo) takeLines() {
	for {
		raw, err := d.out.ReadString('\n')
		if err != nil { // no complete line yet: put it back
			rest := []byte(raw)

			d.out.Reset()
			d.out.Write(rest)

			return
		}

		if line := strings.Trim(raw, "\r\n"); line != "" {
			d.lines = append(d.lines, Format(line))
		}
	}
}
