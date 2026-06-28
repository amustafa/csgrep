package output

import (
	"encoding/json"
	"io"

	"github.com/amustafa/csgrep/search"
	"github.com/amustafa/csgrep/session"
	"github.com/amustafa/csgrep/sorting"
)

type jsonMatch struct {
	SessionID  string   `json:"session_id"`
	ProjectDir string   `json:"project_dir"`
	Timestamp  string   `json:"timestamp"`
	Role       string   `json:"role"`
	Text       string   `json:"text"`
	LineNum    int      `json:"line_num"`
	Score      float64  `json:"score"`
	Offsets    [][2]int `json:"offsets,omitempty"`
	Path       string   `json:"path,omitempty"`
	FilePath   string   `json:"file_path,omitempty"`
	ToolName   string   `json:"tool_name,omitempty"`
}

type jsonSession struct {
	SessionID      string   `json:"session_id"`
	ProjectDir     string   `json:"project_dir"`
	FirstTimestamp string   `json:"first_timestamp"`
	LastTimestamp   string   `json:"last_timestamp"`
	FirstMessage   string   `json:"first_message"`
	LastMessage    string   `json:"last_message"`
	Entrypoint     string   `json:"entrypoint,omitempty"`
	Path           string   `json:"path"`
	ArtifactPaths  []string `json:"artifact_paths,omitempty"`
}

type jsonMessage struct {
	Role      string `json:"role"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
	LineNum   int    `json:"line_num"`
	FilePath  string `json:"file_path,omitempty"`
	ToolName  string `json:"tool_name,omitempty"`
}

type jsonConversation struct {
	SessionID  string        `json:"session_id"`
	ProjectDir string        `json:"project_dir"`
	Messages   []jsonMessage `json:"messages"`
}

func JSON(w io.Writer, matches []search.Match) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	items := make([]jsonMatch, 0, len(matches))
	for _, m := range matches {
		items = append(items, toJSONMatch(m))
	}
	return enc.Encode(items)
}

func toJSONMatch(m search.Match) jsonMatch {
	return jsonMatch{
		SessionID:  m.Session.ID,
		ProjectDir: m.Session.ProjectDir,
		Timestamp:  m.Message.Timestamp.Format("2006-01-02T15:04:05Z"),
		Role:       m.Message.Role,
		Text:       session.CleanText(m.Message.Text),
		LineNum:    m.Message.LineNum,
		Score:      m.Score,
		Offsets:    m.Offsets,
		Path:       m.Session.Path,
		FilePath:   m.Message.FilePath,
		ToolName:   m.Message.ToolName,
	}
}

func JSONSessions(w io.Writer, sessions []session.Session) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	items := make([]jsonSession, 0, len(sessions))
	for _, s := range sessions {
		firstTs := ""
		if !s.FirstTime.IsZero() {
			firstTs = s.FirstTime.Format("2006-01-02T15:04:05Z")
		}
		items = append(items, jsonSession{
			SessionID:      s.ID,
			ProjectDir:     s.ProjectDir,
			FirstTimestamp:  firstTs,
			LastTimestamp:   s.LastTime.Format("2006-01-02T15:04:05Z"),
			FirstMessage:   s.FirstMessage,
			LastMessage:    s.LastMessage,
			Entrypoint:     s.Entrypoint,
			Path:           s.Path,
			ArtifactPaths:  s.ArtifactPaths,
		})
	}
	return enc.Encode(items)
}

type jsonGroup struct {
	GroupKey  string      `json:"group_key"`
	GroupBy   string      `json:"group_by"`
	RankValue any         `json:"rank_value"`
	Matches   []jsonMatch `json:"matches"`
}

func JSONGrouped(w io.Writer, groups []sorting.Group) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	items := make([]jsonGroup, 0, len(groups))
	for _, g := range groups {
		matches := make([]jsonMatch, 0, len(g.Matches))
		for _, m := range g.Matches {
			matches = append(matches, toJSONMatch(m))
		}
		items = append(items, jsonGroup{
			GroupKey:  g.Key,
			GroupBy:   g.GroupBy,
			RankValue: g.RankValue,
			Matches:   matches,
		})
	}
	return enc.Encode(items)
}

func JSONMessages(w io.Writer, s *session.Session, messages []session.Message) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	conv := jsonConversation{
		SessionID:  s.ID,
		ProjectDir: s.ProjectDir,
	}
	for _, m := range messages {
		conv.Messages = append(conv.Messages, jsonMessage{
			Role:      m.Role,
			Text:      m.Text,
			Timestamp: m.Timestamp.Format("2006-01-02T15:04:05Z"),
			LineNum:   m.LineNum,
			FilePath:  m.FilePath,
			ToolName:  m.ToolName,
		})
	}
	return enc.Encode(conv)
}

type jsonLiveSession struct {
	SessionID      string `json:"session_id"`
	ProjectDir     string `json:"project_dir"`
	FirstTimestamp string `json:"first_timestamp,omitempty"`
	LastTimestamp   string `json:"last_timestamp,omitempty"`
	FirstMessage   string `json:"first_message"`
	LastMessage    string `json:"last_message"`
	Entrypoint     string `json:"entrypoint,omitempty"`
	Path           string `json:"path"`
	PID            int    `json:"pid"`
	Active         bool   `json:"active"`
}

func JSONLive(w io.Writer, sessions []session.LiveSession) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	items := make([]jsonLiveSession, 0, len(sessions))
	for _, s := range sessions {
		firstTs := ""
		if !s.FirstTime.IsZero() {
			firstTs = s.FirstTime.Format("2006-01-02T15:04:05Z")
		}
		lastTs := ""
		if !s.LastTime.IsZero() {
			lastTs = s.LastTime.Format("2006-01-02T15:04:05Z")
		}
		items = append(items, jsonLiveSession{
			SessionID:      s.ID,
			ProjectDir:     s.ProjectDir,
			FirstTimestamp:  firstTs,
			LastTimestamp:   lastTs,
			FirstMessage:   s.FirstMessage,
			LastMessage:    s.LastMessage,
			Entrypoint:     s.Entrypoint,
			Path:           s.Path,
			PID:            s.PID,
			Active:         s.Active,
		})
	}
	return enc.Encode(items)
}
