package monitor

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSessionMetadataExtraction tests parsing of session metadata
func TestSessionMetadataExtraction(t *testing.T) {
	// Create a temporary JSONL file with test data
	tmpdir := t.TempDir()
	sessionFile := filepath.Join(tmpdir, "test-session.jsonl")

	testData := `{"type":"user","uuid":"msg1","timestamp":"2026-01-09T14:02:46.229Z","message":{"role":"user","content":"test"},"version":"2.1.1","cwd":"/test/path","sessionId":"test-session","gitBranch":"main","userType":"external","parentUuid":null,"isSidechain":false}
{"type":"assistant","uuid":"msg2","timestamp":"2026-01-09T14:02:52.977Z","message":{"model":"claude-sonnet-4-5-20250929","id":"msg_01","role":"assistant","content":[{"type":"text","text":"response"}],"usage":{"input_tokens":10,"output_tokens":3}},"version":"2.1.1","cwd":"/test/path","sessionId":"test-session","gitBranch":"main"}
{"type":"user","uuid":"msg3","timestamp":"2026-01-09T14:03:00.000Z","message":{"role":"user","content":"second message"},"version":"2.1.1"}
{"type":"assistant","uuid":"msg4","timestamp":"2026-01-09T14:04:00.000Z","message":{"model":"claude-sonnet-4-5-20250929","role":"assistant","content":[]},"version":"2.1.1"}
`

	if err := os.WriteFile(sessionFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Parse the metadata
	metadata, err := GetSessionMetadata(sessionFile)
	if err != nil {
		t.Fatalf("GetSessionMetadata failed: %v", err)
	}

	// Verify results
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"MessageCount", metadata.MessageCount, 4},
		{"UserPrompts", metadata.UserPrompts, 2},
		{"Interruptions", metadata.Interruptions, 0},
		{"GitBranch", metadata.GitBranch, "main"},
		{"IsSidechain", metadata.IsSidechain, false},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.expected)
		}
	}

	// Verify timestamps
	if metadata.Started.IsZero() {
		t.Error("Started time should not be zero")
	}
	if metadata.Ended.IsZero() {
		t.Error("Ended time should not be zero")
	}
	if metadata.Duration <= 0 {
		t.Error("Duration should be positive")
	}
}

// TestSessionInterruptionDetection tests detection of session resumptions
func TestSessionInterruptionDetection(t *testing.T) {
	tmpdir := t.TempDir()
	sessionFile := filepath.Join(tmpdir, "test-interruption.jsonl")

	// Create session with 2-hour gap (should detect as interruption)
	testData := `{"type":"user","timestamp":"2026-01-09T14:00:00.000Z","message":{"role":"user","content":"msg1"},"version":"2.1.1"}
{"type":"assistant","timestamp":"2026-01-09T14:01:00.000Z","message":{"role":"assistant","content":[]},"version":"2.1.1"}
{"type":"user","timestamp":"2026-01-09T16:30:00.000Z","message":{"role":"user","content":"msg2"},"version":"2.1.1"}
{"type":"assistant","timestamp":"2026-01-09T16:31:00.000Z","message":{"role":"assistant","content":[]},"version":"2.1.1"}
`

	if err := os.WriteFile(sessionFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	metadata, err := GetSessionMetadata(sessionFile)
	if err != nil {
		t.Fatalf("GetSessionMetadata failed: %v", err)
	}

	if metadata.Interruptions != 1 {
		t.Errorf("Expected 1 interruption, got %d", metadata.Interruptions)
	}
}

// TestTokenUsageParsing tests extraction of token usage data
func TestTokenUsageParsing(t *testing.T) {
	tmpdir := t.TempDir()
	sessionFile := filepath.Join(tmpdir, "test-tokens.jsonl")

	// Assistant message with token usage
	testData := `{"type":"user","timestamp":"2026-01-09T14:00:00.000Z","message":{"role":"user","content":"analyze"},"version":"2.1.1"}
{"type":"assistant","timestamp":"2026-01-09T14:00:10.000Z","message":{"model":"claude-sonnet-4-5-20250929","role":"assistant","content":[{"type":"text","text":"analysis"}],"usage":{"input_tokens":100,"cache_creation_input_tokens":5000,"cache_read_input_tokens":200,"output_tokens":50}},"version":"2.1.1"}
`

	if err := os.WriteFile(sessionFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// For now, just verify the metadata parses correctly
	metadata, err := GetSessionMetadata(sessionFile)
	if err != nil {
		t.Fatalf("GetSessionMetadata failed: %v", err)
	}

	if metadata.UserPrompts != 1 {
		t.Errorf("Expected 1 user prompt, got %d", metadata.UserPrompts)
	}
}

// TestEventTypeDetection verifies all event types parse correctly
func TestEventTypeDetection(t *testing.T) {
	tmpdir := t.TempDir()
	sessionFile := filepath.Join(tmpdir, "test-events.jsonl")

	testData := `{"type":"user","timestamp":"2026-01-09T14:00:00.000Z","message":{"role":"user","content":"hello"}}
{"type":"assistant","timestamp":"2026-01-09T14:00:10.000Z","message":{"role":"assistant","content":[]}}
{"type":"system","timestamp":"2026-01-09T14:01:00.000Z","subtype":"turn_duration","durationMs":50000}
{"type":"file-history-snapshot","timestamp":"2026-01-09T14:02:00.000Z","snapshot":{"timestamp":"2026-01-09T14:02:00.000Z"}}
`

	if err := os.WriteFile(sessionFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Parse and verify all events are counted
	metadata, err := GetSessionMetadata(sessionFile)
	if err != nil {
		t.Fatalf("GetSessionMetadata failed: %v", err)
	}

	// Should count user + assistant = 2
	if metadata.MessageCount != 2 {
		t.Errorf("Message count: got %d, want 2", metadata.MessageCount)
	}
}
