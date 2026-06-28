package search

import "strings"

type FuzzyMatcher struct {
	pattern   string
	trigrams  map[string]bool
	threshold float64
}

func NewFuzzyMatcher(pattern string, threshold float64) *FuzzyMatcher {
	lower := strings.ToLower(pattern)
	return &FuzzyMatcher{
		pattern:   lower,
		trigrams:  trigramSet(lower),
		threshold: threshold,
	}
}

func (m *FuzzyMatcher) Match(text string) (bool, [][2]int, float64) {
	lower := strings.ToLower(text)
	words := strings.Fields(lower)
	patternLen := len(m.pattern)

	wordStarts := make([]int, len(words))
	wordEnds := make([]int, len(words))
	pos := 0
	for i, word := range words {
		idx := strings.Index(lower[pos:], word)
		if idx < 0 {
			break
		}
		wordStarts[i] = pos + idx
		wordEnds[i] = pos + idx + len(word)
		pos = wordEnds[i]
	}

	bestScore := 0.0
	bestStart := -1
	bestEnd := -1

	for wi := range words {
		for windowSize := 1; windowSize <= 5 && wi+windowSize <= len(words); windowSize++ {
			window := strings.Join(words[wi:wi+windowSize], " ")
			if len(window) < patternLen/3 || len(window) > patternLen*3 {
				continue
			}

			score := trigramSimilarity(m.trigrams, trigramSet(window))
			if score > bestScore {
				bestScore = score
				bestStart = wordStarts[wi]
				bestEnd = wordEnds[wi+windowSize-1]
			}
		}
	}

	if bestScore < m.threshold {
		return false, nil, 0
	}

	var offsets [][2]int
	if bestStart >= 0 && bestEnd > bestStart && bestEnd <= len(text) {
		offsets = [][2]int{{bestStart, bestEnd}}
	}
	return true, offsets, bestScore
}

func trigramSet(s string) map[string]bool {
	set := make(map[string]bool)
	if len(s) < 3 {
		set[s] = true
		return set
	}
	for i := 0; i <= len(s)-3; i++ {
		set[s[i:i+3]] = true
	}
	return set
}

func trigramSimilarity(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	intersection := 0
	for k := range a {
		if b[k] {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

