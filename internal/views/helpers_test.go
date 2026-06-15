package views

import "regexp"

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\][^\x1b]*\x1b\\`)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}
