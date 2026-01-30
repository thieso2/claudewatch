# claudewatch

A real-time TUI monitor for Claude Code CLI instances on macOS. Display CPU usage, memory consumption, working directory, and process information for all running Claude instances.

## Features

- **Real-time monitoring** of CPU and memory usage for Claude instances
- **Working directory detection** using macOS `proc_pidinfo` system call via CGo
- **Interactive table view** with sortable columns
- **MCP helper filtering** to toggle visibility of helper processes
- **Automatic refresh** with configurable intervals
- **Color-coded metrics** for quick visual scanning

## Installation

### Prerequisites

- macOS 10.14 or later
- Go 1.23 or later
- Xcode Command Line Tools (for CGo compilation)

### Build from source

```bash
git clone https://github.com/thies/claudewatch.git
cd claudewatch
go build -o bin/claudewatch ./cmd/claudewatch
```

### Install globally

```bash
go install ./cmd/claudewatch
```

This will install the binary to `$GOPATH/bin/claudewatch` (typically `~/go/bin/claudewatch`).

## Usage

### Basic usage

```bash
./bin/claudewatch
```

### Command-line options

```bash
claudewatch [flags]

Flags:
  -interval duration
    	Refresh interval (default 1s)
  -show-helpers
    	Show MCP helper processes (default false)
```

### Examples

```bash
# Monitor Claude instances with 2-second refresh interval
claudewatch --interval 2s

# Show all instances including MCP helpers
claudewatch --show-helpers

# Combine options
claudewatch --interval 500ms --show-helpers
```

## Keyboard shortcuts

| Key | Action |
|-----|--------|
| `↑` / `k` | Navigate up in process list |
| `↓` / `j` | Navigate down in process list |
| `r` | Manual refresh |
| `f` | Toggle MCP helper processes visibility |
| `q` / `Ctrl+C` | Quit |

## Display columns

- **PID**: Process ID
- **CPU%**: CPU usage percentage
- **MEM**: Memory usage (MB or GB)
- **UPTIME**: Process runtime (days/hours or hours/minutes)
- **WORKDIR**: Current working directory (truncated with ~ for home)
- **COMMAND**: Full command line

## Architecture

### Directory structure

```
claudewatch/
├── cmd/claudewatch/
│   └── main.go                 # Entry point and CLI setup
├── internal/
│   ├── monitor/
│   │   ├── process.go          # Process discovery and filtering
│   │   ├── metrics.go          # Metrics collection utilities
│   │   └── workdir_darwin.go   # macOS proc_pidinfo CGo wrapper
│   ├── ui/
│   │   ├── model.go            # Bubbletea state management
│   │   ├── update.go           # Event handling
│   │   ├── view.go             # UI rendering
│   │   └── table.go            # Table configuration
│   └── types/
│       └── process.go          # ClaudeProcess data structure
└── README.md
```

### Technology stack

- **Bubbletea**: Elm-inspired TUI framework
- **bubble-table**: Sortable table component
- **Lipgloss**: Terminal styling and layout
- **gopsutil v4**: Process metrics collection
- **CGo**: System-level working directory detection

## Implementation notes

### Process detection

Processes are identified as Claude instances if:
- The executable path contains "claude"
- Path matches `/opt/homebrew/*/claude`
- NOT the desktop app (`Claude.app`)

MCP helper processes are identified by the `--claude-in-chrome-mcp` flag in the command line.

### Working directory detection

On macOS, working directory is retrieved using:
1. CGo wrapper around `proc_pidinfo` system call
2. `PROC_PIDVNODEPATHINFO` to get current directory
3. More reliable than alternative methods

If permission is denied, displays "[Permission Denied]" instead of the path.

### Metrics refresh strategy

- **CPU**: Updated every refresh interval (default 1s)
- **Memory**: Updated every refresh interval
- **Working Directory**: Updated every refresh interval

The first CPU call may show "..." as it establishes a baseline.

## Building for distribution

```bash
# Build optimized binary
go build -ldflags="-s -w" -o bin/claudewatch ./cmd/claudewatch

# Build with version info
VERSION=$(git describe --tags --always)
go build -ldflags="-s -w -X main.Version=$VERSION" -o bin/claudewatch ./cmd/claudewatch
```

## Troubleshooting

### Build fails with CGo errors

Ensure Xcode Command Line Tools are installed:
```bash
xcode-select --install
```

### No Claude instances appear

1. Check that Claude CLI is actually running: `ps aux | grep claude`
2. Try manual refresh with `r` key
3. Use `--show-helpers` to see MCP processes

### Permission denied for working directory

Some processes may not allow working directory access. This is expected and displays as "[Permission Denied]".

## Future enhancements

- [ ] Historical metrics graphs
- [ ] Process tree view (parent-child relationships)
- [ ] Export metrics to JSON
- [ ] Alert on process crash
- [ ] Configuration file support (~/.claudewatchrc)
- [ ] Linux support with alternative working dir detection
- [ ] Persistent sorting/filter preferences

## Dependencies

```
github.com/charmbracelet/bubbletea v1.3.10
github.com/charmbracelet/lipgloss v1.1.0
github.com/evertras/bubble-table v0.19.2
github.com/shirou/gopsutil/v4 v4.25.12
```

## License

MIT

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request
