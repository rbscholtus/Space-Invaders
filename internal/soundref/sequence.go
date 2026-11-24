package soundref

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// FrameRate is MAME's (and the board's) video frame rate: 19.968 MHz / 4 /
// 320 / 262 (DESIGN §8). Sequence frames are converted to time with it.
const FrameRate = 19968000.0 / 4 / 320 / 262

// ErrSpec is returned for a malformed sequence spec.
var ErrSpec = errors.New("bad sequence spec")

// Write is one port write in a sequence. MAME's Lua frame callback runs a
// frame late: a write listed at frame f takes effect at the end of frame
// f + 1 (measured with MAME 0.289: the frame-150 write happens at
// 2.536026 s, the end of frame 151). Sample returns that moment.
type Write struct {
	Frame int
	Port  byte
	Value byte
}

// Sample returns the sample at which the write takes effect, at rate.
func (w Write) Sample(rate int) int { return SampleAt(w.Frame+1, rate) }

// Sequence is a named series of sound-port writes, Frames long.
type Sequence struct {
	Name   string
	Frames int
	Writes []Write
}

// SampleAt returns the sample index at the end of frame f, at rate.
func SampleAt(f, rate int) int {
	return int(float64(f)*float64(rate)/FrameRate + 0.5)
}

// ReadSequences reads the spec in testdata/reference/sequences.txt:
// "sound <name> <frames>" lines, each followed by "<frame> <port> <value>"
// lines; # starts a comment.
func ReadSequences(r io.Reader) ([]Sequence, error) {
	var seqs []Sequence

	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line, _, _ := strings.Cut(sc.Text(), "#")

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if len(fields) != 3 {
			return nil, fmt.Errorf("%w: line %d: %q", ErrSpec, n, sc.Text())
		}

		if fields[0] == "sound" {
			frames, err := strconv.Atoi(fields[2])
			if err != nil {
				return nil, fmt.Errorf("%w: line %d: %w", ErrSpec, n, err)
			}

			seqs = append(seqs, Sequence{Name: fields[1], Frames: frames})

			continue
		}

		if len(seqs) == 0 {
			return nil, fmt.Errorf("%w: line %d: a write before any sound", ErrSpec, n)
		}

		frame, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, fmt.Errorf("%w: line %d: %w", ErrSpec, n, err)
		}

		var bytes [2]byte

		for i, f := range fields[1:] {
			v, err := strconv.ParseUint(f, 0, 8)
			if err != nil {
				return nil, fmt.Errorf("%w: line %d: port and value are bytes: %w", ErrSpec, n, err)
			}

			bytes[i] = byte(v) // ParseUint's bitSize 8 bounds v
		}

		last := &seqs[len(seqs)-1]
		last.Writes = append(last.Writes, Write{Frame: frame, Port: bytes[0], Value: bytes[1]})
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("reading the spec: %w", err)
	}

	return seqs, nil
}
