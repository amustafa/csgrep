package session

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/amustafa/csgrep/include"
)

func testdataPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "testdata", name)
}

func TestParseSampleSession(t *testing.T) {
	s, err := Parse(testdataPath("sample-session.jsonl"), ParseOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected session, got nil")
	}
	if s.ID != "sample-session" {
		t.Errorf("ID = %q, want %q", s.ID, "sample-session")
	}
	if s.CWD != "/home/user/myproject" {
		t.Errorf("CWD = %q, want %q", s.CWD, "/home/user/myproject")
	}
	if s.Entrypoint != "cli" {
		t.Errorf("Entrypoint = %q, want %q", s.Entrypoint, "cli")
	}
}

func TestParseFirstMessage(t *testing.T) {
	s, err := Parse(testdataPath("sample-session.jsonl"), ParseOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.FirstMessage != "hello world" {
		t.Errorf("FirstMessage = %q, want %q", s.FirstMessage, "hello world")
	}
}

func TestParseLastMessage(t *testing.T) {
	s, err := Parse(testdataPath("sample-session.jsonl"), ParseOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.LastMessage != "help me with the database migration" {
		t.Errorf("LastMessage = %q, want %q", s.LastMessage, "help me with the database migration")
	}
}

func TestParseClearResetsFirstMessage(t *testing.T) {
	s, err := Parse(testdataPath("session-with-clear.jsonl"), ParseOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.FirstMessage != "new conversation after clear" {
		t.Errorf("FirstMessage = %q, want %q", s.FirstMessage, "new conversation after clear")
	}
}

func TestParseClearResetsMessages(t *testing.T) {
	s, err := Parse(testdataPath("session-with-clear.jsonl"), ParseOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, msg := range s.Messages {
		if msg.Text == "old conversation start" {
			t.Error("messages should not contain pre-clear content")
		}
	}
	if len(s.Messages) != 2 {
		t.Errorf("message count = %d, want 2 (post-clear only)", len(s.Messages))
	}
}

func TestParseMetadataOnly(t *testing.T) {
	s, err := Parse(testdataPath("sample-session.jsonl"), ParseOptions{MetadataOnly: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Messages) != 0 {
		t.Errorf("MetadataOnly should produce no messages, got %d", len(s.Messages))
	}
	if s.FirstMessage == "" {
		t.Error("MetadataOnly should still populate FirstMessage")
	}
}

func TestParseToolContentExcludedByDefault(t *testing.T) {
	s, err := Parse(testdataPath("sample-session.jsonl"), ParseOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, msg := range s.Messages {
		if msg.Role == "user" && msg.Text == "CREATE TABLE users (id INT);" {
			t.Error("tool_result content should be excluded by default")
		}
	}
}

func TestParseToolContentIncluded(t *testing.T) {
	s, err := Parse(testdataPath("sample-session.jsonl"), ParseOptions{
		Include: include.FromAll(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundToolInput := false
	for _, msg := range s.Messages {
		if msg.Role == "assistant" && contains(msg.Text, "file_path") {
			foundToolInput = true
		}
	}
	if !foundToolInput {
		t.Error("tool_use input should be included with ToolOutputs")
	}
}

func TestParseTimestampsAreLocal(t *testing.T) {
	s, err := Parse(testdataPath("sample-session.jsonl"), ParseOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.LastTime.IsZero() {
		t.Fatal("LastTime should not be zero")
	}
	if s.LastTime.Location().String() == "UTC" {
		t.Error("timestamps should be converted to local time")
	}
}

func TestCleanText(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "hello world"},
		{"<b>bold</b> text", "bold text"},
		{"\x1b[31mred\x1b[0m text", "red text"},
		{"  spaces  ", "spaces"},
		{"line1\nline2", "line1 line2"},
		{"<div class=\"foo\">content</div>", "content"},
	}
	for _, tt := range tests {
		got := CleanText(tt.input)
		if got != tt.want {
			t.Errorf("CleanText(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTruncateRunes(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("truncate short = %q", got)
	}
	if got := truncate("hello world", 5); got != "hello..." {
		t.Errorf("truncate long = %q, want %q", got, "hello...")
	}
	emoji := "👋🌍🎉🔥💡✨"
	got := truncate(emoji, 3)
	if got != "👋🌍🎉..." {
		t.Errorf("truncate emoji = %q, want %q", got, "👋🌍🎉...")
	}
}

// Artifact tests

func TestParseArtifactsExcludedByDefault(t *testing.T) {
	s, err := Parse(testdataPath("session-with-artifacts.jsonl"), ParseOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, msg := range s.Messages {
		if msg.Role == "artifact" {
			t.Error("artifacts should not appear without Include.Artifacts")
		}
	}
}

func TestParseArtifactsIncluded(t *testing.T) {
	s, err := Parse(testdataPath("session-with-artifacts.jsonl"), ParseOptions{
		Include: include.IncludeSet{Artifacts: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var artifacts []Message
	for _, msg := range s.Messages {
		if msg.Role == "artifact" {
			artifacts = append(artifacts, msg)
		}
	}

	// Should find Write (main.go), Edit (main.go), NotebookEdit (analysis.ipynb)
	// Temp file (/tmp/debug-output.txt) excluded by default scope
	// Plan file excluded by default scope (it's not under /tmp but IS a regular file)
	if len(artifacts) != 4 {
		t.Errorf("expected 4 artifacts (Write, Edit, Notebook, Plan), got %d", len(artifacts))
		for _, a := range artifacts {
			t.Logf("  artifact: %s %s", a.ToolName, a.FilePath)
		}
	}
}

func TestParseArtifactFields(t *testing.T) {
	s, err := Parse(testdataPath("session-with-artifacts.jsonl"), ParseOptions{
		Include: include.IncludeSet{Artifacts: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var writeArtifact *Message
	for _, msg := range s.Messages {
		if msg.Role == "artifact" && msg.ToolName == "Write" && msg.FilePath == "/home/user/myproject/main.go" {
			writeArtifact = &msg
			break
		}
	}
	if writeArtifact == nil {
		t.Fatal("Write artifact for main.go not found")
	}
	if writeArtifact.FilePath != "/home/user/myproject/main.go" {
		t.Errorf("FilePath = %q", writeArtifact.FilePath)
	}
	if !contains(writeArtifact.Text, "hello world") {
		t.Error("Write artifact text should contain file content")
	}
	if !contains(writeArtifact.Text, "main.go") {
		t.Error("Write artifact text should contain file path")
	}
}

func TestParseArtifactEdit(t *testing.T) {
	s, err := Parse(testdataPath("session-with-artifacts.jsonl"), ParseOptions{
		Include: include.IncludeSet{Artifacts: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var editArtifact *Message
	for _, msg := range s.Messages {
		if msg.Role == "artifact" && msg.ToolName == "Edit" {
			editArtifact = &msg
			break
		}
	}
	if editArtifact == nil {
		t.Fatal("Edit artifact not found")
	}
	if !contains(editArtifact.Text, "hello world") {
		t.Error("Edit artifact should contain old_string")
	}
	if !contains(editArtifact.Text, "hello universe") {
		t.Error("Edit artifact should contain new_string")
	}
}

func TestParseArtifactScopeAll(t *testing.T) {
	s, err := Parse(testdataPath("session-with-artifacts.jsonl"), ParseOptions{
		Include: include.IncludeSet{Artifacts: true, ArtifactScope: "all"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var artifacts []Message
	for _, msg := range s.Messages {
		if msg.Role == "artifact" {
			artifacts = append(artifacts, msg)
		}
	}

	// all scope: Write, Edit, Notebook, Temp, Plan = 5
	if len(artifacts) != 5 {
		t.Errorf("expected 5 artifacts with scope=all, got %d", len(artifacts))
		for _, a := range artifacts {
			t.Logf("  artifact: %s %s", a.ToolName, a.FilePath)
		}
	}
}

func TestParseArtifactScopeTemp(t *testing.T) {
	s, err := Parse(testdataPath("session-with-artifacts.jsonl"), ParseOptions{
		Include: include.IncludeSet{Artifacts: true, ArtifactScope: "tmp"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var artifacts []Message
	for _, msg := range s.Messages {
		if msg.Role == "artifact" {
			artifacts = append(artifacts, msg)
		}
	}
	if len(artifacts) != 1 {
		t.Errorf("expected 1 temp artifact, got %d", len(artifacts))
	}
	if len(artifacts) > 0 && artifacts[0].FilePath != "/tmp/debug-output.txt" {
		t.Errorf("expected /tmp/debug-output.txt, got %s", artifacts[0].FilePath)
	}
}

func TestParseArtifactMatchPath(t *testing.T) {
	s, err := Parse(testdataPath("session-with-artifacts.jsonl"), ParseOptions{
		Include: include.IncludeSet{Artifacts: true, ArtifactMatch: "path"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, msg := range s.Messages {
		if msg.Role == "artifact" {
			if contains(msg.Text, "hello world") {
				t.Error("artifacts:path should not include file content")
			}
			if !contains(msg.Text, "/") {
				t.Error("artifacts:path should contain file path")
			}
		}
	}
}

func TestParseArtifactMatchContent(t *testing.T) {
	s, err := Parse(testdataPath("session-with-artifacts.jsonl"), ParseOptions{
		Include: include.IncludeSet{Artifacts: true, ArtifactMatch: "content"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, msg := range s.Messages {
		if msg.Role == "artifact" && msg.ToolName == "Write" && msg.FilePath == "/home/user/myproject/main.go" {
			if contains(msg.Text, "main.go") {
				t.Error("artifacts:content should not include file path in text")
			}
			if !contains(msg.Text, "hello world") {
				t.Error("artifacts:content should include file content")
			}
		}
	}
}

func TestParseArtifactPaths(t *testing.T) {
	s, err := Parse(testdataPath("session-with-artifacts.jsonl"), ParseOptions{
		Include: include.IncludeSet{Artifacts: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.ArtifactPaths) == 0 {
		t.Error("ArtifactPaths should be populated")
	}
}

func TestParseToolOutputsIncluded(t *testing.T) {
	s, err := Parse(testdataPath("session-with-artifacts.jsonl"), ParseOptions{
		Include: include.IncludeSet{ToolOutputs: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var toolOutputs []Message
	for _, msg := range s.Messages {
		if msg.Role == "tool-output" {
			toolOutputs = append(toolOutputs, msg)
		}
	}
	if len(toolOutputs) == 0 {
		t.Error("tool outputs should be present with ToolOutputs=true")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
