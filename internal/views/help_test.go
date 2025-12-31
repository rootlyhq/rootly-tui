package views

import (
	"strings"
	"testing"
)

func TestNewHelpModel(t *testing.T) {
	m := NewHelpModel()

	if m.Visible {
		t.Error("expected help to be hidden initially")
	}
}

func TestHelpModelToggle(t *testing.T) {
	m := NewHelpModel()

	m.Toggle()
	if !m.Visible {
		t.Error("expected help to be visible after toggle")
	}

	m.Toggle()
	if m.Visible {
		t.Error("expected help to be hidden after second toggle")
	}
}

func TestHelpModelShow(t *testing.T) {
	m := NewHelpModel()

	m.Show()
	if !m.Visible {
		t.Error("expected help to be visible after Show()")
	}
}

func TestHelpModelHide(t *testing.T) {
	m := NewHelpModel()
	m.Visible = true

	m.Hide()
	if m.Visible {
		t.Error("expected help to be hidden after Hide()")
	}
}

func TestHelpModelView(t *testing.T) {
	m := NewHelpModel()
	view := m.View()

	// Check for key sections
	expectedContent := []string{
		"Keyboard Shortcuts",
		"Navigation",
		"j / Down",
		"k / Up",
		"Tab",
		"Actions",
		"Refresh",
		"General",
		"Quit",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(view, expected) {
			t.Errorf("expected help view to contain '%s'", expected)
		}
	}
}

func TestRenderHelpBar(t *testing.T) {
	bar := RenderHelpBar(80, false)

	expectedItems := []string{
		"navigate",
		"switch",
		"refresh",
		"help",
		"quit",
	}

	for _, item := range expectedItems {
		if !strings.Contains(bar, item) {
			t.Errorf("expected help bar to contain '%s'", item)
		}
	}

	// 'open' should not be shown when hasSelection is false
	if strings.Contains(bar, "open") {
		t.Error("expected help bar to NOT contain 'open' when hasSelection is false")
	}
}

func TestRenderHelpBarWithSelection(t *testing.T) {
	bar := RenderHelpBar(80, true)

	// 'open' should be shown when hasSelection is true
	if !strings.Contains(bar, "open") {
		t.Error("expected help bar to contain 'open' when hasSelection is true")
	}
}
