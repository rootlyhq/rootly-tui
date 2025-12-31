package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rootlyhq/rootly-tui/internal/api"
	"github.com/rootlyhq/rootly-tui/internal/styles"
)

type AlertsModel struct {
	alerts      []api.Alert
	cursor      int
	width       int
	height      int
	listWidth   int
	detailWidth int
	loading     bool
	error       string
}

func NewAlertsModel() AlertsModel {
	return AlertsModel{
		alerts: []api.Alert{},
		cursor: 0,
	}
}

func (m AlertsModel) Init() tea.Cmd {
	return nil
}

func (m AlertsModel) Update(msg tea.Msg) (AlertsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.alerts)-1 {
				m.cursor++
			}
			return m, nil

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "g":
			m.cursor = 0
			return m, nil

		case "G":
			if len(m.alerts) > 0 {
				m.cursor = len(m.alerts) - 1
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateDimensions()
	}

	return m, nil
}

func (m *AlertsModel) updateDimensions() {
	if m.width > 0 {
		m.listWidth = int(float64(m.width) * 0.4)
		m.detailWidth = m.width - m.listWidth - 6
	}
}

func (m *AlertsModel) SetAlerts(alerts []api.Alert) {
	m.alerts = alerts
	m.loading = false
	m.error = ""
	if m.cursor >= len(alerts) && len(alerts) > 0 {
		m.cursor = len(alerts) - 1
	}
}

func (m *AlertsModel) SetLoading(loading bool) {
	m.loading = loading
}

func (m *AlertsModel) SetError(err string) {
	m.error = err
	m.loading = false
}

func (m *AlertsModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.updateDimensions()
}

func (m AlertsModel) SelectedAlert() *api.Alert {
	if m.cursor >= 0 && m.cursor < len(m.alerts) {
		return &m.alerts[m.cursor]
	}
	return nil
}

func (m AlertsModel) View() string {
	if m.loading {
		return styles.TextDim.Render("Loading alerts...")
	}

	if m.error != "" {
		return styles.Error.Render("Error: " + m.error)
	}

	if len(m.alerts) == 0 {
		return styles.TextDim.Render("No alerts found")
	}

	contentHeight := m.height - 8
	if contentHeight < 5 {
		contentHeight = 5
	}

	listView := m.renderList(contentHeight)
	detailView := m.renderDetail(contentHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, "  ", detailView)
}

func (m AlertsModel) renderList(height int) string {
	var b strings.Builder

	title := styles.TextBold.Render("ALERTS")
	b.WriteString(title)
	b.WriteString("\n\n")

	maxVisible := height - 4
	if maxVisible < 1 {
		maxVisible = 1
	}

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.alerts) {
		end = len(m.alerts)
	}

	for i := start; i < end; i++ {
		alert := m.alerts[i]

		// Source icon
		source := styles.RenderAlertSource(alert.Source)

		// Short ID (e.g., ABC123)
		shortID := alert.ShortID
		if shortID == "" {
			shortID = "---"
		}

		// Status (padded for alignment)
		status := alert.Status
		if len(status) > 10 {
			status = status[:10]
		}
		statusPadded := fmt.Sprintf("%-10s", status)

		// Summary (truncated)
		titleMaxLen := m.listWidth - 30
		if titleMaxLen < 10 {
			titleMaxLen = 10
		}
		summary := alert.Summary
		if len(summary) > titleMaxLen {
			summary = summary[:titleMaxLen-3] + "..."
		}

		// Format: "[source] ABC123  triggered   Summary here"
		line := fmt.Sprintf("%s %-8s %s %s", source, shortID, styles.RenderStatus(statusPadded), summary)

		if i == m.cursor {
			b.WriteString(styles.ListItemSelected.Width(m.listWidth - 4).Render(line))
		} else {
			b.WriteString(styles.ListItem.Width(m.listWidth - 4).Render(line))
		}
		b.WriteString("\n")
	}

	if len(m.alerts) > maxVisible {
		scrollInfo := fmt.Sprintf("\n%d/%d", m.cursor+1, len(m.alerts))
		b.WriteString(styles.TextDim.Render(scrollInfo))
	}

	content := b.String()
	return styles.ListContainer.Width(m.listWidth).Height(height).Render(content)
}

func (m AlertsModel) renderDetail(height int) string {
	alert := m.SelectedAlert()
	if alert == nil {
		return styles.DetailContainer.Width(m.detailWidth).Height(height).Render(
			styles.TextDim.Render("Select an alert to view details"),
		)
	}

	var b strings.Builder

	// Short ID and Summary as title
	if alert.ShortID != "" {
		b.WriteString(styles.Primary.Bold(true).Render(alert.ShortID))
		b.WriteString("\n")
	}
	b.WriteString(styles.DetailTitle.Render(alert.Summary))
	b.WriteString("\n\n")

	// Status and Source row
	statusBadge := styles.RenderStatus(alert.Status)
	sourceIcon := styles.RenderAlertSource(alert.Source)
	sourceName := styles.AlertSourceName(alert.Source)
	b.WriteString(fmt.Sprintf("Status: %s  Source: %s %s\n\n", statusBadge, sourceIcon, sourceName))

	// Description
	if alert.Description != "" {
		b.WriteString(styles.TextBold.Render("Description"))
		b.WriteString("\n")
		desc := alert.Description
		if len(desc) > 200 {
			desc = desc[:197] + "..."
		}
		b.WriteString(styles.Text.Render(desc))
		b.WriteString("\n\n")
	}

	// Timestamps
	b.WriteString(styles.TextBold.Render("Timeline"))
	b.WriteString("\n")

	if !alert.CreatedAt.IsZero() {
		b.WriteString(m.renderDetailRow("Created", formatAlertTime(alert.CreatedAt)))
	}
	if alert.StartedAt != nil {
		b.WriteString(m.renderDetailRow("Started", formatAlertTime(*alert.StartedAt)))
	}
	if alert.EndedAt != nil {
		b.WriteString(m.renderDetailRow("Ended", formatAlertTime(*alert.EndedAt)))
	}
	b.WriteString("\n")

	// Services
	if len(alert.Services) > 0 {
		b.WriteString(m.renderDetailRow("Services", strings.Join(alert.Services, ", ")))
	}

	// Environments
	if len(alert.Environments) > 0 {
		b.WriteString(m.renderDetailRow("Environments", strings.Join(alert.Environments, ", ")))
	}

	// Groups
	if len(alert.Groups) > 0 {
		b.WriteString(m.renderDetailRow("Groups", strings.Join(alert.Groups, ", ")))
	}

	// External URL
	if alert.ExternalURL != "" {
		b.WriteString("\n")
		b.WriteString(m.renderDetailRow("External URL", alert.ExternalURL))
	}

	// Labels (sorted for consistent display)
	if len(alert.Labels) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.TextBold.Render("Labels"))
		b.WriteString("\n")
		keys := make([]string, 0, len(alert.Labels))
		for k := range alert.Labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteString(m.renderDetailRow(k, alert.Labels[k]))
		}
	}

	content := b.String()
	return styles.DetailContainer.Width(m.detailWidth).Height(height).Render(content)
}

func (m AlertsModel) renderDetailRow(label, value string) string {
	return styles.DetailLabel.Render(label+":") + " " + styles.DetailValue.Render(value) + "\n"
}

func formatAlertTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins < 1 {
			return "just now"
		}
		return fmt.Sprintf("%dm ago", mins)
	}

	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	}

	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}

	return t.Format("Jan 2, 2006 15:04")
}
