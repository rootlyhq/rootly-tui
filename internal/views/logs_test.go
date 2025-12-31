package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rootlyhq/rootly-tui/internal/debug"
)

func TestNewLogsModel(t *testing.T) {
	m := NewLogsModel()

	if m.Visible {
		t.Error("expected logs to be hidden initially")
	}
}

func TestLogsModelToggle(t *testing.T) {
	m := NewLogsModel()

	m.Toggle()
	if !m.Visible {
		t.Error("expected logs to be visible after toggle")
	}

	m.Toggle()
	if m.Visible {
		t.Error("expected logs to be hidden after second toggle")
	}
}

func TestLogsModelShowHide(t *testing.T) {
	m := NewLogsModel()

	m.Show()
	if !m.Visible {
		t.Error("expected logs to be visible after Show()")
	}

	m.Hide()
	if m.Visible {
		t.Error("expected logs to be hidden after Hide()")
	}
}

func TestLogsModelSetDimensions(t *testing.T) {
	m := NewLogsModel()

	m.SetDimensions(100, 50)

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
}

func TestLogsModelRefresh(t *testing.T) {
	// Clear any existing logs
	debug.ClearLogs()

	m := NewLogsModel()
	m.SetDimensions(100, 50)

	// Add some test logs
	debug.Logger.Info("test log 1")
	debug.Logger.Info("test log 2")

	m.Refresh()

	if len(m.logs) < 2 {
		t.Errorf("expected at least 2 logs, got %d", len(m.logs))
	}
}

func TestLogsModelUpdate(t *testing.T) {
	m := NewLogsModel()
	m.Visible = true
	m.SetDimensions(100, 20) // Small height to ensure scrolling is needed

	// Add many logs to scroll through
	debug.ClearLogs()
	for i := 0; i < 200; i++ {
		debug.Logger.Info("test log entry")
	}
	m.Refresh()

	// Reset scroll to beginning
	m.scrollPos = 0

	// Verify we have enough logs to scroll
	if m.totalLines < 20 {
		t.Fatalf("expected at least 20 logs, got %d", m.totalLines)
	}

	// Test scroll down with 'j'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.scrollPos != 1 {
		t.Errorf("expected scroll position 1 after 'j', got %d", m.scrollPos)
	}

	// Test scroll up with 'k'
	m.scrollPos = 10
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.scrollPos != 9 {
		t.Errorf("expected scroll position 9, got %d", m.scrollPos)
	}

	// Test go to top with 'g'
	m.scrollPos = 50
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if m.scrollPos != 0 {
		t.Errorf("expected scroll position 0 after 'g', got %d", m.scrollPos)
	}

	// Test go to bottom with 'G'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	expectedMax := m.totalLines - m.visibleLines()
	if expectedMax < 0 {
		expectedMax = 0
	}
	if m.scrollPos != expectedMax {
		t.Errorf("expected scroll position %d after 'G', got %d", expectedMax, m.scrollPos)
	}

	// Test clear with 'c'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if len(m.logs) != 0 {
		t.Errorf("expected 0 logs after clear, got %d", len(m.logs))
	}
}

func TestLogsModelUpdateWhenHidden(t *testing.T) {
	m := NewLogsModel()
	m.Visible = false
	m.scrollPos = 5

	// Updates should be ignored when hidden
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.scrollPos != 5 {
		t.Error("expected scroll position to remain unchanged when hidden")
	}
}

func TestLogsModelView(t *testing.T) {
	m := NewLogsModel()
	m.SetDimensions(100, 50)

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}

	// Should contain title
	if !containsStr(view, "Debug Logs") {
		t.Error("expected view to contain 'Debug Logs'")
	}
}

func TestLogsModelViewWithLogs(t *testing.T) {
	debug.ClearLogs()
	debug.Logger.Info("test message")

	m := NewLogsModel()
	m.SetDimensions(100, 50)
	m.Refresh()

	view := m.View()

	// Should contain the test message
	if !containsStr(view, "test message") {
		t.Error("expected view to contain log message")
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{123, "123"},
		{-5, "-5"},
		{-123, "-123"},
	}

	for _, tt := range tests {
		result := itoa(tt.input)
		if result != tt.expected {
			t.Errorf("itoa(%d) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestMinInt(t *testing.T) {
	if minInt(5, 10) != 5 {
		t.Error("expected minInt(5, 10) = 5")
	}
	if minInt(10, 5) != 5 {
		t.Error("expected minInt(10, 5) = 5")
	}
	if minInt(5, 5) != 5 {
		t.Error("expected minInt(5, 5) = 5")
	}
}

func TestVisibleLinesMinimum(t *testing.T) {
	m := NewLogsModel()

	// With zero height, should return minimum of 1
	m.SetDimensions(100, 0)
	if m.visibleLines() != 1 {
		t.Errorf("expected visibleLines() = 1 with zero height, got %d", m.visibleLines())
	}

	// With small height, should return minimum of 1
	m.SetDimensions(100, 5)
	if m.visibleLines() != 1 {
		t.Errorf("expected visibleLines() = 1 with small height, got %d", m.visibleLines())
	}

	// With normal height
	m.SetDimensions(100, 30)
	expected := 30 - 8 // height - 8
	if m.visibleLines() != expected {
		t.Errorf("expected visibleLines() = %d, got %d", expected, m.visibleLines())
	}
}

func TestMaxScrollPos(t *testing.T) {
	m := NewLogsModel()
	m.SetDimensions(100, 20)

	// With no logs
	m.totalLines = 0
	if m.maxScrollPos() != 0 {
		t.Errorf("expected maxScrollPos() = 0 with no logs, got %d", m.maxScrollPos())
	}

	// With fewer logs than visible lines
	m.totalLines = 5
	if m.maxScrollPos() != 0 {
		t.Errorf("expected maxScrollPos() = 0 with few logs, got %d", m.maxScrollPos())
	}

	// With more logs than visible lines
	m.totalLines = 100
	expected := 100 - m.visibleLines()
	if m.maxScrollPos() != expected {
		t.Errorf("expected maxScrollPos() = %d, got %d", expected, m.maxScrollPos())
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && containsStrHelper(s, substr))
}

func containsStrHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestColorizeLogEntry(t *testing.T) {
	tests := []struct {
		name  string
		entry string
	}{
		{"error entry", "ERRO rootly-tui Error message"},
		{"error entry alt", "ERROR: something went wrong"},
		{"warning entry", "WARN rootly-tui Warning message"},
		{"info entry", "INFO rootly-tui Info message"},
		{"debug entry", "DEBU rootly-tui Debug message"},
		{"debug entry alt", "DEBUG: some debug info"},
		{"plain entry", "Some plain text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorizeLogEntry(tt.entry)
			// Result should contain the original entry text (even if wrapped in ANSI codes)
			if result == "" {
				t.Error("expected non-empty result")
			}
		})
	}
}
