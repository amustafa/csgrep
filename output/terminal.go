package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/amustafa/csgrep/search"
	"github.com/amustafa/csgrep/session"
	"github.com/amustafa/csgrep/sorting"
	"github.com/fatih/color"
)

type Config struct {
	ShowPath      bool
	ShowArtifacts bool
	ContextBefore int
	ContextAfter  int
	NoGroup       bool
}

var (
	sessionColor = color.New(color.FgMagenta, color.Bold)
	timeColor    = color.New(color.FgGreen)
	dirColor     = color.New(color.FgCyan)
	userColor     = color.New(color.FgBlue, color.Bold)
	assistColor   = color.New(color.FgYellow)
	artifactColor = color.New(color.FgGreen, color.Bold)
	toolOutColor  = color.New(color.FgCyan)
	lineNumColor  = color.New(color.FgGreen)
	dimColor      = color.New(color.Faint)
)

func Terminal(w io.Writer, matches []search.Match, cfg Config) {
	if cfg.NoGroup {
		for _, m := range matches {
			printMatch(w, m, cfg)
		}
		return
	}
	hasContext := cfg.ContextBefore > 0 || cfg.ContextAfter > 0
	grouped := groupBySession(matches)
	for i, group := range grouped {
		if i > 0 {
			fmt.Fprintln(w)
		}
		printSessionHeader(w, group[0].Session, cfg)

		if !hasContext {
			for _, m := range group {
				printMatch(w, m, cfg)
			}
			continue
		}

		printMatchesWithContext(w, group, cfg)
	}
}

func TerminalLive(w io.Writer, sessions []session.LiveSession, cfg Config) {
	activeColor := color.New(color.FgGreen, color.Bold)
	pidColor := color.New(color.FgCyan)

	for _, s := range sessions {
		dot := activeColor.Sprint("●")
		ts := ""
		if !s.LastTime.IsZero() {
			ts = s.LastTime.Format("2006-01-02 15:04")
		}
		fmt.Fprintf(w, "%s %s %s %s\n",
			dot,
			sessionColor.Sprint(s.ID),
			timeColor.Sprintf("(%s)", ts),
			dirColor.Sprint(s.ProjectDir),
		)
		fmt.Fprintf(w, "  %s %s\n", dimColor.Sprint("pid:"), pidColor.Sprintf("%d", s.PID))
		if cfg.ShowPath {
			fmt.Fprintf(w, "  %s %s\n", dimColor.Sprint("path:"), s.Path)
		}
		firstTs := ""
		if !s.FirstTime.IsZero() {
			firstTs = timeColor.Sprintf(" (%s)", s.FirstTime.Format("2006-01-02 15:04"))
		}
		lastTs := ""
		if !s.LastTime.IsZero() {
			lastTs = timeColor.Sprintf(" (%s)", s.LastTime.Format("2006-01-02 15:04"))
		}
		fmt.Fprintf(w, "  %s%s %s\n", dimColor.Sprint("first:"), firstTs, truncateText(s.FirstMessage, 100))
		fmt.Fprintf(w, "  %s %s %s\n", dimColor.Sprint("last:"), lastTs, truncateText(s.LastMessage, 100))
		fmt.Fprintln(w)
	}
}

func TerminalGrouped(w io.Writer, groups []sorting.Group, cfg Config) {
	hasContext := cfg.ContextBefore > 0 || cfg.ContextAfter > 0
	for i, g := range groups {
		if i > 0 {
			fmt.Fprintln(w)
		}
		if len(g.Matches) == 0 {
			continue
		}
		printSessionHeader(w, g.Matches[0].Session, cfg)

		if !hasContext {
			for _, m := range g.Matches {
				printMatch(w, m, cfg)
			}
			continue
		}

		printMatchesWithContext(w, g.Matches, cfg)
	}
}

func TerminalSessions(w io.Writer, sessions []session.Session, cfg Config) {
	for _, s := range sessions {
		printSessionHeader(w, &s, cfg)
		firstTs := ""
		if !s.FirstTime.IsZero() {
			firstTs = timeColor.Sprintf(" (%s)", s.FirstTime.Format("2006-01-02 15:04"))
		}
		lastTs := ""
		if !s.LastTime.IsZero() {
			lastTs = timeColor.Sprintf(" (%s)", s.LastTime.Format("2006-01-02 15:04"))
		}
		fmt.Fprintf(w, "  %s%s %s\n", dimColor.Sprint("first:"), firstTs, truncateText(s.FirstMessage, 100))
		fmt.Fprintf(w, "  %s %s %s\n", dimColor.Sprint("last:"), lastTs, truncateText(s.LastMessage, 100))
		if cfg.ShowArtifacts && len(s.ArtifactPaths) > 0 {
			paths := truncateList(s.ArtifactPaths, 5)
			fmt.Fprintf(w, "  %s %s (%s)\n",
				dimColor.Sprint("artifacts:"),
				artifactColor.Sprintf("%d files", len(s.ArtifactPaths)),
				dirColor.Sprint(strings.Join(paths, ", ")),
			)
		}
		fmt.Fprintln(w)
	}
}

func truncateList(items []string, max int) []string {
	if len(items) <= max {
		return items
	}
	result := make([]string, max)
	copy(result, items[:max])
	result = append(result, fmt.Sprintf("+%d more", len(items)-max))
	return result
}

