package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	htmlTagRe = regexp.MustCompile(`<[^>]+>`)
	ansiRe    = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

type jsonEntry struct {
	Type       string          `json:"type"`
	CWD        string          `json:"cwd"`
	Entrypoint string          `json:"entrypoint"`
	Timestamp  string          `json:"timestamp"`
	Message    json.RawMessage `json:"message"`
}

type messageEnvelope struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type contentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text"`
	Input json.RawMessage `json:"input"`
	Name  string          `json:"name"`
}

func CleanText(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	s = ansiRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func extractText(raw json.RawMessage, includeToolContent bool) string {
	if len(raw) == 0 {
		return ""
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}

	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var parts []string
		for _, b := range blocks {
			switch b.Type {
			case "text":
				if t := strings.TrimSpace(b.Text); t != "" {
					parts = append(parts, t)
				}
			case "tool_use":
				if includeToolContent && len(b.Input) > 0 {
					parts = append(parts, string(b.Input))
				}
			case "tool_result":
				if includeToolContent {
					if t := strings.TrimSpace(b.Text); t != "" {
						parts = append(parts, t)
					}
				}
			}
		}
		return strings.Join(parts, " ")
	}
	return ""
}

func parseTimestamp(s string) time.Time {
	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Local()
		}
	}
	return time.Time{}
}

func Parse(path string, opts ParseOptions) (*Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := &Session{
		ID:   strings.TrimSuffix(filepath.Base(path), ".jsonl"),
		Path: path,
	}

	var firstUserMsg string
	var firstTimestamp time.Time
	var lastUserMsg string
	var lastTimestamp time.Time
	lineNum := 0

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()

		var entry jsonEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		if entry.CWD != "" && s.CWD == "" {
			s.CWD = entry.CWD
		}
		if entry.Entrypoint != "" && s.Entrypoint == "" {
			s.Entrypoint = entry.Entrypoint
		}

		if entry.Type != "user" && entry.Type != "assistant" {
			continue
		}

		var env messageEnvelope
		if err := json.Unmarshal(entry.Message, &env); err != nil {
			continue
		}

		text := extractText(env.Content, opts.IncludeToolContent)
		if text == "" {
			continue
		}

		ts := parseTimestamp(entry.Timestamp)

		if entry.Type == "user" {
			if strings.TrimSpace(text) == "/clear" {
				firstUserMsg = ""
				firstTimestamp = time.Time{}
				if !opts.MetadataOnly {
					s.Messages = nil
				}
				continue
			}
			if firstUserMsg == "" {
				firstUserMsg = text
				firstTimestamp = ts
			}
			lastUserMsg = text
			lastTimestamp = ts
		}

		if !opts.MetadataOnly {
			s.Messages = append(s.Messages, Message{
				Role:      entry.Type,
				Text:      text,
				Timestamp: ts,
				LineNum:   lineNum,
			})
		} else if entry.Type == "assistant" && !ts.IsZero() {
			lastTimestamp = ts
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	if lastUserMsg == "" {
		return nil, nil
	}

	s.FirstMessage = truncate(CleanText(firstUserMsg), 120)
	s.FirstTime = firstTimestamp
	s.LastMessage = truncate(CleanText(lastUserMsg), 120)
	s.LastTime = lastTimestamp

	if s.CWD != "" {
		s.ProjectDir = s.CWD
	} else {
		dirName := filepath.Base(filepath.Dir(path))
		s.ProjectDir = decodeDirName(dirName)
	}

	return s, nil
}

const (
	headReadSize = 50 * 1024
	tailReadSize = 64 * 1024
)

func ParseFast(path string) (*Session, error) {
	s := &Session{
		ID:   strings.TrimSuffix(filepath.Base(path), ".jsonl"),
		Path: path,
	}

	firstMsg, firstTime, cwd, entrypoint, err := parseHead(path)
	if err != nil {
		return nil, err
	}
	if firstMsg == "" {
		return nil, nil
	}

	s.FirstMessage = truncate(CleanText(firstMsg), 120)
	s.FirstTime = firstTime
	s.CWD = cwd
	s.Entrypoint = entrypoint

	lastTime, lastMsg, err := parseTail(path)
	if err == nil && !lastTime.IsZero() {
		s.LastTime = lastTime
	} else {
		s.LastTime = firstTime
	}
	if lastMsg != "" {
		s.LastMessage = truncate(CleanText(lastMsg), 120)
	} else {
		s.LastMessage = s.FirstMessage
	}

	if s.CWD != "" {
		s.ProjectDir = s.CWD
	} else {
		dirName := filepath.Base(filepath.Dir(path))
		s.ProjectDir = decodeDirName(dirName)
	}

	return s, nil
}

func parseHead(path string) (firstMsg string, firstTime time.Time, cwd string, entrypoint string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", time.Time{}, "", "", err
	}
	defer f.Close()

	buf := make([]byte, headReadSize)
	n, _ := f.Read(buf)
	if n == 0 {
		return "", time.Time{}, "", "", nil
	}

	lines := strings.Split(string(buf[:n]), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var entry jsonEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.CWD != "" && cwd == "" {
			cwd = entry.CWD
		}
		if entry.Entrypoint != "" && entrypoint == "" {
			entrypoint = entry.Entrypoint
		}

		if entry.Type != "user" {
			continue
		}

		var env messageEnvelope
		if err := json.Unmarshal(entry.Message, &env); err != nil {
			continue
		}
		text := extractText(env.Content, false)
		if text == "" || strings.TrimSpace(text) == "/clear" {
			continue
		}

		return text, parseTimestamp(entry.Timestamp), cwd, entrypoint, nil
	}

	return "", time.Time{}, cwd, entrypoint, nil
}

func parseTail(path string) (lastTime time.Time, lastMsg string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, "", err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return time.Time{}, "", err
	}

	readSize := int64(tailReadSize)
	if info.Size() < readSize {
		readSize = info.Size()
	}

	_, err = f.Seek(-readSize, 2)
	if err != nil {
		return time.Time{}, "", err
	}

	buf := make([]byte, readSize)
	n, _ := f.Read(buf)
	if n == 0 {
		return time.Time{}, "", nil
	}

	lines := strings.Split(string(buf[:n]), "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if line == "" {
			continue
		}
		var entry jsonEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		ts := parseTimestamp(entry.Timestamp)

		if lastTime.IsZero() && !ts.IsZero() {
			lastTime = ts
		}

		if (entry.Type == "user" || entry.Type == "assistant") && lastMsg == "" {
			var env messageEnvelope
			if err := json.Unmarshal(entry.Message, &env); err != nil {
				continue
			}
			text := extractText(env.Content, false)
			if text != "" && strings.TrimSpace(text) != "/clear" {
				lastMsg = text
			}
		}

		if !lastTime.IsZero() && lastMsg != "" {
			break
		}
	}

	return lastTime, lastMsg, nil
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

func decodeDirName(name string) string {
	name = strings.TrimPrefix(name, "-")
	decoded := "/" + strings.ReplaceAll(name, "-", "/")
	for strings.Contains(decoded, "//") {
		decoded = strings.ReplaceAll(decoded, "//", "/")
	}
	return decoded
}
