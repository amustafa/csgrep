# Structured --include flag instead of per-content-type boolean flags

We chose a single `--include` flag with comma-separated typed values (`artifacts`, `tool-outputs`, with colon sub-selectors like `artifacts:path`) instead of adding separate boolean flags (`--include-artifacts`, `--artifacts-only`, `--include-tool-outputs`, etc.).

The `--all` flag becomes sugar for `--include artifacts,tool-outputs`, preserving backwards compatibility.

The alternative — individual boolean flags — would have been simpler to implement but scales poorly. Each new content type would add 2-3 flags (include, only, sub-selectors), and combinations between them create ambiguity (`--include-artifacts --artifacts-only` — error or override?). The structured value approach handles composition naturally via comma separation and keeps the flag namespace flat. The cost is a slightly more complex parser for the flag value itself, but that's a one-time implementation cost versus ongoing UX debt.
