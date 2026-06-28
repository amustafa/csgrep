package search

import "testing"

func TestFuzzyMatchExact(t *testing.T) {
	m := NewFuzzyMatcher("database", 0.3)
	matched, _, score := m.Match("help with the database migration")
	if !matched {
		t.Fatal("exact substring should match")
	}
	if score < 0.8 {
		t.Errorf("exact match score = %f, want >= 0.8", score)
	}
}

func TestFuzzyMatchTypo(t *testing.T) {
	m := NewFuzzyMatcher("databse", 0.3)
	matched, _, score := m.Match("help with the database migration")
	if !matched {
		t.Fatal("typo should still match")
	}
	if score < 0.3 {
		t.Errorf("typo match score = %f, want >= 0.3", score)
	}
}

func TestFuzzyMatchNoMatch(t *testing.T) {
	m := NewFuzzyMatcher("quantum physics", 0.3)
	matched, _, _ := m.Match("help with the database migration")
	if matched {
		t.Error("unrelated text should not match")
	}
}

func TestFuzzyMatchThreshold(t *testing.T) {
	m := NewFuzzyMatcher("databse", 0.9)
	matched, _, _ := m.Match("help with the database migration")
	if matched {
		t.Error("high threshold should reject approximate match")
	}
}

func TestFuzzyMatchScoreOrdering(t *testing.T) {
	m := NewFuzzyMatcher("migration", 0.3)
	_, _, exactScore := m.Match("the migration is ready")
	_, _, typoScore := m.Match("the migrtion is ready")

	if exactScore <= typoScore {
		t.Errorf("exact score (%f) should be > typo score (%f)", exactScore, typoScore)
	}
}

func TestFuzzyMatchOffsets(t *testing.T) {
	m := NewFuzzyMatcher("hello", 0.3)
	matched, offsets, _ := m.Match("say hello world")
	if !matched {
		t.Fatal("should match")
	}
	if len(offsets) == 0 {
		t.Fatal("should have offsets")
	}
	start := offsets[0][0]
	end := offsets[0][1]
	if start < 0 || end > len("say hello world") || start >= end {
		t.Errorf("invalid offsets: [%d, %d]", start, end)
	}
}

func TestTrigramSet(t *testing.T) {
	set := trigramSet("hello")
	expected := map[string]bool{"hel": true, "ell": true, "llo": true}
	for k := range expected {
		if !set[k] {
			t.Errorf("missing trigram %q", k)
		}
	}
}

func TestTrigramSetShort(t *testing.T) {
	set := trigramSet("hi")
	if !set["hi"] {
		t.Error("short string should be added as-is")
	}
}

func TestTrigramSimilarityIdentical(t *testing.T) {
	a := trigramSet("hello")
	b := trigramSet("hello")
	score := trigramSimilarity(a, b)
	if score != 1.0 {
		t.Errorf("identical sets should have score 1.0, got %f", score)
	}
}

func TestTrigramSimilarityDisjoint(t *testing.T) {
	a := trigramSet("abc")
	b := trigramSet("xyz")
	score := trigramSimilarity(a, b)
	if score != 0.0 {
		t.Errorf("disjoint sets should have score 0.0, got %f", score)
	}
}
