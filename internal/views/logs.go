package views

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"

	"github.com/rootlyhq/rootly-tui/internal/debug"
	"github.com/rootlyhq/rootly-tui/internal/styles"
)

// LogsStatusClearMsg is sent to clear the status message
type LogsStatusClearMsg struct{}

var (
	logDebugStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // Gray
	logInfoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue
	logWarnStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	logErrorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // Red
	logSelectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15"))
)

type LogsModel struct {
	Visible    bool
	width      int
	height     int
	scrollPos  int
	logs       []string
	totalLines int

	// Mouse selection
	selecting    bool
	selectStart  int // Line index where selection started
	selectEnd    int // Line index where selection ended
	hasSelection bool
	dialogTop    int // Y offset of dialog content for mouse coordinate translation

	// Status message
	statusMsg     string
	statusTimeout int
}

func NewLogsModel() LogsModel {
	return LogsModel{}
}

func (m LogsModel) Init() tea.Cmd {
	return nil
}

func (m LogsModel) Update(msg tea.Msg) (LogsModel, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case LogsStatusClearMsg:
		m.statusMsg = ""
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			maxScroll := m.maxScrollPos()
			if m.scrollPos < maxScroll {
				m.scrollPos++
			}
		case "k", "up":
			if m.scrollPos > 0 {
				m.scrollPos--
			}
		case "g":
			m.scrollPos = 0
		case "G":
			m.scrollPos = m.maxScrollPos()
		case "c":
			debug.ClearLogs()
			m.logs = nil
			m.totalLines = 0
			m.scrollPos = 0
			m.clearSelection()
		case "y":
			// Yank/copy selected text or all visible logs
			m.copyToClipboard()
			if m.statusMsg != "" {
				cmd = tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
					return LogsStatusClearMsg{}
				})
			}
		case "a":
			// Select all
			if len(m.logs) > 0 {
				m.selectStart = 0
				m.selectEnd = len(m.logs) - 1
				m.hasSelection = true
			}
		case "esc":
			if m.hasSelection {
				m.clearSelection()
			}
		}

	case tea.MouseMsg:
		m = m.handleMouse(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Calculate dialog top position for mouse translation
		m.dialogTop = (msg.Height - m.visibleLines() - 8) / 2
	}

	return m, cmd
}

func (m *LogsModel) handleMouse(msg tea.MouseMsg) LogsModel {
	// Calculate the line index from mouse Y position
	// Account for dialog position and header
	contentStartY := m.dialogTop + 4 // Title + spacing

	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft {
			lineIdx := m.mouseYToLineIndex(msg.Y, contentStartY)
			if lineIdx >= 0 && lineIdx < len(m.logs) {
				m.selecting = true
				m.selectStart = lineIdx
				m.selectEnd = lineIdx
				m.hasSelection = true
			}
		}

	case tea.MouseActionMotion:
		if m.selecting {
			lineIdx := m.mouseYToLineIndex(msg.Y, contentStartY)
			if lineIdx >= 0 && lineIdx < len(m.logs) {
				m.selectEnd = lineIdx
			}
		}

	case tea.MouseActionRelease:
		if msg.Button == tea.MouseButtonLeft {
			m.selecting = false
		}
	}

	return *m
}

func (m *LogsModel) mouseYToLineIndex(mouseY, contentStartY int) int {
	// Convert mouse Y to log line index
	relativeY := mouseY - contentStartY
	if relativeY < 0 {
		return -1
	}
	lineIdx := m.scrollPos + relativeY
	return lineIdx
}

func (m *LogsModel) clearSelection() {
	m.selecting = false
	m.selectStart = 0
	m.selectEnd = 0
	m.hasSelection = false
}

func (m *LogsModel) getSelectedLines() []string {
	if !m.hasSelection || len(m.logs) == 0 {
		return nil
	}

	start, end := m.selectStart, m.selectEnd
	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if end >= len(m.logs) {
		end = len(m.logs) - 1
	}

	return m.logs[start : end+1]
}

func (m *LogsModel) copyToClipboard() {
	var lines []string

	if m.hasSelection {
		lines = m.getSelectedLines()
	} else {
		// Copy all visible logs if no selection
		visibleLines := m.visibleLines()
		start := m.scrollPos
		end := start + visibleLines
		if end > len(m.logs) {
			end = len(m.logs)
		}
		if start < len(m.logs) {
			lines = m.logs[start:end]
		}
	}

	if len(lines) == 0 {
		return
	}

	// Clean up lines and join
	var cleaned []string
	for _, line := range lines {
		cleaned = append(cleaned, strings.TrimSuffix(line, "\n"))
	}

	text := strings.Join(cleaned, "\n")

	// Copy to clipboard
	if err := clipboard.Init(); err != nil {
		debug.Logger.Error("Failed to initialize clipboard",
			"error", err,
			"hint", "Clipboard requires CGO_ENABLED=1. On Linux, also install xclip/xsel. On headless systems, clipboard is unavailable.",
		)
		m.statusMsg = "Clipboard unavailable (see logs)"
		m.statusTimeout = 3
		return
	}

	clipboard.Write(clipboard.FmtText, []byte(text))
	debug.Logger.Debug("Copied to clipboard", "lines", len(lines), "bytes", len(text))
	m.statusMsg = "Copied!"
	m.statusTimeout = 2
}

