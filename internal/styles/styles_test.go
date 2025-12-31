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
		{"datadog", "[DD]"},
		{"pagerduty", "[PD]"},
		{"grafana", "[GF]"},
		{"slack", "[SL]"},
		{"manual", "[MN]"},
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
