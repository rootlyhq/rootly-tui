package views

import (
	"strings"
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

func TestLogsModelMouseSelection(t *testing.T) {
	debug.ClearLogs()
	for i := 0; i < 20; i++ {
		debug.Logger.Info("test log entry")
	}

	m := NewLogsModel()
	m.Visible = true
	m.SetDimensions(100, 50)
	m.Refresh()

	// Simulate mouse press
	m = m.handleMouse(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		Y:      10,
	})

	if !m.selecting {
		t.Error("expected selecting to be true after mouse press")
	}
	if !m.hasSelection {
		t.Error("expected hasSelection to be true after mouse press")
	}

	// Simulate mouse motion
	m = m.handleMouse(tea.MouseMsg{
		Action: tea.MouseActionMotion,
		Y:      15,
	})

	if m.selectStart == m.selectEnd {
		t.Error("expected selection range to change after motion")
	}

	// Simulate mouse release
	m = m.handleMouse(tea.MouseMsg{
		Action: tea.MouseActionRelease,
		Button: tea.MouseButtonLeft,
	})

	if m.selecting {
		t.Error("expected selecting to be false after mouse release")
	}
}

func TestLogsModelMouseYToLineIndex(t *testing.T) {
	m := NewLogsModel()
	m.SetDimensions(100, 50)
	m.dialogTop = 5
	m.scrollPos = 0

	// Content starts at dialogTop + 4
	contentStartY := m.dialogTop + 4

	// Clicking at content start should return scroll position
	idx := m.mouseYToLineIndex(contentStartY, contentStartY)
	if idx != 0 {
		t.Errorf("expected line index 0, got %d", idx)
	}

	// Clicking above content should return -1
	idx = m.mouseYToLineIndex(contentStartY-1, contentStartY)
	if idx != -1 {
		t.Errorf("expected line index -1 for click above content, got %d", idx)
	}

	// Clicking further down
	m.scrollPos = 5
	idx = m.mouseYToLineIndex(contentStartY+3, contentStartY)
	if idx != 8 {
		t.Errorf("expected line index 8, got %d", idx)
	}
}

func TestLogsModelGetSelectedLines(t *testing.T) {
	m := NewLogsModel()
	m.logs = []string{"line1", "line2", "line3", "line4", "line5"}

	// No selection
	m.hasSelection = false
	lines := m.getSelectedLines()
	if lines != nil {
		t.Error("expected nil when no selection")
	}

	// Normal selection
	m.hasSelection = true
	m.selectStart = 1
	m.selectEnd = 3
	lines = m.getSelectedLines()
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line2" || lines[2] != "line4" {
		t.Error("unexpected selected lines")
	}

	// Reversed selection (end < start)
	m.selectStart = 3
	m.selectEnd = 1
	lines = m.getSelectedLines()
	if len(lines) != 3 {
		t.Errorf("expected 3 lines for reversed selection, got %d", len(lines))
	}

	// Selection beyond bounds
	m.selectStart = -1
	m.selectEnd = 10
	lines = m.getSelectedLines()
	if len(lines) != 5 {
		t.Errorf("expected 5 lines (clamped), got %d", len(lines))
	}
}

func TestLogsModelIsLineSelected(t *testing.T) {
	m := NewLogsModel()

	// No selection
	m.hasSelection = false
	if m.isLineSelected(5) {
		t.Error("expected false when no selection")
	}

	// With selection
	m.hasSelection = true
	m.selectStart = 2
	m.selectEnd = 5

	if m.isLineSelected(1) {
		t.Error("expected line 1 not selected")
	}
	if !m.isLineSelected(2) {
		t.Error("expected line 2 selected")
	}
	if !m.isLineSelected(4) {
		t.Error("expected line 4 selected")
	}
	if !m.isLineSelected(5) {
		t.Error("expected line 5 selected")
	}
	if m.isLineSelected(6) {
		t.Error("expected line 6 not selected")
	}

	// Reversed selection
	m.selectStart = 5
	m.selectEnd = 2
	if !m.isLineSelected(3) {
		t.Error("expected line 3 selected in reversed selection")
	}
}

func TestLogsModelClearSelection(t *testing.T) {
	m := NewLogsModel()
	m.selecting = true
	m.selectStart = 5
	m.selectEnd = 10
	m.hasSelection = true

	m.clearSelection()

	if m.selecting {
		t.Error("expected selecting to be false")
	}
	if m.selectStart != 0 {
		t.Error("expected selectStart to be 0")
	}
	if m.selectEnd != 0 {
		t.Error("expected selectEnd to be 0")
	}
	if m.hasSelection {
		t.Error("expected hasSelection to be false")
	}
}

func TestLogsModelSelectAll(t *testing.T) {
	m := NewLogsModel()
	m.Visible = true
	m.logs = []string{"line1", "line2", "line3"}

	// Press 'a' to select all
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if !m.hasSelection {
		t.Error("expected hasSelection to be true")
	}
	if m.selectStart != 0 {
		t.Errorf("expected selectStart 0, got %d", m.selectStart)
	}
	if m.selectEnd != 2 {
		t.Errorf("expected selectEnd 2, got %d", m.selectEnd)
	}
}

func TestLogsModelEscapeClearsSelection(t *testing.T) {
	m := NewLogsModel()
	m.Visible = true
	m.hasSelection = true
	m.selectStart = 1
	m.selectEnd = 5

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if m.hasSelection {
		t.Error("expected hasSelection to be false after escape")
	}
}

func TestLogsModelViewWithSelection(t *testing.T) {
	debug.ClearLogs()
	debug.Logger.Info("test line 1")
	debug.Logger.Info("test line 2")

	m := NewLogsModel()
	m.SetDimensions(100, 50)
	m.Refresh()
	m.hasSelection = true
	m.selectStart = 0
	m.selectEnd = 0

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestLogsModelCopyToClipboard(t *testing.T) {
	debug.ClearLogs()
	debug.Logger.Info("line to copy")

	m := NewLogsModel()
	m.Visible = true
	m.SetDimensions(100, 50)
	m.Refresh()

	// Select all
	m.hasSelection = true
	m.selectStart = 0
	m.selectEnd = len(m.logs) - 1

	// Copy should set status message (regardless of clipboard availability)
	m.copyToClipboard()

	// Status should be set (either "Copied!" or "Clipboard unavailable")
	if m.statusMsg == "" {
		t.Error("expected status message after copy")
	}
}

func TestLogsModelCopyYKeypress(t *testing.T) {
	debug.ClearLogs()
	debug.Logger.Info("test log")

	m := NewLogsModel()
	m.Visible = true
	m.SetDimensions(100, 50)
	m.Refresh()

	// Press 'y' to copy
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	// Status message should be set
	if m.statusMsg == "" {
		t.Error("expected status message after 'y' keypress")
	}

	// Should return a tick command to clear status
	if cmd == nil {
		t.Error("expected tick command to clear status")
	}
}

func TestLogsModelStatusClearMsg(t *testing.T) {
	m := NewLogsModel()
	m.Visible = true
	m.statusMsg = "Test status"

	// Send clear message
	m, _ = m.Update(LogsStatusClearMsg{})

	if m.statusMsg != "" {
		t.Errorf("expected empty status after LogsStatusClearMsg, got %q", m.statusMsg)
	}
}

func TestLogsModelViewWithStatus(t *testing.T) {
	m := NewLogsModel()
	m.SetDimensions(100, 50)
	m.statusMsg = "Copied!"

	view := m.View()
	if !strings.Contains(view, "Copied!") {
		t.Error("expected view to contain status message")
	}
}
