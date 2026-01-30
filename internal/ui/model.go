package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
	"github.com/thies/claudewatch/internal/monitor"
	"github.com/thies/claudewatch/internal/types"
)

// SessionInfo represents session information for display
type SessionInfo struct {
	ID      string
	Title   string
	Updated string
	Path    string
}

// MessageRow represents a message for display in the message table
type MessageRow struct {
	Index   int
	Role    string
	Content string
	Time    string
}

// ViewMode represents the current view being displayed
type ViewMode int

const (
	ViewProcesses ViewMode = iota
	ViewProjects
	ViewSessions
	ViewSessionDetail
	ViewMessageDetail
)

// ProjectDir represents a project directory with metadata
type ProjectDir struct {
	Name          string
	Path          string
	DisplayName   string // Human-readable project name
	Modified      time.Time
	Sessions      int // Count of session files
}

type MessageFilter int

const (
	FilterAll MessageFilter = iota
	FilterUserOnly
	FilterAssistantOnly
)

// Model represents the main UI state
type Model struct {
	// Main view
	table          table.Model
	processes      []types.ClaudeProcess
	lastUpdate     time.Time
	updateInterval time.Duration
	showHelpers    bool
	quitting       bool
	sortColumn     string
	sortAscending  bool

	// Projects view
	projectsTable    table.Model
	projects         []ProjectDir
	selectedProjIdx  int
	projectsError    string

	// Session view
	viewMode         ViewMode
	selectedProcIdx  int
	selectedProc     *types.ClaudeProcess
	sessionTable     table.Model
	sessions         []SessionInfo
	sessionError     string
	selectedSessionIdx int

	// Session detail view
	selectedSession      *SessionInfo
	sessionStats         interface{} // Will hold *monitor.SessionStats
	messageTable         table.Model
	messages             []MessageRow
	messageError         string
	scrollOffset         int
	messageFilter        MessageFilter // Filter for messages
	filteredMessageCount int           // Count of currently filtered messages
	selectedMessageIdx   int           // Index of selected message for detail view

	// Terminal dimensions
	termWidth  int
	termHeight int

	// Message detail view
	detailMessage        *monitor.Message // Full message being displayed
	detailScrollOffset   int              // Scroll position in message detail
}



// tickMsg is used for periodic updates
type tickMsg time.Time

// processesMsg carries refreshed process data
type processesMsg struct {
	processes []types.ClaudeProcess
	err       error
}

// sessionsMsg carries loaded session data
type sessionsMsg struct {
	sessions []SessionInfo
	err      error
}

// sessionDetailMsg carries loaded session detail data
type sessionDetailMsg struct {
	stats interface{} // *monitor.SessionStats
	err   error
}

// projectsMsg carries loaded project directory data
type projectsMsg struct {
	projects []ProjectDir
	err      error
}

// NewModel creates a new UI model
func NewModel(updateInterval time.Duration, showHelpers bool) Model {
	m := Model{
		updateInterval: updateInterval,
		showHelpers:    showHelpers,
		sortColumn:     "pid",
		sortAscending:  true,
		viewMode:       ViewProcesses,
		selectedProcIdx: 0,
		messageFilter:  FilterAll,
		termWidth:      80,  // Default terminal width
		termHeight:     24,  // Default terminal height
	}

	m.table = createTableWithWidth(m.termWidth)
	m.projectsTable = createProjectsTableWithWidth(m.termWidth)
	m.sessionTable = createSessionTableWithWidth(m.termWidth)
	m.messageTable = createMessageTableWithWidth(m.termWidth)
	return m
}

// Init initializes the model and sets up background tasks
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshProcesses(),
		m.tick(),
	)
}

// refreshProcesses kicks off an asynchronous process discovery
func (m Model) refreshProcesses() tea.Cmd {
	return func() tea.Msg {
		processes, err := monitor.FindClaudeProcesses(m.showHelpers)
		return processesMsg{
			processes: processes,
			err:       err,
		}
	}
}

// tick sends a periodic timer message
func (m Model) tick() tea.Cmd {
	return tea.Tick(m.updateInterval, func(_ time.Time) tea.Msg {
		return tickMsg(time.Now())
	})
}

