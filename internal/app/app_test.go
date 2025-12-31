package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
