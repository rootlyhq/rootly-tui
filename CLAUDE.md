# Rootly TUI

A terminal user interface for viewing Rootly incidents and alerts.

## Tech Stack

- **Go 1.24+**
- **Bubble Tea** - TUI framework (Elm Architecture)
- **Bubbles** - UI components (text input, spinner)
- **Lipgloss** - Terminal styling
- **Log** - Structured logging (charmbracelet/log)
- **rootly-go** - Rootly API client

## Project Structure

```
cmd/rootly-tui/main.go      # Entry point
internal/
  api/client.go             # Rootly API wrapper
  app/                      # Main application model
    app.go                  # tea.Model implementation
    keymap.go               # Keybindings
    messages.go             # Message types
  config/config.go          # Config management (~/.rootly-tui/config.yaml)
  debug/debug.go            # Debug logging utilities
  styles/styles.go          # Lipgloss styles
  views/                    # UI views
    setup.go                # API key setup screen
    incidents.go            # Incidents split view
    alerts.go               # Alerts split view
    help.go                 # Help overlay
```

## Build Commands

```bash
make build    # Build binary to bin/rootly-tui
make run      # Build and run
make dev      # Run without building (go run)
make test     # Run tests
make lint     # Run golangci-lint
```

## Architecture

The app follows the Elm Architecture pattern:
- **Model**: Application state in `internal/app/app.go`
- **Update**: Message handling in `Update()` method
- **View**: Rendering in `View()` method

### Screens
1. **Setup Screen**: First-run API key configuration
2. **Main Screen**: Tabbed view with Incidents and Alerts

### Message Flow
- `IncidentsLoadedMsg` / `AlertsLoadedMsg` - API responses
- `APIKeyValidatedMsg` / `ConfigSavedMsg` - Setup flow
- `tea.KeyMsg` - Keyboard input
- `tea.WindowSizeMsg` - Terminal resize

## Config

Config stored at `~/.rootly-tui/config.yaml`:
```yaml
api_key: "your-api-key"
endpoint: "api.rootly.com"
```

## Key Bindings

- `j/k` or arrows: Navigate
- `Tab`: Switch tabs (Incidents/Alerts)
- `r`: Refresh data
- `?`: Help
- `q`: Quit

## Debug Mode

Enable debug logging to troubleshoot API issues:

```bash
# Debug to stderr
rootly-tui --debug

# Debug to file
rootly-tui --log debug.log

# Debug to stderr, redirect to file
rootly-tui --debug 2> debug.log
```

Debug logs include:
- API requests (method, URL)
- Response status and body length
- JSON parsing errors with prettified response body
- Configuration details

The debug package (`internal/debug/debug.go`) provides:
- `debug.Logger` - Global structured logger
- `debug.Enable()` - Turn on debug logging
- `debug.PrettyJSON()` - Format JSON for readable output
- `debug.SetLogFile()` - Write logs to file
