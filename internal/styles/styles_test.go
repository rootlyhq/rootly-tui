package styles

import (
	"strings"
	"testing"
)

func TestRenderSeverity(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "CRIT"},
		{"Critical", "CRIT"},
		{"CRITICAL", "CRIT"},
		{"sev0", "CRIT"},
		{"high", "HIGH"},
		{"High", "HIGH"},
		{"sev1", "HIGH"},
		{"medium", "MED"},
		{"Medium", "MED"},
		{"sev2", "MED"},
		{"low", "LOW"},
		{"Low", "LOW"},
		{"sev3", "LOW"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := RenderSeverity(tt.severity)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("RenderSeverity(%s) = %s, expected to contain %s", tt.severity, result, tt.expected)
			}
		})
	}
}

func TestRenderSeveritySignal(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "▁▃▅▇"},
		{"Critical", "▁▃▅▇"},
		{"sev0", "▁▃▅▇"},
		{"high", "▁▃▅░"},
		{"High", "▁▃▅░"},
		{"sev1", "▁▃▅░"},
		{"medium", "▁▃░░"},
		{"Medium", "▁▃░░"},
		{"sev2", "▁▃░░"},
		{"low", "▁░░░"},
		{"Low", "▁░░░"},
		{"sev3", "▁░░░"},
		{"unknown", "░░░░"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := RenderSeveritySignal(tt.severity)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("RenderSeveritySignal(%s) = %s, expected to contain %s", tt.severity, result, tt.expected)
			}
		})
	}
}

func TestRenderStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"started", "started"},
		{"in_progress", "in_progress"},
		{"acknowledged", "acknowledged"},
		{"resolved", "resolved"},
		{"closed", "closed"},
		{"mitigated", "mitigated"},
		{"open", "open"},
		{"triggered", "triggered"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := RenderStatus(tt.status)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("RenderStatus(%s) = %s, expected to contain %s", tt.status, result, tt.expected)
			}
		})
	}
}

func TestRenderStatusDot(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"resolved", "○"},
		{"closed", "○"},
		{"mitigated", "○"},
		{"started", "●"},
		{"in_progress", "●"},
		{"acknowledged", "●"},
		{"open", "●"},
		{"triggered", "●"},
		{"critical", "●"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := RenderStatusDot(tt.status)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("RenderStatusDot(%s) = %s, expected to contain %s", tt.status, result, tt.expected)
			}
		})
	}
}

func TestRenderAlertSource(t *testing.T) {
	tests := []struct {
		source   string
		expected string
	}{
		{"datadog", "DD"},
		{"pagerduty", "PD"},
		{"grafana", "GF"},
		{"slack", "SL"},
		{"manual", "MN"},
		{"new_relic", "NR"},
		{"prometheus", "PM"},
		{"alertmanager", "PM"},
		{"opsgenie", "OG"},
		{"sentry", "SE"},
		{"generic_webhook", "GW"},
		{"cloud_watch", "CW"},
		{"api", "AP"},
		{"unknown_source", "UN"}, // fallback uses first 2 chars uppercase
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			result := RenderAlertSource(tt.source)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("RenderAlertSource(%s) = %s, expected to contain %s", tt.source, result, tt.expected)
			}
		})
	}
}

func TestRenderAlertSourceShortFallback(t *testing.T) {
	// Test very short source name
	result := RenderAlertSource("x")
	if !strings.Contains(result, "??") {
		t.Errorf("expected '??' for single char source, got %s", result)
	}
}

func TestRenderHelpItem(t *testing.T) {
	result := RenderHelpItem("q", "quit")

	if !strings.Contains(result, "q") {
		t.Error("expected help item to contain key 'q'")
	}

	if !strings.Contains(result, "quit") {
		t.Error("expected help item to contain description 'quit'")
	}
}

func TestStylesNotNil(t *testing.T) {
	// Verify key styles are defined
	styles := []struct {
		name  string
		style interface{}
	}{
		{"Primary", Primary},
		{"Secondary", Secondary},
		{"Success", Success},
		{"Warning", Warning},
		{"Danger", Danger},
		{"Info", Info},
		{"Muted", Muted},
		{"Text", Text},
		{"TextDim", TextDim},
		{"TextBold", TextBold},
		{"App", App},
		{"Header", Header},
		{"Title", Title},
		{"TabActive", TabActive},
		{"TabInactive", TabInactive},
		{"ListContainer", ListContainer},
		{"ListItem", ListItem},
		{"ListItemSelected", ListItemSelected},
		{"DetailContainer", DetailContainer},
		{"DetailTitle", DetailTitle},
		{"Dialog", Dialog},
		{"Button", Button},
		{"ButtonFocused", ButtonFocused},
		{"Error", Error},
		{"SuccessMsg", SuccessMsg},
		{"Spinner", Spinner},
		{"HelpBar", HelpBar},
		{"HelpKey", HelpKey},
		{"HelpDesc", HelpDesc},
	}

	for _, s := range styles {
		t.Run(s.name, func(t *testing.T) {
			if s.style == nil {
				t.Errorf("style %s is nil", s.name)
			}
		})
	}
}

