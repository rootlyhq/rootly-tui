package app

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rootlyhq/rootly-tui/internal/api"
	"github.com/rootlyhq/rootly-tui/internal/config"
	"github.com/rootlyhq/rootly-tui/internal/styles"
	"github.com/rootlyhq/rootly-tui/internal/views"
)

type Screen int

const (
	ScreenSetup Screen = iota
	ScreenMain
)

type Tab int

const (
	TabIncidents Tab = iota
	TabAlerts
)

type Model struct {
	// Core state
	version   string
	screen    Screen
	activeTab Tab
	keys      KeyMap
	width     int
	height    int

	// Config and API
	cfg       *config.Config
	apiClient *api.Client

	// Views
	setup     views.SetupModel
	incidents views.IncidentsModel
	alerts    views.AlertsModel
	help      views.HelpModel
	logs      views.LogsModel
	spinner   spinner.Model

	// Loading state
	loading        bool
	initialLoading bool
	statusMsg      string
	errorMsg       string
}

func New(version string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Spinner

	m := Model{
		version:   version,
		screen:    ScreenSetup,
		activeTab: TabIncidents,
		keys:      DefaultKeyMap(),
		setup:     views.NewSetupModel(),
		incidents: views.NewIncidentsModel(),
		alerts:    views.NewAlertsModel(),
		help:      views.NewHelpModel(),
		logs:      views.NewLogsModel(),
		spinner:   s,
	}

	// Check if config exists
	if config.Exists() {
		cfg, err := config.Load()
		if err == nil && cfg.IsValid() {
			m.cfg = cfg
			m.screen = ScreenMain
			m.initialLoading = true
		}
	}

	return m
}

func (m Model) Init() tea.Cmd {
	if m.screen == ScreenMain {
		return tea.Batch(
			m.spinner.Tick,
			m.loadData(),
		)
	}
	return m.setup.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Always allow quit
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

		// Handle logs overlay first
		if m.logs.Visible {
			if key.Matches(msg, m.keys.Logs) || msg.String() == "esc" {
				m.logs.Toggle()
				return m, nil
			}
			var cmd tea.Cmd
			m.logs, cmd = m.logs.Update(msg)
			return m, cmd
		}

		// Handle help overlay
		if m.help.Visible {
			if key.Matches(msg, m.keys.Help) || msg.String() == "esc" {
				m.help.Toggle()
				return m, nil
			}
			return m, nil
		}

		// Handle setup screen
		if m.screen == ScreenSetup {
			var cmd tea.Cmd
			m.setup, cmd = m.setup.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		// Handle main screen navigation
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.Toggle()
			return m, nil

		case key.Matches(msg, m.keys.Logs):
			m.logs.Toggle()
			return m, nil

		case key.Matches(msg, m.keys.Tab):
			if m.activeTab == TabIncidents {
				m.activeTab = TabAlerts
			} else {
				m.activeTab = TabIncidents
			}
			return m, nil

		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			m.statusMsg = "Refreshing..."
			// Clear cache on manual refresh
			if m.apiClient != nil {
				m.apiClient.ClearCache()
			}
			return m, m.loadData()

		default:
			// Pass key events to active view
			if m.activeTab == TabIncidents {
				var cmd tea.Cmd
				m.incidents, cmd = m.incidents.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				var cmd tea.Cmd
				m.alerts, cmd = m.alerts.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.incidents.SetDimensions(msg.Width-4, msg.Height-10)
		m.alerts.SetDimensions(msg.Width-4, msg.Height-10)
		m.logs.SetDimensions(msg.Width, msg.Height)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	// Setup screen messages
	case views.APIKeyValidatedMsg:
		m.setup.HandleValidationResult(msg)
		m.setup.SetTesting(false)
		return m, nil

	case views.ConfigSavedMsg:
		if msg.Success {
			// Config saved, load it and switch to main screen
			cfg, err := config.Load()
			if err == nil && cfg.IsValid() {
				m.cfg = cfg
				client, err := api.NewClient(cfg)
				if err == nil {
					m.apiClient = client
					m.screen = ScreenMain
					m.initialLoading = true
					return m, tea.Batch(m.spinner.Tick, m.loadData())
				}
			}
		}
		return m, nil

	// Data loading messages
	case IncidentsLoadedMsg:
		m.loading = false
		m.initialLoading = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
			m.incidents.SetError(msg.Err.Error())
		} else {
			m.incidents.SetIncidents(msg.Incidents)
			m.errorMsg = ""
			m.statusMsg = ""
		}
		return m, nil

	case AlertsLoadedMsg:
		if msg.Err != nil {
			m.alerts.SetError(msg.Err.Error())
		} else {
			m.alerts.SetAlerts(msg.Alerts)
		}
		return m, nil

	case ErrorMsg:
		m.errorMsg = msg.Err.Error()
		m.loading = false
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Setup screen
	if m.screen == ScreenSetup {
		return m.setup.View()
	}

	// Main screen
	var b strings.Builder

	// Header with tabs
	header := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n\n")

	// Main content
	if m.initialLoading {
		b.WriteString(m.spinner.View() + " Loading...")
	} else {
		if m.activeTab == TabIncidents {
			b.WriteString(m.incidents.View())
		} else {
			b.WriteString(m.alerts.View())
		}
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	// Help bar
	b.WriteString("\n")
	b.WriteString(views.RenderHelpBar(m.width))

	// Wrap content
	content := styles.App.Render(b.String())

	// Help overlay
	if m.help.Visible {
		helpDialog := m.help.View()
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpDialog)
	}

	// Logs overlay
	if m.logs.Visible {
		logsDialog := m.logs.View()
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, logsDialog)
	}

	return content
}

