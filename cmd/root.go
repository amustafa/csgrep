package cmd

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/amustafa/csgrep/include"
	"github.com/amustafa/csgrep/output"
	"github.com/amustafa/csgrep/pipe"
	"github.com/amustafa/csgrep/search"
	"github.com/amustafa/csgrep/session"
	"github.com/amustafa/csgrep/sorting"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	flagDir         string
	flagInteractive bool
	flagAfter       string
	flagBefore      string
	flagLimit       int
	flagFixed       bool
	flagFuzzy       bool
	flagIgnoreCase  bool
	flagCaseSens    bool
	flagContextC    int
	flagContextA    int
	flagContextB    int
	flagAll         bool
	flagJSON        bool
	flagNoColor     bool
	flagShowPath    bool
	flagThreshold   float64
	flagGlobal      bool
	flagSort        string
	flagGroupBy     string
	flagNoGroupBy   bool
	flagSessions    bool
	flagInclude     string
)

var rootCmd = &cobra.Command{
	Use:   "csgrep [pattern]",
	Short: "Grep for Claude Code sessions",
	Long: `csgrep is a fast search tool for Claude Code session transcripts.

It searches through your Claude Code session history stored in
~/.claude/projects/, supporting regex, fixed-string, and fuzzy matching
with ripgrep-inspired defaults and output formatting.

By default, csgrep scopes to sessions from the current working directory.
Use -g/--global to search across all projects, or -d to target a
specific directory.`,
	Example: `  csgrep "database migration"              Search current project (regex, smart-case)
  csgrep "auth" -g                         Search across all projects
  csgrep "auth" -d ~/myapp                 Search a specific project
  csgrep -F "exact phrase"                 Fixed-string (literal) search
  csgrep -f "databse migrtion"             Fuzzy search (tolerates typos)
  csgrep "error" --after 3d -n 10          Recent sessions, limit results
  csgrep "TODO" --all                      Include tool call content
  csgrep "auth" -C 2                       Show 2 messages of context
  csgrep "config" --json                   Machine-readable JSON output`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagDir, "dir", "d", "", "filter to sessions from this directory (absolute path or substring)")
	rootCmd.PersistentFlags().BoolVarP(&flagGlobal, "global", "g", false, "search all projects instead of current directory")
	rootCmd.PersistentFlags().BoolVar(&flagInteractive, "interactive", false, "only show interactive CLI sessions")
	rootCmd.PersistentFlags().StringVar(&flagAfter, "after", "", "sessions after this date (YYYY-MM-DD, or relative: 3d, 1w, 2h)")
	rootCmd.PersistentFlags().StringVar(&flagBefore, "before", "", "sessions before this date (YYYY-MM-DD, or relative: 3d, 1w)")
	rootCmd.PersistentFlags().IntVarP(&flagLimit, "limit", "n", 0, "show only the N most recent results")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&flagShowPath, "path", false, "show full file path to session JSONL")

	rootCmd.Flags().BoolVarP(&flagFixed, "fixed-strings", "F", false, "treat pattern as a literal string")
	rootCmd.Flags().BoolVarP(&flagFuzzy, "fuzzy", "f", false, "use fuzzy (trigram) matching")
	rootCmd.Flags().BoolVarP(&flagIgnoreCase, "ignore-case", "i", false, "force case-insensitive matching")
	rootCmd.Flags().BoolVarP(&flagCaseSens, "case-sensitive", "s", false, "force case-sensitive matching")
	rootCmd.Flags().IntVarP(&flagContextC, "context", "C", 0, "show N messages of context around each match")
	rootCmd.Flags().IntVarP(&flagContextA, "after-context", "A", 0, "show N messages after each match")
	rootCmd.Flags().IntVarP(&flagContextB, "before-context", "B", 0, "show N messages before each match")
	rootCmd.Flags().BoolVarP(&flagAll, "all", "a", false, "include tool call/result content in search")
	rootCmd.Flags().Float64Var(&flagThreshold, "threshold", 0.3, "fuzzy match threshold (0.0-1.0)")
	rootCmd.Flags().BoolVar(&flagSessions, "sessions", false, "pipe mode: use piped input as session scope (search all messages)")
	rootCmd.Flags().StringVar(&flagInclude, "include", "", "content types to include: artifacts, artifacts:path, artifacts:content, artifacts:all, artifacts:tmp, artifacts:plans, tool-outputs")

	rootCmd.Flags().StringVar(&flagSort, "sort", "", "sort results by field:dir pairs (e.g. timestamp:asc,score:desc)")
	rootCmd.Flags().StringVar(&flagGroupBy, "group-by", "session_id", "group results by field (session_id, project_dir, role)")
	rootCmd.Flags().BoolVar(&flagNoGroupBy, "no-group-by", false, "disable grouping, output flat results")
}

