package styles

import "github.com/charmbracelet/lipgloss"

// Color palette (Rootly brand colors)
var (
	ColorPrimary    = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary  = lipgloss.Color("#6366F1") // Indigo
	ColorSuccess    = lipgloss.Color("#10B981") // Green
	ColorWarning    = lipgloss.Color("#F59E0B") // Amber
	ColorDanger     = lipgloss.Color("#EF4444") // Red
	ColorInfo       = lipgloss.Color("#4D96FF") // Blue
	ColorMuted      = lipgloss.Color("#6B7280") // Gray
	ColorText       = lipgloss.Color("#F9FAFB") // Light
	ColorTextDim    = lipgloss.Color("#9CA3AF") // Dimmed
	ColorBackground = lipgloss.Color("#1F2937") // Dark
	ColorBorder     = lipgloss.Color("#374151") // Border gray
	ColorHighlight  = lipgloss.Color("#8B5CF6") // Lighter purple
)

// Severity colors
var (
	ColorCritical = lipgloss.Color("#DC2626") // Dark red
	ColorHigh     = lipgloss.Color("#EA580C") // Orange
	ColorMedium   = lipgloss.Color("#CA8A04") // Yellow
	ColorLow      = lipgloss.Color("#2563EB") // Blue
)

// Text styles
var (
	Primary   = lipgloss.NewStyle().Foreground(ColorPrimary)
	Secondary = lipgloss.NewStyle().Foreground(ColorSecondary)
	Success   = lipgloss.NewStyle().Foreground(ColorSuccess)
	Warning   = lipgloss.NewStyle().Foreground(ColorWarning)
	Danger    = lipgloss.NewStyle().Foreground(ColorDanger)
	Info      = lipgloss.NewStyle().Foreground(ColorInfo)
	Muted     = lipgloss.NewStyle().Foreground(ColorMuted)
	Text      = lipgloss.NewStyle().Foreground(ColorText)
	TextDim   = lipgloss.NewStyle().Foreground(ColorTextDim)
	TextBold  = lipgloss.NewStyle().Foreground(ColorText).Bold(true)
)

// Layout styles
var (
	App = lipgloss.NewStyle().
		Padding(1, 2)

	Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText).
		Background(ColorPrimary).
		Padding(0, 2).
		MarginBottom(1)

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText)

	// Tab styles
	TabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText).
			Background(ColorPrimary).
			Padding(0, 2)

	TabInactive = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Background(ColorBorder).
			Padding(0, 2)

	// List styles
	ListContainer = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	ListItem = lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1)

	ListItemSelected = lipgloss.NewStyle().
				Foreground(ColorText).
				Background(ColorPrimary).
				Bold(true).
				Padding(0, 1)

	// Detail pane styles
	DetailContainer = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	DetailTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	DetailLabel = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Width(15)

	DetailValue = lipgloss.NewStyle().
			Foreground(ColorText)

	// Status badges
	StatusActive = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorSuccess).
			Padding(0, 1).
			Bold(true)

	StatusInProgress = lipgloss.NewStyle().
				Foreground(ColorText).
				Background(ColorWarning).
				Padding(0, 1).
				Bold(true)

	StatusResolved = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorMuted).
			Padding(0, 1).
			Bold(true)

	// Severity badges
	SeverityCritical = lipgloss.NewStyle().
				Foreground(ColorText).
				Background(ColorCritical).
				Padding(0, 1).
				Bold(true)

	SeverityHigh = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorHigh).
			Padding(0, 1).
			Bold(true)

	SeverityMedium = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(ColorMedium).
			Padding(0, 1).
			Bold(true)

	SeverityLow = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorLow).
			Padding(0, 1).
			Bold(true)

	// Input styles
	InputLabel = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true).
			MarginBottom(1)

	InputField = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1).
			Width(50)

	InputFieldFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1).
				Width(50)

	// Button styles
	Button = lipgloss.NewStyle().
		Foreground(ColorText).
		Background(ColorMuted).
		Padding(0, 2).
		MarginRight(1)

	ButtonFocused = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorPrimary).
			Padding(0, 2).
			MarginRight(1)

	// Dialog styles
	Dialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(2, 4).
		Width(60)

	DialogTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	// Help bar
	HelpBar = lipgloss.NewStyle().
		Foreground(ColorTextDim).
		MarginTop(1)

	HelpKey = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	HelpDesc = lipgloss.NewStyle().
			Foreground(ColorTextDim)

	// Status bar
	StatusBar = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			MarginTop(1)

	// Error/Success messages
	Error = lipgloss.NewStyle().
		Foreground(ColorDanger).
		Bold(true)

	SuccessMsg = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// Spinner
	Spinner = lipgloss.NewStyle().
		Foreground(ColorPrimary)

	// Status indicators
	DotActive = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			SetString("●")

	DotWarning = lipgloss.NewStyle().
			Foreground(ColorWarning).
			SetString("●")

	DotDanger = lipgloss.NewStyle().
			Foreground(ColorDanger).
			SetString("●")

	DotMuted = lipgloss.NewStyle().
			Foreground(ColorMuted).
			SetString("○")
)

