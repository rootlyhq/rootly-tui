package views

import (
	"fmt"
	"html"
	"strings"
	"testing"
)

func TestOAuthCallbackHTMLEscaping(t *testing.T) {
	tests := []struct {
		name           string
		errParam       string
		errDescription string
	}{
		{
			name:           "script tag in error param",
			errParam:       `<script>alert('xss')</script>`,
			errDescription: "some description",
		},
		{
			name:           "img onerror in error description",
			errParam:       "invalid_grant",
			errDescription: `<img onerror="fetch('https://evil.com')" src=x>`,
		},
		{
			name:           "svg onload in both params",
			errParam:       `"><svg onload=alert(1)>`,
			errDescription: `'><iframe src="javascript:alert(1)">`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			escaped := fmt.Sprintf(
				`<p>%s: %s</p>`,
				html.EscapeString(tt.errParam),
				html.EscapeString(tt.errDescription),
			)

			if strings.Contains(escaped, tt.errParam) && strings.ContainsAny(tt.errParam, "<>\"'&") {
				t.Errorf("error param not escaped: found raw %q in output", tt.errParam)
			}
			if strings.Contains(escaped, tt.errDescription) && strings.ContainsAny(tt.errDescription, "<>\"'&") {
				t.Errorf("error description not escaped: found raw %q in output", tt.errDescription)
			}

			if strings.Contains(tt.errParam, "<") && !strings.Contains(escaped, "&lt;") {
				t.Error("expected < in errParam to be escaped to &lt;")
			}
			if strings.Contains(tt.errDescription, "<") && !strings.Contains(escaped, "&lt;") {
				t.Error("expected < in errDescription to be escaped to &lt;")
			}
		})
	}
}
