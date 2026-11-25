package font

import "strings"

// Wrap breaks text into lines of at most cols characters, at spaces where
// it can; a word longer than a line is cut. Newlines in text start new
// lines.
func Wrap(text string, cols int) []string {
	var lines []string

	for para := range strings.SplitSeq(text, "\n") {
		line := ""

		for word := range strings.FieldsSeq(para) {
			for len([]rune(word)) > cols { // too long for any line: cut it
				if line != "" {
					lines = append(lines, line)
					line = ""
				}

				r := []rune(word)
				lines = append(lines, string(r[:cols]))
				word = string(r[cols:])
			}

			switch {
			case line == "":
				line = word
			case len([]rune(line))+1+len([]rune(word)) <= cols:
				line += " " + word
			default:
				lines = append(lines, line)
				line = word
			}
		}

		lines = append(lines, line)
	}

	return lines
}
