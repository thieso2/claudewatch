package monitor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSessionFile(t *testing.T) {
	// Load the fixture file
	fixturePath := filepath.Join("testdata", "sample_session.jsonl")

	// Check if fixture exists
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skipf("Fixture file not found: %s", fixturePath)
	}

	stats, err := ParseSessionFile(fixturePath)
	if err != nil {
		t.Fatalf("ParseSessionFile failed: %v", err)
	}

	// Basic sanity checks
	if stats == nil {
		t.Fatal("ParseSessionFile returned nil stats")
	}

	if stats.FilePath != fixturePath {
		t.Errorf("FilePath mismatch: got %s, want %s", stats.FilePath, fixturePath)
	}

	t.Logf("Parsed session with %d total messages", stats.TotalMessages)
	t.Logf("  User messages: %d", stats.UserMessages)
	t.Logf("  Assistant messages: %d", stats.AssistantMessages)
	t.Logf("  Message history length: %d", len(stats.MessageHistory))
}

func TestMessageTypeDetection(t *testing.T) {
	fixturePath := filepath.Join("testdata", "sample_session.jsonl")

	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skipf("Fixture file not found: %s", fixturePath)
	}

	stats, err := ParseSessionFile(fixturePath)
	if err != nil {
		t.Fatalf("ParseSessionFile failed: %v", err)
	}

	// Count message types
	promptCount := 0
	assistantCount := 0
	toolResultCount := 0

	for _, msg := range stats.MessageHistory {
		t.Logf("Message: Type=%q, Role=%q, Content=[%d chars], Tool=%q",
			msg.Type, msg.Role, len(msg.Content), msg.ToolName)

		switch msg.Type {
		case "prompt":
			promptCount++
		case "assistant_response":
			assistantCount++
		case "tool_result":
			toolResultCount++
		}
	}

	t.Logf("\nMessage type breakdown:")
	t.Logf("  Prompts: %d", promptCount)
	t.Logf("  Assistant responses: %d", assistantCount)
	t.Logf("  Tool results: %d", toolResultCount)

	if promptCount == 0 && assistantCount == 0 && toolResultCount == 0 {
		t.Error("No message types were detected - parser may not be extracting types correctly")
	}
}

func TestToolInformationExtraction(t *testing.T) {
	fixturePath := filepath.Join("testdata", "sample_session.jsonl")

	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skipf("Fixture file not found: %s", fixturePath)
	}

	stats, err := ParseSessionFile(fixturePath)
	if err != nil {
		t.Fatalf("ParseSessionFile failed: %v", err)
	}

	// Look for tool uses
	toolsFound := 0
	for _, msg := range stats.MessageHistory {
		if msg.Type == "tool_result" || msg.ToolName != "" {
			toolsFound++
			t.Logf("Tool found: %s", msg.ToolName)
			if msg.ToolInput != "" {
				t.Logf("  Input: %s", msg.ToolInput)
			}
		}
	}

	t.Logf("Tools found: %d", toolsFound)
	if toolsFound == 0 {
		t.Log("Warning: No tool information extracted. Check if fixture contains tool uses.")
	}
}

func TestSessionStats(t *testing.T) {
	fixturePath := filepath.Join("testdata", "sample_session.jsonl")

	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skipf("Fixture file not found: %s", fixturePath)
	}

	stats, err := ParseSessionFile(fixturePath)
	if err != nil {
		t.Fatalf("ParseSessionFile failed: %v", err)
	}

	// Check timestamps
	if stats.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if stats.LastActivity.IsZero() {
		t.Error("LastActivity is zero")
	}

	if !stats.CreatedAt.IsZero() && !stats.LastActivity.IsZero() {
		if stats.CreatedAt.After(stats.LastActivity) {
			t.Error("CreatedAt is after LastActivity")
		}
	}

	// Check summary
	summary := stats.GetSummary()
	t.Logf("Summary: %s", summary)
	if summary == "" {
		t.Error("GetSummary returned empty string")
	}

	// Check detailed stats
	detailed := stats.GetDetailedStats()
	t.Logf("Detailed: %s", detailed)
	if detailed == "" {
		t.Error("GetDetailedStats returned empty string")
	}
}

func TestMessageContent(t *testing.T) {
	fixturePath := filepath.Join("testdata", "sample_session.jsonl")

	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skipf("Fixture file not found: %s", fixturePath)
	}

	stats, err := ParseSessionFile(fixturePath)
	if err != nil {
		t.Fatalf("ParseSessionFile failed: %v", err)
	}

	for i, msg := range stats.MessageHistory {
		if msg.Content == "" {
			t.Errorf("Message %d has empty content", i)
		}

		if msg.Timestamp.IsZero() {
			t.Errorf("Message %d has zero timestamp", i)
		}

		if msg.Role == "" {
			t.Errorf("Message %d has empty role", i)
		}
	}
}
