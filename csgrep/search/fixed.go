package search

import "strings"

type FixedMatcher struct {
	pattern       string
	caseSensitive bool
	lowerPattern  string
}

func NewFixedMatcher(pattern string, caseSensitive bool) *FixedMatcher {
	return &FixedMatcher{
		pattern:       pattern,
		caseSensitive: caseSensitive,
		lowerPattern:  strings.ToLower(pattern),
	}
}

func (m *FixedMatcher) Match(text string) (bool, [][2]int, float64) {
	var offsets [][2]int
	searchText := text
	searchPattern := m.pattern
	if !m.caseSensitive {
		searchText = strings.ToLower(text)
		searchPattern = m.lowerPattern
	}

	start := 0
	for {
		idx := strings.Index(searchText[start:], searchPattern)
		if idx < 0 {
			break
		}
		absIdx := start + idx
		offsets = append(offsets, [2]int{absIdx, absIdx + len(searchPattern)})
		start = absIdx + len(searchPattern)
	}

	if len(offsets) == 0 {
		return false, nil, 0
	}
	return true, offsets, 1.0
}
