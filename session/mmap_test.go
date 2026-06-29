package session

import (
	"testing"

	"github.com/amustafa/csgrep/include"
)

func forceMmap(t *testing.T) {
	t.Helper()
	old := mmapMinSize
	mmapMinSize = 0
	t.Cleanup(func() { mmapMinSize = old })
}

func TestMmapParseSampleSession(t *testing.T) {
	forceMmap(t)
	TestParseSampleSession(t)
}

func TestMmapParseFirstMessage(t *testing.T) {
	forceMmap(t)
	TestParseFirstMessage(t)
}

func TestMmapParseLastMessage(t *testing.T) {
	forceMmap(t)
	TestParseLastMessage(t)
}

func TestMmapParseClearResetsFirstMessage(t *testing.T) {
	forceMmap(t)
	TestParseClearResetsFirstMessage(t)
}

func TestMmapParseClearResetsMessages(t *testing.T) {
	forceMmap(t)
	TestParseClearResetsMessages(t)
}

func TestMmapParseMetadataOnly(t *testing.T) {
	forceMmap(t)
	TestParseMetadataOnly(t)
}

func TestMmapParseToolContentExcludedByDefault(t *testing.T) {
	forceMmap(t)
	TestParseToolContentExcludedByDefault(t)
}

func TestMmapParseToolContentIncluded(t *testing.T) {
	forceMmap(t)
	TestParseToolContentIncluded(t)
}

func TestMmapParseTimestampsAreLocal(t *testing.T) {
	forceMmap(t)
	TestParseTimestampsAreLocal(t)
}

func TestMmapParityMessageCount(t *testing.T) {
	files := []string{"sample-session.jsonl", "session-with-clear.jsonl", "agent-session.jsonl"}
	optSets := []ParseOptions{
		{},
		{MetadataOnly: true},
		{Include: include.FromAll()},
	}

	for _, file := range files {
		for _, opts := range optSets {
			mmapMinSize = 1<<30
			scannerResult, err := Parse(testdataPath(file), opts)
			if err != nil {
				t.Fatalf("scanner parse %s: %v", file, err)
			}

			mmapMinSize = 0
			mmapResult, err := Parse(testdataPath(file), opts)
			if err != nil {
				t.Fatalf("mmap parse %s: %v", file, err)
			}

			mmapMinSize = mmapThreshold

			if scannerResult == nil && mmapResult == nil {
				continue
			}
			if (scannerResult == nil) != (mmapResult == nil) {
				t.Errorf("%s: nil mismatch scanner=%v mmap=%v", file, scannerResult == nil, mmapResult == nil)
				continue
			}
			if scannerResult.ID != mmapResult.ID {
				t.Errorf("%s: ID mismatch %q vs %q", file, scannerResult.ID, mmapResult.ID)
			}
			if scannerResult.FirstMessage != mmapResult.FirstMessage {
				t.Errorf("%s: FirstMessage mismatch %q vs %q", file, scannerResult.FirstMessage, mmapResult.FirstMessage)
			}
			if scannerResult.LastMessage != mmapResult.LastMessage {
				t.Errorf("%s: LastMessage mismatch %q vs %q", file, scannerResult.LastMessage, mmapResult.LastMessage)
			}
			if len(scannerResult.Messages) != len(mmapResult.Messages) {
				t.Errorf("%s: message count mismatch %d vs %d", file, len(scannerResult.Messages), len(mmapResult.Messages))
			}
		}
	}
}
