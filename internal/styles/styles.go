package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

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

// Pastel colors for status text
var (
	ColorPastelRed    = lipgloss.Color("#F87171") // Soft red
	ColorPastelYellow = lipgloss.Color("#FBBF24") // Soft amber
	ColorPastelGreen  = lipgloss.Color("#34D399") // Soft green
	ColorPastelGray   = lipgloss.Color("#9CA3AF") // Soft gray
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
		Padding(0, 2).
		MarginBottom(1)

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText)

	// Tab styles
	TabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Underline(true).
			Padding(0, 2)

	TabInactive = lipgloss.NewStyle().
			Foreground(ColorTextDim).
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
				Foreground(ColorHighlight).
				Bold(true).
				Padding(0, 1)

	// Detail pane styles
	DetailContainer = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	DetailContainerFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
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

	// Status styles (pastel text colors, no background)
	StatusActive = lipgloss.NewStyle().
			Foreground(ColorPastelRed)

	StatusInProgress = lipgloss.NewStyle().
				Foreground(ColorPastelYellow)

	StatusResolved = lipgloss.NewStyle().
			Foreground(ColorPastelGreen)

	StatusMuted = lipgloss.NewStyle().
			Foreground(ColorPastelGray)

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
	// Normalize status for comparison
	s := strings.ToLower(strings.TrimSpace(status))
	switch s {
	// Active/urgent - needs attention (red)
	case "open", "triggered", "firing", "critical":
		return StatusActive.Render(status)
	// In progress/mitigated - being worked on (yellow)
	case "started", "in_progress", "acknowledged", "investigating", "identified", "monitoring", "mitigated":
		return StatusInProgress.Render(status)
	// Resolved - completed successfully (green)
	case "resolved", "fixed":
		return StatusResolved.Render(status)
	// Closed/cancelled - done but neutral (gray)
	case "closed", "cancelled", "canceled", "suppressed":
		return StatusMuted.Render(status)
	default:
		return StatusMuted.Render(status)
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

// AlertSourceIcon returns just the emoji icon for an alert source
func AlertSourceIcon(source string) string {
	switch source {
	case "datadog":
		return "🐶"
	case "pagerduty":
		return "📟"
	case "grafana":
		return "📊"
	case "new_relic":
		return "🔮"
	case "prometheus", "alertmanager":
		return "🔥"
	case "opsgenie":
		return "🔔"
	case "sentry":
		return "🐛"
	case "splunk":
		return "📈"
	case "honeycomb":
		return "🍯"
	case "chronosphere":
		return "⏱️"
	case "cloud_watch", "cloudwatch":
		return "☁️"
	case "azure":
		return "☁️"
	case "google_cloud":
		return "☁️"
	case "slack":
		return "💬"
	case "email":
		return "📧"
	case "generic_webhook":
		return "🔗"
	case "api":
		return "🔌"
	case "manual":
		return "✋"
	case "jira":
		return "📋"
	case "zendesk":
		return "🎫"
	case "rollbar":
		return "🪵"
	case "bugsnag", "bug_snag":
		return "🐞"
	default:
		return "📡"
	}
}

func RenderAlertSource(source string) string {
	icon := AlertSourceIcon(source)
	switch source {
	// Major monitoring platforms
	case "datadog":
		return Info.Render(icon + "DD")
	case "pagerduty":
		return Success.Render(icon + "PD")
	case "grafana":
		return Warning.Render(icon + "GF")
	case "new_relic":
		return Info.Render(icon + "NR")
	case "prometheus", "alertmanager":
		return Danger.Render(icon + "PM")
	case "opsgenie":
		return Info.Render(icon + "OG")
	case "sentry":
		return Danger.Render(icon + "SE")
	case "splunk":
		return Success.Render(icon + "SP")
	case "honeycomb":
		return Warning.Render(icon + "HC")
	case "chronosphere":
		return Info.Render(icon + "CS")

	// Cloud providers
	case "cloud_watch", "cloudwatch":
		return Warning.Render(icon + "CW")
	case "azure":
		return Info.Render(icon + "AZ")
	case "google_cloud":
		return Info.Render(icon + "GC")

	// Communication
	case "slack":
		return Primary.Render(icon + "SL")
	case "email":
		return Muted.Render(icon + "EM")

	// Other
	case "generic_webhook":
		return Muted.Render(icon + "GW")
	case "api":
		return Muted.Render(icon + "AP")
	case "manual":
		return Muted.Render(icon + "MN")
	case "jira":
		return Info.Render(icon + "JI")
	case "zendesk":
		return Success.Render(icon + "ZD")
	case "rollbar":
		return Danger.Render(icon + "RB")
	case "bugsnag", "bug_snag":
		return Danger.Render(icon + "BS")

	default:
		// Fallback: first 2 chars uppercase
		if len(source) >= 2 {
			return Muted.Render(icon + strings.ToUpper(source[:2]))
		}
		return Muted.Render(icon + "??")
	}
}

// AlertSourceName returns the human-readable name for an alert source
func AlertSourceName(source string) string {
	switch source {
	case "datadog":
		return "Datadog"
	case "pagerduty":
		return "PagerDuty"
	case "grafana":
		return "Grafana"
	case "new_relic":
		return "New Relic"
	case "prometheus":
		return "Prometheus"
	case "alertmanager":
		return "Alertmanager"
	case "opsgenie":
		return "OpsGenie"
	case "sentry":
		return "Sentry"
	case "splunk":
		return "Splunk"
	case "honeycomb":
		return "Honeycomb"
	case "chronosphere":
		return "Chronosphere"
	case "cloud_watch", "cloudwatch":
		return "CloudWatch"
	case "azure":
		return "Azure"
	case "google_cloud":
		return "Google Cloud"
	case "slack":
		return "Slack"
	case "email":
		return "Email"
	case "generic_webhook":
		return "Generic Webhook"
	case "api":
		return "API"
	case "manual":
		return "Manual"
	case "jira":
		return "Jira"
	case "zendesk":
		return "Zendesk"
	case "rollbar":
		return "Rollbar"
	case "bugsnag", "bug_snag":
		return "BugSnag"
	default:
		return source
	}
}

func RenderHelpItem(key, desc string) string {
	return HelpKey.Render(key) + " " + HelpDesc.Render(desc)
}

// RenderLink renders a clickable hyperlink using OSC 8 escape sequences
// Most modern terminals support this (iTerm2, Kitty, Windows Terminal, etc.)
func RenderLink(url, text string) string {
	if text == "" {
		text = url
	}
	// OSC 8 hyperlink format: \x1b]8;;URL\x1b\\TEXT\x1b]8;;\x1b\\
	return "\x1b]8;;" + url + "\x1b\\" + Info.Underline(true).Render(text) + "\x1b]8;;\x1b\\"
}

// RenderURL renders a URL as a clickable link (URL is both the link and display text)
func RenderURL(url string) string {
	return RenderLink(url, url)
}

// RenderEmail renders an email as a clickable mailto link
func RenderEmail(email string) string {
	return RenderLink("mailto:"+email, email)
}

// RenderNameWithEmail renders a name with email in format "Name [email]" where email is clickable
func RenderNameWithEmail(name, email string) string {
	if email == "" {
		return name
	}
	return name + " [" + RenderEmail(email) + "]"
}

// markdownRenderer is a cached glamour renderer for dark terminals
var markdownRenderer *glamour.TermRenderer

// getMarkdownRenderer returns a cached glamour renderer
func getMarkdownRenderer(width int) *glamour.TermRenderer {
	if markdownRenderer == nil {
		// Build style JSON using ColorInfo constant for consistent link styling
		styleJSON := fmt.Sprintf(`{
			"document": {"margin": 0},
			"paragraph": {"margin": 0},
			"link": {"color": "%s", "underline": true},
			"link_text": {"color": "%s", "underline": true}
		}`, ColorInfo, ColorInfo)

		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width),
			glamour.WithStylesFromJSONBytes([]byte(styleJSON)),
		)
		if err != nil {
			return nil
		}
		markdownRenderer = r
	}
	return markdownRenderer
}

// RenderMarkdown renders markdown text for terminal display using glamour
// Falls back to plain text if rendering fails
func RenderMarkdown(text string, width int) string {
	if text == "" {
		return ""
	}

	// Use a reasonable default width
	if width <= 0 {
		width = 80
	}

	r := getMarkdownRenderer(width)
	if r == nil {
		return text
	}

	rendered, err := r.Render(text)
	if err != nil {
		return text
	}

	// Trim extra whitespace that glamour adds
	return strings.TrimSpace(rendered)
}
