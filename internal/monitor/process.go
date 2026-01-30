package monitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/thies/claudewatch/internal/types"
)

// FindClaudeProcesses discovers all running Claude instances and returns their metrics
func FindClaudeProcesses(showHelpers bool) ([]types.ClaudeProcess, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	var claudeProcesses []types.ClaudeProcess

	for _, proc := range processes {
		// Skip processes that aren't Claude
		if !isClaudeProcess(proc) {
			continue
		}

		// Get command line for helper detection
		cmdline, err := proc.Cmdline()
		if err != nil {
			continue
		}

		isHelper := isClaudeHelperProcess(cmdline)

		// Skip helpers unless explicitly requested
		if isHelper && !showHelpers {
			continue
		}

		// Collect metrics
		claudeProc, err := collectMetrics(proc, isHelper)
		if err != nil {
			continue // Skip processes we can't collect metrics for
		}

		claudeProcesses = append(claudeProcesses, claudeProc)
	}

	return claudeProcesses, nil
}

// isClaudeProcess checks if a process is a Claude instance
func isClaudeProcess(proc *process.Process) bool {
	exe, err := proc.Exe()
	if err != nil {
		return false
	}

	// Check for Claude executable paths
	if strings.Contains(exe, "claude") && !strings.Contains(exe, "Claude.app") {
		return true
	}

	return false
}

// isClaudeHelperProcess checks if a process is a Claude MCP helper
func isClaudeHelperProcess(cmdline string) bool {
	return strings.Contains(cmdline, "--claude-in-chrome-mcp") ||
		strings.Contains(cmdline, "--mcp")
}

// collectMetrics gathers CPU, memory, and other metrics for a process
func collectMetrics(proc *process.Process, isHelper bool) (types.ClaudeProcess, error) {
	pid := proc.Pid

	// CPU: Get CPU percentage (non-blocking)
	cpuPercent, err := proc.CPUPercent()
	if err != nil {
		cpuPercent = 0
	}

	// Memory: Get RSS in bytes and convert to MB
	memInfo, err := proc.MemoryInfo()
	var memoryMB float64
	if err == nil {
		memoryMB = float64(memInfo.RSS) / 1024 / 1024
	}

	// Working directory: Use CGo proc_pidinfo on macOS
	workDir, err := getWorkingDir(pid)
	if err != nil {
		workDir = "[Permission Denied]"
	}

	// Command line and timing info
	cmdline, _ := proc.Cmdline()
	createTime, _ := proc.CreateTime()
	var uptime time.Duration
	if createTime > 0 {
		uptime = time.Since(time.UnixMilli(createTime))
	}

	return types.ClaudeProcess{
		PID:        pid,
		CPUPercent: cpuPercent,
		MemoryMB:   memoryMB,
		WorkingDir: workDir,
		Command:    cmdline,
		Uptime:     uptime,
		StartTime:  time.UnixMilli(createTime),
		IsHelper:   isHelper,
	}, nil
}
