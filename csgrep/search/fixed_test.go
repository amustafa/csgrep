package search

import "testing"

func TestFixedMatchBasic(t *testing.T) {
	m := NewFixedMatcher("hello", false)
	matched, offsets, _ := m.Match("say hello world")
	if !matched {
		t.Fatal("should match")
	}
	if len(offsets) != 1 || offsets[0] != [2]int{4, 9} {
		t.Errorf("offsets = %v, want [[4,9]]", offsets)
	}
}

func TestFixedMatchCaseInsensitive(t *testing.T) {
	m := NewFixedMatcher("hello", false)
	matched, offsets, _ := m.Match("HELLO world")
	if !matched {
		t.Fatal("case-insensitive should match")
	}
	if offsets[0] != [2]int{0, 5} {
		t.Errorf("offsets = %v, want [[0,5]]", offsets)
	}
}

func TestFixedMatchCaseSensitive(t *testing.T) {
	m := NewFixedMatcher("hello", true)
	matched, _, _ := m.Match("HELLO world")
	if matched {
		t.Error("case-sensitive should not match")
	}
}

func TestFixedMatchMultiple(t *testing.T) {
	m := NewFixedMatcher("ab", false)
	matched, offsets, _ := m.Match("ababab")
	if !matched {
		t.Fatal("should match")
	}
	if len(offsets) != 3 {
		t.Errorf("should find 3 matches, got %d", len(offsets))
	}
}

func TestFixedMatchNoMatch(t *testing.T) {
	m := NewFixedMatcher("xyz", false)
	matched, _, _ := m.Match("hello world")
	if matched {
		t.Error("should not match")
	}
}

func TestFixedMatchSpecialChars(t *testing.T) {
	m := NewFixedMatcher("[test]", false)
	matched, _, _ := m.Match("this is a [test] string")
	if !matched {
		t.Error("should match literal brackets")
	}
}