func (m Model) renderHeader() string {
	title := styles.Title.Render("Rootly TUI")

	// Tab indicators
	var incidentsTab, alertsTab string
	if m.activeTab == TabIncidents {
		incidentsTab = styles.TabActive.Render("Incidents")
		alertsTab = styles.TabInactive.Render("Alerts")
	} else {
		incidentsTab = styles.TabInactive.Render("Incidents")
		alertsTab = styles.TabActive.Render("Alerts")
	}
	tabs := incidentsTab + " " + alertsTab

	// Version
	version := styles.TextDim.Render("v" + m.version)

	// Calculate spacing
	leftPart := title + "  "
	leftWidth := lipgloss.Width(leftPart)
	tabsWidth := lipgloss.Width(tabs)
	rightWidth := lipgloss.Width(version)
	spacing := m.width - leftWidth - tabsWidth - rightWidth - 10

	if spacing < 1 {
		spacing = 1
	}

	return styles.Header.Width(m.width).Render(
		leftPart + strings.Repeat(" ", spacing/2) + tabs + strings.Repeat(" ", spacing/2) + version,
	)
}

func (m Model) renderStatusBar() string {
	if m.errorMsg != "" {
		return styles.Error.Render("Error: " + m.errorMsg)
	}
	if m.statusMsg != "" {
		return styles.StatusBar.Render(m.statusMsg)
	}
	if m.loading {
		return m.spinner.View() + " " + styles.StatusBar.Render("Loading...")
	}
	return ""
}

func (m Model) loadData() tea.Cmd {
	return tea.Batch(
		m.loadIncidents(),
		m.loadAlerts(),
	)
}

func (m Model) loadIncidents() tea.Cmd {
	return func() tea.Msg {
		if m.apiClient == nil {
			// Try to create client from config
			cfg, err := config.Load()
			if err != nil {
				return IncidentsLoadedMsg{Err: err}
			}
			client, err := api.NewClient(cfg)
			if err != nil {
				return IncidentsLoadedMsg{Err: err}
			}
			m.apiClient = client
		}

		ctx := context.Background()
		incidents, err := m.apiClient.ListIncidents(ctx)
		if err != nil {
			return IncidentsLoadedMsg{Err: err}
		}

		return IncidentsLoadedMsg{Incidents: incidents}
	}
}

func (m Model) loadAlerts() tea.Cmd {
	return func() tea.Msg {
		if m.apiClient == nil {
			cfg, err := config.Load()
			if err != nil {
				return AlertsLoadedMsg{Err: err}
			}
			client, err := api.NewClient(cfg)
			if err != nil {
				return AlertsLoadedMsg{Err: err}
			}
			m.apiClient = client
		}

		ctx := context.Background()
		alerts, err := m.apiClient.ListAlerts(ctx)
		if err != nil {
			return AlertsLoadedMsg{Err: err}
		}

		return AlertsLoadedMsg{Alerts: alerts}
	}
}
