package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
	"github.com/thies/claudewatch/internal/monitor"
	"github.com/thies/claudewatch/internal/types"
)

// Model represents the main UI state
type Model struct {
	table          table.Model
	processes      []types.ClaudeProcess
	lastUpdate     time.Time
	updateInterval time.Duration
	showHelpers    bool
	quitting       bool
	sortColumn     string
	sortAscending  bool
}

// tickMsg is used for periodic updates
type tickMsg time.Time

// processesMsg carries refreshed process data
type processesMsg struct {
	processes []types.ClaudeProcess
	err       error
}

// NewModel creates a new UI model
func NewModel(updateInterval time.Duration, showHelpers bool) Model {
	m := Model{
		updateInterval: updateInterval,
		showHelpers:    showHelpers,
		sortColumn:     "pid",
		sortAscending:  true,
	}

	m.table = createTable()
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
