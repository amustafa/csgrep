package include

import (
	"fmt"
	"sort"
	"strings"
)

type IncludeSet struct {
	Artifacts     bool
	ArtifactScope string // "" = non-temp (default), "all", "tmp", "plans"
	ArtifactMatch string // "" = path+content (default), "path", "content"
	ToolOutputs   bool
}

var validValues = map[string]bool{
	"artifacts":         true,
	"artifacts:path":    true,
	"artifacts:content": true,
	"artifacts:all":     true,
	"artifacts:tmp":     true,
	"artifacts:plans":   true,
	"tool-outputs":      true,
}

func Parse(raw string) (IncludeSet, error) {
	var s IncludeSet
	for _, token := range strings.Split(raw, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if !validValues[token] {
			return IncludeSet{}, fmt.Errorf("unknown --include value %q (valid: %s)", token, validList())
		}
		switch token {
		case "artifacts":
			s.Artifacts = true
		case "artifacts:path":
			s.Artifacts = true
			s.ArtifactMatch = "path"
		case "artifacts:content":
			s.Artifacts = true
			s.ArtifactMatch = "content"
		case "artifacts:all":
			s.Artifacts = true
			s.ArtifactScope = "all"
		case "artifacts:tmp":
			s.Artifacts = true
			s.ArtifactScope = "tmp"
		case "artifacts:plans":
			s.Artifacts = true
			s.ArtifactScope = "plans"
		case "tool-outputs":
			s.ToolOutputs = true
		}
	}
	return s, nil
}

func FromAll() IncludeSet {
	return IncludeSet{
		Artifacts:   true,
		ToolOutputs: true,
	}
}

func (s IncludeSet) IsEmpty() bool {
	return !s.Artifacts && !s.ToolOutputs
}

func (s IncludeSet) MatchesScope(filePath string) bool {
	isTemp := strings.HasPrefix(filePath, "/tmp/") || strings.HasPrefix(filePath, "/var/tmp/")
	isPlan := strings.Contains(filePath, "/.claude/plans/")

	switch s.ArtifactScope {
	case "all":
		return true
	case "tmp":
		return isTemp
	case "plans":
		return isPlan
	default:
		return !isTemp
	}
}

func validList() string {
	names := make([]string, 0, len(validValues))
	for k := range validValues {
		names = append(names, k)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func (s IncludeSet) ShouldIncludeToolContent() bool {
	return s.Artifacts || s.ToolOutputs
}
