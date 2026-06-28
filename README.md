# csgrep — Search your Claude Code history

Ever wonder what you asked Claude last week, or need to find that migration plan from a few sessions ago? Claude Code stores every conversation as JSONL files, but they're not easy to search through by hand.

csgrep is a fast CLI tool that searches across all your Claude Code sessions — regex, fixed-string, or fuzzy — with colored output, smart-case, and parallel scanning. Think [ripgrep](https://github.com/BurntSushi/ripgrep), but for `~/.claude/projects/`.

## Installation

```bash
go install github.com/amustafa/csgrep@latest
```

Requires Go 1.21+.

### Development

If you want to hack on csgrep locally:

```bash
git clone https://github.com/amustafa/csgrep.git
cd csgrep
cp .envrc.example .envrc   # configure CSGREP_LINK_DIR, then `direnv allow`
make build                 # build to bin/csgrep
make link                  # symlink to CSGREP_LINK_DIR
```

## Quick Start

```bash
# Search current project sessions (regex, smart-case)
csgrep "database migration"

# Search across ALL projects
csgrep "auth" -g

# List recent sessions for this project
csgrep list -n 10

# View a full session conversation
csgrep show a1b2c3d4
```

## Default Scope

By default, csgrep scopes to sessions from the current working directory (walking up parent directories to find the matching session folder). This means running `csgrep list` inside your project shows only that project's sessions.

Use `-g` / `--global` to search across all projects, or `-d <path>` to target a specific one.

## Usage

### Search (default command)

```bash
csgrep <pattern> [flags]
```

The default command. Searches message content using regex with smart-case.

```bash
csgrep "auth middleware"                 # regex, current project
csgrep "auth" -g                         # search all projects
csgrep "Auth"                            # case-sensitive (has uppercase)
csgrep -F "exact [literal] string"       # fixed-string match
csgrep -f "databse migrtion"             # fuzzy search (tolerates typos)
csgrep "error" -d ~/workspace/myapp      # search a specific project
csgrep "TODO" -d myapp                   # substring match on project dir
csgrep "bug" --after 3d -n 5             # last 3 days, top 5 results
csgrep "config" --interactive            # only interactive CLI sessions
csgrep "migration" -C 2                  # show 2 messages of context
csgrep "auth" --all                      # include tool call content
csgrep "deploy" --json                   # JSON output for scripting
```

### List

```bash
csgrep list [flags]
```

Enumerate sessions with metadata — session ID, timestamps, project directory, and the first user message (after the last `/clear`).

```bash
csgrep list                              # current project sessions
csgrep list -g                           # all sessions across projects
csgrep list -d ~/workspace/myapp         # sessions for a specific project
csgrep list -d ftron                     # substring match on project dir
csgrep list --interactive                # only interactive CLI sessions
csgrep list --after 1w -n 10             # last week, top 10
csgrep list --json                       # JSON output
csgrep list --path                       # show full JSONL file paths
```

### Show

```bash
csgrep show <session-id> [pattern] [flags]
```

Display a full session conversation. Output is automatically piped through `$PAGER` (or `less`). Supports partial session ID matching.

```bash
csgrep show a1b2c3d4                     # show full conversation
csgrep show a1b2c3d4 "auth"              # highlight matches
csgrep show a1b2c3d4 --role user         # only user messages
csgrep show a1b2c3d4 --all               # include tool calls/results
csgrep show a1b2c3d4 --json              # JSON output
```

## Flags Reference

### Global Flags (all commands)

| Flag | Short | Description |
|------|-------|-------------|
| `--dir <path>` | `-d` | Filter to sessions from this directory. Absolute paths do exact matching; plain strings do substring matching. |
| `--global` | `-g` | Search all projects instead of current directory |
| `--interactive` | | Only show interactive CLI sessions (skip agent/automated) |
| `--after <date>` | | Sessions after this date. Supports `YYYY-MM-DD` or relative: `3d`, `1w`, `2h`, `1m` |
| `--before <date>` | | Sessions before this date (same format as `--after`) |
| `--limit N` | `-n` | Show only the N most recent results |
| `--json` | | Output in JSON format |
| `--no-color` | | Disable colored output |
| `--path` | | Show full file path to session JSONL |

### Search Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--fixed-strings` | `-F` | Treat pattern as a literal string (not regex) |
| `--fuzzy` | `-f` | Use fuzzy (trigram) matching — tolerates typos |
| `--ignore-case` | `-i` | Force case-insensitive matching |
| `--case-sensitive` | `-s` | Force case-sensitive matching |
| `--context N` | `-C` | Show N messages of context around each match |
| `--after-context N` | `-A` | Show N messages after each match |
| `--before-context N` | `-B` | Show N messages before each match |
| `--all` | `-a` | Include tool call/result content in search |
| `--threshold` | | Fuzzy match threshold, 0.0–1.0 (default: 0.3) |

### Show Flags

| Flag | Description |
|------|-------------|
| `--role <role>` | Filter messages by role (`user` or `assistant`) |
| `--all` / `-a` | Include tool call/result content |

## Shell Completion

csgrep supports shell completion via cobra. Generate and install for your shell:

```bash
# Bash
csgrep completion bash > ~/.local/share/bash-completion/completions/csgrep

# Zsh
csgrep completion zsh > "${fpath[1]}/_csgrep"

# Fish
csgrep completion fish > ~/.config/fish/completions/csgrep.fish
```

## Output Format

### Terminal (default)

```
a1b2c3d4-... (2026-06-25 14:30) /home/user/workspace/myapp
  first: (2026-06-25 12:00) help me set up auth middleware
  last:  (2026-06-25 14:30) looks good, ship it
```

For search results:
```
a1b2c3d4-... (2026-06-25 14:30) /home/user/workspace/myapp
  [user]   L42: ...the auth middleware should handle JWT...
  [assist] L87: ...I'll configure the auth middleware with RS256...
```

- Session ID in **magenta**, timestamp in **green**, directory in **cyan**
- Matched text highlighted in **red bold**
- Fuzzy matches show a similarity score: `[0.85]`
- Context messages (`-C`) shown dimmed with `──` separators between groups

### JSON (`--json`)

```json
[
  {
    "session_id": "a1b2c3d4-...",
    "project_dir": "/home/user/workspace/myapp",
    "timestamp": "2026-06-25T14:30:00Z",
    "role": "user",
    "text": "the auth middleware should handle JWT...",
    "line_num": 42,
    "score": 1.0,
    "offsets": [[4, 8]]
  }
]
```

## How It Works

Claude Code stores session transcripts as JSONL files in `~/.claude/projects/<encoded-dir>/`. Each line contains a JSON object with message type, content, timestamps, and metadata.

csgrep parallelizes scanning across sessions using a goroutine worker pool (one per CPU core), parsing JSONL and matching in parallel, then sorting results by timestamp before output.

### Smart-Case (ripgrep behavior)

- Pattern `auth` → case-insensitive (matches "Auth", "AUTH", "auth")
- Pattern `Auth` → case-sensitive (only matches "Auth")
- Override with `-i` (always insensitive) or `-s` (always sensitive)

### Fuzzy Matching

Uses trigram similarity (Jaccard index over 3-character ngrams). A threshold of 0.3 means at least 30% trigram overlap is required. Results are ranked by score, best matches first.

### First Message

The "first message" shown in `list` output is the first user message after the last `/clear` command — reflecting the effective start of the conversation, not the literal first line of the session file.

## Performance

On a typical session history (~2000 sessions, ~150k JSONL lines):

- `csgrep list`: scans all sessions in under 2 seconds
- `csgrep "pattern"`: parallel regex search completes in under 1 second
- Compared to the Python predecessor: ~2.5x faster wall-clock time

## License

[MIT](LICENSE)
