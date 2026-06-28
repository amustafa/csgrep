package pipe

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/amustafa/csgrep/search"
	"github.com/amustafa/csgrep/session"
)

type PipedMatch struct {
	SessionID  string   `json:"session_id"`
	ProjectDir string   `json:"project_dir"`
	Timestamp  string   `json:"timestamp"`
	Role       string   `json:"role"`
	Text       string   `json:"text"`
	LineNum    int      `json:"line_num"`
	Score      float64  `json:"score"`
	Offsets    [][2]int `json:"offsets,omitempty"`
	Path       string   `json:"path,omitempty"`
}

type PipedSession struct {
	SessionID      string `json:"session_id"`
	ProjectDir     string `json:"project_dir"`
	FirstTimestamp string `json:"first_timestamp"`
	LastTimestamp   string `json:"last_timestamp"`
	FirstMessage   string `json:"first_message"`
	LastMessage    string `json:"last_message"`
	Entrypoint     string `json:"entrypoint,omitempty"`
	Path           string `json:"path"`
}

func StdinIsPiped() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}

func StdoutIsPiped() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}

type PipedGroup struct {
	GroupKey string       `json:"group_key"`
	GroupBy  string       `json:"group_by"`
	Matches []PipedMatch `json:"matches"`
}

func ReadStdin() (matches []PipedMatch, sessions []PipedSession, err error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, nil, err
	}
	if len(data) == 0 {
		return nil, nil, nil
	}

	if err := json.Unmarshal(data, &matches); err == nil && len(matches) > 0 {
		if matches[0].Role != "" {
			return matches, nil, nil
		}
	}
	matches = nil

	var groups []PipedGroup
	if err := json.Unmarshal(data, &groups); err == nil && len(groups) > 0 {
		if len(groups[0].Matches) > 0 {
			for _, g := range groups {
				matches = append(matches, g.Matches...)
			}
			return matches, nil, nil
		}
	}
	matches = nil

	if err := json.Unmarshal(data, &sessions); err == nil && len(sessions) > 0 {
		if sessions[0].SessionID != "" {
			return nil, sessions, nil
		}
	}

	return nil, nil, nil
}

func MatchesToSearchMatches(piped []PipedMatch) []search.Match {
	sessionMap := make(map[string]*session.Session)
	var results []search.Match

	for _, pm := range piped {
		s, exists := sessionMap[pm.SessionID]
		if !exists {
			s = &session.Session{
				ID:         pm.SessionID,
				ProjectDir: pm.ProjectDir,
				Path:       pm.Path,
			}
			sessionMap[pm.SessionID] = s
		}

		ts, _ := time.Parse("2006-01-02T15:04:05Z", pm.Timestamp)
		results = append(results, search.Match{
			Session: s,
			Message: session.Message{
				Role:      pm.Role,
				Text:      pm.Text,
				Timestamp: ts,
				LineNum:   pm.LineNum,
			},
			Score:   pm.Score,
			Offsets: pm.Offsets,
		})
	}
	return results
}

func SessionIDs(sessions []PipedSession) []string {
	seen := make(map[string]bool)
	var ids []string
	for _, s := range sessions {
		if !seen[s.SessionID] {
			seen[s.SessionID] = true
			ids = append(ids, s.SessionID)
		}
	}
	return ids
}

func SessionIDsFromMatches(matches []PipedMatch) []string {
	seen := make(map[string]bool)
	var ids []string
	for _, m := range matches {
		if !seen[m.SessionID] {
			seen[m.SessionID] = true
			ids = append(ids, m.SessionID)
		}
	}
	return ids
}

func SessionPaths(sessions []PipedSession) []string {
	seen := make(map[string]bool)
	var paths []string
	for _, s := range sessions {
		if s.Path != "" && !seen[s.Path] {
			seen[s.Path] = true
			paths = append(paths, s.Path)
		}
	}
	return paths
}

func MatchPaths(matches []PipedMatch) []string {
	seen := make(map[string]bool)
	var paths []string
	for _, m := range matches {
		if m.Path != "" && !seen[m.Path] {
			seen[m.Path] = true
			paths = append(paths, m.Path)
		}
	}
	return paths
}
