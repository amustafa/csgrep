# Benchmark Results

System: Linux, Go 1.25, ripgrep 14.1.1, 5 runs per operation.
Dataset: ~2700 session files, ~636 MB total JSONL.

## Scanner vs rg pre-filter

| Operation | Scanner | rg pre-filter | Speedup |
|-----------|---------|---------------|---------|
| list (all) | 5793ms | 6817ms | (n/a — list doesn't use rg) |
| list (project) | 218ms | 316ms | (n/a) |
| search: "auth" (global) | 828ms | 736ms | 1.1x |
| search: "migration" (global) | 761ms | 616ms | 1.2x |
| search: "database" (global) | 724ms | 468ms | **1.5x** |
| search: "MADV_SEQUENTIAL" (rare) | 699ms | 76ms | **9.2x** |
| search: fuzzy "databse" (global) | 2687ms | 2471ms | (fuzzy skips rg) |

## How it works

With `CSGREP_USE_RG=1`, csgrep shells out to `rg -l` (ripgrep in file-list mode)
to find which JSONL files contain the pattern before parsing them. This skips
JSON parsing entirely for files that can't possibly match.

- **Rare patterns** see the biggest wins — rg scans 636 MB in ~35ms and eliminates
  most files, so csgrep only parses a handful.
- **Common patterns** that appear in most files see modest improvement since fewer
  files are skipped.
- **Fuzzy search** skips the rg pre-filter (rg doesn't do trigram matching).
- **List** is unaffected — it uses head/tail reads, not full parsing.

## Previous experiments

### mmap (reverted)

Memory-mapped I/O was tested and reverted. The `parseState` struct refactor
needed to share code between scanner and mmap paths caused 4x more heap objects
and 18% more GC cycles, negating the allocation savings. The bottleneck is
`json.Unmarshal`, not file I/O.
