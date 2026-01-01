package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rootlyhq/rootly-tui/internal/api"
	"github.com/rootlyhq/rootly-tui/internal/config"
	"github.com/rootlyhq/rootly-tui/internal/i18n"
	"github.com/rootlyhq/rootly-tui/internal/styles"
)

type SetupField int

const (
	FieldEndpoint SetupField = iota
	FieldAPIKey
	FieldTimezone
	FieldLanguage
	FieldButtons
)

const (
	testResultSuccess = "success"
	testResultError   = "error"
)

type SetupModel struct {
	endpoint      textinput.Model
	apiKey        textinput.Model
	timezones     []string // Available timezones from system
	timezoneIndex int      // Index into timezones
	languages     []string // Available languages
	languageIndex int      // Index into languages
	spinner       spinner.Model
	focusIndex    SetupField
	buttonIndex   int // 0 = Test, 1 = Save
	testing       bool
	saving        bool
	testResult    string
	testError     string
	width         int
	height        int
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
	return NewSetupModelWithConfig(nil)
}

// NewSetupModelWithConfig creates a setup model pre-populated with existing config values
func NewSetupModelWithConfig(cfg *config.Config) SetupModel {
	endpointInput := textinput.New()
	endpointInput.Placeholder = "api.rootly.com"
	endpointInput.Focus()
	endpointInput.Width = 45

	apiKeyInput := textinput.New()
	apiKeyInput.Placeholder = "Enter your API key"
	apiKeyInput.EchoMode = textinput.EchoPassword
	apiKeyInput.EchoCharacter = '*'
	apiKeyInput.Width = 45

	// Load available timezones from system
	timezones := config.ListTimezones()

	// Load available languages
	languages := i18n.ListLanguages()

	// Default values
	tzIndex := 0
	langIndex := 0

	if cfg != nil && cfg.IsValid() {
		// Use existing config values
		endpointInput.SetValue(cfg.Endpoint)
		apiKeyInput.SetValue(cfg.APIKey)

		// Find timezone index
		for i, tz := range timezones {
			if tz == cfg.Timezone {
				tzIndex = i
				break
			}
		}

		// Find language index
		langIndex = i18n.LanguageIndex(cfg.Language)
	} else {
		// Use defaults for new setup
		endpointInput.SetValue(config.DefaultEndpoint)

		// Detect timezone and find index in list
		detectedTZ := config.DetectTimezone()
		for i, tz := range timezones {
			if tz == detectedTZ {
				tzIndex = i
				break
			}
		}

		// Detect language and find index in list
		detectedLang := i18n.DetectLanguage()
		langIndex = i18n.LanguageIndex(string(detectedLang))
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Spinner

	return SetupModel{
		endpoint:      endpointInput,
		apiKey:        apiKeyInput,
		timezones:     timezones,
		timezoneIndex: tzIndex,
		languages:     languages,
		languageIndex: langIndex,
		spinner:       s,
		focusIndex:    FieldEndpoint,
		buttonIndex:   0,
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

		case "left", "h":
			if m.focusIndex == FieldButtons && m.buttonIndex > 0 {
				m.buttonIndex--
			} else if m.focusIndex == FieldTimezone && m.timezoneIndex > 0 {
				m.timezoneIndex--
			} else if m.focusIndex == FieldLanguage && m.languageIndex > 0 {
				m.languageIndex--
				// Update active language immediately for UI preview
				i18n.SetLanguage(i18n.Language(m.languages[m.languageIndex]))
			}
			return m, nil

		case "right", "l":
			if m.focusIndex == FieldButtons && m.buttonIndex < 1 {
				m.buttonIndex++
			} else if m.focusIndex == FieldTimezone && m.timezoneIndex < len(m.timezones)-1 {
				m.timezoneIndex++
			} else if m.focusIndex == FieldLanguage && m.languageIndex < len(m.languages)-1 {
				m.languageIndex++
				// Update active language immediately for UI preview
				i18n.SetLanguage(i18n.Language(m.languages[m.languageIndex]))
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
				}
				// Save and continue
				if m.testResult == testResultSuccess {
					m.saving = true
					return m, m.doSaveConfig()
				}
				return m, nil
			}
			// Move to next field on enter
			m.focusIndex++
			if m.focusIndex > FieldButtons {
				m.focusIndex = FieldButtons
			}
			m.updateFocus()
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

	// Update text inputs (timezone is a selector, not text input)
	var cmd tea.Cmd
	switch m.focusIndex {
	case FieldEndpoint:
		m.endpoint, cmd = m.endpoint.Update(msg)
		cmds = append(cmds, cmd)
	case FieldAPIKey:
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
	// Capture timezone value
	timezone := ""
	if m.timezoneIndex >= 0 && m.timezoneIndex < len(m.timezones) {
		timezone = m.timezones[m.timezoneIndex]
	}

	// Capture language value
	language := string(i18n.DefaultLanguage)
	if m.languageIndex >= 0 && m.languageIndex < len(m.languages) {
		language = m.languages[m.languageIndex]
	}

	return func() tea.Msg {
		cfg := &config.Config{
			Endpoint: m.endpoint.Value(),
			APIKey:   m.apiKey.Value(),
			Timezone: timezone,
			Language: language,
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
		m.testResult = testResultSuccess
		m.testError = ""
	} else {
		m.testResult = testResultError
		m.testError = msg.Error
	}
}

func (m SetupModel) IsTesting() bool {
	return m.testing
}

func (m *SetupModel) SetTesting(testing bool) {
	m.testing = testing
}

func (m *SetupModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

func (m SetupModel) View() string {
	var b strings.Builder

	// Title
	title := styles.DialogTitle.Render(i18n.T("setup.welcome"))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Description
	desc := styles.TextDim.Render(i18n.T("setup.enter_credentials"))
	b.WriteString(desc)
	b.WriteString("\n\n")

	// Endpoint field
	endpointLabel := styles.InputLabel.Render(i18n.T("setup.api_endpoint"))
	b.WriteString(endpointLabel)
	b.WriteString("\n")
	if m.focusIndex == FieldEndpoint {
		b.WriteString(styles.InputFieldFocused.Render(m.endpoint.View()))
	} else {
		b.WriteString(styles.InputField.Render(m.endpoint.View()))
	}
	b.WriteString("\n\n")

	// API Key field
	apiKeyLabel := styles.InputLabel.Render(i18n.T("setup.api_key"))
	b.WriteString(apiKeyLabel)
	b.WriteString("\n")
	if m.focusIndex == FieldAPIKey {
		b.WriteString(styles.InputFieldFocused.Render(m.apiKey.View()))
	} else {
		b.WriteString(styles.InputField.Render(m.apiKey.View()))
	}
	b.WriteString("\n\n")

	// Timezone selector
	timezoneLabel := styles.InputLabel.Render(i18n.T("setup.timezone"))
	b.WriteString(timezoneLabel)
	b.WriteString(" ")
	b.WriteString(styles.TextDim.Render(i18n.T("setup.use_arrows")))
	b.WriteString("\n")

	// Show the selected timezone with navigation hints
	selectedTZ := "UTC"
	if m.timezoneIndex >= 0 && m.timezoneIndex < len(m.timezones) {
		selectedTZ = m.timezones[m.timezoneIndex]
	}
	tzDisplay := fmt.Sprintf("◀ %s ▶", selectedTZ)
	if m.focusIndex == FieldTimezone {
		b.WriteString(styles.InputFieldFocused.Render(tzDisplay))
	} else {
		b.WriteString(styles.InputField.Render(tzDisplay))
	}
	b.WriteString("\n\n")

	// Language selector
	languageLabel := styles.InputLabel.Render(i18n.T("setup.language"))
	b.WriteString(languageLabel)
	b.WriteString(" ")
	b.WriteString(styles.TextDim.Render(i18n.T("setup.use_arrows")))
	b.WriteString("\n")

	// Show the selected language with navigation hints
	selectedLang := i18n.LanguageName(string(i18n.DefaultLanguage))
	if m.languageIndex >= 0 && m.languageIndex < len(m.languages) {
		selectedLang = i18n.LanguageName(m.languages[m.languageIndex])
	}
	langDisplay := fmt.Sprintf("◀ %s ▶", selectedLang)
	if m.focusIndex == FieldLanguage {
		b.WriteString(styles.InputFieldFocused.Render(langDisplay))
	} else {
		b.WriteString(styles.InputField.Render(langDisplay))
	}
	b.WriteString("\n\n")

	// Test result
	if m.testing {
		b.WriteString(m.spinner.View() + " " + i18n.T("setup.testing_connection"))
		b.WriteString("\n\n")
	} else if m.testResult == testResultSuccess {
		b.WriteString(styles.SuccessMsg.Render(i18n.T("setup.connection_success")))
		b.WriteString("\n\n")
	} else if m.testResult == testResultError {
		b.WriteString(styles.Error.Render(i18n.T("common.error") + ": " + m.testError))
		b.WriteString("\n\n")
	}

	// Buttons
	var testBtn, saveBtn string
	if m.focusIndex == FieldButtons && m.buttonIndex == 0 {
		testBtn = styles.ButtonFocused.Render(i18n.T("setup.test_connection"))
	} else {
		testBtn = styles.Button.Render(i18n.T("setup.test_connection"))
	}

	if m.focusIndex == FieldButtons && m.buttonIndex == 1 {
		if m.testResult == testResultSuccess {
			saveBtn = styles.ButtonFocused.Render(i18n.T("setup.save_and_continue"))
		} else {
			saveBtn = styles.Button.Render(i18n.T("setup.save_and_continue"))
		}
	} else {
		saveBtn = styles.Button.Render(i18n.T("setup.save_and_continue"))
	}

	b.WriteString(testBtn + "  " + saveBtn)
	b.WriteString("\n\n")

	// Help
	help := styles.HelpBar.Render(i18n.T("setup.help"))
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
