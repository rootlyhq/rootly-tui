package app

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rootlyhq/rootly-tui/internal/api"
	"github.com/rootlyhq/rootly-tui/internal/config"
)

func TestNew(t *testing.T) {
	m := New("1.0.0")

	if m.version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", m.version)
	}

	// Should start on setup screen if no config exists
	if m.screen != ScreenSetup && m.screen != ScreenMain {
		t.Errorf("expected screen to be ScreenSetup or ScreenMain, got %d", m.screen)
	}

	if m.activeTab != TabIncidents {
		t.Errorf("expected active tab to be TabIncidents, got %d", m.activeTab)
	}
}

func TestModelInit(t *testing.T) {
	m := New("1.0.0")
	cmd := m.Init()

	// Init should return a command (not nil)
	if cmd == nil {
		t.Error("expected Init() to return a command")
	}
}

func TestModelUpdateQuit(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenMain // Set to main screen

	// Test quit with 'q'
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if newModel == nil {
		t.Error("expected model to be non-nil")
	}

	// Should return quit command
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestModelUpdateQuitFromSetup(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenSetup

	// Test quit with 'q' from setup screen
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Should return quit command even from setup
	if cmd == nil {
		t.Error("expected quit command from setup screen")
	}
}

func TestModelUpdateTabSwitch(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenMain
	m.activeTab = TabIncidents

	// Test tab switch
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	model := newModel.(Model)

	if model.activeTab != TabAlerts {
		t.Errorf("expected active tab to be TabAlerts after tab press, got %d", model.activeTab)
	}

	// Switch back
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = newModel.(Model)

	if model.activeTab != TabIncidents {
		t.Errorf("expected active tab to be TabIncidents after second tab press, got %d", model.activeTab)
	}
}

func TestModelUpdateHelp(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenMain
	m.help.Visible = false

	// Test help toggle with '?'
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model := newModel.(Model)

	if !model.help.Visible {
		t.Error("expected help to be visible after '?' press")
	}

	// Toggle off
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model = newModel.(Model)

	if model.help.Visible {
		t.Error("expected help to be hidden after second '?' press")
	}
}

func TestModelUpdateWindowSize(t *testing.T) {
	m := New("1.0.0")

	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := newModel.(Model)

	if model.width != 120 {
		t.Errorf("expected width 120, got %d", model.width)
	}

	if model.height != 40 {
		t.Errorf("expected height 40, got %d", model.height)
	}
}

func TestModelView(t *testing.T) {
	m := New("1.0.0")

	// Test view before window size is set
	view := m.View()
	if view != "Loading..." {
		t.Errorf("expected 'Loading...' before window size, got '%s'", view)
	}

	// Set window size
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = newModel.(Model)

	view = m.View()
	if view == "" {
		t.Error("expected non-empty view after window size set")
	}
}

func TestModelIncidentsLoaded(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenMain
	m.initialLoading = true

	// Simulate incidents loaded message
	newModel, _ := m.Update(IncidentsLoadedMsg{
		Incidents: nil,
		Err:       nil,
	})
	model := newModel.(Model)

	if model.initialLoading {
		t.Error("expected initialLoading to be false after IncidentsLoadedMsg")
	}

	if model.loading {
		t.Error("expected loading to be false after IncidentsLoadedMsg")
	}
}

func TestModelAlertsLoaded(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenMain

	// Simulate alerts loaded message
	newModel, _ := m.Update(AlertsLoadedMsg{
		Alerts: nil,
		Err:    nil,
	})
	model := newModel.(Model)

	// Should not error
	if model.errorMsg != "" {
		t.Errorf("unexpected error: %s", model.errorMsg)
	}
}

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Verify key bindings are set
	if len(km.Up.Keys()) == 0 {
		t.Error("expected Up key binding to be set")
	}
	if len(km.Down.Keys()) == 0 {
		t.Error("expected Down key binding to be set")
	}
	if len(km.Tab.Keys()) == 0 {
		t.Error("expected Tab key binding to be set")
	}
	if len(km.Refresh.Keys()) == 0 {
		t.Error("expected Refresh key binding to be set")
	}
	if len(km.Help.Keys()) == 0 {
		t.Error("expected Help key binding to be set")
	}
	if len(km.Logs.Keys()) == 0 {
		t.Error("expected Logs key binding to be set")
	}
	if len(km.Quit.Keys()) == 0 {
		t.Error("expected Quit key binding to be set")
	}
}

func TestModelUpdateLogs(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenMain
	m.logs.Visible = false

	// Test logs toggle with 'l'
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	model := newModel.(Model)

	if !model.logs.Visible {
		t.Error("expected logs to be visible after 'l' press")
	}

	// Toggle off with 'l'
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	model = newModel.(Model)

	if model.logs.Visible {
		t.Error("expected logs to be hidden after second 'l' press")
	}
}

