package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

// createTable initializes the bubble-table model with columns and styling
func createTable() table.Model {
	columns := []table.Column{
		table.NewColumn("pid", "PID", 8),
		table.NewColumn("cpu", "CPU%", 10),
		table.NewColumn("mem", "MEM", 12),
		table.NewColumn("uptime", "UPTIME", 12),
		table.NewColumn("workdir", "WORKDIR", 60),
		table.NewColumn("cmd", "COMMAND", 100),
	}

	t := table.New(columns).
		WithPageSize(20).
		WithBaseStyle(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")),
		).
		Focused(true)

	return t
}

// createTableWithWidth creates a table with columns sized for the given width
func createTableWithWidth(width int) table.Model {
	// Use fixed large widths that will fill most screens
	return createTable()
}

// styleHighCPU applies red styling to high CPU values
func styleHighCPU(cpu string) string {
	// This would be applied in view.go when rendering
	return cpu
}

// styleWarningMemory applies yellow styling to high memory values
func styleWarningMemory(mem string) string {
	// This would be applied in view.go when rendering
	return mem
}

// createSessionTable initializes the session table
func createSessionTable() table.Model {
	columns := []table.Column{
		table.NewColumn("id", "SESSION ID", 40),
		table.NewColumn("title", "TITLE", 80),
		table.NewColumn("updated", "UPDATED", 20),
	}

	t := table.New(columns).
		WithPageSize(20).
		WithBaseStyle(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")),
		).
		Focused(true)

	return t
}

// createSessionTableWithWidth creates a session table with columns sized for the given width
func createSessionTableWithWidth(width int) table.Model {
	// Use fixed large widths that will fill most screens
	return createSessionTable()
}

// createMessageTable initializes the message table
func createMessageTable() table.Model {
	columns := []table.Column{
		table.NewColumn("role", "ROLE", 12),
		table.NewColumn("content", "MESSAGE", 200),
		table.NewColumn("time", "TIME", 12),
	}

	t := table.New(columns).
		WithPageSize(15).
		WithBaseStyle(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")),
		).
		Focused(true)

	return t
}

// createMessageTableWithWidth creates a message table with columns sized for the given width
func createMessageTableWithWidth(width int) table.Model {
	// Use fixed large widths that will fill most screens
	return createMessageTable()
}
