package monitor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// SessionEntry represents a single entry in a session JSONL file
type SessionEntry struct {
	Type      string                 `json:"type"`
	Timestamp string                 `json:"timestamp"`
	Message   *struct {
		Role    string        `json:"role"`
		Content interface{}   `json:"content"` // Can be string or array
	} `json:"message"`
	Data map[string]interface{} `json:"data"`
}

// Message represents a user message or response
type Message struct {
	Role      string
	Content   string
	Timestamp time.Time
}

// SessionStats contains aggregated session statistics
type SessionStats struct {
	FilePath        string
	CreatedAt       time.Time
	LastActivity    time.Time
	Duration        time.Duration
	TotalMessages   int
	UserMessages    int
	AssistantMessages int
	CompactCount    int
	MessageHistory  []Message
	ErrorCount      int
}

// ParseSessionFile reads and parses a JSONL session file
func ParseSessionFile(filePath string) (*SessionStats, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open session file: %w", err)
	}
	defer file.Close()

	stats := &SessionStats{
		FilePath:       filePath,
		MessageHistory: []Message{},
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		var entry SessionEntry

		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip malformed lines
		}

		// Parse timestamp
		var timestamp time.Time
		if entry.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339Nano, entry.Timestamp); err == nil {
				timestamp = t
			} else if t, err := time.Parse(time.RFC3339, entry.Timestamp); err == nil {
				timestamp = t
			}
		}

		// Update creation and activity times
		if stats.CreatedAt.IsZero() || timestamp.Before(stats.CreatedAt) {
			stats.CreatedAt = timestamp
		}
		if timestamp.After(stats.LastActivity) {
			stats.LastActivity = timestamp
		}

		// Process different entry types
		switch entry.Type {
		case "user", "assistant":
			// Message entry
			if entry.Message != nil && entry.Message.Role != "" {
				stats.TotalMessages++
				if entry.Message.Role == "user" {
					stats.UserMessages++
				} else if entry.Message.Role == "assistant" {
					stats.AssistantMessages++
				}

				// Extract message content - can be string or array
				var contentStr string
				if content, ok := entry.Message.Content.(string); ok {
					contentStr = content
				} else if contentArr, ok := entry.Message.Content.([]interface{}); ok {
					// For array content, try to extract text
					for _, item := range contentArr {
						if itemMap, ok := item.(map[string]interface{}); ok {
							if itemContent, ok := itemMap["content"].(string); ok {
								contentStr = itemContent
								break
							}
						}
					}
				}

				if contentStr != "" {
					msg := Message{
						Role:      entry.Message.Role,
						Content:   contentStr,
						Timestamp: timestamp,
					}
					stats.MessageHistory = append(stats.MessageHistory, msg)
				}
			}

		case "compact":
			stats.CompactCount++

		case "error":
			stats.ErrorCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading session file: %w", err)
	}

	// Calculate duration
	if !stats.CreatedAt.IsZero() && !stats.LastActivity.IsZero() {
		stats.Duration = stats.LastActivity.Sub(stats.CreatedAt)
	}

	return stats, nil
}

// GetSummary returns a human-readable summary of session stats
func (s *SessionStats) GetSummary() string {
	duration := formatDuration(s.Duration)
	return fmt.Sprintf(
		"Started: %s | Duration: %s | Messages: %d (User: %d, AI: %d) | Compacts: %d",
		s.CreatedAt.Format("2006-01-02 15:04"),
		duration,
		s.TotalMessages,
		s.UserMessages,
		s.AssistantMessages,
		s.CompactCount,
	)
}

// formatDuration converts a duration to human-readable format
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "0s"
	}

	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// GetMessageSummary returns a brief summary of a message
func (m Message) GetMessageSummary() string {
	// Truncate content to 80 characters
	content := m.Content
	if len(content) > 80 {
		content = content[:77] + "..."
	}

	// Replace newlines with spaces for single-line display
	for _, c := range content {
		if c == '\n' {
			content = content[:len(content)-1] + " "
			break
		}
	}

	return fmt.Sprintf("[%s] %s", m.Role, content)
}
