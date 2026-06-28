package search

import (
	"regexp"
	"strings"
)

type RegexMatcher struct {
	re *regexp.Regexp
}

func NewRegexMatcher(pattern string, caseSensitive bool) (*RegexMatcher, error) {
	if !caseSensitive {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexMatcher{re: re}, nil
}

func (m *RegexMatcher) Match(text string) (bool, [][2]int, float64) {
	locs := m.re.FindAllStringIndex(text, -1)
	if locs == nil {
		return false, nil, 0
	}
	offsets := make([][2]int, len(locs))
	for i, loc := range locs {
		offsets[i] = [2]int{loc[0], loc[1]}
	}
	return true, offsets, 1.0
}

func SmartCasePattern(pattern string) bool {
	for _, r := range pattern {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

// used by show command to highlight
func HighlightText(text string, offsets [][2]int, hiStart, hiEnd string) string {
	if len(offsets) == 0 {
		return text
	}
	var b strings.Builder
	last := 0
	for _, o := range offsets {
		if o[0] > last {
			b.WriteString(text[last:o[0]])
		}
		b.WriteString(hiStart)
		b.WriteString(text[o[0]:o[1]])
		b.WriteString(hiEnd)
		last = o[1]
	}
	if last < len(text) {
		b.WriteString(text[last:])
	}
	return b.String()
}
