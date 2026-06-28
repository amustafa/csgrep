package cmd

import (
	"fmt"
	"os"

	"github.com/amustafa/csgrep/output"
	"github.com/amustafa/csgrep/pipe"
	"github.com/amustafa/csgrep/session"
	"github.com/spf13/cobra"
)

var liveCmd = &cobra.Command{
	Use:   "live",
	Short: "List active Claude Code sessions",
	Long: `Show Claude Code sessions attached to running processes.

Scans /proc for running claude processes and matches them to session
files. Shows session metadata with PID and active status indicators.`,
	Example: `  csgrep live                              List all active sessions
  csgrep live -d ftron                     Active sessions for a project
  csgrep live --json                       JSON output with PIDs
  csgrep live | csgrep "auth"              Search within active sessions`,
	RunE: runLive,
}

func init() {
	rootCmd.AddCommand(liveCmd)
}

func runLive(cmd *cobra.Command, args []string) error {
	applyColorConfig()
	filter := session.Filter{
		Interactive: flagInteractive,
		Limit:       flagLimit,
	}
	if flagDir != "" {
		filter.Dir = flagDir
	}

	liveSessions := session.MatchLiveSessions(filter)

	if flagLimit > 0 && len(liveSessions) > flagLimit {
		liveSessions = liveSessions[:flagLimit]
	}

	useJSON := flagJSON || pipe.StdoutIsPiped()
	if useJSON {
		return output.JSONLive(os.Stdout, liveSessions)
	}
	output.TerminalLive(os.Stdout, liveSessions, output.Config{
		ShowPath: flagShowPath,
	})

	fmt.Fprintf(os.Stderr, "%d active sessions\n", len(liveSessions))
	return nil
}
