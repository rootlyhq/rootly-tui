package views

import (
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rootlyhq/rootly-tui/internal/styles"
)

const (
	welcomeTickRate   = 12 * time.Millisecond
	typeSpeed         = 8
	gradientHoldTicks = 10
	shimmerWidth      = 14
	shimmerSpeed      = 1.2
	subtitleDelay     = 5
	panelDelay        = 15
)

var asciiLogo = []string{
	`  ██████╗  ██████╗  ██████╗ ████████╗██╗  ██╗   ██╗`,
	`  ██╔══██╗██╔═══██╗██╔═══██╗╚══██╔══╝██║  ╚██╗ ██╔╝`,
	`  ██████╔╝██║   ██║██║   ██║   ██║   ██║   ╚████╔╝ `,
	`  ██╔══██╗██║   ██║██║   ██║   ██║   ██║    ╚██╔╝  `,
	`  ██║  ██║╚██████╔╝╚██████╔╝   ██║   ███████╗██║   `,
	`  ╚═╝  ╚═╝ ╚═════╝  ╚═════╝    ╚═╝   ╚══════╝╚═╝   `,
}

// Gradient palette from purple to indigo to cyan
var gradientColors = []lipgloss.Color{
	"#7C3AED", // purple
	"#7C3AED",
	"#6D5AED",
	"#6366F1", // indigo
	"#6366F1",
	"#5B7FF1",
	"#4D96FF", // blue
	"#4D96FF",
	"#3DBDF1",
	"#2DD4BF", // teal
}

// Shimmer colors (bright highlights)
var shimmerColors = []lipgloss.Color{
	"#C4B5FD", // light purple
	"#DDD6FE",
	"#EDE9FE",
	"#F5F3FF", // near white
	"#FFFFFF", // white
	"#F5F3FF",
	"#EDE9FE",
	"#DDD6FE",
	"#C4B5FD",
}

// Pre-computed styles to avoid per-character allocations on every render frame.
var (
	gradientStyles []lipgloss.Style
	shimmerStyles  []lipgloss.Style
	cursorStyle    = lipgloss.NewStyle().Foreground(styles.ColorPrimary).Bold(true)
)

func init() {
	gradientStyles = make([]lipgloss.Style, len(gradientColors))
	for i, c := range gradientColors {
		gradientStyles[i] = lipgloss.NewStyle().Foreground(c).Bold(true)
	}
	shimmerStyles = make([]lipgloss.Style, len(shimmerColors))
	for i, c := range shimmerColors {
		shimmerStyles[i] = lipgloss.NewStyle().Foreground(c).Bold(true)
	}
}

type welcomeTickMsg time.Time

type WelcomeModel struct {
	tick       int
	totalChars int
	logoWidth  int
}

func NewWelcomeModel() WelcomeModel {
	total := 0
	maxWidth := 0
	for _, line := range asciiLogo {
		runes := []rune(line)
		total += len(runes)
		if len(runes) > maxWidth {
			maxWidth = len(runes)
		}
	}
	return WelcomeModel{
		totalChars: total,
		logoWidth:  maxWidth,
	}
}

func (m WelcomeModel) Init() tea.Cmd {
	return m.tickCmd()
}

func (m WelcomeModel) tickCmd() tea.Cmd {
	return tea.Tick(welcomeTickRate, func(t time.Time) tea.Msg {
		return welcomeTickMsg(t)
	})
}

func (m WelcomeModel) Update(msg tea.Msg) (WelcomeModel, tea.Cmd) {
	if _, ok := msg.(welcomeTickMsg); ok {
		m.tick++
		// Stop ticking once the panel is shown and one shimmer cycle completes
		if m.ShowPanel() && m.ticksSinceTypingDone() > panelDelay+int(float64(m.logoWidth+shimmerWidth*2)/shimmerSpeed)+gradientHoldTicks {
			return m, nil
		}
		return m, m.tickCmd()
	}
	return m, nil
}

func (m WelcomeModel) charsRevealed() int {
	n := m.tick * typeSpeed
	if n > m.totalChars {
		return m.totalChars
	}
	return n
}

func (m WelcomeModel) typingDone() bool {
	return m.charsRevealed() >= m.totalChars
}

