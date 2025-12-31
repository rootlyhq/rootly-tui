package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSetupModel(t *testing.T) {
	m := NewSetupModel()

	if m.focusIndex != FieldEndpoint {
		t.Errorf("expected focus on endpoint, got %v", m.focusIndex)
	}

	if m.buttonIndex != 0 {
		t.Errorf("expected button index 0, got %d", m.buttonIndex)
	}

	if m.testing {
		t.Error("expected testing to be false")
	}

	if m.saving {
		t.Error("expected saving to be false")
	}
}

func TestSetupModelInit(t *testing.T) {
	m := NewSetupModel()
	cmd := m.Init()

	if cmd == nil {
		t.Error("expected non-nil command from Init")
	}
}

func TestSetupModelUpdateNavigation(t *testing.T) {
	m := NewSetupModel()

	// Tab to API key field
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focusIndex != FieldAPIKey {
		t.Errorf("expected focus on API key after tab, got %v", m.focusIndex)
	}

	// Tab to timezone field
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focusIndex != FieldTimezone {
		t.Errorf("expected focus on timezone after tab, got %v", m.focusIndex)
	}

	// Tab to buttons
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focusIndex != FieldButtons {
		t.Errorf("expected focus on buttons after tab, got %v", m.focusIndex)
	}

	// Tab wraps to endpoint
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focusIndex != FieldEndpoint {
		t.Errorf("expected focus on endpoint after wrap, got %v", m.focusIndex)
	}

	// Shift+Tab goes back
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.focusIndex != FieldButtons {
		t.Errorf("expected focus on buttons after shift+tab, got %v", m.focusIndex)
	}
}

func TestSetupModelUpdateButtonNavigation(t *testing.T) {
	m := NewSetupModel()

	// Navigate to buttons
	m.focusIndex = FieldButtons
	m.buttonIndex = 0

	// Right moves to save button
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.buttonIndex != 1 {
		t.Errorf("expected button index 1 after right, got %d", m.buttonIndex)
	}

	// Right at end stays at end
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.buttonIndex != 1 {
		t.Errorf("expected button index to stay 1, got %d", m.buttonIndex)
	}

	// Left moves back
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if m.buttonIndex != 0 {
		t.Errorf("expected button index 0 after left, got %d", m.buttonIndex)
	}

	// Left at start stays at start
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if m.buttonIndex != 0 {
		t.Errorf("expected button index to stay 0, got %d", m.buttonIndex)
	}
}

func TestSetupModelUpdateEnterMovesToNext(t *testing.T) {
	m := NewSetupModel()
	m.focusIndex = FieldEndpoint

	// Enter moves to next field
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.focusIndex != FieldAPIKey {
		t.Errorf("expected focus on API key after enter, got %v", m.focusIndex)
	}

	// Enter moves to timezone
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.focusIndex != FieldTimezone {
		t.Errorf("expected focus on timezone after enter, got %v", m.focusIndex)
	}

	// Enter moves to buttons
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.focusIndex != FieldButtons {
		t.Errorf("expected focus on buttons after enter, got %v", m.focusIndex)
	}
}

func TestSetupModelUpdateIgnoredWhileTesting(t *testing.T) {
	m := NewSetupModel()
	m.testing = true
	m.focusIndex = FieldEndpoint

	// Keys should be ignored while testing
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focusIndex != FieldEndpoint {
		t.Error("expected focus to stay on endpoint while testing")
	}
}

func TestSetupModelUpdateIgnoredWhileSaving(t *testing.T) {
	m := NewSetupModel()
	m.saving = true
	m.focusIndex = FieldEndpoint

	// Keys should be ignored while saving
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focusIndex != FieldEndpoint {
		t.Error("expected focus to stay on endpoint while saving")
	}
}

func TestSetupModelUpdateWindowSize(t *testing.T) {
	m := NewSetupModel()

	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
}

func TestSetupModelHandleValidationResult(t *testing.T) {
	m := NewSetupModel()
	m.testing = true

	// Success case
	m.HandleValidationResult(APIKeyValidatedMsg{Valid: true})

	if m.testing {
		t.Error("expected testing to be false after validation")
	}
	if m.testResult != testResultSuccess {
		t.Errorf("expected test result success, got %s", m.testResult)
	}
	if m.testError != "" {
		t.Errorf("expected empty test error, got %s", m.testError)
	}

	// Error case
	m.testing = true
	m.HandleValidationResult(APIKeyValidatedMsg{Valid: false, Error: "invalid key"})

	if m.testing {
		t.Error("expected testing to be false after validation")
	}
	if m.testResult != testResultError {
		t.Errorf("expected test result error, got %s", m.testResult)
	}
	if m.testError != "invalid key" {
		t.Errorf("expected test error 'invalid key', got %s", m.testError)
	}
}

func TestSetupModelIsTesting(t *testing.T) {
	m := NewSetupModel()

	if m.IsTesting() {
		t.Error("expected IsTesting to be false initially")
	}

	m.testing = true
	if !m.IsTesting() {
		t.Error("expected IsTesting to be true")
	}
}

func TestSetupModelSetTesting(t *testing.T) {
	m := NewSetupModel()

	m.SetTesting(true)
	if !m.testing {
		t.Error("expected testing to be true")
	}

	m.SetTesting(false)
	if m.testing {
		t.Error("expected testing to be false")
	}
}

