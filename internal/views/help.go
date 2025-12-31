package views

import (
	"strings"

	"github.com/rootlyhq/rootly-tui/internal/i18n"
	"github.com/rootlyhq/rootly-tui/internal/styles"
)

type HelpModel struct {
	Visible bool
}

func NewHelpModel() HelpModel {
	return HelpModel{Visible: false}
}

func (m *HelpModel) Toggle() {
	m.Visible = !m.Visible
}

func (m *HelpModel) Show() {
	m.Visible = true
}

func (m *HelpModel) Hide() {
	m.Visible = false
}

func (m HelpModel) View() string {
	var b strings.Builder

	b.WriteString(styles.DialogTitle.Render(i18n.T("keyboard_shortcuts")))
	b.WriteString("\n\n")

	// Navigation section
	b.WriteString(styles.TextBold.Render(i18n.T("navigation")))
	b.WriteString("\n")
	b.WriteString(renderHelpLine("j / Down", i18n.T("move_down")))
	b.WriteString(renderHelpLine("k / Up", i18n.T("move_up")))
	b.WriteString(renderHelpLine("g", i18n.T("go_to_first")))
	b.WriteString(renderHelpLine("G", i18n.T("go_to_last")))
	b.WriteString(renderHelpLine("[", i18n.T("previous_page")))
	b.WriteString(renderHelpLine("]", i18n.T("next_page")))
	b.WriteString(renderHelpLine("Tab", i18n.T("switch_tabs")))
	b.WriteString("\n")

	// Actions section
	b.WriteString(styles.TextBold.Render(i18n.T("actions")))
	b.WriteString("\n")
	b.WriteString(renderHelpLine("r", i18n.T("refresh_data")))
	b.WriteString(renderHelpLine("Enter", i18n.T("view_details")))
	b.WriteString(renderHelpLine("o", i18n.T("open_url")))
	b.WriteString("\n")

	// Sorting section
	b.WriteString(styles.TextBold.Render(i18n.T("sorting")))
	b.WriteString("\n")
	b.WriteString(renderHelpLine("S", i18n.T("open_sort_menu")))
	b.WriteString("\n")

	// General section
	b.WriteString(styles.TextBold.Render(i18n.T("general")))
	b.WriteString("\n")
	b.WriteString(renderHelpLine("l", i18n.T("view_logs")))
	b.WriteString(renderHelpLine("s", i18n.T("open_setup")))
	b.WriteString(renderHelpLine("A", i18n.T("view_about")))
	b.WriteString(renderHelpLine("?", i18n.T("toggle_help")))
	b.WriteString(renderHelpLine("q / Ctrl+C", i18n.T("quit")))
	b.WriteString("\n\n")

	b.WriteString(styles.TextDim.Render(i18n.T("press_to_close")))

	return styles.Dialog.Render(b.String())
}

func renderHelpLine(key, desc string) string {
	keyStyle := styles.HelpKey.Width(12)
	return keyStyle.Render(key) + " " + styles.HelpDesc.Render(desc) + "\n"
}

// RenderHelpBar renders the bottom help bar
// hasSelection indicates whether an incident or alert is currently selected
// isLoading indicates whether data is currently being loaded
// isIncidentsTab indicates whether we're on the incidents tab (for sorting hints)
func RenderHelpBar(width int, hasSelection, isLoading, isIncidentsTab bool) string {
	items := []string{
		styles.RenderHelpItem("j/k", i18n.T("navigate")),
		styles.RenderHelpItem("[/]", i18n.T("page_nav")),
		styles.RenderHelpItem("Tab", i18n.T("switch")),
	}
	if !isLoading {
		items = append(items, styles.RenderHelpItem("r", i18n.T("refresh")))
		if hasSelection {
			items = append(items, styles.RenderHelpItem("o", i18n.T("open")))
		}
	}
	// Show sorting hint only on incidents tab
	if isIncidentsTab {
		items = append(items, styles.RenderHelpItem("S", i18n.T("sort")))
	}
	items = append(items,
		styles.RenderHelpItem("l", i18n.T("logs")),
		styles.RenderHelpItem("s", i18n.T("setup")),
		styles.RenderHelpItem("A", i18n.T("about")),
		styles.RenderHelpItem("?", i18n.T("help")),
		styles.RenderHelpItem("q", i18n.T("quit_action")),
	)

	return styles.HelpBar.Width(width).Render(strings.Join(items, "  "))
}
