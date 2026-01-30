package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thies/claudewatch/internal/monitor"
	"github.com/thies/claudewatch/internal/ui"
)

func main() {
	// Parse CLI flags
	interval := flag.Duration("interval", 1*time.Second, "Refresh interval")
	showHelpers := flag.Bool("show-helpers", false, "Show MCP helper processes")
	processMode := flag.Bool("p", false, "Show processes (CLI mode)")
	sessionsDir := flag.String("d", "", "Show sessions for directory (CLI mode)")
	inspectFile := flag.String("i", "", "Inspect session file (CLI mode)")
	flag.Parse()

	// Handle CLI modes
	if *processMode {
		cliShowProcesses(*showHelpers)
		return
	}

	if *sessionsDir != "" {
		cliShowSessions(*sessionsDir)
		return
	}

	if *inspectFile != "" {
		cliInspectSession(*inspectFile)
		return
	}

	// Run TUI mode
	model := ui.NewModel(*interval, *showHelpers)
	program := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// cliShowProcesses displays all Claude processes in CLI mode
func cliShowProcesses(showHelpers bool) {
	processes, err := monitor.FindClaudeProcesses(showHelpers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(processes) == 0 {
		fmt.Println("No Claude processes found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PID\tCPU%\tMEM\tUPTIME\tWORKDIR\tCOMMAND")
	fmt.Fprintln(w, "---\t----\t---\t------\t-------\t-------")

	for _, proc := range processes {
		fmt.Fprintf(w, "%d\t%.1f%%\t%.2fM\t%v\t%s\t%s\n",
			proc.PID,
			proc.CPUPercent,
			proc.MemoryMB,
			proc.Uptime,
			proc.WorkingDir,
			truncateCmd(proc.Command, 50),
		)
	}
	w.Flush()
}

// cliShowSessions displays all sessions for a directory in CLI mode
func cliShowSessions(dir string) {
	sessions, err := monitor.FindSessionsForDirectory(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Printf("No sessions found for directory: %s\n", dir)
		return
	}

	fmt.Printf("Found %d sessions for: %s\n\n", len(sessions), dir)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SESSION ID\tTITLE\tUPDATED\tFILE")
	fmt.Fprintln(w, "----------\t-----\t-------\t----")

	for _, sess := range sessions {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			sess.ID[:8]+"...",
			sess.GetSessionInfo(),
			sess.GetSessionTime(),
			sess.FilePath,
		)
	}
	w.Flush()
}

// cliInspectSession displays detailed information about a session in CLI mode
func cliInspectSession(filePath string) {
	stats, err := monitor.ParseSessionFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing session: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== SESSION DETAILS ===")
	fmt.Printf("File: %s\n", stats.FilePath)
	fmt.Println()

	fmt.Println("=== STATISTICS ===")
	fmt.Printf("Started:      %s\n", stats.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Last Activity: %s\n", stats.LastActivity.Format("2006-01-02 15:04:05"))
	fmt.Printf("Duration:     %v\n", stats.Duration)
	fmt.Printf("Claude Version: %s\n", stats.ClaudeVersion)
	fmt.Println()

	fmt.Println("=== MESSAGE COUNTS ===")
	fmt.Printf("Total Messages:      %d\n", stats.TotalMessages)
	fmt.Printf("User Prompts:        %d\n", stats.UserMessages)
	fmt.Printf("Claude Responses:    %d\n", stats.AssistantMessages)
	fmt.Println()

	fmt.Println("=== EVENTS ===")
	fmt.Printf("Progress Events:     %d\n", stats.ProgressEvents)
	fmt.Printf("System Events:       %d\n", stats.SystemEvents)
	fmt.Printf("File Snapshots:      %d\n", stats.FileSnapshots)
	fmt.Printf("Queue Operations:    %d\n", stats.QueueOperations)
	fmt.Printf("Errors:              %d\n", stats.ErrorCount)
	fmt.Println()

	if len(stats.MessageHistory) > 0 {
		fmt.Println("=== CONVERSATION ===")
		for i, msg := range stats.MessageHistory {
			role := msg.Role
			if msg.Role == "user" {
				role = "👤 user"
			} else if msg.Role == "assistant" {
				role = "🤖 assistant"
			}

			content := msg.Content
			if len(content) > 100 {
				content = content[:97] + "..."
			}
			// Replace newlines for display
			for j := 0; j < len(content); j++ {
				if content[j] == '\n' {
					content = content[:j] + " " + content[j+1:]
				}
			}

			fmt.Printf("[%d] %s (%s): %s\n", i+1, role, msg.Timestamp.Format("15:04:05"), content)
		}
	}
}

func truncateCmd(cmd string, maxLen int) string {
	if len(cmd) <= maxLen {
		return cmd
	}
	return cmd[:maxLen-3] + "..."
}
