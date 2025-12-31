package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rootlyhq/rootly-tui/internal/api"
)

func TestNewIncidentsModel(t *testing.T) {
	m := NewIncidentsModel()

	if m.cursor != 0 {
		t.Errorf("expected cursor to be 0, got %d", m.cursor)
	}

	if len(m.incidents) != 0 {
		t.Errorf("expected no incidents initially, got %d", len(m.incidents))
	}
}

func TestIncidentsModelSetIncidents(t *testing.T) {
	m := NewIncidentsModel()
	incidents := api.MockIncidents()
	pagination := api.PaginationInfo{CurrentPage: 1, HasNext: true, HasPrev: false}

	m.SetIncidents(incidents, pagination)

	if len(m.incidents) != len(incidents) {
		t.Errorf("expected %d incidents, got %d", len(incidents), len(m.incidents))
	}

	if m.loading {
		t.Error("expected loading to be false after SetIncidents")
	}

	if m.currentPage != 1 {
		t.Errorf("expected currentPage 1, got %d", m.currentPage)
	}
}

func TestIncidentsModelSetLoading(t *testing.T) {
	m := NewIncidentsModel()

	m.SetLoading(true)
	if !m.loading {
		t.Error("expected loading to be true")
	}

	m.SetLoading(false)
	if m.loading {
		t.Error("expected loading to be false")
	}
}

func TestIncidentsModelSetError(t *testing.T) {
	m := NewIncidentsModel()

	m.SetError("test error")

	if m.error != "test error" {
		t.Errorf("expected error 'test error', got '%s'", m.error)
	}

	if m.loading {
		t.Error("expected loading to be false after SetError")
	}
}

func TestIncidentsModelSelectedIncident(t *testing.T) {
	m := NewIncidentsModel()
	incidents := api.MockIncidents()
	m.SetIncidents(incidents, api.PaginationInfo{CurrentPage: 1})

	// Test initial selection
	selected := m.SelectedIncident()
	if selected == nil {
		t.Fatal("expected selected incident to be non-nil")
	}

	if selected.ID != incidents[0].ID {
		t.Errorf("expected selected ID '%s', got '%s'", incidents[0].ID, selected.ID)
	}
}

func TestIncidentsModelNavigation(t *testing.T) {
	m := NewIncidentsModel()
	incidents := api.MockIncidents()
	m.SetIncidents(incidents, api.PaginationInfo{CurrentPage: 1})
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

	// Test move up at top (should stay at 0)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 at top, got %d", m.cursor)
	}

	// Test go to bottom
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if m.cursor != len(incidents)-1 {
		t.Errorf("expected cursor %d at bottom, got %d", len(incidents)-1, m.cursor)
	}

	// Test go to top
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 at top, got %d", m.cursor)
	}
}

func TestIncidentsModelView(t *testing.T) {
	m := NewIncidentsModel()
	m.SetDimensions(100, 30)

	// Test empty view
	view := m.View()
	if !strings.Contains(view, "No incidents found") {
		t.Error("expected 'No incidents found' in empty view")
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

	// Test with incidents
	m.SetIncidents(api.MockIncidents(), api.PaginationInfo{CurrentPage: 1})
	view = m.View()
	if !strings.Contains(view, "INCIDENTS") {
		t.Error("expected 'INCIDENTS' in view with data")
	}
}

func TestIncidentsModelCursorBounds(t *testing.T) {
	m := NewIncidentsModel()
	incidents := api.MockIncidents()
	m.SetIncidents(incidents, api.PaginationInfo{CurrentPage: 1})
	m.SetDimensions(100, 30)

	// Move cursor to last item
	for i := 0; i < len(incidents)+5; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}

	// Cursor should not exceed last index
	if m.cursor >= len(incidents) {
		t.Errorf("cursor exceeded bounds: got %d, max should be %d", m.cursor, len(incidents)-1)
	}
}
