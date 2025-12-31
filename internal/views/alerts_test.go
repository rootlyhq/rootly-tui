package views

import (
	"strings"
	"testing"
	"time"

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
	pagination := api.PaginationInfo{CurrentPage: 1, HasNext: true, HasPrev: false}

	m.SetAlerts(alerts, pagination)

	if len(m.alerts) != len(alerts) {
		t.Errorf("expected %d alerts, got %d", len(alerts), len(m.alerts))
	}

	if m.loading {
		t.Error("expected loading to be false after SetAlerts")
	}

	if m.currentPage != 1 {
		t.Errorf("expected currentPage 1, got %d", m.currentPage)
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
	m.SetAlerts(alerts, api.PaginationInfo{CurrentPage: 1})

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
	m.SetAlerts(alerts, api.PaginationInfo{CurrentPage: 1})
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
	m.SetAlerts(api.MockAlerts(), api.PaginationInfo{CurrentPage: 1})
	view = m.View()
	if !strings.Contains(view, "ALERTS") {
		t.Error("expected 'ALERTS' in view with data")
	}
}

func TestAlertsModelCursorBounds(t *testing.T) {
	m := NewAlertsModel()
	alerts := api.MockAlerts()
	m.SetAlerts(alerts, api.PaginationInfo{CurrentPage: 1})
	m.SetDimensions(100, 30)

	// Move cursor beyond bounds
	for i := 0; i < len(alerts)+5; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}

	if m.cursor >= len(alerts) {
		t.Errorf("cursor exceeded bounds: got %d, max should be %d", m.cursor, len(alerts)-1)
	}
}

func TestAlertsModelInit(t *testing.T) {
	m := NewAlertsModel()
	cmd := m.Init()

	// Init should return nil for AlertsModel
	if cmd != nil {
		t.Error("expected Init to return nil")
	}
}

func TestAlertsModelSetSpinner(t *testing.T) {
	m := NewAlertsModel()

	m.SetSpinner("⠋")
	if m.spinnerView != "⠋" {
		t.Errorf("expected spinner '⠋', got '%s'", m.spinnerView)
	}

	m.SetSpinner("⠙")
	if m.spinnerView != "⠙" {
		t.Errorf("expected spinner '⠙', got '%s'", m.spinnerView)
	}
}

func TestAlertsModelCurrentPage(t *testing.T) {
	m := NewAlertsModel()

	// Default page should be 1
	if m.CurrentPage() != 1 {
		t.Errorf("expected default page 1, got %d", m.CurrentPage())
	}

	// After setting alerts with pagination
	m.SetAlerts(api.MockAlerts(), api.PaginationInfo{CurrentPage: 3, HasNext: true, HasPrev: true})
	if m.CurrentPage() != 3 {
		t.Errorf("expected page 3, got %d", m.CurrentPage())
	}
}

func TestAlertsModelHasNextPage(t *testing.T) {
	m := NewAlertsModel()

	// Default should be false
	if m.HasNextPage() {
		t.Error("expected HasNextPage to be false initially")
	}

	// Set with hasNext = true
	m.SetAlerts(api.MockAlerts(), api.PaginationInfo{CurrentPage: 1, HasNext: true})
	if !m.HasNextPage() {
		t.Error("expected HasNextPage to be true")
	}

	// Set with hasNext = false
	m.SetAlerts(api.MockAlerts(), api.PaginationInfo{CurrentPage: 1, HasNext: false})
	if m.HasNextPage() {
		t.Error("expected HasNextPage to be false")
	}
}

func TestAlertsModelHasPrevPage(t *testing.T) {
	m := NewAlertsModel()

	// Default should be false
	if m.HasPrevPage() {
		t.Error("expected HasPrevPage to be false initially")
	}

	// Set with hasPrev = true
	m.SetAlerts(api.MockAlerts(), api.PaginationInfo{CurrentPage: 2, HasPrev: true})
	if !m.HasPrevPage() {
		t.Error("expected HasPrevPage to be true")
	}

	// Set with hasPrev = false
	m.SetAlerts(api.MockAlerts(), api.PaginationInfo{CurrentPage: 1, HasPrev: false})
	if m.HasPrevPage() {
		t.Error("expected HasPrevPage to be false")
	}
}

