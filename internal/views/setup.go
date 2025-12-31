package views

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rootlyhq/rootly-tui/internal/api"
	"github.com/rootlyhq/rootly-tui/internal/config"
	"github.com/rootlyhq/rootly-tui/internal/styles"
)

type SetupField int

const (
	FieldEndpoint SetupField = iota
	FieldAPIKey
	FieldButtons
)

type SetupModel struct {
	endpoint    textinput.Model
	apiKey      textinput.Model
	spinner     spinner.Model
	focusIndex  SetupField
	buttonIndex int // 0 = Test, 1 = Save
	testing     bool
	saving      bool
	testResult  string
	testError   string
	width       int
	height      int
}

type APIKeyValidatedMsg struct {
	Valid bool
	Error string
}

type ConfigSavedMsg struct {
	Success bool
	Error   string
}

func NewSetupModel() SetupModel {
	endpointInput := textinput.New()
	endpointInput.Placeholder = "api.rootly.com"
	endpointInput.SetValue(config.DefaultEndpoint)
	endpointInput.Focus()
	endpointInput.Width = 45

	apiKeyInput := textinput.New()
	apiKeyInput.Placeholder = "Enter your API key"
	apiKeyInput.EchoMode = textinput.EchoPassword
	apiKeyInput.EchoCharacter = '*'
	apiKeyInput.Width = 45

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Spinner

	return SetupModel{
		endpoint:    endpointInput,
		apiKey:      apiKeyInput,
		spinner:     s,
		focusIndex:  FieldEndpoint,
		buttonIndex: 0,
	}
}

func (m SetupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SetupModel) Update(msg tea.Msg) (SetupModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.testing || m.saving {
			return m, nil
		}

		switch msg.String() {
		case "tab", "down":
			m.focusIndex++
			if m.focusIndex > FieldButtons {
				m.focusIndex = FieldEndpoint
			}
			m.updateFocus()
			return m, nil

		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < FieldEndpoint {
				m.focusIndex = FieldButtons
			}
			m.updateFocus()
			return m, nil

		case "left":
			if m.focusIndex == FieldButtons && m.buttonIndex > 0 {
				m.buttonIndex--
			}
			return m, nil

		case "right":
			if m.focusIndex == FieldButtons && m.buttonIndex < 1 {
				m.buttonIndex++
			}
			return m, nil

		case "enter":
			if m.focusIndex == FieldButtons {
				if m.buttonIndex == 0 {
					// Test connection
					m.testing = true
					m.testResult = ""
					m.testError = ""
					return m, tea.Batch(m.spinner.Tick, m.doTestConnection())
				} else {
					// Save and continue
					if m.testResult == "success" {
						m.saving = true
						return m, m.doSaveConfig()
					}
				}
			} else {
				// Move to next field on enter
				m.focusIndex++
				if m.focusIndex > FieldButtons {
					m.focusIndex = FieldButtons
				}
				m.updateFocus()
			}
			return m, nil
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update text inputs
	var cmd tea.Cmd
	if m.focusIndex == FieldEndpoint {
		m.endpoint, cmd = m.endpoint.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.focusIndex == FieldAPIKey {
		m.apiKey, cmd = m.apiKey.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *SetupModel) updateFocus() {
	m.endpoint.Blur()
	m.apiKey.Blur()

	switch m.focusIndex {
	case FieldEndpoint:
		m.endpoint.Focus()
	case FieldAPIKey:
		m.apiKey.Focus()
	}
}

func (m SetupModel) doTestConnection() tea.Cmd {
	return func() tea.Msg {
		cfg := &config.Config{
			Endpoint: m.endpoint.Value(),
			APIKey:   m.apiKey.Value(),
		}

		client, err := api.NewClient(cfg)
		if err != nil {
			return APIKeyValidatedMsg{Valid: false, Error: err.Error()}
		}

		ctx := context.Background()
		if err := client.ValidateAPIKey(ctx); err != nil {
			return APIKeyValidatedMsg{Valid: false, Error: err.Error()}
		}

		return APIKeyValidatedMsg{Valid: true}
	}
}

func (m SetupModel) doSaveConfig() tea.Cmd {
	return func() tea.Msg {
		cfg := &config.Config{
			Endpoint: m.endpoint.Value(),
			APIKey:   m.apiKey.Value(),
		}

		if err := config.Save(cfg); err != nil {
			return ConfigSavedMsg{Success: false, Error: err.Error()}
		}

		return ConfigSavedMsg{Success: true}
	}
}

func (m *SetupModel) HandleValidationResult(msg APIKeyValidatedMsg) {
	m.testing = false
	if msg.Valid {
		m.testResult = "success"
		m.testError = ""
	} else {
		m.testResult = "error"
		m.testError = msg.Error
	}
}

func (m SetupModel) IsTesting() bool {
	return m.testing
}

func (m *SetupModel) SetTesting(testing bool) {
	m.testing = testing
}

func (m SetupModel) View() string {
	var b strings.Builder

	// Title
	title := styles.DialogTitle.Render("Welcome to Rootly TUI")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Description
	desc := styles.TextDim.Render("Enter your Rootly API credentials to get started.")
	b.WriteString(desc)
	b.WriteString("\n\n")

	// Endpoint field
	endpointLabel := styles.InputLabel.Render("API Endpoint")
	b.WriteString(endpointLabel)
	b.WriteString("\n")
	if m.focusIndex == FieldEndpoint {
		b.WriteString(styles.InputFieldFocused.Render(m.endpoint.View()))
	} else {
		b.WriteString(styles.InputField.Render(m.endpoint.View()))
	}
	b.WriteString("\n\n")

	// API Key field
	apiKeyLabel := styles.InputLabel.Render("API Key")
	b.WriteString(apiKeyLabel)
	b.WriteString("\n")
	if m.focusIndex == FieldAPIKey {
		b.WriteString(styles.InputFieldFocused.Render(m.apiKey.View()))
	} else {
		b.WriteString(styles.InputField.Render(m.apiKey.View()))
	}
	b.WriteString("\n\n")

	// Test result
	if m.testing {
		b.WriteString(m.spinner.View() + " Testing connection...")
		b.WriteString("\n\n")
	} else if m.testResult == "success" {
		b.WriteString(styles.SuccessMsg.Render("Connection successful!"))
		b.WriteString("\n\n")
	} else if m.testResult == "error" {
		b.WriteString(styles.Error.Render("Error: " + m.testError))
		b.WriteString("\n\n")
	}

	// Buttons
	var testBtn, saveBtn string
	if m.focusIndex == FieldButtons && m.buttonIndex == 0 {
		testBtn = styles.ButtonFocused.Render("Test Connection")
	} else {
		testBtn = styles.Button.Render("Test Connection")
	}

	if m.focusIndex == FieldButtons && m.buttonIndex == 1 {
		if m.testResult == "success" {
			saveBtn = styles.ButtonFocused.Render("Save & Continue")
		} else {
			saveBtn = styles.Button.Render("Save & Continue")
		}
	} else {
		saveBtn = styles.Button.Render("Save & Continue")
	}

	b.WriteString(testBtn + "  " + saveBtn)
	b.WriteString("\n\n")

	// Help
	help := styles.HelpBar.Render("Tab/Arrow keys to navigate, Enter to select, q to quit")
	b.WriteString(help)

	// Wrap in dialog
	content := b.String()
	dialog := styles.Dialog.Width(60).Render(content)

	// Center on screen
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
	}

	return dialog
}
