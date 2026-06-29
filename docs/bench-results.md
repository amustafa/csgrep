# Benchmark Results

System: Linux, Go 1.25, ripgrep 14.1.1, 5 runs per operation.
Dataset: ~2700 session files, ~636 MB total JSONL.

## Default (with rg) vs CSGREP_NO_DEPS=1

| Operation | No deps | With rg | Speedup |
|-----------|---------|---------|---------|
| list (all, ~2700 sessions) | 8295ms | 1181ms | **7.0x** |
| list (project, 66 sessions) | 202ms | 37ms | **5.5x** |
| search: "auth" (global) | 752ms | 793ms | ~1x |
| search: "migration" (global) | 808ms | 581ms | 1.4x |
| search: "database" (global) | 716ms | 507ms | 1.4x |
| search: "MADV_SEQUENTIAL" (rare) | 753ms | 91ms | **8.3x** |
| search: fuzzy "databse" (global) | 2571ms | 2603ms | ~1x |

## How it works

By default (when `rg` is installed), csgrep uses external tools for acceleration:

**Search**: `rg -l` scans all 636 MB in ~35ms to find which files contain the
pattern. csgrep then only JSON-parses matching files. Rare patterns see the
biggest wins (8x) since most files are skipped entirely.

**List**: Uses `ParseFast` (head/tail reads only, ~114KB per file) instead of
full-file parsing. `rg` identifies the ~5 files containing `/clear` so only
those get full-parsed for correct first-message extraction.

Set `CSGREP_NO_DEPS=1` to disable external tools and use pure Go parsing.

## Previous experiments

### mmap (reverted)

Memory-mapped I/O was tested and reverted. The `parseState` struct refactor
needed to share code between scanner and mmap paths caused 4x more heap objects
and 18% more GC cycles, negating the allocation savings. The bottleneck is
`json.Unmarshal`, not file I/O.