func (m *LogsModel) isLineSelected(lineIdx int) bool {
	if !m.hasSelection {
		return false
	}

	start, end := m.selectStart, m.selectEnd
	if start > end {
		start, end = end, start
	}

	return lineIdx >= start && lineIdx <= end
}

func (m *LogsModel) visibleLines() int {
	lines := m.height - 8 // Account for header, footer, borders
	if lines < 1 {
		return 1
	}
	return lines
}

func (m *LogsModel) maxScrollPos() int {
	maxScroll := m.totalLines - m.visibleLines()
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func (m *LogsModel) Toggle() {
	m.Visible = !m.Visible
	if m.Visible {
		m.Refresh()
	}
}

func (m *LogsModel) Show() {
	m.Visible = true
	m.Refresh()
}

func (m *LogsModel) Hide() {
	m.Visible = false
}

func (m *LogsModel) Refresh() {
	m.logs = debug.GetLogs()
	m.totalLines = len(m.logs)
	// Auto-scroll to bottom on refresh
	m.scrollPos = m.maxScrollPos()
}

func (m *LogsModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

func (m LogsModel) View() string {
	var b strings.Builder

	// Title
	title := styles.DialogTitle.Render("Debug Logs")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Log entries
	visibleLines := m.visibleLines()
	if visibleLines < 1 {
		visibleLines = 1
	}

	if len(m.logs) == 0 {
		b.WriteString(styles.TextDim.Render("No logs yet. Logs are captured automatically."))
		b.WriteString("\n")
	} else {
		start := m.scrollPos
		end := start + visibleLines
		if end > len(m.logs) {
			end = len(m.logs)
		}

		for i := start; i < end; i++ {
			entry := m.logs[i]
			// Trim trailing newline
			entry = strings.TrimSuffix(entry, "\n")
			// Truncate long lines
			maxLen := m.width - 10
			if maxLen > 0 && len(entry) > maxLen {
				entry = entry[:maxLen-3] + "..."
			}
			// Check if line is selected
			if m.isLineSelected(i) {
				b.WriteString(logSelectedStyle.Render(entry))
			} else {
				// Colorize based on log level
				b.WriteString(colorizeLogEntry(entry))
			}
			b.WriteString("\n")
		}
	}

	// Scroll indicator
	if m.totalLines > visibleLines {
		b.WriteString("\n")
		// Scroll info is rendered below using itoa
		b.WriteString(styles.TextDim.Render(
			"  Showing lines " + itoa(m.scrollPos+1) + "-" +
				itoa(minInt(m.scrollPos+visibleLines, m.totalLines)) +
				" of " + itoa(m.totalLines),
		))
	}

	b.WriteString("\n\n")

	// Status message (if any)
	if m.statusMsg != "" {
		b.WriteString(styles.Success.Render(m.statusMsg))
		b.WriteString("\n\n")
	}

	// Help
	help := styles.HelpBar.Render("j/k scroll • g/G top/bottom • a select all • y copy • c clear • l/Esc close")
	b.WriteString(help)

	// Wrap in dialog
	content := b.String()
	dialogWidth := m.width - 4
	if dialogWidth < 40 {
		dialogWidth = 40
	}
	dialog := styles.Dialog.Width(dialogWidth).Render(content)

	return dialog
}

// Simple int to string conversion
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// colorizeLogEntry applies color based on log level
func colorizeLogEntry(entry string) string {
	// charmbracelet/log format: "LEVL prefix message key=value..."
	// Look for level indicators at the start
	upperEntry := strings.ToUpper(entry)

	switch {
	case strings.Contains(upperEntry, "ERRO") || strings.Contains(upperEntry, "ERROR"):
		return logErrorStyle.Render(entry)
	case strings.Contains(upperEntry, "WARN"):
		return logWarnStyle.Render(entry)
	case strings.Contains(upperEntry, "INFO"):
		return logInfoStyle.Render(entry)
	case strings.Contains(upperEntry, "DEBU") || strings.Contains(upperEntry, "DEBUG"):
		return logDebugStyle.Render(entry)
	default:
		return styles.Text.Render(entry)
	}
}
