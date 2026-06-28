# Results Sorting & Grouping

## Summary

Add user-controllable sorting and grouping to csgrep, enabling it to serve as a backend for search UIs. Today sorting is hardcoded by match mode; this feature exposes it as flags and adds first-class grouped output.

## Flags

### `--sort field:dir[,field:dir,...]`

Multi-field sort. Each entry is a field name and direction separated by a colon.

**Valid fields:** `timestamp`, `score`

**Valid directions:** `asc`, `desc`

**Examples:**
```
--sort timestamp:asc
--sort score:desc,timestamp:asc
```

**Defaults when omitted (current behavior preserved):**
- Fuzzy search (`--fuzzy`): `score:desc`
- Regex / fixed-string search: `timestamp:desc`

**Validation:**
- Unknown field names → error with message listing valid fields
- `score` without `--fuzzy` → error: `"--sort score requires --fuzzy"`
- Missing direction → error (both field and direction are required)

### `--group-by field`

Group results by a field. Defaults to `session_id` when omitted.

**Valid fields:** `session_id`, `project_dir`, `role`

**Group ranking:** each group is ranked by the "best" value of the primary sort field in the sort direction. For `timestamp:desc`, the group with the newest match ranks first. For `score:desc`, the group with the highest score ranks first.

**Within-group ordering:** matches inside each group are sorted by the full `--sort` specification.

### `--no-group-by`

Disables grouping. Results are sorted flat.

**Mutually exclusive** with `--group-by`. If both are provided → error: `"cannot use --group-by and --no-group-by together"`

## Flag Interactions

| Flags provided | Behavior |
|---|---|
| Neither | Group by `session_id`, mode-based sort default |
| `--sort` only | Group by `session_id`, user-specified sort |
| `--group-by` only | Group by specified field, mode-based sort default |
| `--sort` + `--group-by` | Group by specified field, user-specified sort |
| `--no-group-by` | Flat sort, mode-based sort default |
| `--no-group-by` + `--sort` | Flat sort, user-specified sort |

## JSON Output

### Grouped (default, or with `--group-by`)

```json
[
  {
    "group_key": "abc123",
    "group_by": "session_id",
    "rank_value": "2026-06-27T14:00:00Z",
    "matches": [
      {
        "session_id": "abc123",
        "project_dir": "/home/user/myapp",
        "timestamp": "2026-06-27T14:00:00Z",
        "role": "user",
        "text": "matched text...",
        "line_num": 42,
        "score": 1.0,
        "offsets": [[5, 12]],
        "path": "/home/user/.claude/projects/..."
      }
    ]
  }
]
```

- `group_key`: the value of the grouped field for this group
- `group_by`: the field name used for grouping (self-describing output)
- `rank_value`: the representative value used to order this group among other groups
- `matches`: array of match objects, sorted per `--sort` specification
- Match objects are unchanged — no field hoisting to group level

### Flat (`--no-group-by`)

```json
[
  {
    "session_id": "abc123",
    "project_dir": "/home/user/myapp",
    "timestamp": "2026-06-27T14:00:00Z",
    "role": "user",
    "text": "matched text...",
    "line_num": 42,
    "score": 1.0,
    "offsets": [[5, 12]],
    "path": "/home/user/.claude/projects/..."
  }
]
```

Same shape as today's `--json` output.

### Terminal Output

No structural change. Terminal output already renders grouped by session. The `--sort` flag controls ordering; `--group-by` controls grouping; rendering continues to use session headers and colored output.

## Backward Compatibility

- No flags → identical behavior to today (grouped by session, mode-based sort)
- Existing `--json` output shape changes to nested format by default (grouped)
- `--json --no-group-by` produces the previous flat shape
