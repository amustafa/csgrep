package session

import "time"

type Message struct {
	Role      string
	Text      string
	Timestamp time.Time
	LineNum   int
}

type Session struct {
	ID           string
	Path         string
	ProjectDir   string
	CWD          string
	Entrypoint   string
	FirstMessage string
	FirstTime    time.Time
	LastMessage  string
	LastTime     time.Time
	Messages     []Message
}

type ParseOptions struct {
	MetadataOnly       bool
	IncludeToolContent bool
}

type Filter struct {
	Dir         string
	Interactive bool
	After       time.Time
	Before      time.Time
	Limit       int
}

func (f Filter) Matches(s *Session) bool {
	if f.Interactive {
		if s.Entrypoint != "" && s.Entrypoint != "cli" {
			return false
		}
	}
	if !f.After.IsZero() && s.LastTime.Before(f.After) {
		return false
	}
	if !f.Before.IsZero() && s.LastTime.After(f.Before) {
		return false
	}
	return true
}
