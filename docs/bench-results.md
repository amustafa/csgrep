# Benchmark Results

Measured on 2724 session files (636 MB total).
System: Linux, Go 1.25.

## Wall-clock (5 runs each)

| Operation | Scanner | Mmap |
|-----------|---------|------|
| list (all, 2724 sessions) | 7478ms | 8306ms |
| list (project, 66 sessions) | 200ms | 219ms |
| search: "auth" (global) | 754ms | 692ms |
| search: "migration" (global) | 775ms | 740ms |
| search: "database" (global) | 732ms | 724ms |
| search: fuzzy "databse" (global) | 2564ms | 2744ms |

## Heap allocation (search "auth", 2233 files)

| Metric | Scanner | Mmap | Delta |
|--------|---------|------|-------|
| Wall clock | 683ms | 677ms | -1% |
| Total allocs | 5950 MB | 5634 MB | -5% |
| Heap in use | 300.8 MB | 372.7 MB | +24% |
| Heap objects | 75,937 | 298,070 | +293% |
| GC cycles | 71 | 84 | +18% |
| GC pause total | 29ms | 34ms | +15% |

## Conclusion

Mmap saved ~300 MB in total allocations (no scanner buffers) but the
`parseState` struct refactor caused 4x more heap objects, more GC cycles,
and longer GC pauses. Net effect was negative. **Mmap was reverted.**

The bottleneck is `json.Unmarshal` per line, not file I/O. Future
optimization should target the JSON parsing layer (e.g. `jsoniter`,
hand-rolled field extraction, or pre-filtering lines before unmarshaling).