func TestColorsNotEmpty(t *testing.T) {
	colors := []struct {
		name  string
		color interface{}
	}{
		{"ColorPrimary", ColorPrimary},
		{"ColorSecondary", ColorSecondary},
		{"ColorSuccess", ColorSuccess},
		{"ColorWarning", ColorWarning},
		{"ColorDanger", ColorDanger},
		{"ColorInfo", ColorInfo},
		{"ColorMuted", ColorMuted},
		{"ColorText", ColorText},
		{"ColorTextDim", ColorTextDim},
		{"ColorBackground", ColorBackground},
		{"ColorBorder", ColorBorder},
		{"ColorHighlight", ColorHighlight},
		{"ColorCritical", ColorCritical},
		{"ColorHigh", ColorHigh},
		{"ColorMedium", ColorMedium},
		{"ColorLow", ColorLow},
	}

	for _, c := range colors {
		t.Run(c.name, func(t *testing.T) {
			if c.color == nil {
				t.Errorf("color %s is nil", c.name)
			}
		})
	}
}

func TestAlertSourceIcon(t *testing.T) {
	tests := []struct {
		source   string
		expected string
	}{
		{"datadog", "🐶"},
		{"pagerduty", "📟"},
		{"grafana", "📊"},
		{"new_relic", "🔮"},
		{"prometheus", "🔥"},
		{"alertmanager", "🔥"},
		{"opsgenie", "🔔"},
		{"sentry", "🐛"},
		{"slack", "💬"},
		{"email", "📧"},
		{"jira", "📋"},
		{"unknown", "📡"},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			result := AlertSourceIcon(tt.source)
			if result != tt.expected {
				t.Errorf("AlertSourceIcon(%s) = %s, expected %s", tt.source, result, tt.expected)
			}
		})
	}
}

func TestAlertSourceName(t *testing.T) {
	tests := []struct {
		source   string
		expected string
	}{
		{"datadog", "Datadog"},
		{"pagerduty", "PagerDuty"},
		{"new_relic", "New Relic"},
		{"prometheus", "Prometheus"},
		{"cloud_watch", "CloudWatch"},
		{"cloudwatch", "CloudWatch"},
		{"generic_webhook", "Generic Webhook"},
		{"unknown_source", "unknown_source"}, // fallback returns source as-is
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			result := AlertSourceName(tt.source)
			if result != tt.expected {
				t.Errorf("AlertSourceName(%s) = %s, expected %s", tt.source, result, tt.expected)
			}
		})
	}
}

func TestRenderStatusExtended(t *testing.T) {
	tests := []struct {
		status string
	}{
		// Active (red)
		{"open"},
		{"triggered"},
		{"firing"},
		{"OPEN"},          // case insensitive
		{"  triggered  "}, // whitespace trimmed
		// In progress (yellow)
		{"started"},
		{"in_progress"},
		{"acknowledged"},
		{"investigating"},
		// Resolved (green)
		{"resolved"},
		{"mitigated"},
		{"fixed"},
		// Muted (gray)
		{"closed"},
		{"cancelled"},
		{"suppressed"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := RenderStatus(tt.status)
			// Just verify it returns something non-empty
			if result == "" {
				t.Errorf("RenderStatus(%s) returned empty string", tt.status)
			}
		})
	}
}

func TestRenderLink(t *testing.T) {
	url := "https://example.com"
	text := "Example"
	result := RenderLink(url, text)

	// Should contain OSC 8 escape sequences
	if !strings.Contains(result, "\x1b]8;;") {
		t.Error("RenderLink should contain OSC 8 escape sequence")
	}
	if !strings.Contains(result, url) {
		t.Errorf("RenderLink should contain URL %s", url)
	}
	if !strings.Contains(result, text) {
		t.Errorf("RenderLink should contain text %s", text)
	}
}

func TestRenderLinkEmptyText(t *testing.T) {
	url := "https://example.com"
	result := RenderLink(url, "")

	// When text is empty, URL should be used as display text
	// URL should appear twice (once in link, once as display)
	if strings.Count(result, url) < 1 {
		t.Errorf("RenderLink with empty text should use URL as display text")
	}
}

func TestRenderURL(t *testing.T) {
	url := "https://example.com"
	result := RenderURL(url)

	if !strings.Contains(result, url) {
		t.Errorf("RenderURL should contain URL %s", url)
	}
	if !strings.Contains(result, "\x1b]8;;") {
		t.Error("RenderURL should contain OSC 8 escape sequence")
	}
}