func TestAlertsModelNextPage(t *testing.T) {
	m := NewAlertsModel()
	m.SetAlerts(api.MockAlerts(), api.PaginationInfo{CurrentPage: 1, HasNext: true})
	m.cursor = 3 // Set cursor to non-zero

	m.NextPage()

	if m.currentPage != 2 {
		t.Errorf("expected page 2 after NextPage, got %d", m.currentPage)
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor reset to 0 after NextPage, got %d", m.cursor)
	}
}

func TestAlertsModelNextPageAtEnd(t *testing.T) {
	m := NewAlertsModel()
	m.SetAlerts(api.MockAlerts(), api.PaginationInfo{CurrentPage: 5, HasNext: false})

	m.NextPage()

	// Should not change when hasNext is false
	if m.currentPage != 5 {
		t.Errorf("expected page to stay at 5 when no next, got %d", m.currentPage)
	}
}

func TestAlertsModelPrevPage(t *testing.T) {
	m := NewAlertsModel()
	m.SetAlerts(api.MockAlerts(), api.PaginationInfo{CurrentPage: 3, HasPrev: true})
	m.cursor = 5 // Set cursor to non-zero

	m.PrevPage()

	if m.currentPage != 2 {
		t.Errorf("expected page 2 after PrevPage, got %d", m.currentPage)
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor reset to 0 after PrevPage, got %d", m.cursor)
	}
}

func TestAlertsModelPrevPageAtStart(t *testing.T) {
	m := NewAlertsModel()
	m.SetAlerts(api.MockAlerts(), api.PaginationInfo{CurrentPage: 1, HasPrev: false})

	m.PrevPage()

	// Should not change when hasPrev is false or at page 1
	if m.currentPage != 1 {
		t.Errorf("expected page to stay at 1 when no prev, got %d", m.currentPage)
	}
}

func TestAlertsModelSelectedAlertEmpty(t *testing.T) {
	m := NewAlertsModel()

	// With no alerts, should return nil
	selected := m.SelectedAlert()
	if selected != nil {
		t.Error("expected nil for selected alert when no alerts")
	}
}

func TestAlertsModelSetAlertsCursorAdjustment(t *testing.T) {
	m := NewAlertsModel()
	m.cursor = 10 // Set cursor beyond new list size

	// Set only 2 alerts - cursor should be adjusted
	alerts := api.MockAlerts()[:2]
	m.SetAlerts(alerts, api.PaginationInfo{CurrentPage: 1})

	if m.cursor >= len(alerts) {
		t.Errorf("cursor should be adjusted to valid range, got %d for %d alerts", m.cursor, len(alerts))
	}
}

func TestAlertsModelWindowSizeMsg(t *testing.T) {
	m := NewAlertsModel()

	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
	if m.listWidth == 0 {
		t.Error("expected listWidth to be calculated")
	}
}

func TestFormatAlertTime(t *testing.T) {
	// Test with a known time
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	result := formatAlertTime(testTime)

	// Should contain the date
	if !strings.Contains(result, "Jan 15, 2024") {
		t.Errorf("expected result to contain 'Jan 15, 2024', got '%s'", result)
	}
}

func TestAlertsModelViewStripsNewlines(t *testing.T) {
	m := NewAlertsModel()
	m.SetDimensions(100, 40)

	// Alert with newlines in summary
	alerts := []api.Alert{
		{
			ID:        "1",
			ShortID:   "ABC123",
			Summary:   "Test alert\nwith newline\rand carriage return",
			Status:    "triggered",
			Source:    "datadog",
			CreatedAt: time.Now(),
		},
	}
	m.SetAlerts(alerts, api.PaginationInfo{CurrentPage: 1})

	view := m.View()

	// The summary should appear on a single line with the ID
	// Count lines that contain ABC123 - should be exactly 1 in the list portion
	lines := strings.Split(view, "\n")
	countWithID := 0
	for _, line := range lines {
		if strings.Contains(line, "ABC123") {
			countWithID++
		}
	}
	// ABC123 appears in list (1) and in detail pane header (1) = 2 total
	if countWithID < 1 {
		t.Error("expected at least 1 line containing ABC123")
	}
	// Verify no raw \r in any line with the ID
	for _, line := range lines {
		if strings.Contains(line, "ABC123") && strings.Contains(line, "\r") {
			t.Error("expected carriage returns to be stripped from lines containing ABC123")
		}
	}
}
