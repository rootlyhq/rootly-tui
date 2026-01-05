# Rootly TUI

A terminal user interface for viewing Rootly incidents and alerts.

![Go Version](https://img.shields.io/github/go-mod/go-version/rootlyhq/rootly-tui)
![License](https://img.shields.io/github/license/rootlyhq/rootly-tui)
![Release](https://img.shields.io/github/v/release/rootlyhq/rootly-tui)

## Features

- View and navigate incidents with full details
- View and navigate alerts with full details
- Split-pane interface with list and detail views
- Press Enter to load extended details (roles, causes, responders, etc.)
- Keyboard-driven navigation
- Configurable API endpoint (supports self-hosted Rootly)
- Internationalization with 12 supported languages
- Persistent caching for faster startup
- In-app debug log viewer

## Supported Languages

- English (US/UK)
- Spanish (Espanol)
- French (Francais)
- German (Deutsch)
- Chinese Simplified (简体中文)
- Hindi (हिन्दी)
- Arabic (العربية)
- Bengali (বাংলা)
- Portuguese Brazilian (Portugues)
- Russian (Русский)
- Japanese (日本語)

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

On first run, you'll be prompted to enter your Rootly API credentials, select your timezone, and choose your preferred language.

Configuration is stored at `~/.rootly-tui/config.yaml`:

```yaml
api_key: "your-api-key"
endpoint: "api.rootly.com"  # Optional: defaults to api.rootly.com
timezone: "America/Los_Angeles"
language: "en_US"
```

### Getting an API Key

1. Log in to your Rootly account
2. Navigate to **Settings** > **API Keys**
3. Create a new API key with read permissions for incidents and alerts

## Usage

```bash
# Run the TUI
rootly-tui

# Check version
rootly-tui --version

# Enable debug logging (outputs to stderr)
rootly-tui --debug

# Write debug logs to a file
rootly-tui --log debug.log
```

### Debug Mode

Debug mode logs API requests, responses, and parsing details. This is useful for troubleshooting connection issues or unexpected behavior.

```bash
# Debug to stderr (visible after exiting TUI)
rootly-tui --debug 2> debug.log

# Or write directly to a file
rootly-tui --log debug.log
```

Debug logs include:
- API endpoint configuration
- HTTP request method and URL
- Response status codes and body length
- JSON parsing results and errors (with prettified JSON)

### In-App Log Viewer

Press `l` at any time to open the in-app log viewer. Logs are always captured in memory (up to 1000 entries) even without `--debug` mode.

Log viewer controls:
- `j/k` - Scroll up/down
- `g/G` - Jump to top/bottom
- `a` - Select all logs
- `y` - Copy selected logs to clipboard
- `c` - Clear logs
- `l` or `Esc` - Close viewer

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `g` | Go to first item |
| `G` | Go to last item |
| `h` / `←` | Previous page |
| `l` / `→` | Next page (when not in logs view) |
| `Tab` | Switch between Incidents and Alerts |
| `Enter` | Load detailed view for selected item |
| `o` | Open item URL in browser |
| `r` | Refresh data |
| `l` | View debug logs |
| `s` | Open setup screen |
| `?` | Toggle help |
| `q` / `Esc` | Quit (or return from overlay/setup) |

## Screenshots

```
┌─────────────────────────────────────────────────────────────┐
│  Rootly                          [Incidents] Alerts   v0.1.0│
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────────┬──────────────────────────────────┐│
│  │ INCIDENTS            │ Incident Details                 ││
│  │                      │                                  ││
│  │ ████ CRIT INC-123    │ Database Connection Failure      ││
│  │   in_progress        │ Status: In Progress              ││
│  │ ███  HIGH INC-122    │ Severity: Critical               ││
│  │   resolved           │                                  ││
│  │ ██   MED  INC-121    │ Timeline:                        ││
│  │   resolved           │ Started: 10:30 AM                ││
│  │                      │ Detected: 10:32 AM               ││
│  │                      │                                  ││
│  │                      │ Services: api, database          ││
│  │                      │ Teams: Platform, SRE             ││
│  │                      │                                  ││
│  │                      │ Press Enter for more details     ││
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
make coverage   # Run tests with coverage report
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
- [Log](https://github.com/charmbracelet/log) - Structured logging
- [go-i18n](https://github.com/nicksnyder/go-i18n) - Internationalization
- [BoltDB](https://github.com/etcd-io/bbolt) - Persistent cache
- [rootly-go](https://github.com/rootlyhq/rootly-go) - Rootly API client

## License

MIT License - see [LICENSE](LICENSE.txt) for details.
