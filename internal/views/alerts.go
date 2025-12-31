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
	// Pagination state
	currentPage int
	hasNext     bool
	hasPrev     bool
	// Loading spinner (passed from app)
	spinnerView string
	// Detail loading state
	detailLoading bool
}

func NewAlertsModel() AlertsModel {
	return AlertsModel{
		alerts:      []api.Alert{},
		cursor:      0,
		currentPage: 1,
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

func (m *AlertsModel) SetAlerts(alerts []api.Alert, pagination api.PaginationInfo) {
	m.alerts = alerts
	m.loading = false
	m.error = ""
	m.currentPage = pagination.CurrentPage
	m.hasNext = pagination.HasNext
	m.hasPrev = pagination.HasPrev
	if m.cursor >= len(alerts) && len(alerts) > 0 {
		m.cursor = len(alerts) - 1
	}
}

func (m *AlertsModel) SetLoading(loading bool) {
	m.loading = loading
}

func (m *AlertsModel) SetSpinner(spinner string) {
	m.spinnerView = spinner
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

// Pagination methods
func (m AlertsModel) CurrentPage() int {
	return m.currentPage
}

func (m AlertsModel) HasNextPage() bool {
	return m.hasNext
}

func (m AlertsModel) HasPrevPage() bool {
	return m.hasPrev
}

func (m *AlertsModel) NextPage() {
	if m.hasNext {
		m.currentPage++
		m.cursor = 0
	}
}

func (m *AlertsModel) PrevPage() {
	if m.hasPrev && m.currentPage > 1 {
		m.currentPage--
		m.cursor = 0
	}
}

func (m AlertsModel) SelectedAlert() *api.Alert {
	if m.cursor >= 0 && m.cursor < len(m.alerts) {
		return &m.alerts[m.cursor]
	}
	return nil
}

func (m AlertsModel) SelectedIndex() int {
	return m.cursor
}

func (m *AlertsModel) SetDetailLoading(loading bool) {
	m.detailLoading = loading
}

func (m *AlertsModel) UpdateAlertDetail(index int, alert *api.Alert) {
	if index >= 0 && index < len(m.alerts) && alert != nil {
		m.alerts[index] = *alert
	}
}

func (m AlertsModel) View() string {
	contentHeight := m.height - 8
	if contentHeight < 5 {
		contentHeight = 5
	}

	if m.loading {
		// Show loading within the layout structure to prevent jarring shift
		loadingMsg := fmt.Sprintf("%s Loading page %d...", m.spinnerView, m.currentPage)
		listContent := styles.TextBold.Render("ALERTS") + "\n\n" + styles.TextDim.Render(loadingMsg)
		listView := styles.ListContainer.Width(m.listWidth).Height(contentHeight).Render(listContent)
		detailView := styles.DetailContainer.Width(m.detailWidth).Height(contentHeight).Render("")
		return lipgloss.JoinHorizontal(lipgloss.Top, listView, "  ", detailView)
	}

	if m.error != "" {
		return styles.Error.Render("Error: " + m.error)
	}

	if len(m.alerts) == 0 {
		return styles.TextDim.Render("No alerts found")
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

		// Source icon (emoji only in list)
		source := styles.AlertSourceIcon(alert.Source)

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
		// Account for: selector(2) + emoji(4) + space(1) + shortID(8) + space(1) + status(10) + space(1) + padding(8)
		titleMaxLen := m.listWidth - 35
		if titleMaxLen < 10 {
			titleMaxLen = 10
		}
		summary := strings.ReplaceAll(alert.Summary, "\n", " ")
		summary = strings.ReplaceAll(summary, "\r", "")
		if len(summary) > titleMaxLen {
			summary = summary[:titleMaxLen-3] + "..."
		}

		// Format: "▶ [source] ABC123  triggered   Summary here" (▶ for selected)
		// Single line only - no wrapping
		if i == m.cursor {
			line := fmt.Sprintf("▶ %s %-8s %s %s", source, shortID, statusPadded, summary)
			b.WriteString(styles.ListItemSelected.Width(m.listWidth - 4).MaxWidth(m.listWidth - 4).Render(line))
		} else {
			line := fmt.Sprintf("  %s %-8s %s %s", source, shortID, styles.RenderStatus(statusPadded), summary)
			b.WriteString(styles.ListItem.Width(m.listWidth - 4).MaxWidth(m.listWidth - 4).Render(line))
		}
		b.WriteString("\n")
	}

	// Scroll and pagination indicator
	var footer strings.Builder
	footer.WriteString("\n")

	// Page navigation indicators
	if m.hasPrev {
		footer.WriteString(styles.TextDim.Render("← ["))
	} else {
		footer.WriteString(styles.TextDim.Render("  "))
	}
	footer.WriteString(fmt.Sprintf(" Page %d ", m.currentPage))
	if m.hasNext {
		footer.WriteString(styles.TextDim.Render("] →"))
	}

	// Item count
	if len(m.alerts) > 0 {
		footer.WriteString(styles.TextDim.Render(fmt.Sprintf("  (%d-%d)", m.cursor+1, len(m.alerts))))
	}

	b.WriteString(footer.String())

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

	// Title line: [SHORT_ID] Summary (strip newlines for single-line display)
	summaryClean := strings.ReplaceAll(alert.Summary, "\n", " ")
	summaryClean = strings.ReplaceAll(summaryClean, "\r", "")
	if alert.ShortID != "" {
		b.WriteString(styles.Primary.Bold(true).Render("[" + alert.ShortID + "]"))
		b.WriteString(" ")
	}
	b.WriteString(styles.DetailTitle.Render(summaryClean))
	b.WriteString("\n\n")

	// Status and Source row
	statusBadge := styles.RenderStatus(alert.Status)
	sourceIcon := styles.AlertSourceIcon(alert.Source)
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

	// Teams (Groups)
	if len(alert.Groups) > 0 {
		b.WriteString(m.renderDetailRow("Teams", strings.Join(alert.Groups, ", ")))
	}

	// Links section
	rootlyURL := ""
	if alert.ShortID != "" {
		rootlyURL = fmt.Sprintf("https://rootly.com/account/alerts/%s", alert.ShortID)
	}
	if rootlyURL != "" || alert.ExternalURL != "" {
		b.WriteString("\n")
		b.WriteString(styles.TextBold.Render("Links"))
		b.WriteString("\n")
		if rootlyURL != "" {
			b.WriteString(m.renderLinkRow("Rootly", rootlyURL))
		}
		if alert.ExternalURL != "" {
			b.WriteString(m.renderLinkRow("Source", alert.ExternalURL))
		}
	}

	// Extended info (populated when DetailLoaded is true)
	if alert.DetailLoaded {
		// Urgency
		if alert.Urgency != "" {
			b.WriteString(m.renderDetailRow("Urgency", alert.Urgency))
		}

		// Responders
		if len(alert.Responders) > 0 {
			b.WriteString("\n")
			b.WriteString(styles.TextBold.Render("Responders"))
			b.WriteString("\n")
			for _, responder := range alert.Responders {
				b.WriteString(styles.Text.Render("  • " + responder + "\n"))
			}
		}
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

	// Show hint if detail not loaded
	if !alert.DetailLoaded {
		b.WriteString("\n")
		b.WriteString(styles.TextDim.Render("Press Enter for more details"))
	}

	content := b.String()
	return styles.DetailContainer.Width(m.detailWidth).Height(height).Render(content)
}

func (m AlertsModel) renderDetailRow(label, value string) string {
	return styles.DetailLabel.Render(label+":") + " " + styles.DetailValue.Render(value) + "\n"
}

func (m AlertsModel) renderLinkRow(label, url string) string {
	return styles.DetailLabel.Render(label+":") + " " + styles.RenderLink(url, url) + "\n"
}

func formatAlertTime(t time.Time) string {
	// Convert to local timezone
	local := t.Local()
	localStr := local.Format("Jan 2, 2006 15:04 MST")

	// If not UTC, also show UTC equivalent
	_, offset := local.Zone()
	if offset != 0 {
		utcStr := t.UTC().Format("15:04 UTC")
		return localStr + " (" + utcStr + ")"
	}
	return localStr
}
