package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// View renders the UI
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if len(m.processes) == 0 {
		return m.renderEmpty()
	}

	return m.renderWithTable()
}

// renderEmpty displays a message when no processes are found
func (m Model) renderEmpty() string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("11")).
		Render("claudewatch")

	content := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("No Claude instances found.")

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("Press 'r' to refresh or 'q' to quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

// renderWithTable displays the full UI with the process table
func (m Model) renderWithTable() string {
	// Header with title and status
	headerTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("11")).
		Render("claudewatch")

	status := fmt.Sprintf("%d instances", len(m.processes))
	if m.showHelpers {
		status += " (including helpers)"
	}
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))
	statusText := statusStyle.Render(status)

	timestamp := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render(fmt.Sprintf("Updated: %s", m.lastUpdate.Format("15:04:05")))

	headerLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		headerTitle,
		"  ",
		statusText,
		"  |  ",
		timestamp,
	)

	// Table
	tableView := m.table.View()

	// Footer with help text
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	helpText := "↑/↓: Navigate  |  r: Refresh  |  f: Toggle helpers  |  q: Quit"
	footer := footerStyle.Render(helpText)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine,
		"",
		tableView,
		"",
		footer,
	)
}
