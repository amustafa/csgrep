package include

import (
	"testing"
)

func TestParseValid(t *testing.T) {
	tests := []struct {
		input string
		want  IncludeSet
	}{
		{"artifacts", IncludeSet{Artifacts: true}},
		{"tool-outputs", IncludeSet{ToolOutputs: true}},
		{"artifacts,tool-outputs", IncludeSet{Artifacts: true, ToolOutputs: true}},
		{"artifacts:path", IncludeSet{Artifacts: true, ArtifactMatch: "path"}},
		{"artifacts:content", IncludeSet{Artifacts: true, ArtifactMatch: "content"}},
		{"artifacts:all", IncludeSet{Artifacts: true, ArtifactScope: "all"}},
		{"artifacts:tmp", IncludeSet{Artifacts: true, ArtifactScope: "tmp"}},
		{"artifacts:plans", IncludeSet{Artifacts: true, ArtifactScope: "plans"}},
	}
	for _, tt := range tests {
		got, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Parse(%q) = %+v, want %+v", tt.input, got, tt.want)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	invalid := []string{
		"foo",
		"artefacts",
		"artifacts:foo",
		"artifacts,foo",
		"tool-output",
	}
	for _, input := range invalid {
		_, err := Parse(input)
		if err == nil {
			t.Errorf("Parse(%q) should error, got nil", input)
		}
	}
}

func TestParseWhitespace(t *testing.T) {
	got, err := Parse(" artifacts , tool-outputs ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Artifacts || !got.ToolOutputs {
		t.Errorf("should handle whitespace: %+v", got)
	}
}

func TestFromAll(t *testing.T) {
	s := FromAll()
	if !s.Artifacts || !s.ToolOutputs {
		t.Error("FromAll should enable both")
	}
	if s.ArtifactScope != "" || s.ArtifactMatch != "" {
		t.Error("FromAll should use defaults")
	}
}

func TestIsEmpty(t *testing.T) {
	if !(IncludeSet{}).IsEmpty() {
		t.Error("zero value should be empty")
	}
	if (IncludeSet{Artifacts: true}).IsEmpty() {
		t.Error("artifacts set should not be empty")
	}
	if (IncludeSet{ToolOutputs: true}).IsEmpty() {
		t.Error("tool-outputs set should not be empty")
	}
}

func TestMatchesScope(t *testing.T) {
	tests := []struct {
		scope string
		path  string
		want  bool
	}{
		{"", "/home/user/project/main.go", true},
		{"", "/tmp/foo.txt", false},
		{"", "/var/tmp/bar.txt", false},
		{"", "/home/user/.claude/plans/plan.md", true},
		{"all", "/tmp/foo.txt", true},
		{"all", "/home/user/project/main.go", true},
		{"all", "/home/user/.claude/plans/plan.md", true},
		{"tmp", "/tmp/foo.txt", true},
		{"tmp", "/var/tmp/bar.txt", true},
		{"tmp", "/home/user/project/main.go", false},
		{"plans", "/home/user/.claude/plans/some-plan.md", true},
		{"plans", "/home/user/project/main.go", false},
		{"plans", "/tmp/foo.txt", false},
	}
	for _, tt := range tests {
		s := IncludeSet{Artifacts: true, ArtifactScope: tt.scope}
		got := s.MatchesScope(tt.path)
		if got != tt.want {
			t.Errorf("scope=%q path=%q: got %v, want %v", tt.scope, tt.path, got, tt.want)
		}
	}
}

func TestShouldIncludeToolContent(t *testing.T) {
	if (IncludeSet{}).ShouldIncludeToolContent() {
		t.Error("empty should not include tool content")
	}
	if !(IncludeSet{Artifacts: true}).ShouldIncludeToolContent() {
		t.Error("artifacts should include tool content")
	}
	if !(IncludeSet{ToolOutputs: true}).ShouldIncludeToolContent() {
		t.Error("tool-outputs should include tool content")
	}
}
