package views

import (
	"strings"

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

	b.WriteString(styles.DialogTitle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	// Navigation section
	b.WriteString(styles.TextBold.Render("Navigation"))
	b.WriteString("\n")
	b.WriteString(renderHelpLine("j / Down", "Move cursor down"))
	b.WriteString(renderHelpLine("k / Up", "Move cursor up"))
	b.WriteString(renderHelpLine("g", "Go to first item"))
	b.WriteString(renderHelpLine("G", "Go to last item"))
	b.WriteString(renderHelpLine("Tab", "Switch between tabs"))
	b.WriteString("\n")

	// Actions section
	b.WriteString(styles.TextBold.Render("Actions"))
	b.WriteString("\n")
	b.WriteString(renderHelpLine("r", "Refresh data"))
	b.WriteString(renderHelpLine("Enter", "View details / Select"))
	b.WriteString("\n")

	// General section
	b.WriteString(styles.TextBold.Render("General"))
	b.WriteString("\n")
	b.WriteString(renderHelpLine("?", "Toggle this help"))
	b.WriteString(renderHelpLine("q / Ctrl+C", "Quit"))
	b.WriteString("\n\n")

	b.WriteString(styles.TextDim.Render("Press ? or Esc to close"))

	return styles.Dialog.Render(b.String())
}

func renderHelpLine(key, desc string) string {
	keyStyle := styles.HelpKey.Width(12)
	return keyStyle.Render(key) + " " + styles.HelpDesc.Render(desc) + "\n"
}

// RenderHelpBar renders the bottom help bar
func RenderHelpBar(width int) string {
	items := []string{
		styles.RenderHelpItem("j/k", "navigate"),
		styles.RenderHelpItem("Tab", "switch"),
		styles.RenderHelpItem("r", "refresh"),
		styles.RenderHelpItem("?", "help"),
		styles.RenderHelpItem("q", "quit"),
	}

	return styles.HelpBar.Width(width).Render(strings.Join(items, "  "))
}