func applyColorConfig() {
	if flagNoColor {
		color.NoColor = true
	}
}

func parseRelativeDate(s string) (time.Time, error) {
	now := time.Now()
	if len(s) < 2 {
		return time.Time{}, fmt.Errorf("invalid relative date: %s", s)
	}
	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	var n int
	if _, err := fmt.Sscanf(numStr, "%d", &n); err != nil {
		return time.Time{}, fmt.Errorf("invalid relative date: %s", s)
	}
	switch unit {
	case 'h':
		return now.Add(-time.Duration(n) * time.Hour), nil
	case 'd':
		return now.AddDate(0, 0, -n), nil
	case 'w':
		return now.AddDate(0, 0, -n*7), nil
	case 'm':
		return now.AddDate(0, -n, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unknown time unit '%c' (use h, d, w, m)", unit)
	}
}

func parseDate(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return parseRelativeDate(s)
}

func buildFilter() session.Filter {
	dir := flagDir
	if dir == "" && !flagGlobal {
		if cwd, err := os.Getwd(); err == nil {
			dir = cwd
		}
	}
	f := session.Filter{
		Dir:         dir,
		Interactive: flagInteractive,
		Limit:       flagLimit,
	}
	if flagAfter != "" {
		if t, err := parseDate(flagAfter); err == nil {
			f.After = t
		} else {
			fmt.Fprintf(os.Stderr, "warning: could not parse --after %q: %v\n", flagAfter, err)
		}
	}
	if flagBefore != "" {
		if t, err := parseDate(flagBefore); err == nil {
			f.Before = t
		} else {
			fmt.Fprintf(os.Stderr, "warning: could not parse --before %q: %v\n", flagBefore, err)
		}
	}
	return f
}

func buildMatcher(pattern string) (search.Matcher, error) {
	if flagFuzzy {
		return search.NewFuzzyMatcher(pattern, flagThreshold), nil
	}
	if flagFixed {
		caseSensitive := flagCaseSens || (!flagIgnoreCase && hasUpperCase(pattern))
		return search.NewFixedMatcher(pattern, caseSensitive), nil
	}
	caseSensitive := flagCaseSens || (!flagIgnoreCase && hasUpperCase(pattern))
	return search.NewRegexMatcher(pattern, caseSensitive)
}

func hasUpperCase(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func buildIncludeSet() (include.IncludeSet, error) {
	if flagAll && flagInclude != "" {
		return include.IncludeSet{}, fmt.Errorf("cannot use --all and --include together")
	}
	if flagAll {
		return include.FromAll(), nil
	}
	if flagInclude != "" {
		return include.Parse(flagInclude)
	}
	return include.IncludeSet{}, nil
}

func runSearch(cmd *cobra.Command, args []string) error {
	applyColorConfig()
	pattern := args[0]

	if flagNoGroupBy && cmd.Flags().Changed("group-by") {
		return fmt.Errorf("cannot use --group-by and --no-group-by together")
	}

	var sortFields []sorting.SortField
	if flagSort != "" {
		var err error
		sortFields, err = sorting.ParseSort(flagSort, flagFuzzy)
		if err != nil {
			return err
		}
	} else {
		sortFields = sorting.DefaultSortFields(flagFuzzy)
	}

	if !flagNoGroupBy {
		if err := sorting.ParseGroupBy(flagGroupBy); err != nil {
			return err
		}
	}

	matcher, err := buildMatcher(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	useJSON := flagJSON || pipe.StdoutIsPiped()

	inc, err := buildIncludeSet()
	if err != nil {
		return err
	}

	var matches []search.Match
	if pipe.StdinIsPiped() {
		matches, err = searchFromPipe(matcher, inc)
		if err != nil {
			return err
		}
	} else {
		filter := buildFilter()
		files := session.FindFiles(filter)

		if os.Getenv("CSGREP_USE_RG") == "1" && !flagFuzzy && session.RgAvailable() {
			before := len(files)
			files = session.FilterWithRg(files, pattern)
			fmt.Fprintf(os.Stderr, "Searching %d sessions (rg filtered %d → %d)...\n", len(files), before, len(files))
		} else {
			fmt.Fprintf(os.Stderr, "Searching %d sessions...\n", len(files))
		}

		opts := search.Options{
			Include: inc,
			Workers: runtime.NumCPU(),
		}
		matches = search.Run(files, matcher, opts)
	}

	cfg := sorting.SortConfig{
		Fields:  sortFields,
		GroupBy: flagGroupBy,
		NoGroup: flagNoGroupBy,
	}

	contextBefore := flagContextB
	contextAfter := flagContextA
	if flagContextC > 0 {
		contextBefore = flagContextC
		contextAfter = flagContextC
	}

	if flagNoGroupBy {
		sorting.SortMatches(matches, sortFields)
		if flagLimit > 0 && len(matches) > flagLimit {
			matches = matches[:flagLimit]
		}
		if useJSON {
			return output.JSON(os.Stdout, matches)
		}
		output.Terminal(os.Stdout, matches, output.Config{
			ShowPath:      flagShowPath,
			ContextBefore: contextBefore,
			ContextAfter:  contextAfter,
			NoGroup:       true,
		})
	} else {
		groups := sorting.GroupAndSort(matches, cfg)
		if flagLimit > 0 {
			groups = limitGroups(groups, flagLimit)
		}
		if useJSON {
			return output.JSONGrouped(os.Stdout, groups)
		}
		output.TerminalGrouped(os.Stdout, groups, output.Config{
			ShowPath:      flagShowPath,
			ContextBefore: contextBefore,
			ContextAfter:  contextAfter,
		})
	}

	fmt.Fprintf(os.Stderr, "%d matches across %d sessions\n", len(matches), countSessions(matches))
	return nil
}

func searchFromPipe(matcher search.Matcher, inc include.IncludeSet) ([]search.Match, error) {
	pipedMatches, pipedSessions, err := pipe.ReadStdin()
	if err != nil {
		return nil, fmt.Errorf("reading pipe: %w", err)
	}

	if pipedMatches != nil && !flagSessions {
		fmt.Fprintf(os.Stderr, "Filtering %d piped matches...\n", len(pipedMatches))
		incoming := pipe.MatchesToSearchMatches(pipedMatches)
		var filtered []search.Match
		for _, m := range incoming {
			matched, offsets, score := matcher.Match(m.Message.Text)
			if matched {
				filtered = append(filtered, search.Match{
					Session:      m.Session,
					Message:      m.Message,
					MessageIndex: m.MessageIndex,
					Score:        score,
					Offsets:      offsets,
				})
			}
		}
		return filtered, nil
	}

	var files []string
	if pipedSessions != nil {
		files = pipe.SessionPaths(pipedSessions)
	} else if pipedMatches != nil {
		files = pipe.MatchPaths(pipedMatches)
	}

	if len(files) == 0 {
		return nil, nil
	}

	fmt.Fprintf(os.Stderr, "Searching %d sessions from pipe...\n", len(files))
	opts := search.Options{
		Include: inc,
		Workers: runtime.NumCPU(),
	}
	return search.Run(files, matcher, opts), nil
}

func limitGroups(groups []sorting.Group, limit int) []sorting.Group {
	total := 0
	for i, g := range groups {
		total += len(g.Matches)
		if total >= limit {
			g.Matches = g.Matches[:len(g.Matches)-(total-limit)]
			return groups[:i+1]
		}
	}
	return groups
}

func countSessions(matches []search.Match) int {
	seen := make(map[string]bool)
	for _, m := range matches {
		seen[m.Session.ID] = true
	}
	return len(seen)
}