func (m WelcomeModel) ticksSinceTypingDone() int {
	typingTicks := m.totalChars / typeSpeed
	if m.tick <= typingTicks {
		return 0
	}
	return m.tick - typingTicks
}

// ShowSubtitle returns true when the subtitle should be visible.
func (m WelcomeModel) ShowSubtitle() bool {
	return m.ticksSinceTypingDone() >= subtitleDelay
}

// ShowPanel returns true when the auth panel should be visible.
func (m WelcomeModel) ShowPanel() bool {
	return m.ticksSinceTypingDone() >= panelDelay
}

// SubtitleOpacity returns 0.0-1.0 for subtitle fade-in.
func (m WelcomeModel) SubtitleOpacity() float64 {
	since := m.ticksSinceTypingDone() - subtitleDelay
	if since <= 0 {
		return 0
	}
	if since >= 20 {
		return 1
	}
	return float64(since) / 20.0
}

// PanelOpacity returns 0.0-1.0 for panel fade-in.
func (m WelcomeModel) PanelOpacity() float64 {
	since := m.ticksSinceTypingDone() - panelDelay
	if since <= 0 {
		return 0
	}
	if since >= 15 {
		return 1
	}
	return float64(since) / 15.0
}

func (m WelcomeModel) View() string {
	revealed := m.charsRevealed()
	var b strings.Builder

	charIdx := 0
	for lineNum, line := range asciiLogo {
		runes := []rune(line)
		for i, r := range runes {
			if charIdx >= revealed {
				// Cursor at the typing position
				if charIdx == revealed {
					b.WriteString(cursorStyle.Render("█"))
				}
				// Rest is empty
				break
			}

			if r == ' ' {
				b.WriteRune(r)
			} else {
				b.WriteString(m.styleForChar(i).Render(string(r)))
			}
			charIdx++
		}
		if lineNum < len(asciiLogo)-1 {
			b.WriteString("\n")
		}
		// If we ran out of revealed chars mid-line, skip remaining lines
		if charIdx >= revealed {
			// Add empty lines for remaining logo lines
			for j := lineNum + 1; j < len(asciiLogo); j++ {
				b.WriteString("\n")
			}
			break
		}
	}

	return b.String()
}

func (m WelcomeModel) styleForChar(colIdx int) lipgloss.Style {
	if !m.typingDone() {
		return columnGradientStyle(colIdx, m.logoWidth)
	}

	since := m.ticksSinceTypingDone()
	if since < gradientHoldTicks {
		return columnGradientStyle(colIdx, m.logoWidth)
	}

	// Shimmer phase: a bright band sweeps left to right repeatedly
	shimmerTick := float64(since-gradientHoldTicks) * shimmerSpeed
	shimmerPos := math.Mod(shimmerTick, float64(m.logoWidth+shimmerWidth*2)) - float64(shimmerWidth)

	dist := math.Abs(float64(colIdx) - shimmerPos)
	if dist < float64(shimmerWidth)/2 {
		idx := int((1.0 - dist/(float64(shimmerWidth)/2)) * float64(len(shimmerColors)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(shimmerColors) {
			idx = len(shimmerColors) - 1
		}
		return shimmerStyles[idx]
	}

	return columnGradientStyle(colIdx, m.logoWidth)
}

func columnGradientStyle(col, width int) lipgloss.Style {
	if width <= 0 {
		return gradientStyles[0]
	}
	idx := col * (len(gradientColors) - 1) / width
	if idx < 0 {
		idx = 0
	}
	if idx >= len(gradientColors) {
		idx = len(gradientColors) - 1
	}
	return gradientStyles[idx]
}

// RenderSubtitle renders the subtitle with fade-in effect.
func (m WelcomeModel) RenderSubtitle() string {
	opacity := m.SubtitleOpacity()
	if opacity <= 0 {
		return ""
	}

	text := "Terminal User Interface"

	// Fade in by revealing characters progressively
	runes := []rune(text)
	charsToShow := int(float64(len(runes)) * opacity)
	if charsToShow > len(runes) {
		charsToShow = len(runes)
	}

	if opacity >= 1.0 {
		return lipgloss.NewStyle().
			Foreground(styles.ColorTextDim).
			Render(text)
	}

	visible := string(runes[:charsToShow])
	return lipgloss.NewStyle().
		Foreground(styles.ColorTextDim).
		Render(visible)
}
