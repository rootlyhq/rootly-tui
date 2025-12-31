package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rootlyhq/rootly-tui/internal/api"
	"github.com/rootlyhq/rootly-tui/internal/styles"
)

type IncidentsModel struct {
	incidents   []api.Incident
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
}

func NewIncidentsModel() IncidentsModel {
	return IncidentsModel{
		incidents:   []api.Incident{},
		cursor:      0,
		currentPage: 1,
	}
}

func (m IncidentsModel) Init() tea.Cmd {
	return nil
}

func (m IncidentsModel) Update(msg tea.Msg) (IncidentsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.incidents)-1 {
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
			if len(m.incidents) > 0 {
				m.cursor = len(m.incidents) - 1
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

func (m *IncidentsModel) updateDimensions() {
	if m.width > 0 {
		m.listWidth = int(float64(m.width) * 0.4)
		m.detailWidth = m.width - m.listWidth - 6 // Account for borders and padding
	}
}

func (m *IncidentsModel) SetIncidents(incidents []api.Incident, pagination api.PaginationInfo) {
	m.incidents = incidents
	m.loading = false
	m.error = ""
	m.currentPage = pagination.CurrentPage
	m.hasNext = pagination.HasNext
	m.hasPrev = pagination.HasPrev
	if m.cursor >= len(incidents) && len(incidents) > 0 {
		m.cursor = len(incidents) - 1
	}
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
		m.cursor = 0
	}
}

func (m *IncidentsModel) PrevPage() {
	if m.hasPrev && m.currentPage > 1 {
		m.currentPage--
		m.cursor = 0
	}
}

func (m IncidentsModel) SelectedIncident() *api.Incident {
	if m.cursor >= 0 && m.cursor < len(m.incidents) {
		return &m.incidents[m.cursor]
	}
	return nil
}

func (m IncidentsModel) View() string {
	// Calculate available height for content
	contentHeight := m.height - 8 // Account for header, help bar, etc.
	if contentHeight < 5 {
		contentHeight = 5
	}

	if m.loading {
		// Show loading within the layout structure to prevent jarring shift
		loadingMsg := fmt.Sprintf("%s Loading page %d...", m.spinnerView, m.currentPage)
		listContent := styles.TextBold.Render("INCIDENTS") + "\n\n" + styles.TextDim.Render(loadingMsg)
		listView := styles.ListContainer.Width(m.listWidth).Height(contentHeight).Render(listContent)
		detailView := styles.DetailContainer.Width(m.detailWidth).Height(contentHeight).Render("")
		return lipgloss.JoinHorizontal(lipgloss.Top, listView, "  ", detailView)
	}

	if m.error != "" {
		return styles.Error.Render("Error: " + m.error)
	}

	if len(m.incidents) == 0 {
		return styles.TextDim.Render("No incidents found")
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
	title := styles.TextBold.Render("INCIDENTS")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Calculate visible range
	maxVisible := height - 4
	if maxVisible < 1 {
		maxVisible = 1
	}

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.incidents) {
		end = len(m.incidents)
	}

	// Render items
	for i := start; i < end; i++ {
		inc := m.incidents[i]

		// Sequential ID (e.g., INC-123)
		seqID := inc.SequentialID
		if seqID == "" {
			seqID = "INC-?"
		}

		// Status (padded for alignment)
		status := inc.Status
		if len(status) > 12 {
			status = status[:12]
		}
		statusPadded := fmt.Sprintf("%-12s", status)

		// Title (truncated)
		titleMaxLen := m.listWidth - 35
		if titleMaxLen < 10 {
			titleMaxLen = 10
		}
		title := inc.Summary
		if title == "" {
			title = inc.Title
		}
		if len(title) > titleMaxLen {
			title = title[:titleMaxLen-3] + "..."
		}

		// Format: "▶ ▁▃▅▇ INC-123  started      Title here" (▶ for selected)
		// Selected items use plain text so selection color applies uniformly
		if i == m.cursor {
			sevPlain := severitySignalPlain(inc.Severity)
			line := fmt.Sprintf("▶ %s %-8s %s %s", sevPlain, seqID, statusPadded, title)
			b.WriteString(styles.ListItemSelected.Width(m.listWidth - 4).Render(line))
		} else {
			sev := styles.RenderSeveritySignal(inc.Severity)
			line := fmt.Sprintf("  %s %-8s %s %s", sev, seqID, styles.RenderStatus(statusPadded), title)
			b.WriteString(styles.ListItem.Width(m.listWidth - 4).Render(line))
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
	if len(m.incidents) > 0 {
		footer.WriteString(styles.TextDim.Render(fmt.Sprintf("  (%d-%d)", m.cursor+1, len(m.incidents))))
	}

	b.WriteString(footer.String())

	content := b.String()
	return styles.ListContainer.Width(m.listWidth).Height(height).Render(content)
}

func (m IncidentsModel) renderDetail(height int) string {
	inc := m.SelectedIncident()
	if inc == nil {
		return styles.DetailContainer.Width(m.detailWidth).Height(height).Render(
			styles.TextDim.Render("Select an incident to view details"),
		)
	}

	var b strings.Builder

	// Title line: [INC-XXX] Title
	title := inc.Title
	if title == "" {
		title = inc.Summary
	}
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
	b.WriteString(fmt.Sprintf("Status: %s  Severity: %s %s\n\n", statusBadge, sevSignal, sevBadge))

	// Description (shows the summary if different from title)
	if inc.Summary != "" && inc.Summary != inc.Title {
		b.WriteString(styles.TextBold.Render("Description"))
		b.WriteString("\n")
		b.WriteString(styles.TextDim.Render(inc.Summary))
		b.WriteString("\n\n")
	}

	// Timeline
	b.WriteString(styles.TextBold.Render("Timeline"))
	b.WriteString("\n")

	if !inc.CreatedAt.IsZero() {
		b.WriteString(m.renderDetailRow("Created", formatTime(inc.CreatedAt)))
	}
	if inc.StartedAt != nil {
		b.WriteString(m.renderDetailRow("Started", formatTime(*inc.StartedAt)))
	}
	if inc.DetectedAt != nil {
		b.WriteString(m.renderDetailRow("Detected", formatTime(*inc.DetectedAt)))
	}
	if inc.AcknowledgedAt != nil {
		b.WriteString(m.renderDetailRow("Acknowledged", formatTime(*inc.AcknowledgedAt)))
	}
	if inc.MitigatedAt != nil {
		b.WriteString(m.renderDetailRow("Mitigated", formatTime(*inc.MitigatedAt)))
	}
	if inc.ResolvedAt != nil {
		b.WriteString(m.renderDetailRow("Resolved", formatTime(*inc.ResolvedAt)))
	}
	b.WriteString("\n")

	// Services
	if len(inc.Services) > 0 {
		b.WriteString(m.renderDetailRow("Services", strings.Join(inc.Services, ", ")))
	}

	// Environments
	if len(inc.Environments) > 0 {
		b.WriteString(m.renderDetailRow("Environments", strings.Join(inc.Environments, ", ")))
	}

	// Teams
	if len(inc.Teams) > 0 {
		b.WriteString(m.renderDetailRow("Teams", strings.Join(inc.Teams, ", ")))
	}

	// External links (clickable)
	if inc.SlackChannelURL != "" || inc.JiraIssueURL != "" {
		b.WriteString("\n")
		b.WriteString(styles.TextBold.Render("Links"))
		b.WriteString("\n")
		if inc.SlackChannelURL != "" {
			b.WriteString(m.renderLinkRow("Slack", inc.SlackChannelURL))
		}
		if inc.JiraIssueURL != "" {
			b.WriteString(m.renderLinkRow("Jira", inc.JiraIssueURL))
		}
	}

	content := b.String()
	return styles.DetailContainer.Width(m.detailWidth).Height(height).Render(content)
}

func (m IncidentsModel) renderDetailRow(label, value string) string {
	return styles.DetailLabel.Render(label+":") + " " + styles.DetailValue.Render(value) + "\n"
}

func (m IncidentsModel) renderLinkRow(label, url string) string {
	return styles.DetailLabel.Render(label+":") + " " + styles.RenderLink(url, url) + "\n"
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
