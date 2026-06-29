#!/usr/bin/env bash
# bench.sh — benchmark csgrep list and search operations
# Usage: ./bench.sh [label]
# Results are appended to docs/bench-results.md

set -euo pipefail

LABEL="${1:-$(git rev-parse --short HEAD)}"
RUNS=5
OUTFILE="docs/bench-results.md"

go build -o bin/csgrep .

hyperfine_check() {
    if ! command -v hyperfine &>/dev/null; then
        echo "hyperfine not found, using manual timing" >&2
        return 1
    fi
    return 0
}

time_cmd() {
    local label="$1"
    shift
    if hyperfine_check 2>/dev/null; then
        result=$(hyperfine --warmup 1 --runs "$RUNS" --export-json /dev/stdout "$@" 2>/dev/null)
        mean=$(echo "$result" | python3 -c "import sys,json; print(f'{json.load(sys.stdin)[\"results\"][0][\"mean\"]*1000:.0f}')")
        stddev=$(echo "$result" | python3 -c "import sys,json; print(f'{json.load(sys.stdin)[\"results\"][0][\"stddev\"]*1000:.0f}')")
        echo "| $label | ${mean}ms | ±${stddev}ms | $RUNS |"
    else
        total=0
        for i in $(seq 1 $RUNS); do
            ms=$( { TIMEFORMAT='%3R'; time "$@" >/dev/null 2>&1; } 2>&1 )
            ms_int=$(echo "$ms * 1000" | bc | cut -d. -f1)
            total=$((total + ms_int))
        done
        mean=$((total / RUNS))
        echo "| $label | ${mean}ms | - | $RUNS |"
    fi
}

echo ""
echo "## Benchmark: $LABEL ($(date -u +%Y-%m-%dT%H:%M:%SZ))"
echo ""
echo "| Operation | Mean | Stddev | Runs |"
echo "|-----------|------|--------|------|"

time_cmd "list (all)" bin/csgrep list -g
time_cmd "list (project)" bin/csgrep list
time_cmd "search: \"auth\" (global)" bin/csgrep "auth" -g -n 50
time_cmd "search: \"migration\" (global)" bin/csgrep "migration" -g -n 50
time_cmd "search: \"database\" (global)" bin/csgrep "database" -g -n 50
time_cmd "search: fuzzy \"databse\" (global)" bin/csgrep -f "databse" -g -n 50
