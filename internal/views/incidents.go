package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/evertras/bubble-table/table"

	"github.com/rootlyhq/rootly-tui/internal/api"
	"github.com/rootlyhq/rootly-tui/internal/i18n"
	"github.com/rootlyhq/rootly-tui/internal/styles"
)

// renderBulletList renders a section with a bold title and bullet list using lipgloss/list
func renderBulletList(title string, items []string) string {
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

// Column keys for incidents table
const (
	colKeyIndicator = "indicator"
	colKeySev       = "sev"
	colKeyID        = "id"
	colKeyStatus    = "status"
	colKeyTitle     = "title"
)

// Row indicator for selected row
const rowIndicator = "▶"

type IncidentsModel struct {
	incidents   []api.Incident
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
	// Detail loading state - tracks which incident ID is currently loading (empty = not loading)
	detailLoadingID string
	// Detail viewport for scrollable content
	detailViewport      viewport.Model
	detailViewportReady bool
	detailFocused       bool // Whether detail pane has focus (for scrolling)
	// Table for list view
	table table.Model
}

// borderNoDividers creates a rounded border without vertical column dividers
func borderNoDividers() table.Border {
	return table.Border{
		Top:    "─",
		Left:   "│",
		Right:  "│",
		Bottom: "─",

		TopRight:    "╮",
		TopLeft:     "╭",
		BottomRight: "╯",
		BottomLeft:  "╰",

		TopJunction:    "─",
		LeftJunction:   "│",
		RightJunction:  "│",
		BottomJunction: "─",
		InnerJunction:  " ",

		InnerDivider: " ", // Space instead of vertical line between columns
	}
}

func NewIncidentsModel() IncidentsModel {
	// Define table columns with i18n headers using evertras/bubble-table
	columns := []table.Column{
		table.NewColumn(colKeyIndicator, "", 2), // Selection indicator column
		table.NewColumn(colKeySev, i18n.T("col_sev"), 4),
		table.NewColumn(colKeyID, i18n.T("col_id"), 10),
		table.NewColumn(colKeyStatus, i18n.T("status"), 12),
		table.NewFlexColumn(colKeyTitle, i18n.T("col_title"), 1), // Flex to fill remaining space
	}

	t := table.New(columns).
		Focused(true).
		Border(borderNoDividers()).
		WithBaseStyle(lipgloss.NewStyle().Foreground(styles.ColorText)).
		HighlightStyle(lipgloss.NewStyle()). // No background highlight, arrow shows selection
		HeaderStyle(lipgloss.NewStyle().Bold(true).Foreground(styles.ColorText))

	return IncidentsModel{
		incidents:   []api.Incident{},
		currentPage: 1,
		table:       t,
	}
}

func (m IncidentsModel) Init() tea.Cmd {
	return nil
}

func (m IncidentsModel) Update(msg tea.Msg) (IncidentsModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Track previous cursor position to detect changes
	prevCursor := m.table.GetHighlightedRowIndex()

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

		// Handle navigation keys ourselves to prevent table's wrap-around behavior
		switch msg.String() {
		case "j", "down":
			cursor := m.table.GetHighlightedRowIndex()
			if cursor < len(m.incidents)-1 {
				m.table = m.table.WithHighlightedRow(cursor + 1)
				m.updateRowIndicators()
				m.updateViewportContent()
			}
			return m, nil
		case "k", "up":
			cursor := m.table.GetHighlightedRowIndex()
			if cursor > 0 {
				m.table = m.table.WithHighlightedRow(cursor - 1)
				m.updateRowIndicators()
				m.updateViewportContent()
			}
			return m, nil
		case "g":
			// Go to first row
			m.table = m.table.WithHighlightedRow(0)
			m.updateRowIndicators()
			m.updateViewportContent()
			return m, nil
		case "G":
			// Go to last row
			if len(m.incidents) > 0 {
				m.table = m.table.WithHighlightedRow(len(m.incidents) - 1)
				m.updateRowIndicators()
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

	// Forward other messages to table
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	// Update detail viewport if cursor changed
	if m.table.GetHighlightedRowIndex() != prevCursor {
		m.updateViewportContent()
	}

	// Forward mouse messages to viewport for scrolling
	if m.detailViewportReady && m.detailFocused {
		m.detailViewport, cmd = m.detailViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// updateRowIndicators updates the arrow indicator to show on the current row
func (m *IncidentsModel) updateRowIndicators() {
	if len(m.incidents) == 0 {
		return
	}
	cursor := m.table.GetHighlightedRowIndex()
	rows := make([]table.Row, len(m.incidents))
	for i, inc := range m.incidents {
		seqID := inc.SequentialID
		if seqID == "" {
			seqID = "INC-?"
		}
		status := inc.Status
		if len(status) > 12 {
			status = status[:12]
		}
		title := inc.Summary
		if title == "" {
			title = inc.Title
		}
		title = strings.ReplaceAll(title, "\n", " ")
		title = strings.ReplaceAll(title, "\r", "")

		sevCell := table.NewStyledCell(severitySignalPlain(inc.Severity), severityStyle(inc.Severity))
		statusCell := table.NewStyledCell(status, statusStyle(status))

		indicator := ""
		if i == cursor {
			indicator = rowIndicator
		}

		rows[i] = table.NewRow(table.RowData{
			colKeyIndicator: indicator,
			colKeySev:       sevCell,
			colKeyID:        seqID,
			colKeyStatus:    statusCell,
			colKeyTitle:     title,
		})
	}
	m.table = m.table.WithRows(rows)
}

// updateViewportContent updates the viewport content when data changes
func (m *IncidentsModel) updateViewportContent() {
	if !m.detailViewportReady {
		return
	}
	inc := m.SelectedIncident()
	if inc == nil {
		return
	}
	content := m.generateDetailContent(inc)
	m.detailViewport.SetContent(content)
	m.detailViewport.GotoTop()
}

func (m *IncidentsModel) updateDimensions() {
	if m.width > 0 {
		m.listWidth = int(float64(m.width) * 0.35)
		m.detailWidth = m.width - m.listWidth - 6 // Account for borders and padding

		// Update viewport dimensions
		contentHeight := m.height - 8
		if contentHeight < 5 {
			contentHeight = 5
		}

		// Update table dimensions using fluent API
		// Account for: title (2 lines), footer (2 lines), container borders (2)
		tableHeight := contentHeight - 6
		if tableHeight < 3 {
			tableHeight = 3
		}
		m.table = m.table.WithTargetWidth(m.listWidth - 4).WithMinimumHeight(tableHeight)

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

func (m *IncidentsModel) SetIncidents(incidents []api.Incident, pagination api.PaginationInfo) {
	m.incidents = incidents
	m.loading = false
	m.error = ""
	m.currentPage = pagination.CurrentPage
	m.hasNext = pagination.HasNext
	m.hasPrev = pagination.HasPrev

	// Build table rows from incidents with styled cells
	rows := make([]table.Row, len(incidents))
	cursor := m.table.GetHighlightedRowIndex()
	for i, inc := range incidents {
		seqID := inc.SequentialID
		if seqID == "" {
			seqID = "INC-?"
		}
		status := inc.Status
		if len(status) > 12 {
			status = status[:12]
		}
		title := inc.Summary
		if title == "" {
			title = inc.Title
		}
		title = strings.ReplaceAll(title, "\n", " ")
		title = strings.ReplaceAll(title, "\r", "")

		// Create styled cells using evertras/bubble-table
		sevCell := table.NewStyledCell(severitySignalPlain(inc.Severity), severityStyle(inc.Severity))
		statusCell := table.NewStyledCell(status, statusStyle(status))

		// Show indicator for highlighted row
		indicator := ""
		if i == cursor {
			indicator = rowIndicator
		}

		rows[i] = table.NewRow(table.RowData{
			colKeyIndicator: indicator,
			colKeySev:       sevCell,
			colKeyID:        seqID,
			colKeyStatus:    statusCell,
			colKeyTitle:     title,
		})
	}
	m.table = m.table.WithRows(rows)

	// Adjust cursor if needed
	if cursor >= len(incidents) && len(incidents) > 0 {
		m.table = m.table.WithHighlightedRow(len(incidents) - 1)
	}
	m.updateViewportContent()
}

func (m *IncidentsModel) SetLoading(loading bool) {
	m.loading = loading
}

func (m *IncidentsModel) SetSpinner(spinner string) {
	m.spinnerView = spinner
}

func (m *IncidentsModel) SetError(err string) {
	m.error = err
	m.loading = false
}

func (m *IncidentsModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.updateDimensions()
}

// Pagination methods
func (m IncidentsModel) CurrentPage() int {
	return m.currentPage
}

func (m IncidentsModel) HasNextPage() bool {
	return m.hasNext
}

func (m IncidentsModel) HasPrevPage() bool {
	return m.hasPrev
}

func (m *IncidentsModel) NextPage() {
	if m.hasNext {
		m.currentPage++
		m.table = m.table.WithHighlightedRow(0)
	}
}

func (m *IncidentsModel) PrevPage() {
	if m.hasPrev && m.currentPage > 1 {
		m.currentPage--
		m.table = m.table.WithHighlightedRow(0)
	}
}

func (m IncidentsModel) SelectedIncident() *api.Incident {
	cursor := m.table.GetHighlightedRowIndex()
	if cursor >= 0 && cursor < len(m.incidents) {
		return &m.incidents[cursor]
	}
	return nil
}

func (m IncidentsModel) SelectedIndex() int {
	return m.table.GetHighlightedRowIndex()
}

func (m *IncidentsModel) SetDetailLoading(id string) {
	m.detailLoadingID = id
}

func (m *IncidentsModel) ClearDetailLoading() {
	m.detailLoadingID = ""
}

func (m IncidentsModel) IsDetailLoading() bool {
	return m.detailLoadingID != ""
}

// IsLoadingIncident returns true if the specified incident ID is currently loading
func (m IncidentsModel) IsLoadingIncident(id string) bool {
	return m.detailLoadingID == id
}

// SetDetailFocused sets focus on the detail pane for scrolling
func (m *IncidentsModel) SetDetailFocused(focused bool) {
	m.detailFocused = focused
}

// IsDetailFocused returns whether the detail pane has focus
func (m IncidentsModel) IsDetailFocused() bool {
	return m.detailFocused
}

func (m *IncidentsModel) UpdateIncidentDetail(index int, incident *api.Incident) {
	if index >= 0 && index < len(m.incidents) && incident != nil {
		m.incidents[index] = *incident
		// Update viewport content without resetting scroll (detail just loaded)
		if m.detailViewportReady && index == m.table.GetHighlightedRowIndex() {
			content := m.generateDetailContent(incident)
			m.detailViewport.SetContent(content)
		}
	}
}

func (m IncidentsModel) View() string {
	// Calculate available height for content
	contentHeight := m.height - 8 // Account for header, help bar, etc.
	if contentHeight < 5 {
		contentHeight = 5
	}

	if m.loading {
		// Show loading within the layout structure to prevent jarring shift
		loadingMsg := fmt.Sprintf("%s %s", m.spinnerView, i18n.Tf("loading_page", map[string]any{"Page": m.currentPage}))
		listContent := styles.TextBold.Render(i18n.T("incidents")) + "\n\n" + styles.TextDim.Render(loadingMsg)
		listView := styles.ListContainer.Width(m.listWidth).Height(contentHeight).Render(listContent)
		detailView := styles.DetailContainer.Width(m.detailWidth).Height(contentHeight).Render("")
		return lipgloss.JoinHorizontal(lipgloss.Top, listView, "  ", detailView)
	}

	if m.error != "" {
		return styles.Error.Render(i18n.T("error") + ": " + m.error)
	}

	if len(m.incidents) == 0 {
		return styles.TextDim.Render(i18n.T("no_incidents"))
	}

	// Build list view
	listView := m.renderList(contentHeight)

	// Build detail view
	detailView := m.renderDetail(contentHeight)

	// Join horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, listView, "  ", detailView)
}

func (m IncidentsModel) renderList(height int) string {
	var b strings.Builder

	// Title
	title := styles.TextBold.Render(i18n.T("incidents"))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render table
	b.WriteString(m.table.View())
	b.WriteString("\n")

	// Page navigation footer
	var footer strings.Builder
	if m.hasPrev {
		footer.WriteString(styles.TextDim.Render("← ["))
	} else {
		footer.WriteString(styles.TextDim.Render("  "))
	}
	fmt.Fprintf(&footer, " %s %d ", i18n.T("page"), m.currentPage)
	if m.hasNext {
		footer.WriteString(styles.TextDim.Render("] →"))
	}

	// Item count
	if len(m.incidents) > 0 {
		footer.WriteString(styles.TextDim.Render(fmt.Sprintf("  (%d-%d)", m.table.GetHighlightedRowIndex()+1, len(m.incidents))))
	}

	b.WriteString(footer.String())

	content := b.String()
	return styles.ListContainer.Width(m.listWidth).Height(height).Render(content)
}

func (m IncidentsModel) renderDetail(height int) string {
	inc := m.SelectedIncident()
	if inc == nil {
		return styles.DetailContainer.Width(m.detailWidth).Height(height).Render(
			styles.TextDim.Render(i18n.T("select_incident")),
		)
	}

	// Render with or without viewport
	if !m.detailViewportReady {
		content := m.generateDetailContent(inc)
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

func (m IncidentsModel) generateDetailContent(inc *api.Incident) string {
	var b strings.Builder

	// Title line: [INC-XXX] Title (strip newlines for single-line display)
	title := inc.Title
	if title == "" {
		title = inc.Summary
	}
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", "")
	if inc.SequentialID != "" {
		b.WriteString(styles.Primary.Bold(true).Render("[" + inc.SequentialID + "]"))
		b.WriteString(" ")
	}
	b.WriteString(styles.DetailTitle.Render(title))
	b.WriteString("\n\n")

	// Status and Severity row
	statusBadge := styles.RenderStatus(inc.Status)
	sevSignal := styles.RenderSeveritySignal(inc.Severity)
	sevBadge := styles.RenderSeverity(inc.Severity)
	b.WriteString(fmt.Sprintf("%s: %s  %s: %s %s", i18n.T("status"), statusBadge, i18n.T("severity"), sevSignal, sevBadge))

	// Show creator if available (from detail view)
	if inc.CreatedByName != "" {
		creatorInfo := styles.RenderNameWithEmail(inc.CreatedByName, inc.CreatedByEmail)
		fmt.Fprintf(&b, "  %s: %s", i18n.T("created_by"), creatorInfo)
	}
	b.WriteString("\n\n")

	// Links section (high up for quick access)
	rootlyURL := inc.ShortURL
	if rootlyURL == "" {
		rootlyURL = inc.URL
	}
	if rootlyURL == "" && inc.ID != "" {
		rootlyURL = fmt.Sprintf("https://rootly.com/account/incidents/%s", inc.ID)
	}
	if inc.SlackChannelURL != "" || inc.JiraIssueURL != "" || rootlyURL != "" {
		b.WriteString(styles.TextBold.Render(i18n.T("links")))
		b.WriteString("\n")
		if rootlyURL != "" {
			b.WriteString(m.renderLinkRow(i18n.T("rootly"), rootlyURL))
		}
		if inc.SlackChannelURL != "" {
			b.WriteString(m.renderLinkRow(i18n.T("slack"), inc.SlackChannelURL))
		}
		if inc.JiraIssueURL != "" {
			b.WriteString(m.renderLinkRow(i18n.T("jira"), inc.JiraIssueURL))
		}
		b.WriteString("\n")
	}

	// Description (shows the summary if different from title, rendered as markdown)
	summaryClean := strings.ReplaceAll(inc.Summary, "\r", "")
	titleClean := strings.ReplaceAll(title, "\n", " ")
	titleClean = strings.ReplaceAll(titleClean, "\r", "")
	if summaryClean != "" && strings.TrimSpace(summaryClean) != strings.TrimSpace(titleClean) {
		b.WriteString(styles.TextBold.Render(i18n.T("description")))
		b.WriteString("\n")
		// Render as markdown, use detail width minus padding
		descWidth := m.detailWidth - 4
		if descWidth < 40 {
			descWidth = 40
		}
		b.WriteString(styles.RenderMarkdown(summaryClean, descWidth))
		b.WriteString("\n\n")
	}

	// Timeline
	b.WriteString(styles.TextBold.Render(i18n.T("timeline")))
	b.WriteString("\n")

	if !inc.CreatedAt.IsZero() {
		b.WriteString(m.renderDetailRow(i18n.T("created"), formatTime(inc.CreatedAt)))
	}
	if inc.StartedAt != nil {
		b.WriteString(m.renderDetailRow(i18n.T("started"), formatTime(*inc.StartedAt)))
	}
	if inc.DetectedAt != nil {
		b.WriteString(m.renderDetailRow(i18n.T("detected"), formatTime(*inc.DetectedAt)))
	}
	if inc.AcknowledgedAt != nil {
		b.WriteString(m.renderDetailRow(i18n.T("acknowledged"), formatTime(*inc.AcknowledgedAt)))
	}
	if inc.MitigatedAt != nil {
		b.WriteString(m.renderDetailRow(i18n.T("mitigated"), formatTime(*inc.MitigatedAt)))
	}
	if inc.ResolvedAt != nil {
		b.WriteString(m.renderDetailRow(i18n.T("resolved"), formatTime(*inc.ResolvedAt)))
	}
	b.WriteString("\n")

	// Services, Environments, Teams
	b.WriteString(renderBulletList(i18n.T("services"), inc.Services))
	b.WriteString(renderBulletList(i18n.T("environments"), inc.Environments))
	b.WriteString(renderBulletList(i18n.T("teams"), inc.Teams))

	// Extended info (populated when DetailLoaded is true)
	if inc.DetailLoaded {
		// Roles (Commander, Communicator, etc.)
		if len(inc.Roles) > 0 {
			b.WriteString(styles.TextBold.Render(i18n.T("roles")))
			b.WriteString("\n")
			for _, role := range inc.Roles {
				userName := strings.TrimSpace(role.UserName)
				if userName == "" {
					continue
				}
				roleName := strings.TrimSpace(role.Name)
				userEmail := strings.TrimSpace(role.UserEmail)
				b.WriteString(styles.DetailLabel.Render(roleName + ":"))
				b.WriteString(" ")
				b.WriteString(styles.RenderNameWithEmail(userName, userEmail))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}

		// Causes, Types, Functionalities
		b.WriteString(renderBulletList(i18n.T("causes"), inc.Causes))
		b.WriteString(renderBulletList(i18n.T("types"), inc.IncidentTypes))
		b.WriteString(renderBulletList(i18n.T("functionalities"), inc.Functionalities))
	}

	// Show loading spinner or hint if detail not loaded
	if m.IsLoadingIncident(inc.ID) {
		b.WriteString("\n")
		fmt.Fprintf(&b, "%s %s", m.spinnerView, i18n.T("loading_details"))
	} else if !inc.DetailLoaded {
		b.WriteString("\n")
		b.WriteString(styles.TextDim.Render(i18n.T("press_enter_details")))
	}

	return b.String()
}

func (m IncidentsModel) renderDetailRow(label, value string) string {
	return styles.DetailLabel.Render(label+":") + " " + styles.DetailValue.Render(value) + "\n"
}

func (m IncidentsModel) renderLinkRow(label, url string) string {
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

// severitySignalPlain returns plain signal bars without color styling
func severitySignalPlain(severity string) string {
	switch severity {
	case "critical", "Critical", "CRITICAL", "sev0", "SEV0":
		return "▁▃▅▇"
	case "high", "High", "HIGH", "sev1", "SEV1":
		return "▁▃▅░"
	case "medium", "Medium", "MEDIUM", "sev2", "SEV2":
		return "▁▃░░"
	case "low", "Low", "LOW", "sev3", "SEV3":
		return "▁░░░"
	default:
		return "░░░░"
	}
}

// severityStyle returns the lipgloss style for a severity level
func severityStyle(severity string) lipgloss.Style {
	switch severity {
	case "critical", "Critical", "CRITICAL", "sev0", "SEV0":
		return lipgloss.NewStyle().Foreground(styles.ColorCritical).Bold(true)
	case "high", "High", "HIGH", "sev1", "SEV1":
		return lipgloss.NewStyle().Foreground(styles.ColorHigh).Bold(true)
	case "medium", "Medium", "MEDIUM", "sev2", "SEV2":
		return lipgloss.NewStyle().Foreground(styles.ColorMedium).Bold(true)
	case "low", "Low", "LOW", "sev3", "SEV3":
		return lipgloss.NewStyle().Foreground(styles.ColorLow).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(styles.ColorMuted)
	}
}

// statusStyle returns the lipgloss style for a status
func statusStyle(status string) lipgloss.Style {
	s := strings.ToLower(strings.TrimSpace(status))
	switch s {
	case "open", "triggered", "firing", "critical":
		return lipgloss.NewStyle().Foreground(styles.ColorPastelRed)
	case "started", "in_progress", "acknowledged", "investigating", "identified", "monitoring", "mitigated":
		return lipgloss.NewStyle().Foreground(styles.ColorPastelYellow)
	case "resolved", "fixed":
		return lipgloss.NewStyle().Foreground(styles.ColorPastelGreen)
	default:
		return lipgloss.NewStyle().Foreground(styles.ColorPastelGray)
	}
}

func formatTime(t time.Time) string {
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
