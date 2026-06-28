package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/amustafa/csgrep/output"
	"github.com/amustafa/csgrep/pipe"
	"github.com/amustafa/csgrep/session"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Claude Code sessions",
	Long: `Enumerate Claude Code sessions with metadata.

Displays session ID, timestamp, project directory, and the first user
message (after the last /clear). Results are sorted by most recent first.`,
	Example: `  csgrep list                              List sessions for current project
  csgrep list -g                           List all sessions across projects
  csgrep list -d ~/workspace/myapp         List sessions for a specific project
  csgrep list -d ftron                     Substring match on project dir
  csgrep list --interactive                Only interactive CLI sessions
  csgrep list --after 1w -n 10             Last week's sessions, top 10
  csgrep list --json                       Machine-readable output`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	applyColorConfig()
	filter := buildFilter()

	files := session.FindFiles(filter)
	fmt.Fprintf(os.Stderr, "Scanning %d sessions...\n", len(files))

	parseOpts := session.ParseOptions{
		MetadataOnly: true,
	}

	var sessions []session.Session
	for _, f := range files {
		s, err := session.Parse(f, parseOpts)
		if err != nil || s == nil {
			continue
		}
		if !filter.Matches(s) {
			continue
		}
		sessions = append(sessions, *s)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastTime.After(sessions[j].LastTime)
	})

	if flagLimit > 0 && len(sessions) > flagLimit {
		sessions = sessions[:flagLimit]
	}

	useJSON := flagJSON || pipe.StdoutIsPiped()
	if useJSON {
		return output.JSONSessions(os.Stdout, sessions)
	}
	output.TerminalSessions(os.Stdout, sessions, output.Config{
		ShowPath: flagShowPath,
	})

	fmt.Fprintf(os.Stderr, "%d sessions found\n", len(sessions))
	return nil
}
