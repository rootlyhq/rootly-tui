# Rootly TUI

A terminal user interface for viewing Rootly incidents and alerts.

![Go Version](https://img.shields.io/github/go-mod/go-version/rootlyhq/rootly-tui)
![License](https://img.shields.io/github/license/rootlyhq/rootly-tui)
![Release](https://img.shields.io/github/v/release/rootlyhq/rootly-tui)

## Features

- View and navigate incidents with full details
- View and navigate alerts with full details
- Split-pane interface with list and detail views
- Keyboard-driven navigation
- Configurable API endpoint (supports self-hosted Rootly)

## Installation

### Homebrew (macOS/Linux)

```bash
brew install rootlyhq/tap/rootly-tui
```

### Go Install

```bash
go install github.com/rootlyhq/rootly-tui/cmd/rootly-tui@latest
```

### Download Binary

Download the latest release from the [Releases](https://github.com/rootlyhq/rootly-tui/releases) page.

### Build from Source

```bash
git clone https://github.com/rootlyhq/rootly-tui.git
cd rootly-tui
make build
./bin/rootly-tui
```

## Configuration

On first run, you'll be prompted to enter your Rootly API credentials.

Configuration is stored at `~/.rootly-tui/config.yaml`:

```yaml
api_key: "your-api-key"
endpoint: "api.rootly.com"  # Optional: defaults to api.rootly.com
```

### Getting an API Key

1. Log in to your Rootly account
2. Navigate to **Settings** > **API Keys**
3. Create a new API key with read permissions

## Usage

```bash
# Run the TUI
rootly-tui

# Check version
rootly-tui --version
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `g` | Go to first item |
| `G` | Go to last item |
| `Tab` | Switch between Incidents and Alerts |
| `r` | Refresh data |
| `?` | Toggle help |
| `q` / `Ctrl+C` | Quit |

## Screenshots

```
┌─────────────────────────────────────────────────────────────┐
│  Rootly TUI                    [Incidents] Alerts     v0.1.0│
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────────┬──────────────────────────────────┐│
│  │ INCIDENTS            │ Incident Details                 ││
│  │                      │                                  ││
│  │ ● CRIT Database down │ Database Connection Failure      ││
│  │   HIGH API latency   │ Status: In Progress              ││
│  │   MED  Deploy failed │ Severity: Critical               ││
│  │                      │                                  ││
│  │                      │ Started: 10:30 AM                ││
│  │                      │ Detected: 10:32 AM               ││
│  │                      │                                  ││
│  │                      │ Services: api, database          ││
│  │                      │ Teams: Platform, SRE             ││
│  └──────────────────────┴──────────────────────────────────┘│
├─────────────────────────────────────────────────────────────┤
│  j/k navigate  Tab switch  r refresh  ? help  q quit       │
└─────────────────────────────────────────────────────────────┘
```

## Development

### Prerequisites

- Go 1.24+
- Make

### Build

```bash
make build      # Build binary
make run        # Build and run
make dev        # Run with go run
```

### Test

```bash
make test       # Run tests
make lint       # Run linter
make check      # Format, lint, and test
```

### Release

Releases are automated via GoReleaser when a new tag is pushed:

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [rootly-go](https://github.com/rootlyhq/rootly-go) - Rootly API client

## License

MIT License - see [LICENSE](LICENSE) for details.
