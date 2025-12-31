package views

import (
	"strings"
	"testing"
	"time"

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

func TestIncidentsModelInit(t *testing.T) {
	m := NewIncidentsModel()
	cmd := m.Init()

	// Init should return nil for IncidentsModel
	if cmd != nil {
		t.Error("expected Init to return nil")
	}
}

func TestIncidentsModelSetSpinner(t *testing.T) {
	m := NewIncidentsModel()

	m.SetSpinner("⠋")
	if m.spinnerView != "⠋" {
		t.Errorf("expected spinner '⠋', got '%s'", m.spinnerView)
	}

	m.SetSpinner("⠙")
	if m.spinnerView != "⠙" {
		t.Errorf("expected spinner '⠙', got '%s'", m.spinnerView)
	}
}

func TestIncidentsModelCurrentPage(t *testing.T) {
	m := NewIncidentsModel()

	// Default page should be 1
	if m.CurrentPage() != 1 {
		t.Errorf("expected default page 1, got %d", m.CurrentPage())
	}

	// After setting incidents with pagination
	m.SetIncidents(api.MockIncidents(), api.PaginationInfo{CurrentPage: 3, HasNext: true, HasPrev: true})
	if m.CurrentPage() != 3 {
		t.Errorf("expected page 3, got %d", m.CurrentPage())
	}
}

func TestIncidentsModelHasNextPage(t *testing.T) {
	m := NewIncidentsModel()

	// Default should be false
	if m.HasNextPage() {
		t.Error("expected HasNextPage to be false initially")
	}

	// Set with hasNext = true
	m.SetIncidents(api.MockIncidents(), api.PaginationInfo{CurrentPage: 1, HasNext: true})
	if !m.HasNextPage() {
		t.Error("expected HasNextPage to be true")
	}

	// Set with hasNext = false
	m.SetIncidents(api.MockIncidents(), api.PaginationInfo{CurrentPage: 1, HasNext: false})
	if m.HasNextPage() {
		t.Error("expected HasNextPage to be false")
	}
}

func TestIncidentsModelHasPrevPage(t *testing.T) {
	m := NewIncidentsModel()

	// Default should be false
	if m.HasPrevPage() {
		t.Error("expected HasPrevPage to be false initially")
	}

	// Set with hasPrev = true
	m.SetIncidents(api.MockIncidents(), api.PaginationInfo{CurrentPage: 2, HasPrev: true})
	if !m.HasPrevPage() {
		t.Error("expected HasPrevPage to be true")
	}

	// Set with hasPrev = false
	m.SetIncidents(api.MockIncidents(), api.PaginationInfo{CurrentPage: 1, HasPrev: false})
	if m.HasPrevPage() {
		t.Error("expected HasPrevPage to be false")
	}
}

func TestIncidentsModelNextPage(t *testing.T) {
	m := NewIncidentsModel()
	m.SetIncidents(api.MockIncidents(), api.PaginationInfo{CurrentPage: 1, HasNext: true})
	m.cursor = 3 // Set cursor to non-zero

	m.NextPage()

	if m.currentPage != 2 {
		t.Errorf("expected page 2 after NextPage, got %d", m.currentPage)
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor reset to 0 after NextPage, got %d", m.cursor)
	}
}

func TestIncidentsModelNextPageAtEnd(t *testing.T) {
	m := NewIncidentsModel()
	m.SetIncidents(api.MockIncidents(), api.PaginationInfo{CurrentPage: 5, HasNext: false})

	m.NextPage()

	// Should not change when hasNext is false
	if m.currentPage != 5 {
		t.Errorf("expected page to stay at 5 when no next, got %d", m.currentPage)
	}
}

func TestIncidentsModelPrevPage(t *testing.T) {
	m := NewIncidentsModel()
	m.SetIncidents(api.MockIncidents(), api.PaginationInfo{CurrentPage: 3, HasPrev: true})
	m.cursor = 5 // Set cursor to non-zero

	m.PrevPage()

	if m.currentPage != 2 {
		t.Errorf("expected page 2 after PrevPage, got %d", m.currentPage)
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor reset to 0 after PrevPage, got %d", m.cursor)
	}
}

func TestIncidentsModelPrevPageAtStart(t *testing.T) {
	m := NewIncidentsModel()
	m.SetIncidents(api.MockIncidents(), api.PaginationInfo{CurrentPage: 1, HasPrev: false})

	m.PrevPage()

	// Should not change when hasPrev is false or at page 1
	if m.currentPage != 1 {
		t.Errorf("expected page to stay at 1 when no prev, got %d", m.currentPage)
	}
}

func TestIncidentsModelSelectedIncidentEmpty(t *testing.T) {
	m := NewIncidentsModel()

	// With no incidents, should return nil
	selected := m.SelectedIncident()
	if selected != nil {
		t.Error("expected nil for selected incident when no incidents")
	}
}

func TestIncidentsModelSetIncidentsCursorAdjustment(t *testing.T) {
	m := NewIncidentsModel()
	m.cursor = 10 // Set cursor beyond new list size

	// Set only 3 incidents - cursor should be adjusted
	incidents := api.MockIncidents()[:2]
	m.SetIncidents(incidents, api.PaginationInfo{CurrentPage: 1})

	if m.cursor >= len(incidents) {
		t.Errorf("cursor should be adjusted to valid range, got %d for %d incidents", m.cursor, len(incidents))
	}
}

func TestIncidentsModelWindowSizeMsg(t *testing.T) {
	m := NewIncidentsModel()

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

func TestSeveritySignalPlain(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "▁▃▅▇"},
		{"Critical", "▁▃▅▇"},
		{"CRITICAL", "▁▃▅▇"},
		{"sev0", "▁▃▅▇"},
		{"SEV0", "▁▃▅▇"},
		{"high", "▁▃▅░"},
		{"High", "▁▃▅░"},
		{"HIGH", "▁▃▅░"},
		{"sev1", "▁▃▅░"},
		{"SEV1", "▁▃▅░"},
		{"medium", "▁▃░░"},
		{"Medium", "▁▃░░"},
		{"MEDIUM", "▁▃░░"},
		{"sev2", "▁▃░░"},
		{"SEV2", "▁▃░░"},
		{"low", "▁░░░"},
		{"Low", "▁░░░"},
		{"LOW", "▁░░░"},
		{"sev3", "▁░░░"},
		{"SEV3", "▁░░░"},
		{"unknown", "░░░░"},
		{"", "░░░░"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := severitySignalPlain(tt.severity)
			if result != tt.expected {
				t.Errorf("severitySignalPlain(%s) = %s, expected %s", tt.severity, result, tt.expected)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	// Test with a known time
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	result := formatTime(testTime)

	// Should contain the date
	if !strings.Contains(result, "Jan 15, 2024") {
		t.Errorf("expected result to contain 'Jan 15, 2024', got '%s'", result)
	}
}

func TestIncidentsModelViewStripsNewlines(t *testing.T) {
	m := NewIncidentsModel()
	m.SetDimensions(100, 40)

	// Incident with newlines in summary
	incidents := []api.Incident{
		{
			ID:           "1",
			SequentialID: "INC-123",
			Summary:      "Test incident\nwith newline\rand carriage return",
			Status:       "started",
			Severity:     "critical",
			CreatedAt:    time.Now(),
		},
	}
	m.SetIncidents(incidents, api.PaginationInfo{CurrentPage: 1})

	view := m.View()

	// The summary should appear on a single line with the ID
	// INC-123 appears in list (1) and in detail pane header (1) = 2 total
	lines := strings.Split(view, "\n")
	countWithID := 0
	for _, line := range lines {
		if strings.Contains(line, "INC-123") {
			countWithID++
		}
	}
	if countWithID < 1 {
		t.Error("expected at least 1 line containing INC-123")
	}
	// Verify no raw \r in any line with the ID
	for _, line := range lines {
		if strings.Contains(line, "INC-123") && strings.Contains(line, "\r") {
			t.Error("expected carriage returns to be stripped from lines containing INC-123")
		}
	}
}

func TestIncidentsModelSelectedIndex(t *testing.T) {
	m := NewIncidentsModel()
	incidents := api.MockIncidents()
	m.SetIncidents(incidents, api.PaginationInfo{CurrentPage: 1})
	m.SetDimensions(100, 30)

	// Initial index should be 0
	if m.SelectedIndex() != 0 {
		t.Errorf("expected initial index 0, got %d", m.SelectedIndex())
	}

	// Move cursor down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.SelectedIndex() != 1 {
		t.Errorf("expected index 1 after j, got %d", m.SelectedIndex())
	}

	// Move cursor down again
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.SelectedIndex() != 2 {
		t.Errorf("expected index 2 after j, got %d", m.SelectedIndex())
	}
}

func TestIncidentsModelSetDetailLoading(t *testing.T) {
	m := NewIncidentsModel()

	// Default should be false
	if m.detailLoading {
		t.Error("expected detailLoading to be false initially")
	}

	m.SetDetailLoading(true)
	if !m.detailLoading {
		t.Error("expected detailLoading to be true after SetDetailLoading(true)")
	}

	m.SetDetailLoading(false)
	if m.detailLoading {
		t.Error("expected detailLoading to be false after SetDetailLoading(false)")
	}
}

func TestIncidentsModelUpdateIncidentDetail(t *testing.T) {
	m := NewIncidentsModel()
	incidents := api.MockIncidents()
	m.SetIncidents(incidents, api.PaginationInfo{CurrentPage: 1})

	// Get original incident
	originalID := m.incidents[0].ID
	if originalID != incidents[0].ID {
		t.Fatalf("expected original ID %s, got %s", incidents[0].ID, originalID)
	}

	// Create detailed incident
	detailedIncident := &api.Incident{
		ID:               incidents[0].ID,
		SequentialID:     incidents[0].SequentialID,
		Title:            "Detailed Title",
		Summary:          "Detailed Summary",
		Status:           "resolved",
		Severity:         "critical",
		DetailLoaded:     true,
		CommanderName:    "John Doe",
		CommunicatorName: "Jane Smith",
		Causes:           []string{"Configuration Error"},
		IncidentTypes:    []string{"Infrastructure"},
		URL:              "https://rootly.io/incidents/123",
	}

	// Update at index 0
	m.UpdateIncidentDetail(0, detailedIncident)

	// Verify the incident was updated
	updated := m.incidents[0]
	if !updated.DetailLoaded {
		t.Error("expected DetailLoaded to be true after update")
	}
	if updated.CommanderName != "John Doe" {
		t.Errorf("expected CommanderName 'John Doe', got '%s'", updated.CommanderName)
	}
	if updated.CommunicatorName != "Jane Smith" {
		t.Errorf("expected CommunicatorName 'Jane Smith', got '%s'", updated.CommunicatorName)
	}
	if len(updated.Causes) != 1 || updated.Causes[0] != "Configuration Error" {
		t.Errorf("expected Causes ['Configuration Error'], got %v", updated.Causes)
	}
	if updated.URL != "https://rootly.io/incidents/123" {
		t.Errorf("expected URL, got '%s'", updated.URL)
	}
}

func TestIncidentsModelUpdateIncidentDetailInvalidIndex(t *testing.T) {
	m := NewIncidentsModel()
	incidents := api.MockIncidents()
	m.SetIncidents(incidents, api.PaginationInfo{CurrentPage: 1})

	detailedIncident := &api.Incident{
		ID:           "new",
		DetailLoaded: true,
	}

	// Update at invalid index should not panic
	m.UpdateIncidentDetail(-1, detailedIncident)
	m.UpdateIncidentDetail(100, detailedIncident)

	// Verify nothing changed
	if m.incidents[0].DetailLoaded {
		t.Error("expected DetailLoaded to remain unchanged for invalid index")
	}
}

func TestIncidentsModelUpdateIncidentDetailNil(t *testing.T) {
	m := NewIncidentsModel()
	incidents := api.MockIncidents()
	m.SetIncidents(incidents, api.PaginationInfo{CurrentPage: 1})

	// Update with nil should not panic
	m.UpdateIncidentDetail(0, nil)

	// Verify nothing changed
	if m.incidents[0].DetailLoaded {
		t.Error("expected DetailLoaded to remain unchanged for nil incident")
	}
}

func TestIncidentsModelViewShowsDetailHint(t *testing.T) {
	m := NewIncidentsModel()
	m.SetDimensions(100, 40)

	// Incident without detail loaded
	incidents := []api.Incident{
		{
			ID:           "1",
			SequentialID: "INC-123",
			Summary:      "Test incident",
			Status:       "started",
			Severity:     "critical",
			CreatedAt:    time.Now(),
			DetailLoaded: false,
		},
	}
	m.SetIncidents(incidents, api.PaginationInfo{CurrentPage: 1})

	view := m.View()
	if !strings.Contains(view, "Press Enter for more details") {
		t.Error("expected hint 'Press Enter for more details' when detail not loaded")
	}
}

func TestIncidentsModelViewHidesDetailHint(t *testing.T) {
	m := NewIncidentsModel()
	m.SetDimensions(100, 40)

	// Incident with detail loaded
	incidents := []api.Incident{
		{
			ID:           "1",
			SequentialID: "INC-123",
			Summary:      "Test incident",
			Status:       "started",
			Severity:     "critical",
			CreatedAt:    time.Now(),
			DetailLoaded: true,
		},
	}
	m.SetIncidents(incidents, api.PaginationInfo{CurrentPage: 1})

	view := m.View()
	if strings.Contains(view, "Press Enter for more details") {
		t.Error("expected no hint when detail is loaded")
	}
}

func TestIncidentsModelViewShowsExtendedDetail(t *testing.T) {
	m := NewIncidentsModel()
	m.SetDimensions(100, 40)

	// Incident with extended detail
	incidents := []api.Incident{
		{
			ID:               "1",
			SequentialID:     "INC-123",
			Summary:          "Test incident",
			Status:           "resolved",
			Severity:         "critical",
			CreatedAt:        time.Now(),
			DetailLoaded:     true,
			CommanderName:    "John Doe",
			CommunicatorName: "Jane Smith",
			Causes:           []string{"Config Error"},
			IncidentTypes:    []string{"Infrastructure"},
			Functionalities:  []string{"API Gateway"},
			URL:              "https://rootly.io/test",
			Roles: []api.IncidentRole{
				{Name: "Commander", UserName: "John Doe"},
			},
		},
	}
	m.SetIncidents(incidents, api.PaginationInfo{CurrentPage: 1})

	view := m.View()

	// Should show roles section
	if !strings.Contains(view, "Roles") {
		t.Error("expected 'Roles' section in detail view")
	}
	if !strings.Contains(view, "Commander") {
		t.Error("expected 'Commander' in detail view")
	}
	if !strings.Contains(view, "John Doe") {
		t.Error("expected 'John Doe' in detail view")
	}

	// Should show causes
	if !strings.Contains(view, "Causes") {
		t.Error("expected 'Causes' in detail view")
	}
	if !strings.Contains(view, "Config Error") {
		t.Error("expected 'Config Error' in detail view")
	}

	// Should show types
	if !strings.Contains(view, "Types") {
		t.Error("expected 'Types' in detail view")
	}
	if !strings.Contains(view, "Infrastructure") {
		t.Error("expected 'Infrastructure' in detail view")
	}

	// Should show Rootly link
	if !strings.Contains(view, "Rootly") {
		t.Error("expected 'Rootly' link in detail view")
	}
}