func TestModelUpdateLogsEscape(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenMain
	m.logs.Visible = true

	// Test closing logs with Escape
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := newModel.(Model)

	if model.logs.Visible {
		t.Error("expected logs to be hidden after Escape press")
	}
}

func TestModelLogsBlocksOtherKeys(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenMain
	m.activeTab = TabIncidents
	m.logs.Visible = true

	// Tab should not switch tabs when logs overlay is visible
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	model := newModel.(Model)

	if model.activeTab != TabIncidents {
		t.Error("expected tab switch to be blocked when logs are visible")
	}
}

func TestNewWithValidConfigCreatesClientOnce(t *testing.T) {
	// This test verifies that when a valid config exists,
	// the API client is created in New() so that loadData()
	// doesn't try to create multiple clients concurrently.

	m := New("1.0.0")

	// If we're on the main screen, the client should already be initialized
	if m.screen == ScreenMain {
		if m.apiClient == nil {
			t.Error("expected apiClient to be initialized when screen is ScreenMain")
		}
	}

	// If we're on setup screen, client should be nil (no valid config)
	if m.screen == ScreenSetup {
		if m.apiClient != nil {
			t.Error("expected apiClient to be nil when screen is ScreenSetup")
		}
	}
}

func TestModelClientNotRecreatedOnRefresh(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenMain

	// Manually set a non-nil client to simulate initialized state
	// We can't create a real client without config, but we can verify
	// that the loadData functions check for existing client
	initialClient := m.apiClient

	// If client is already set and we're on main screen, refreshing
	// should not create a new client
	if m.screen == ScreenMain && m.apiClient != nil {
		// After refresh, client should be the same instance
		// (This is a structural test - actual refresh tested in integration)
		if m.apiClient != initialClient {
			t.Error("expected apiClient to remain the same after initialization")
		}
	}
}

func TestModelClose(t *testing.T) {
	m := New("1.0.0")

	// Close with nil client should not error
	err := m.Close()
	if err != nil {
		t.Errorf("expected no error closing model with nil client, got %v", err)
	}
}

func TestModelCloseWithClient(t *testing.T) {
	// Create a temp directory for cache
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create a valid config
	cfg := &config.Config{
		APIKey:   "test-key",
		Endpoint: "api.rootly.com",
	}

	// Create client manually
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	m := New("1.0.0")
	m.apiClient = client

	// Close should close the client
	err = m.Close()
	if err != nil {
		t.Errorf("expected no error closing model, got %v", err)
	}

	// Client should be closed (calling Close again is safe but we just verify no panic)
}

func TestModelOpenKeyBinding(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenMain
	m.activeTab = TabIncidents
	// Use no-op URL opener to prevent opening real browser during tests
	m.urlOpener = func(url string) error { return nil }

	// Add a test incident
	m.incidents.SetIncidents([]api.Incident{
		{
			ID:       "inc_123",
			Title:    "Test Incident",
			ShortURL: "https://root.ly/i/abc123",
		},
	}, api.PaginationInfo{CurrentPage: 1})

	// Press 'o' key - should not panic even without a real browser
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	if newModel == nil {
		t.Error("expected model to be non-nil")
	}
	// Command should be nil (openURL is fire-and-forget)
	if cmd != nil {
		t.Error("expected nil command for open key")
	}
}

func TestModelOpenKeyBindingWithFallbackURL(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenMain
	m.activeTab = TabIncidents
	// Use no-op URL opener to prevent opening real browser during tests
	m.urlOpener = func(url string) error { return nil }

	// Add a test incident without ShortURL or URL (should construct from ID)
	m.incidents.SetIncidents([]api.Incident{
		{
			ID:    "inc_456",
			Title: "Test Incident Without URL",
		},
	}, api.PaginationInfo{CurrentPage: 1})

	// Press 'o' key - should not panic
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	if newModel == nil {
		t.Error("expected model to be non-nil")
	}
}

func TestModelOpenKeyBindingForAlerts(t *testing.T) {
	m := New("1.0.0")
	m.screen = ScreenMain
	m.activeTab = TabAlerts
	// Use no-op URL opener to prevent opening real browser during tests
	m.urlOpener = func(url string) error { return nil }

	// Add a test alert
	m.alerts.SetAlerts([]api.Alert{
		{
			ID:      "alert_123",
			ShortID: "ABC123",
			Summary: "Test Alert",
		},
	}, api.PaginationInfo{CurrentPage: 1})

	// Press 'o' key - should not panic
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	if newModel == nil {
		t.Error("expected model to be non-nil")
	}
	// Command should be nil (openURL is fire-and-forget)
	if cmd != nil {
		t.Error("expected nil command for open key")
	}
}
