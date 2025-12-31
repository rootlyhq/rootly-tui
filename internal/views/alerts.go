package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"

	"github.com/rootlyhq/rootly-tui/internal/api"
	"github.com/rootlyhq/rootly-tui/internal/i18n"
	"github.com/rootlyhq/rootly-tui/internal/styles"
)

// renderAlertBulletList renders a section with a bold title and bullet list using lipgloss/list
func renderAlertBulletList(title string, items []string) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(styles.TextBold.Render(title))
	b.WriteString("\n")
	// Convert []string to []any for list.New
	anyItems := make([]any, len(items))
	for i, item := range items {
		anyItems[i] = item
	}
	l := list.New(anyItems...).
		Enumerator(list.Bullet).
		ItemStyle(styles.DetailValue)
	b.WriteString(l.String())
	b.WriteString("\n\n") // Blank line after section
	return b.String()
}

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
	// Detail loading state - tracks which alert ID is currently loading (empty = not loading)
	detailLoadingID string
	// Detail viewport for scrollable content
	detailViewport      viewport.Model
	detailViewportReady bool
	detailFocused       bool // Whether detail pane has focus (for scrolling)
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
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// When detail is focused, handle scrolling keys
		if m.detailFocused {
			switch msg.String() {
			case "esc", "q":
				// Return focus to list
				m.detailFocused = false
				return m, nil
			case "j", "down":
				if m.detailViewportReady {
					m.detailViewport.ScrollDown(3)
				}
				return m, nil
			case "k", "up":
				if m.detailViewportReady {
					m.detailViewport.ScrollUp(3)
				}
				return m, nil
			case "g":
				if m.detailViewportReady {
					m.detailViewport.GotoTop()
				}
				return m, nil
			case "G":
				if m.detailViewportReady {
					m.detailViewport.GotoBottom()
				}
				return m, nil
			case "d", "pgdown":
				if m.detailViewportReady {
					m.detailViewport.HalfPageDown()
				}
				return m, nil
			case "u", "pgup":
				if m.detailViewportReady {
					m.detailViewport.HalfPageUp()
				}
				return m, nil
			}
			// Forward other keys to viewport
			if m.detailViewportReady {
				m.detailViewport, cmd = m.detailViewport.Update(msg)
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// Normal list navigation when detail is not focused
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.alerts)-1 {
				m.cursor++
				m.updateViewportContent() // Update content when cursor changes
			}
			return m, nil

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.updateViewportContent() // Update content when cursor changes
			}
			return m, nil

		case "g":
			m.cursor = 0
			m.updateViewportContent()
			return m, nil

		case "G":
			if len(m.alerts) > 0 {
				m.cursor = len(m.alerts) - 1
				m.updateViewportContent()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateDimensions()
		m.updateViewportContent()
	}

	// Forward mouse messages to viewport for scrolling
	if m.detailViewportReady && m.detailFocused {
		m.detailViewport, cmd = m.detailViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// updateViewportContent updates the viewport content when data changes
func (m *AlertsModel) updateViewportContent() {
	if !m.detailViewportReady {
		return
	}
	alert := m.SelectedAlert()
	if alert == nil {
		return
	}
	content := m.generateDetailContent(alert)
	m.detailViewport.SetContent(content)
	m.detailViewport.GotoTop()
}

// SetDetailFocused sets focus on the detail pane for scrolling
func (m *AlertsModel) SetDetailFocused(focused bool) {
	m.detailFocused = focused
}

// IsDetailFocused returns whether the detail pane has focus
func (m AlertsModel) IsDetailFocused() bool {
	return m.detailFocused
}

func (m *AlertsModel) updateDimensions() {
	if m.width > 0 {
		m.listWidth = int(float64(m.width) * 0.35)
		m.detailWidth = m.width - m.listWidth - 6

		// Update viewport dimensions
		contentHeight := m.height - 8
		if contentHeight < 5 {
			contentHeight = 5
		}
		// Account for detail container borders/padding
		viewportHeight := contentHeight - 4
		viewportWidth := m.detailWidth - 4
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		if viewportWidth < 20 {
			viewportWidth = 20
		}

		if !m.detailViewportReady {
			m.detailViewport = viewport.New(viewportWidth, viewportHeight)
			m.detailViewport.MouseWheelEnabled = true
			m.detailViewportReady = true
		} else {
			m.detailViewport.Width = viewportWidth
			m.detailViewport.Height = viewportHeight
		}
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
	m.updateViewportContent()
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

func (m *AlertsModel) SetDetailLoading(id string) {
	m.detailLoadingID = id
}

func (m *AlertsModel) ClearDetailLoading() {
	m.detailLoadingID = ""
}

func (m AlertsModel) IsDetailLoading() bool {
	return m.detailLoadingID != ""
}

// IsLoadingAlert returns true if the specified alert ID is currently loading
func (m AlertsModel) IsLoadingAlert(id string) bool {
	return m.detailLoadingID == id
}

func (m *AlertsModel) UpdateAlertDetail(index int, alert *api.Alert) {
	if index >= 0 && index < len(m.alerts) && alert != nil {
		m.alerts[index] = *alert
		// Update viewport content without resetting scroll (detail just loaded)
		if m.detailViewportReady && index == m.cursor {
			content := m.generateDetailContent(alert)
			m.detailViewport.SetContent(content)
		}
	}
}

func (m AlertsModel) View() string {
	contentHeight := m.height - 8
	if contentHeight < 5 {
		contentHeight = 5
	}

	if m.loading {
		// Show loading within the layout structure to prevent jarring shift
		loadingMsg := fmt.Sprintf("%s %s", m.spinnerView, i18n.Tf("loading_page", map[string]interface{}{"Page": m.currentPage}))
		listContent := styles.TextBold.Render(i18n.T("alerts")) + "\n\n" + styles.TextDim.Render(loadingMsg)
		listView := styles.ListContainer.Width(m.listWidth).Height(contentHeight).Render(listContent)
		detailView := styles.DetailContainer.Width(m.detailWidth).Height(contentHeight).Render("")
		return lipgloss.JoinHorizontal(lipgloss.Top, listView, "  ", detailView)
	}

	if m.error != "" {
		return styles.Error.Render(i18n.T("error") + ": " + m.error)
	}

	if len(m.alerts) == 0 {
		return styles.TextDim.Render(i18n.T("no_alerts"))
	}

	listView := m.renderList(contentHeight)
	detailView := m.renderDetail(contentHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, "  ", detailView)
}

func (m AlertsModel) renderList(height int) string {
	var b strings.Builder

	title := styles.TextBold.Render(i18n.T("alerts"))
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
	footer.WriteString(fmt.Sprintf(" %s %d ", i18n.T("page"), m.currentPage))
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
			styles.TextDim.Render(i18n.T("select_alert")),
		)
	}

	// Render with or without viewport
	if !m.detailViewportReady {
		content := m.generateDetailContent(alert)
		return styles.DetailContainer.Width(m.detailWidth).Height(height).Render(content)
	}

	// Add scroll indicator if content is scrollable
	var footer string
	if m.detailViewport.TotalLineCount() > m.detailViewport.VisibleLineCount() {
		scrollPercent := int(m.detailViewport.ScrollPercent() * 100)
		if m.detailFocused {
			footer = styles.Primary.Render(fmt.Sprintf("─── %d%% (j/k scroll, Esc to exit) ───", scrollPercent))
		} else {
			footer = styles.TextDim.Render(fmt.Sprintf("─── %d%% (Enter to scroll) ───", scrollPercent))
		}
	}

	// Use viewport for rendering
	viewportContent := m.detailViewport.View()
	if footer != "" {
		viewportContent = viewportContent + "\n" + footer
	}

	// Use focused style when detail has focus
	containerStyle := styles.DetailContainer
	if m.detailFocused {
		containerStyle = styles.DetailContainerFocused
	}
	return containerStyle.Width(m.detailWidth).Height(height).Render(viewportContent)
}

func (m AlertsModel) generateDetailContent(alert *api.Alert) string {
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
	b.WriteString(fmt.Sprintf("%s: %s  %s: %s %s\n\n", i18n.T("status"), statusBadge, i18n.T("source"), sourceIcon, sourceName))

	// Description (rendered as markdown)
	if alert.Description != "" {
		b.WriteString(styles.TextBold.Render(i18n.T("description")))
		b.WriteString("\n")
		// Render as markdown, use detail width minus padding
		descWidth := m.detailWidth - 4
		if descWidth < 40 {
			descWidth = 40
		}
		b.WriteString(styles.RenderMarkdown(alert.Description, descWidth))
		b.WriteString("\n\n")
	}

	// Timestamps
	b.WriteString(styles.TextBold.Render(i18n.T("timeline")))
	b.WriteString("\n")

	if !alert.CreatedAt.IsZero() {
		b.WriteString(m.renderDetailRow(i18n.T("created"), formatAlertTime(alert.CreatedAt)))
	}
	if alert.StartedAt != nil {
		b.WriteString(m.renderDetailRow(i18n.T("started"), formatAlertTime(*alert.StartedAt)))
	}
	if alert.EndedAt != nil {
		b.WriteString(m.renderDetailRow(i18n.T("ended"), formatAlertTime(*alert.EndedAt)))
	}
	b.WriteString("\n")

	// Services, Environments, Teams
	b.WriteString(renderAlertBulletList(i18n.T("services"), alert.Services))
	b.WriteString(renderAlertBulletList(i18n.T("environments"), alert.Environments))
	b.WriteString(renderAlertBulletList(i18n.T("teams"), alert.Groups))

	// Extended info (populated when DetailLoaded is true)
	if alert.DetailLoaded {
		// Urgency
		if alert.Urgency != "" {
			b.WriteString(m.renderDetailRow(i18n.T("urgency"), alert.Urgency))
		}

		// Responders
		if len(alert.Responders) > 0 {
			b.WriteString("\n")
			b.WriteString(styles.TextBold.Render(i18n.T("responders")))
			b.WriteString("\n")
			for _, responder := range alert.Responders {
				b.WriteString(styles.Text.Render("  • " + responder + "\n"))
			}
		}
	}

	// Labels (sorted for consistent display)
	if len(alert.Labels) > 0 {
		b.WriteString(styles.TextBold.Render(i18n.T("labels")))
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

	// Links section (after Labels)
	rootlyURL := ""
	if alert.ShortID != "" {
		rootlyURL = fmt.Sprintf("https://rootly.com/account/alerts/%s", alert.ShortID)
	}
	if rootlyURL != "" || alert.ExternalURL != "" {
		b.WriteString("\n")
		b.WriteString(styles.TextBold.Render(i18n.T("links")))
		b.WriteString("\n")
		if rootlyURL != "" {
			b.WriteString(m.renderLinkRow(i18n.T("rootly"), rootlyURL))
		}
		if alert.ExternalURL != "" {
			b.WriteString(m.renderLinkRow(i18n.T("source"), alert.ExternalURL))
		}
	}

	// Show loading spinner or hint if detail not loaded
	if m.IsLoadingAlert(alert.ID) {
		b.WriteString("\n")
		_, _ = fmt.Fprintf(&b, "%s %s", m.spinnerView, i18n.T("loading_details"))
	} else if !alert.DetailLoaded {
		b.WriteString("\n")
		b.WriteString(styles.TextDim.Render(i18n.T("press_enter_details")))
	}

	return b.String()
}

func (m AlertsModel) renderDetailRow(label, value string) string {
	return styles.DetailLabel.Render(label+":") + " " + styles.DetailValue.Render(value) + "\n"
}

func (m AlertsModel) renderLinkRow(label, url string) string {
	// Calculate available width for URL display
	// Account for label, colon, space, container padding, and border (~20 chars)
	maxURLLen := m.detailWidth - len(label) - 20
	if maxURLLen < 20 {
		maxURLLen = 20
	}

	displayURL := url
	if len(displayURL) > maxURLLen {
		displayURL = displayURL[:maxURLLen-3] + "..."
	}

	return styles.DetailLabel.Render(label+":") + " " + styles.RenderLink(url, displayURL) + "\n"
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
