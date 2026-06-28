package session

import (
	"testing"
	"time"
)

func TestFilterMatchesInteractive(t *testing.T) {
	f := Filter{Interactive: true}
	cli := &Session{Entrypoint: "cli"}
	agent := &Session{Entrypoint: "agent"}
	empty := &Session{Entrypoint: ""}

	if !f.Matches(cli) {
		t.Error("should match cli sessions")
	}
	if f.Matches(agent) {
		t.Error("should not match agent sessions")
	}
	if !f.Matches(empty) {
		t.Error("should match sessions with no entrypoint")
	}
}

func TestFilterMatchesDateRange(t *testing.T) {
	now := time.Now()
	f := Filter{
		After:  now.Add(-24 * time.Hour),
		Before: now.Add(24 * time.Hour),
	}

	recent := &Session{LastTime: now}
	old := &Session{LastTime: now.Add(-48 * time.Hour)}
	future := &Session{LastTime: now.Add(48 * time.Hour)}

	if !f.Matches(recent) {
		t.Error("should match session within range")
	}
	if f.Matches(old) {
		t.Error("should not match session before After")
	}
	if f.Matches(future) {
		t.Error("should not match session after Before")
	}
}

func TestFilterMatchesNoConstraints(t *testing.T) {
	f := Filter{}
	s := &Session{Entrypoint: "agent", LastTime: time.Now()}
	if !f.Matches(s) {
		t.Error("empty filter should match everything")
	}
}
