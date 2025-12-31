package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rootlyhq/rootly-tui/internal/api"
)

func TestNewAlertsModel(t *testing.T) {
	m := NewAlertsModel()

	if m.cursor != 0 {
		t.Errorf("expected cursor to be 0, got %d", m.cursor)
	}

	if len(m.alerts) != 0 {
		t.Errorf("expected no alerts initially, got %d", len(m.alerts))
	}
}

func TestAlertsModelSetAlerts(t *testing.T) {
	m := NewAlertsModel()
	alerts := api.MockAlerts()

	m.SetAlerts(alerts)

	if len(m.alerts) != len(alerts) {
		t.Errorf("expected %d alerts, got %d", len(alerts), len(m.alerts))
	}

	if m.loading {
		t.Error("expected loading to be false after SetAlerts")
	}
}

func TestAlertsModelSetLoading(t *testing.T) {
	m := NewAlertsModel()

	m.SetLoading(true)
	if !m.loading {
		t.Error("expected loading to be true")
	}

	m.SetLoading(false)
	if m.loading {
		t.Error("expected loading to be false")
	}
}

func TestAlertsModelSetError(t *testing.T) {
	m := NewAlertsModel()

	m.SetError("test error")

	if m.error != "test error" {
		t.Errorf("expected error 'test error', got '%s'", m.error)
	}

	if m.loading {
		t.Error("expected loading to be false after SetError")
	}
}

func TestAlertsModelSelectedAlert(t *testing.T) {
	m := NewAlertsModel()
	alerts := api.MockAlerts()
	m.SetAlerts(alerts)

	selected := m.SelectedAlert()
	if selected == nil {
		t.Fatal("expected selected alert to be non-nil")
	}

	if selected.ID != alerts[0].ID {
		t.Errorf("expected selected ID '%s', got '%s'", alerts[0].ID, selected.ID)
	}
}

func TestAlertsModelNavigation(t *testing.T) {
	m := NewAlertsModel()
	alerts := api.MockAlerts()
	m.SetAlerts(alerts)
	m.SetDimensions(100, 30)

	// Test move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after 'j', got %d", m.cursor)
	}

	// Test move up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after 'k', got %d", m.cursor)
	}

	// Test move up at top
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 at top, got %d", m.cursor)
	}

	// Test go to bottom
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if m.cursor != len(alerts)-1 {
		t.Errorf("expected cursor %d at bottom, got %d", len(alerts)-1, m.cursor)
	}

	// Test go to top
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 at top, got %d", m.cursor)
	}
}

func TestAlertsModelView(t *testing.T) {
	m := NewAlertsModel()
	m.SetDimensions(100, 30)

	// Test empty view
	view := m.View()
	if !strings.Contains(view, "No alerts found") {
		t.Error("expected 'No alerts found' in empty view")
	}

	// Test loading view
	m.SetLoading(true)
	view = m.View()
	if !strings.Contains(view, "Loading") {
		t.Error("expected 'Loading' in loading view")
	}

	// Test error view
	m.SetLoading(false)
	m.SetError("API error")
	view = m.View()
	if !strings.Contains(view, "API error") {
		t.Error("expected error message in view")
	}

	// Test with alerts
	m.SetAlerts(api.MockAlerts())
	view = m.View()
	if !strings.Contains(view, "ALERTS") {
		t.Error("expected 'ALERTS' in view with data")
	}
}

func TestAlertsModelCursorBounds(t *testing.T) {
	m := NewAlertsModel()
	alerts := api.MockAlerts()
	m.SetAlerts(alerts)
	m.SetDimensions(100, 30)

	// Move cursor beyond bounds
	for i := 0; i < len(alerts)+5; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}

	if m.cursor >= len(alerts) {
		t.Errorf("cursor exceeded bounds: got %d, max should be %d", m.cursor, len(alerts)-1)
	}
}
