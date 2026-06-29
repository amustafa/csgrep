#!/usr/bin/env bash
# bench.sh — benchmark csgrep list and search operations
# Usage: ./bench.sh [label]
# Results are appended to docs/bench-results.md

set -euo pipefail

LABEL="${1:-$(git rev-parse --short HEAD)}"
RUNS=5

go build -o bin/csgrep .

time_cmd() {
    local label="$1"
    shift
    total=0
    for i in $(seq 1 $RUNS); do
        ms=$( { TIMEFORMAT='%3R'; time "$@" >/dev/null 2>&1; } 2>&1 )
        ms_int=$(echo "$ms * 1000" | bc | cut -d. -f1)
        total=$((total + ms_int))
    done
    mean=$((total / RUNS))
    echo "| $label | ${mean}ms | $RUNS |"
}

echo ""
echo "## $LABEL ($(date -u +%Y-%m-%dT%H:%M:%SZ))"
echo ""
echo "| Operation | Mean | Runs |"
echo "|-----------|------|------|"

time_cmd "list (all)" bin/csgrep list -g
time_cmd "list (project)" bin/csgrep list
time_cmd "search: \"auth\" (global)" bin/csgrep "auth" -g -n 50
time_cmd "search: \"migration\" (global)" bin/csgrep "migration" -g -n 50
time_cmd "search: \"database\" (global)" bin/csgrep "database" -g -n 50
time_cmd "search: \"MADV_SEQUENTIAL\" (global, rare)" bin/csgrep "MADV_SEQUENTIAL" -g -n 50
time_cmd "search: fuzzy \"databse\" (global)" bin/csgrep -f "databse" -g -n 50
