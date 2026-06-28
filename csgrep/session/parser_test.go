package session

import (
	"path/filepath"
	"runtime"
	"testing"
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
	s, err := Parse(testdataPath("sample-session.jsonl"), ParseOptions{IncludeToolContent: true})
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
		t.Error("tool_use input should be included with IncludeToolContent")
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