func TestSetupModelView(t *testing.T) {
	m := NewSetupModel()
	m.width = 100
	m.height = 50

	view := m.View()

	// Should contain key elements
	if !containsStr(view, "Welcome to Rootly TUI") {
		t.Error("expected view to contain welcome message")
	}
	if !containsStr(view, "API Endpoint") {
		t.Error("expected view to contain API Endpoint label")
	}
	if !containsStr(view, "API Key") {
		t.Error("expected view to contain API Key label")
	}
	if !containsStr(view, "Timezone") {
		t.Error("expected view to contain Timezone label")
	}
	if !containsStr(view, "Test Connection") {
		t.Error("expected view to contain Test Connection button")
	}
	if !containsStr(view, "Save & Continue") {
		t.Error("expected view to contain Save & Continue button")
	}
}

func TestSetupModelViewWithTestResult(t *testing.T) {
	m := NewSetupModel()
	m.width = 100
	m.height = 50

	// Success state
	m.testResult = testResultSuccess
	view := m.View()
	if !containsStr(view, "Connection successful") {
		t.Error("expected view to contain success message")
	}

	// Error state
	m.testResult = testResultError
	m.testError = "connection failed"
	view = m.View()
	if !containsStr(view, "Error") {
		t.Error("expected view to contain error message")
	}
}

func TestSetupModelViewWhileTesting(t *testing.T) {
	m := NewSetupModel()
	m.width = 100
	m.height = 50
	m.testing = true

	view := m.View()
	if !containsStr(view, "Testing connection") {
		t.Error("expected view to contain testing message")
	}
}

func TestSetupModelViewWithoutDimensions(t *testing.T) {
	m := NewSetupModel()
	// Don't set width/height

	view := m.View()

	// Should still render without centering
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestSetupModelUpdateFocus(t *testing.T) {
	m := NewSetupModel()

	// Initially endpoint is focused
	if !m.endpoint.Focused() {
		t.Error("expected endpoint to be focused initially")
	}

	// Focus API key
	m.focusIndex = FieldAPIKey
	m.updateFocus()
	if m.endpoint.Focused() {
		t.Error("expected endpoint to be blurred")
	}
	if !m.apiKey.Focused() {
		t.Error("expected API key to be focused")
	}

	// Focus timezone (it's a selector, not a text input, so just check text inputs are blurred)
	m.focusIndex = FieldTimezone
	m.updateFocus()
	if m.apiKey.Focused() {
		t.Error("expected API key to be blurred")
	}
	if m.endpoint.Focused() {
		t.Error("expected endpoint to be blurred when timezone focused")
	}

	// Focus buttons (all inputs blurred)
	m.focusIndex = FieldButtons
	m.updateFocus()
	if m.endpoint.Focused() {
		t.Error("expected endpoint to be blurred")
	}
	if m.apiKey.Focused() {
		t.Error("expected API key to be blurred")
	}
}

func TestSetupModelEnterOnTestButton(t *testing.T) {
	m := NewSetupModel()
	m.focusIndex = FieldButtons
	m.buttonIndex = 0

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !m.testing {
		t.Error("expected testing to be true after pressing enter on test button")
	}
	if cmd == nil {
		t.Error("expected non-nil command")
	}
}

func TestSetupModelEnterOnSaveButtonWithoutSuccess(t *testing.T) {
	m := NewSetupModel()
	m.focusIndex = FieldButtons
	m.buttonIndex = 1
	m.testResult = "" // Not yet tested

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.saving {
		t.Error("expected saving to be false when test not successful")
	}
	if cmd != nil {
		t.Error("expected nil command when test not successful")
	}
}

func TestSetupModelEnterOnSaveButtonWithSuccess(t *testing.T) {
	m := NewSetupModel()
	m.focusIndex = FieldButtons
	m.buttonIndex = 1
	m.testResult = testResultSuccess

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !m.saving {
		t.Error("expected saving to be true")
	}
	if cmd == nil {
		t.Error("expected non-nil command")
	}
}

func TestSetupModelViewButtonStates(t *testing.T) {
	m := NewSetupModel()
	m.width = 100
	m.height = 50

	// Test button focused
	m.focusIndex = FieldButtons
	m.buttonIndex = 0
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}

	// Save button focused with success
	m.buttonIndex = 1
	m.testResult = testResultSuccess
	view = m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestSetupModelSetDimensions(t *testing.T) {
	m := NewSetupModel()

	m.SetDimensions(120, 60)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 60 {
		t.Errorf("expected height 60, got %d", m.height)
	}
}

func TestSetupModelTimezoneNavigation(t *testing.T) {
	m := NewSetupModel()
	m.focusIndex = FieldTimezone

	// Save initial index
	initialIndex := m.timezoneIndex

	// Move right (if there are more timezones)
	if initialIndex < len(m.timezones)-1 {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
		if m.timezoneIndex != initialIndex+1 {
			t.Errorf("expected timezone index %d after right, got %d", initialIndex+1, m.timezoneIndex)
		}

		// Move left
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
		if m.timezoneIndex != initialIndex {
			t.Errorf("expected timezone index %d after left, got %d", initialIndex, m.timezoneIndex)
		}
	}

	// Test 'h' and 'l' keys
	m.timezoneIndex = 1
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if m.timezoneIndex != 0 {
		t.Errorf("expected timezone index 0 after 'h', got %d", m.timezoneIndex)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if m.timezoneIndex != 1 {
		t.Errorf("expected timezone index 1 after 'l', got %d", m.timezoneIndex)
	}
}

func TestSetupModelTimezoneAtBounds(t *testing.T) {
	m := NewSetupModel()
	m.focusIndex = FieldTimezone

	// Move to start
	m.timezoneIndex = 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if m.timezoneIndex != 0 {
		t.Error("expected timezone index to stay at 0 at left bound")
	}

	// Move to end
	m.timezoneIndex = len(m.timezones) - 1
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.timezoneIndex != len(m.timezones)-1 {
		t.Error("expected timezone index to stay at end at right bound")
	}
}
