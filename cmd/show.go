package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/amustafa/csgrep/output"
	"github.com/amustafa/csgrep/pipe"
	"github.com/amustafa/csgrep/search"
	"github.com/amustafa/csgrep/session"
	"github.com/spf13/cobra"
)

var (
	flagRole string
)

var showCmd = &cobra.Command{
	Use:   "show [session-id] [pattern]",
	Short: "Display a full session conversation",
	Long: `Show the complete conversation for a Claude Code session.

Displays all messages in chronological order. An optional pattern
argument highlights matching text within the conversation. Output
is automatically piped through your $PAGER (or less) when it
exceeds the terminal height.

When receiving piped input, shows full conversations for all sessions
in the pipe, with matched terms highlighted.`,
	Example: `  csgrep show a1b2c3d4                     Show full conversation
  csgrep show a1b2c3d4 "auth"              Show with highlighted matches
  csgrep show a1b2c3d4 --role user         Only show user messages
  csgrep show a1b2c3d4 --all               Include tool call content
  csgrep "auth" | csgrep show              Show sessions from piped matches
  csgrep list -n 3 | csgrep show           Show sessions from piped list`,
	Args: cobra.RangeArgs(0, 2),
	RunE: runShow,
}

func init() {
	showCmd.Flags().StringVar(&flagRole, "role", "", "filter by message role (user, assistant)")
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
	applyColorConfig()

	if pipe.StdinIsPiped() {
		return runShowFromPipe(args)
	}

	if len(args) == 0 {
		return fmt.Errorf("session ID required (or pipe input from another csgrep command)")
	}
	return runShowDirect(args)
}

func runShowDirect(args []string) error {
	sessionID := args[0]

	var highlightPattern string
	if len(args) > 1 {
		highlightPattern = args[1]
	}

	file := session.FindByID(sessionID)
	if file == "" {
		return fmt.Errorf("session %q not found", sessionID)
	}

	s, err := parseSessionForShow(file)
	if err != nil {
		return err
	}

	messages := filterByRole(s.Messages)
	highlighter := buildHighlighter(highlightPattern)

	return renderConversation(s, messages, highlighter)
}

func runShowFromPipe(args []string) error {
	pipedMatches, pipedSessions, err := pipe.ReadStdin()
	if err != nil {
		return fmt.Errorf("reading pipe: %w", err)
	}

	var highlightPattern string
	if len(args) > 0 {
		highlightPattern = args[0]
	}

	var sessionIDs []string
	if pipedMatches != nil {
		sessionIDs = pipe.SessionIDsFromMatches(pipedMatches)
		if highlightPattern == "" {
			highlightPattern = inferPatternFromMatches(pipedMatches)
		}
	} else if pipedSessions != nil {
		sessionIDs = pipe.SessionIDs(pipedSessions)
	}

	if len(sessionIDs) == 0 {
		return fmt.Errorf("no sessions found in piped input")
	}

	highlighter := buildHighlighter(highlightPattern)

	fmt.Fprintf(os.Stderr, "Showing %d sessions...\n", len(sessionIDs))

	var allSessions []*session.Session
	var allMessages [][]session.Message
	for _, id := range sessionIDs {
		file := session.FindByID(id)
		if file == "" {
			continue
		}
		s, err := parseSessionForShow(file)
		if err != nil {
			continue
		}
		allSessions = append(allSessions, s)
		allMessages = append(allMessages, filterByRole(s.Messages))
	}

	if flagJSON {
		for _, i := range allSessions {
			idx := indexOf(allSessions, i)
			if err := output.JSONMessages(os.Stdout, i, allMessages[idx]); err != nil {
				return err
			}
		}
		return nil
	}

	return renderMultiConversation(allSessions, allMessages, highlighter)
}

func parseSessionForShow(file string) (*session.Session, error) {
	inc, err := buildIncludeSet()
	if err != nil {
		return nil, err
	}
	parseOpts := session.ParseOptions{
		Include: inc,
	}
	s, err := session.Parse(file, parseOpts)
	if err != nil {
		return nil, fmt.Errorf("parsing session: %w", err)
	}
	if s == nil {
		return nil, fmt.Errorf("session is empty")
	}
	return s, nil
}

func filterByRole(messages []session.Message) []session.Message {
	if flagRole == "" {
		return messages
	}
	var filtered []session.Message
	for _, m := range messages {
		if strings.EqualFold(m.Role, flagRole) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func buildHighlighter(pattern string) search.Matcher {
	if pattern == "" {
		return nil
	}
	caseSensitive := hasUpperCase(pattern)
	m, _ := search.NewRegexMatcher(pattern, caseSensitive)
	return m
}

func inferPatternFromMatches(matches []pipe.PipedMatch) string {
	return ""
}

func renderConversation(s *session.Session, messages []session.Message, highlighter search.Matcher) error {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}

	if !isTerminal() || flagJSON || flagNoColor {
		if flagJSON {
			return output.JSONMessages(os.Stdout, s, messages)
		}
		output.TerminalConversation(os.Stdout, s, messages, highlighter)
		return nil
	}

	pagerCmd := exec.Command(pager, "-R")
	pagerCmd.Stdout = os.Stdout
	pagerCmd.Stderr = os.Stderr
	w, err := pagerCmd.StdinPipe()
	if err != nil {
		output.TerminalConversation(os.Stdout, s, messages, highlighter)
		return nil
	}
	if err := pagerCmd.Start(); err != nil {
		output.TerminalConversation(os.Stdout, s, messages, highlighter)
		return nil
	}
	output.TerminalConversation(w, s, messages, highlighter)
	w.Close()
	return pagerCmd.Wait()
}

func renderMultiConversation(sessions []*session.Session, messages [][]session.Message, highlighter search.Matcher) error {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}

	renderAll := func(w *os.File) {
		for i, s := range sessions {
			if i > 0 {
				fmt.Fprintf(w, "\n%s\n\n", strings.Repeat("═", 60))
			}
			output.TerminalConversation(w, s, messages[i], highlighter)
		}
	}

	if !isTerminal() || flagNoColor {
		renderAll(os.Stdout)
		return nil
	}

	pagerCmd := exec.Command(pager, "-R")
	pagerCmd.Stdout = os.Stdout
	pagerCmd.Stderr = os.Stderr
	w, err := pagerCmd.StdinPipe()
	if err != nil {
		renderAll(os.Stdout)
		return nil
	}
	if err := pagerCmd.Start(); err != nil {
		renderAll(os.Stdout)
		return nil
	}
	for i, s := range sessions {
		if i > 0 {
			fmt.Fprintf(w, "\n%s\n\n", strings.Repeat("═", 60))
		}
		output.TerminalConversation(w, s, messages[i], highlighter)
	}
	w.Close()
	return pagerCmd.Wait()
}

func indexOf(sessions []*session.Session, s *session.Session) int {
	for i, ss := range sessions {
		if ss == s {
			return i
		}
	}
	return 0
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
