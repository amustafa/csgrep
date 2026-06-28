# csgrep

A CLI tool for searching Claude Code session transcripts stored as JSONL files.

## Language

### Content types

**Artifact**:
A file that Claude wrote or modified during a session, extracted from `Write`, `Edit`, or `NotebookEdit` tool calls in the session transcript.
_Avoid_: output, result, file change, diff

**Tool Output**:
The result returned from a tool call — stdout from `Bash`, file content from `Read`, etc.
_Avoid_: tool result, tool response, tool return

**Include Set**:
The set of content types beyond conversational messages to include in search or display. Controls which artifacts and tool outputs are considered.
_Avoid_: content filter, search scope

### Session concepts

**Message**:
A single unit of content in a session — a user message, assistant response, artifact, or tool output. The universal type that flows through search, sort, and display.
_Avoid_: entry, line, record

**Session**:
A complete Claude Code conversation stored as a JSONL file. Contains messages, tool calls, and metadata.
_Avoid_: conversation, transcript, log

### Artifact scopes

**Non-temp artifact**:
An artifact whose file path is not under `/tmp/`, `/var/tmp/`, or other temporary directories. The default scope when searching artifacts.

**Plan artifact**:
An artifact written to `~/.claude/plans/`. Excluded from default artifact scope because paths carry no meaningful signal.

## Relationships

- A **Session** contains zero or more **Messages**
- An **Artifact** is a **Message** with a file path and tool name
- A **Tool Output** is a **Message** with a tool name
- An **Include Set** determines which **Messages** are produced during parsing

## Example dialogue

> **Dev:** "I want to find the session where Claude wrote the timeout handler."
> **Domain expert:** "Search with `--include artifacts` — that searches the content of all non-temp **Artifacts**. If you only care about file names, use `--include artifacts:path`."
> **Dev:** "What about finding a stack trace from a build failure?"
> **Domain expert:** "That's a **Tool Output** from a Bash call. Use `--include tool-outputs`, or just `--all` to search everything."

## Flagged ambiguities

- "artifact" outside this project often means "claude.ai rendered panel" or "build output" — in csgrep it strictly means files written by Claude via Write/Edit/NotebookEdit tool calls.
- `--all` was originally "include tool content." It now means "include artifacts and tool-outputs" — same practical effect, but the mental model shifted from a raw content dump to structured content types.
