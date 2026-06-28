package search

import "testing"

func TestRegexMatchBasic(t *testing.T) {
	m, err := NewRegexMatcher("hello", false)
	if err != nil {
		t.Fatal(err)
	}
	matched, offsets, score := m.Match("say hello world")
	if !matched {
		t.Fatal("should match")
	}
	if len(offsets) != 1 || offsets[0] != [2]int{4, 9} {
		t.Errorf("offsets = %v, want [[4,9]]", offsets)
	}
	if score != 1.0 {
		t.Errorf("score = %f, want 1.0", score)
	}
}

func TestRegexMatchCaseInsensitive(t *testing.T) {
	m, _ := NewRegexMatcher("hello", false)
	matched, _, _ := m.Match("HELLO WORLD")
	if !matched {
		t.Error("case-insensitive match should find HELLO")
	}
}

func TestRegexMatchCaseSensitive(t *testing.T) {
	m, _ := NewRegexMatcher("hello", true)
	matched, _, _ := m.Match("HELLO WORLD")
	if matched {
		t.Error("case-sensitive match should not find HELLO")
	}
}

func TestRegexMatchMultiple(t *testing.T) {
	m, _ := NewRegexMatcher("a", false)
	matched, offsets, _ := m.Match("abracadabra")
	if !matched {
		t.Fatal("should match")
	}
	if len(offsets) != 5 {
		t.Errorf("should find 5 matches, got %d", len(offsets))
	}
}

func TestRegexMatchNoMatch(t *testing.T) {
	m, _ := NewRegexMatcher("xyz", false)
	matched, _, _ := m.Match("hello world")
	if matched {
		t.Error("should not match")
	}
}

func TestRegexMatchPattern(t *testing.T) {
	m, _ := NewRegexMatcher(`\d+`, false)
	matched, offsets, _ := m.Match("error on line 42")
	if !matched {
		t.Fatal("should match digits")
	}
	if offsets[0] != [2]int{14, 16} {
		t.Errorf("offsets = %v, want [[14,16]]", offsets)
	}
}

func TestRegexInvalidPattern(t *testing.T) {
	_, err := NewRegexMatcher("[invalid", true)
	if err == nil {
		t.Error("should return error for invalid regex")
	}
}

func TestSmartCasePattern(t *testing.T) {
	if SmartCasePattern("hello") {
		t.Error("all lowercase should return false")
	}
	if !SmartCasePattern("Hello") {
		t.Error("mixed case should return true")
	}
	if SmartCasePattern("123") {
		t.Error("digits only should return false")
	}
}

func TestHighlightText(t *testing.T) {
	text := "hello world"
	offsets := [][2]int{{0, 5}}
	got := HighlightText(text, offsets, "[", "]")
	if got != "[hello] world" {
		t.Errorf("got %q, want %q", got, "[hello] world")
	}
}

func TestHighlightTextMultiple(t *testing.T) {
	text := "aXbXc"
	offsets := [][2]int{{1, 2}, {3, 4}}
	got := HighlightText(text, offsets, "<", ">")
	if got != "a<X>b<X>c" {
		t.Errorf("got %q, want %q", got, "a<X>b<X>c")
	}
}

func TestHighlightTextEmpty(t *testing.T) {
	got := HighlightText("hello", nil, "[", "]")
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}
