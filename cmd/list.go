package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/amustafa/csgrep/include"
	"github.com/amustafa/csgrep/output"
	"github.com/amustafa/csgrep/pipe"
	"github.com/amustafa/csgrep/session"
	"github.com/spf13/cobra"
)

var flagHas string

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
  csgrep list --has artifacts              Only sessions that wrote files
  csgrep list --include artifacts          Show artifact summary per session
  csgrep list --json                       Machine-readable output`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVar(&flagHas, "has", "", "filter to sessions containing: artifacts")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	applyColorConfig()
	filter := buildFilter()

	needArtifacts := false
	if flagHas == "artifacts" {
		needArtifacts = true
	} else if flagHas != "" {
		return fmt.Errorf("unknown --has value %q (valid: artifacts)", flagHas)
	}

	var listInc include.IncludeSet
	if flagInclude != "" {
		var err error
		listInc, err = include.Parse(flagInclude)
		if err != nil {
			return err
		}
		if listInc.Artifacts {
			needArtifacts = true
		}
	}

	files := session.FindFiles(filter)
	fmt.Fprintf(os.Stderr, "Scanning %d sessions...\n", len(files))

	useFastPath := !needArtifacts && os.Getenv("CSGREP_NO_DEPS") != "1" && session.RgAvailable()

	var sessions []session.Session
	if useFastPath {
		clearFiles := session.FindClearFiles(files)
		for _, f := range files {
			var s *session.Session
			var err error
			if clearFiles[f] {
				s, err = session.Parse(f, session.ParseOptions{MetadataOnly: true})
			} else {
				s, err = session.ParseFast(f)
			}
			if err != nil || s == nil {
				continue
			}
			if !filter.Matches(s) {
				continue
			}
			sessions = append(sessions, *s)
		}
	} else {
		parseOpts := session.ParseOptions{
			MetadataOnly: !needArtifacts,
		}
		if needArtifacts {
			parseOpts.Include = include.IncludeSet{Artifacts: true}
		}
		for _, f := range files {
			s, err := session.Parse(f, parseOpts)
			if err != nil || s == nil {
				continue
			}
			if !filter.Matches(s) {
				continue
			}
			if flagHas == "artifacts" && len(s.ArtifactPaths) == 0 {
				continue
			}
			sessions = append(sessions, *s)
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastTime.After(sessions[j].LastTime)
	})

	if flagLimit > 0 && len(sessions) > flagLimit {
		sessions = sessions[:flagLimit]
	}

	showArtifacts := listInc.Artifacts

	useJSON := flagJSON || pipe.StdoutIsPiped()
	if useJSON {
		return output.JSONSessions(os.Stdout, sessions)
	}
	output.TerminalSessions(os.Stdout, sessions, output.Config{
		ShowPath:      flagShowPath,
		ShowArtifacts: showArtifacts,
	})

	fmt.Fprintf(os.Stderr, "%d sessions found\n", len(sessions))
	return nil
}
