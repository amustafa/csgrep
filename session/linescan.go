package session

import "bytes"

// scanLines iterates over newline-delimited lines in data, calling fn for each non-empty line.
// line is a sub-slice of data (zero-copy). lineNum is 1-based.
// If fn returns false, scanning stops early.
func scanLines(data []byte, fn func(line []byte, lineNum int) bool) {
	offset := 0
	lineNum := 0
	for offset < len(data) {
		lineNum++
		end := bytes.IndexByte(data[offset:], '\n')
		var line []byte
		if end == -1 {
			line = data[offset:]
			offset = len(data)
		} else {
			line = data[offset : offset+end]
			offset += end + 1
		}
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		if len(line) == 0 {
			continue
		}
		if !fn(line, lineNum) {
			return
		}
	}
}
