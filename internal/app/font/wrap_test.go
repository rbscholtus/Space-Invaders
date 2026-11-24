package font

import (
	"slices"
	"testing"
)

func TestWrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text string
		cols int
		want []string
	}{
		{"ROM missing: invaders.f", 28, []string{"ROM missing: invaders.f"}},
		{"drop your invaders.zip on this window", 12, []string{"drop your", "invaders.zip", "on this", "window"}},
		{"crc 0123456789abcdef0123 end", 8, []string{"crc", "01234567", "89abcdef", "0123 end"}},
		{"two\nlines", 28, []string{"two", "lines"}},
		{"", 28, []string{""}},
	}
	for _, tt := range tests {
		if got := Wrap(tt.text, tt.cols); !slices.Equal(got, tt.want) {
			t.Errorf("Wrap(%q, %d) = %q, want %q", tt.text, tt.cols, got, tt.want)
		}
	}
}
