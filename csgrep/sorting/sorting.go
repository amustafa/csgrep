package sorting

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/amustafa/csgrep/search"
)

var validSortFields = map[string]bool{
	"timestamp": true,
	"score":     true,
}

var validGroupByFields = map[string]bool{
	"session_id":  true,
	"project_dir": true,
	"role":        true,
}

type SortField struct {
	Field string
	Desc  bool
}

type SortConfig struct {
	Fields  []SortField
	GroupBy string
	NoGroup bool
}

type Group struct {
	Key       string
	GroupBy   string
	RankValue any
	Matches   []search.Match
}

func ParseSort(raw string, isFuzzy bool) ([]SortField, error) {
	var fields []SortField
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		pieces := strings.SplitN(part, ":", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("invalid sort field %q (expected field:dir, e.g. timestamp:asc)", part)
		}
		field, dir := pieces[0], pieces[1]
		if !validSortFields[field] {
			names := sortedKeys(validSortFields)
			return nil, fmt.Errorf("unknown sort field %q (valid: %s)", field, strings.Join(names, ", "))
		}
		if field == "score" && !isFuzzy {
			return nil, fmt.Errorf("--sort score requires --fuzzy")
		}
		var desc bool
		switch dir {
		case "asc":
			desc = false
		case "desc":
			desc = true
		default:
			return nil, fmt.Errorf("invalid sort direction %q (valid: asc, desc)", dir)
		}
		fields = append(fields, SortField{Field: field, Desc: desc})
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("--sort requires at least one field:dir pair")
	}
	return fields, nil
}

func ParseGroupBy(groupBy string) error {
	if !validGroupByFields[groupBy] {
		names := sortedKeys(validGroupByFields)
		return fmt.Errorf("unknown group-by field %q (valid: %s)", groupBy, strings.Join(names, ", "))
	}
	return nil
}

func DefaultSortFields(isFuzzy bool) []SortField {
	if isFuzzy {
		return []SortField{{Field: "score", Desc: true}}
	}
	return []SortField{{Field: "timestamp", Desc: true}}
}

func SortMatches(matches []search.Match, fields []SortField) {
	sort.SliceStable(matches, func(i, j int) bool {
		for _, f := range fields {
			cmp := compareField(matches[i], matches[j], f.Field)
			if cmp == 0 {
				continue
			}
			if f.Desc {
				return cmp > 0
			}
			return cmp < 0
		}
		return false
	})
}

func GroupAndSort(matches []search.Match, cfg SortConfig) []Group {
	if cfg.NoGroup {
		SortMatches(matches, cfg.Fields)
		return nil
	}

	groupOrder := make([]string, 0)
	groupMap := make(map[string][]search.Match)
	for _, m := range matches {
		key := groupKeyValue(m, cfg.GroupBy)
		if _, exists := groupMap[key]; !exists {
			groupOrder = append(groupOrder, key)
		}
		groupMap[key] = append(groupMap[key], m)
	}

	groups := make([]Group, 0, len(groupOrder))
	for _, key := range groupOrder {
		gMatches := groupMap[key]
		SortMatches(gMatches, cfg.Fields)
		rank := bestValue(gMatches, cfg.Fields[0])
		groups = append(groups, Group{
			Key:       key,
			GroupBy:   cfg.GroupBy,
			RankValue: rank,
			Matches:   gMatches,
		})
	}

	sortGroups(groups, cfg.Fields[0])
	return groups
}

func compareField(a, b search.Match, field string) int {
	switch field {
	case "timestamp":
		ta, tb := a.Message.Timestamp, b.Message.Timestamp
		if ta.Equal(tb) {
			return 0
		}
		if ta.Before(tb) {
			return -1
		}
		return 1
	case "score":
		sa, sb := a.Score, b.Score
		if sa == sb {
			return 0
		}
		if sa < sb {
			return -1
		}
		return 1
	}
	return 0
}

func groupKeyValue(m search.Match, field string) string {
	switch field {
	case "session_id":
		return m.Session.ID
	case "project_dir":
		return m.Session.ProjectDir
	case "role":
		return m.Message.Role
	}
	return ""
}

func bestValue(matches []search.Match, primary SortField) any {
	if len(matches) == 0 {
		return nil
	}
	best := matches[0]
	for _, m := range matches[1:] {
		cmp := compareField(m, best, primary.Field)
		if primary.Desc && cmp > 0 {
			best = m
		} else if !primary.Desc && cmp < 0 {
			best = m
		}
	}
	switch primary.Field {
	case "timestamp":
		return best.Message.Timestamp.Format(time.RFC3339)
	case "score":
		return best.Score
	}
	return nil
}

func sortGroups(groups []Group, primary SortField) {
	sort.SliceStable(groups, func(i, j int) bool {
		cmp := compareRankValues(groups[i].RankValue, groups[j].RankValue, primary.Field)
		if primary.Desc {
			return cmp > 0
		}
		return cmp < 0
	})
}

func compareRankValues(a, b any, field string) int {
	switch field {
	case "timestamp":
		ta, _ := time.Parse(time.RFC3339, a.(string))
		tb, _ := time.Parse(time.RFC3339, b.(string))
		if ta.Equal(tb) {
			return 0
		}
		if ta.Before(tb) {
			return -1
		}
		return 1
	case "score":
		sa, sb := a.(float64), b.(float64)
		if sa == sb {
			return 0
		}
		if sa < sb {
			return -1
		}
		return 1
	}
	return 0
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