func TerminalConversation(w io.Writer, s *session.Session, messages []session.Message, highlighter search.Matcher) {
	printSessionHeader(w, s, Config{})
	fmt.Fprintln(w)

	for _, msg := range messages {
		roleTag := formatRole(msg.Role)
		text := msg.Text
		if highlighter != nil {
			matched, offsets, _ := highlighter.Match(text)
			if matched {
				text = search.HighlightText(text, offsets, hiStart, hiEnd)
			}
		}

		ts := ""
		if !msg.Timestamp.IsZero() {
			ts = timeColor.Sprintf(" (%s)", msg.Timestamp.Format("15:04"))
		}

		fmt.Fprintf(w, "%s%s %s L%s\n",
			roleTag,
			ts,
			dimColor.Sprint("─"),
			lineNumColor.Sprintf("%d", msg.LineNum),
		)

		lines := strings.Split(text, "\n")
		for _, line := range lines {
			fmt.Fprintf(w, "  %s\n", line)
		}
		fmt.Fprintln(w)
	}
}

func printMatchesWithContext(w io.Writer, group []search.Match, cfg Config) {
	s := group[0].Session
	msgCount := len(s.Messages)
	if msgCount == 0 {
		return
	}

	matchIndices := make(map[int]search.Match)
	for _, m := range group {
		matchIndices[m.MessageIndex] = m
	}

	var spans []span
	for _, m := range group {
		lo := m.MessageIndex - cfg.ContextBefore
		hi := m.MessageIndex + cfg.ContextAfter
		if lo < 0 {
			lo = 0
		}
		if hi >= msgCount {
			hi = msgCount - 1
		}
		spans = append(spans, span{lo, hi})
	}

	merged := mergeSpans(spans)

	for si, sp := range merged {
		if si > 0 {
			fmt.Fprintf(w, "  %s\n", dimColor.Sprint("──"))
		}
		for idx := sp.start; idx <= sp.end; idx++ {
			msg := s.Messages[idx]
			if m, isMatch := matchIndices[idx]; isMatch {
				printMatch(w, m, cfg)
			} else {
				printContextLine(w, msg)
			}
		}
	}
}

func printContextLine(w io.Writer, msg session.Message) {
	text := session.CleanText(msg.Text)
	snippet := snippetAround(text, 120)
	role := fmt.Sprintf("[%-11s]", msg.Role)
	fmt.Fprintf(w, "  %s\n",
		dimColor.Sprintf("%s L%d: %s", role, msg.LineNum, snippet),
	)
}

type span struct{ start, end int }

func mergeSpans(spans []span) []span {
	if len(spans) == 0 {
		return nil
	}
	sorted := make([]span, len(spans))
	copy(sorted, spans)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j].start < sorted[j-1].start; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	merged := []span{sorted[0]}
	for _, s := range sorted[1:] {
		last := &merged[len(merged)-1]
		if s.start <= last.end+1 {
			if s.end > last.end {
				last.end = s.end
			}
		} else {
			merged = append(merged, s)
		}
	}
	return merged
}

func printSessionHeader(w io.Writer, s *session.Session, cfg Config) {
	ts := ""
	if !s.LastTime.IsZero() {
		ts = s.LastTime.Format("2006-01-02 15:04")
	}
	fmt.Fprintf(w, "%s %s %s\n",
		sessionColor.Sprint(s.ID),
		timeColor.Sprintf("(%s)", ts),
		dirColor.Sprint(s.ProjectDir),
	)
	if cfg.ShowPath {
		fmt.Fprintf(w, "  %s %s\n", dimColor.Sprint("path:"), s.Path)
	}
}

func printMatch(w io.Writer, m search.Match, cfg Config) {
	roleTag := formatRole(m.Message.Role)
	text := m.Message.Text

	snippet := snippetAround(text, 120)

	if len(m.Offsets) > 0 {
		snippet = highlightInText(text, m.Offsets, 120)
	}

	scoreStr := ""
	if m.Score < 1.0 {
		scoreStr = dimColor.Sprintf(" [%.2f]", m.Score)
	}

	filePrefix := ""
	if m.Message.FilePath != "" {
		filePrefix = dirColor.Sprint(m.Message.FilePath) + ": "
	}

	fmt.Fprintf(w, "  %s L%s: %s%s%s\n",
		roleTag,
		lineNumColor.Sprintf("%d", m.Message.LineNum),
		filePrefix,
		snippet,
		scoreStr,
	)
}

const (
	hiStart = "\033[1;31m"
	hiEnd   = "\033[0m"
)

func highlightInText(text string, offsets [][2]int, maxLen int) string {
	if maxLen > 0 && len([]rune(text)) > maxLen {
		runes := []rune(text)
		text = string(runes[:maxLen]) + "..."
		var valid [][2]int
		for _, o := range offsets {
			if o[0] < len(text) {
				end := o[1]
				if end > len(text) {
					end = len(text)
				}
				valid = append(valid, [2]int{o[0], end})
			}
		}
		offsets = valid
	}
	return search.HighlightText(text, offsets, hiStart, hiEnd)
}

func formatRole(role string) string {
	switch role {
	case "user":
		return userColor.Sprintf("[user]       ")
	case "assistant":
		return assistColor.Sprintf("[assist]     ")
	case "artifact":
		return artifactColor.Sprintf("[artifact]   ")
	case "tool-output":
		return toolOutColor.Sprintf("[tool-output]")
	default:
		return dimColor.Sprintf("[%-11s]", role)
	}
}

func snippetAround(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

func truncateText(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

func groupBySession(matches []search.Match) [][]search.Match {
	order := make([]string, 0)
	groups := make(map[string][]search.Match)
	for _, m := range matches {
		id := m.Session.ID
		if _, exists := groups[id]; !exists {
			order = append(order, id)
		}
		groups[id] = append(groups[id], m)
	}
	result := make([][]search.Match, len(order))
	for i, id := range order {
		result[i] = groups[id]
	}
	return result
}