// Helper functions

// Signal bar severity indicators
var (
	SignalCritical = lipgloss.NewStyle().Foreground(ColorCritical).Bold(true)
	SignalHigh     = lipgloss.NewStyle().Foreground(ColorHigh).Bold(true)
	SignalMedium   = lipgloss.NewStyle().Foreground(ColorMedium).Bold(true)
	SignalLow      = lipgloss.NewStyle().Foreground(ColorLow).Bold(true)
)

func RenderSeverity(severity string) string {
	switch severity {
	case "critical", "Critical", "CRITICAL", "sev0", "SEV0":
		return SeverityCritical.Render("CRIT")
	case "high", "High", "HIGH", "sev1", "SEV1":
		return SeverityHigh.Render("HIGH")
	case "medium", "Medium", "MEDIUM", "sev2", "SEV2":
		return SeverityMedium.Render("MED")
	case "low", "Low", "LOW", "sev3", "SEV3":
		return SeverityLow.Render("LOW")
	default:
		return Muted.Render(severity)
	}
}

// RenderSeveritySignal renders severity as signal bars (▁▃▅▇)
func RenderSeveritySignal(severity string) string {
	switch severity {
	case "critical", "Critical", "CRITICAL", "sev0", "SEV0":
		return SignalCritical.Render("▁▃▅▇")
	case "high", "High", "HIGH", "sev1", "SEV1":
		return SignalHigh.Render("▁▃▅░")
	case "medium", "Medium", "MEDIUM", "sev2", "SEV2":
		return SignalMedium.Render("▁▃░░")
	case "low", "Low", "LOW", "sev3", "SEV3":
		return SignalLow.Render("▁░░░")
	default:
		return Muted.Render("░░░░")
	}
}

func RenderStatus(status string) string {
	switch status {
	case "started", "in_progress", "acknowledged":
		return StatusInProgress.Render(status)
	case "resolved", "closed", "mitigated":
		return StatusResolved.Render(status)
	case "open", "triggered":
		return StatusActive.Render(status)
	default:
		return Muted.Render(status)
	}
}

func RenderStatusDot(status string) string {
	switch status {
	case "resolved", "closed", "mitigated":
		return DotMuted.String()
	case "started", "in_progress", "acknowledged":
		return DotWarning.String()
	case "open", "triggered", "critical":
		return DotDanger.String()
	default:
		return DotActive.String()
	}
}

func RenderAlertSource(source string) string {
	switch source {
	case "datadog":
		return Info.Render("[DD]")
	case "pagerduty":
		return Success.Render("[PD]")
	case "grafana":
		return Warning.Render("[GF]")
	case "slack":
		return Primary.Render("[SL]")
	case "manual":
		return Muted.Render("[MN]")
	default:
		return Muted.Render("[" + source[:2] + "]")
	}
}

func RenderHelpItem(key, desc string) string {
	return HelpKey.Render(key) + " " + HelpDesc.Render(desc)
}
