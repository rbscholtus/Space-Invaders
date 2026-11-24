package soundref

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// Reference is the committed measurement of one reference recording (DESIGN
// §14.5): the recording itself stays local.
type Reference struct {
	Features

	Sound      string `json:"sound"`
	Source     string `json:"source"`     // what recorded it, e.g. "MAME 0.289"
	WAVSHA256  string `json:"wavSha256"`  // of the recording measured
	StartFrame int    `json:"startFrame"` // measurement starts when this frame's write takes effect
}

// WriteReference writes ref as JSON, with values rounded to what the tests
// need and one array per line, so the files stay small and diffs readable.
func WriteReference(w io.Writer, ref Reference) error {
	round := func(x []float64, decimals float64) []float64 {
		out := make([]float64, len(x))
		scale := math.Pow(10, decimals)

		for i, v := range x {
			out[i] = math.Round(v*scale) / scale
		}

		return out
	}

	head, err := json.Marshal(struct {
		Sound      string `json:"sound"`
		Source     string `json:"source"`
		WAVSHA256  string `json:"wavSha256"`
		StartFrame int    `json:"startFrame"`
		SampleRate int    `json:"sampleRate"`
		Hop        int    `json:"hop"`
	}{ref.Sound, ref.Source, ref.WAVSHA256, ref.StartFrame, ref.SampleRate, ref.Hop})
	if err != nil {
		return fmt.Errorf("encoding %s: %w", ref.Sound, err)
	}

	var buf bytes.Buffer

	buf.Write(head[:len(head)-1]) // without the closing brace

	for _, field := range []struct {
		name     string
		values   []float64
		decimals float64
	}{{"rms", ref.RMS, 5}, {"pitch", ref.Pitch, 2}, {"centroid", ref.Centroid, 1}} {
		values, err := json.Marshal(round(field.values, field.decimals))
		if err != nil {
			return fmt.Errorf("encoding %s: %w", ref.Sound, err)
		}

		fmt.Fprintf(&buf, ",\n%q: %s", field.name, values)
	}

	buf.WriteString("\n}\n")

	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("writing %s: %w", ref.Sound, err)
	}

	return nil
}

// ReadReference reads a Reference written by WriteReference.
func ReadReference(r io.Reader) (Reference, error) {
	var ref Reference
	if err := json.NewDecoder(r).Decode(&ref); err != nil {
		return Reference{}, fmt.Errorf("reading a reference: %w", err)
	}

	return ref, nil
}
