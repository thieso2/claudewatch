package types

import "time"

// ClaudeProcess represents a monitored Claude instance with its metrics
type ClaudeProcess struct {
	PID        int32
	CPUPercent float64
	MemoryMB   float64
	WorkingDir string
	Command    string
	Uptime     time.Duration
	StartTime  time.Time
	IsHelper   bool // MCP helper vs main instance
}
