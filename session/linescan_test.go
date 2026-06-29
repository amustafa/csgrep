package session

import (
	"testing"
)

func collectLines(data []byte) (lines []string, lineNums []int) {
	scanLines(data, func(line []byte, lineNum int) bool {
		lines = append(lines, string(line))
		lineNums = append(lineNums, lineNum)
		return true
	})
	return
}

func TestScanLinesBasic(t *testing.T) {
	lines, nums := collectLines([]byte("aaa\nbbb\nccc\n"))
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if lines[0] != "aaa" || lines[1] != "bbb" || lines[2] != "ccc" {
		t.Errorf("lines = %v", lines)
	}
	if nums[0] != 1 || nums[1] != 2 || nums[2] != 3 {
		t.Errorf("lineNums = %v", nums)
	}
}

func TestScanLinesNoTrailingNewline(t *testing.T) {
	lines, _ := collectLines([]byte("aaa\nbbb"))
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	if lines[1] != "bbb" {
		t.Errorf("last line = %q, want %q", lines[1], "bbb")
	}
}

func TestScanLinesEmptyLines(t *testing.T) {
	lines, nums := collectLines([]byte("aaa\n\nbbb\n\n"))
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2 (empty lines skipped)", len(lines))
	}
	if nums[0] != 1 || nums[1] != 3 {
		t.Errorf("lineNums = %v, want [1, 3]", nums)
	}
}

func TestScanLinesCRLF(t *testing.T) {
	lines, _ := collectLines([]byte("aaa\r\nbbb\r\n"))
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	if lines[0] != "aaa" || lines[1] != "bbb" {
		t.Errorf("lines = %v (\\r not stripped)", lines)
	}
}

func TestScanLinesEarlyStop(t *testing.T) {
	var count int
	scanLines([]byte("a\nb\nc\nd\n"), func(line []byte, lineNum int) bool {
		count++
		return count < 2
	})
	if count != 2 {
		t.Errorf("got %d calls, want 2 (stopped after second)", count)
	}
}

func TestScanLinesSingleLine(t *testing.T) {
	lines, _ := collectLines([]byte("only"))
	if len(lines) != 1 || lines[0] != "only" {
		t.Errorf("lines = %v, want [\"only\"]", lines)
	}
}

func TestScanLinesEmpty(t *testing.T) {
	lines, _ := collectLines([]byte{})
	if len(lines) != 0 {
		t.Errorf("got %d lines from empty input, want 0", len(lines))
	}
}
