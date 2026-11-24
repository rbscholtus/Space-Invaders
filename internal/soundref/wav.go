// Package soundref measures sound for the synth's reference tests (DESIGN
// §14.5): it reads and writes WAV files, reads the shared sequence spec, and
// measures loudness, pitch and brightness per 10 ms so that two renderings
// of a sound can be compared with tolerances.
//
// It uses only the standard library. It is used by tests and by
// cmd/soundref, never by the app.
package soundref

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// ErrWAV is returned for WAV files this package doesn't read: anything but
// 16-bit PCM.
var ErrWAV = errors.New("not a 16-bit PCM WAV file")

const (
	pcmFormat     = 1
	bitsPerSample = 16
	fullScale     = 32768
)

// ReadWAV reads 16-bit PCM audio, averaging the channels, as samples in
// [-1, 1), and returns its sample rate.
func ReadWAV(r io.Reader) ([]float32, int, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, 0, fmt.Errorf("reading the WAV: %w", err)
	}

	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, 0, ErrWAV
	}

	var (
		channels, bits, format int
		rate                   int
		pcm                    []byte
	)

	for pos := 12; pos+8 <= len(data); {
		id := string(data[pos : pos+4])
		size := int(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		body := data[pos+8 : min(pos+8+size, len(data))]

		switch id {
		case "fmt ":
			if len(body) < 16 {
				return nil, 0, ErrWAV
			}

			format = int(binary.LittleEndian.Uint16(body[0:2]))
			channels = int(binary.LittleEndian.Uint16(body[2:4]))
			rate = int(binary.LittleEndian.Uint32(body[4:8]))
			bits = int(binary.LittleEndian.Uint16(body[14:16]))
		case "data":
			pcm = body
		}

		pos += 8 + size + size&1 // chunks are padded to even sizes
	}

	if format != pcmFormat || bits != bitsPerSample || channels < 1 || rate <= 0 || pcm == nil {
		return nil, 0, ErrWAV
	}

	frames := len(pcm) / (2 * channels)
	out := make([]float32, frames)

	for i := range out {
		var sum int

		for c := range channels {
			sum += int(int16(binary.LittleEndian.Uint16(pcm[(i*channels+c)*2:]))) //nolint:gosec // G115: a 16-bit sample
		}

		out[i] = float32(sum) / float32(channels) / fullScale
	}

	return out, rate, nil
}

// WriteWAV writes samples (clipped to [-1, 1]) as mono 16-bit PCM.
func WriteWAV(w io.Writer, samples []float32, rate int) error {
	const headerSize = 44

	buf := make([]byte, headerSize+2*len(samples))
	copy(buf[0:], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:], uint32(len(buf)-8)) //nolint:gosec // G115: test-sized files
	copy(buf[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(buf[16:], 16)
	binary.LittleEndian.PutUint16(buf[20:], pcmFormat)
	binary.LittleEndian.PutUint16(buf[22:], 1)
	binary.LittleEndian.PutUint32(buf[24:], uint32(rate))   //nolint:gosec // G115: a sample rate
	binary.LittleEndian.PutUint32(buf[28:], uint32(rate*2)) //nolint:gosec // G115: a byte rate
	binary.LittleEndian.PutUint16(buf[32:], 2)
	binary.LittleEndian.PutUint16(buf[34:], bitsPerSample)
	copy(buf[36:], "data")
	binary.LittleEndian.PutUint32(buf[40:], uint32(2*len(samples))) //nolint:gosec // G115: test-sized files

	for i, s := range samples {
		v := math.Round(float64(max(-1, min(1, s))) * (fullScale - 1))
		pcm := uint16(int16(v)) //nolint:gosec // G115: two's complement, as WAV stores it
		binary.LittleEndian.PutUint16(buf[headerSize+2*i:], pcm)
	}

	if _, err := w.Write(buf); err != nil {
		return fmt.Errorf("writing the WAV: %w", err)
	}

	return nil
}
