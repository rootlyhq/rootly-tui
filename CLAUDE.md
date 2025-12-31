# Rootly TUI

A terminal user interface for viewing Rootly incidents and alerts.

## Tech Stack

- **Go 1.24+**
- **Bubble Tea** - TUI framework (Elm Architecture)
- **Bubbles** - UI components (text input, spinner)
- **Lipgloss** - Terminal styling
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