// loadSessions loads sessions for the currently selected process
func (m Model) loadSessions() tea.Cmd {
	if m.selectedProc == nil {
		return nil
	}

	return func() tea.Msg {
		sessions, err := monitor.FindSessionsForDirectory(m.selectedProc.WorkingDir)
		if err != nil {
			return sessionsMsg{
				err: err,
			}
		}

		// Convert to SessionInfo for display
		sessionInfos := make([]SessionInfo, len(sessions))
		for i, s := range sessions {
			sessionInfos[i] = SessionInfo{
				ID:      s.ID,
				Title:   s.GetSessionInfo(),
				Updated: s.GetSessionTime(),
				Path:    s.FilePath,
			}
		}

		return sessionsMsg{sessions: sessionInfos}
	}
}

// loadSessionDetail loads detailed stats for a session file
func (m Model) loadSessionDetail() tea.Cmd {
	if m.selectedSessionIdx < 0 || m.selectedSessionIdx >= len(m.sessions) {
		return nil
	}

	selectedSession := &m.sessions[m.selectedSessionIdx]

	return func() tea.Msg {
		stats, err := monitor.ParseSessionFile(selectedSession.Path)
		if err != nil {
			return sessionDetailMsg{
				err: err,
			}
		}

		return sessionDetailMsg{stats: stats}
	}
}

// loadProjects kicks off an asynchronous project directory loading
func (m Model) loadProjects() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.getProjectDirs()
		return projectsMsg{
			projects: projects,
			err:      err,
		}
	}
}

// getProjectDirs returns all project directories sorted by modification time (newest first)
func (m Model) getProjectDirs() ([]ProjectDir, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot get home directory: %w", err)
	}

	projectsPath := filepath.Join(home, ".claude", "projects")
	entries, err := os.ReadDir(projectsPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read projects directory: %w", err)
	}

	var projects []ProjectDir

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Count JSONL files in this directory
		sessionCount := 0
		dirPath := filepath.Join(projectsPath, entry.Name())
		sessionEntries, err := os.ReadDir(dirPath)
		if err == nil {
			for _, se := range sessionEntries {
				if !se.IsDir() && strings.HasSuffix(se.Name(), ".jsonl") {
					sessionCount++
				}
			}
		}

		// Try to get the original path from sessions-index.json
		displayName := entry.Name()
		indexPath := filepath.Join(dirPath, "sessions-index.json")
		if indexData, err := os.ReadFile(indexPath); err == nil {
			// Extract originalPath from JSON
			if origPath := extractOriginalPath(string(indexData)); origPath != "" {
				displayName = formatProjectPath(origPath, home)
			}
		}

		projects = append(projects, ProjectDir{
			Name:        entry.Name(),
			Path:        dirPath,
			DisplayName: displayName,
			Modified:    info.ModTime(),
			Sessions:    sessionCount,
		})
	}

	// Sort by modification time (newest first)
	for i := 0; i < len(projects); i++ {
		for j := i + 1; j < len(projects); j++ {
			if projects[j].Modified.After(projects[i].Modified) {
				projects[i], projects[j] = projects[j], projects[i]
			}
		}
	}

	return projects, nil
}

// extractOriginalPath extracts the originalPath value from a JSON string
func extractOriginalPath(jsonStr string) string {
	// Look for "originalPath": "..."
	// Simple string search approach
	idx := strings.Index(jsonStr, `"originalPath"`)
	if idx < 0 {
		return ""
	}

	// Find the opening quote after the colon
	colonIdx := strings.Index(jsonStr[idx:], ":")
	if colonIdx < 0 {
		return ""
	}

	quoteIdx := strings.Index(jsonStr[idx+colonIdx:], `"`)
	if quoteIdx < 0 {
		return ""
	}

	// Find the closing quote
	startIdx := idx + colonIdx + quoteIdx + 1
	endIdx := strings.Index(jsonStr[startIdx:], `"`)
	if endIdx < 0 {
		return ""
	}

	return jsonStr[startIdx : startIdx+endIdx]
}

// formatProjectPath converts an absolute path to a user-friendly display format
func formatProjectPath(path string, home string) string {
	// Replace /Users/username with ~/
	path = strings.ReplaceAll(path, home, "~")
	// Replace encoded dashes with slashes (if needed for any path separators)
	return path
}
