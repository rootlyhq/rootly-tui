package views

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/oauth2"

	"github.com/rootlyhq/rootly-tui/internal/api"
	"github.com/rootlyhq/rootly-tui/internal/config"
	"github.com/rootlyhq/rootly-tui/internal/debug"
	"github.com/rootlyhq/rootly-tui/internal/i18n"
	"github.com/rootlyhq/rootly-tui/internal/oauth"
	"github.com/rootlyhq/rootly-tui/internal/styles"
)

// Panel identifiers
type Panel int

const (
	PanelConnection Panel = iota
	PanelConfig
)

// Auth method
type AuthMethod int

const (
	AuthMethodOAuth AuthMethod = iota
	AuthMethodAPIKey
)

// Connection panel fields
type ConnectionField int

const (
	ConnFieldAuthMethod ConnectionField = iota
	ConnFieldEndpoint
	ConnFieldAPIKey
	ConnFieldButtons
)

// Config panel fields
type ConfigField int

const (
	ConfigFieldTimezone ConfigField = iota
	ConfigFieldLanguage
	ConfigFieldLayout
	ConfigFieldButton
)

const (
	testResultSuccess = "success"
	testResultError   = "error"
)

// SetupField is kept for backward compatibility with tests
type SetupField int

const (
	FieldEndpoint SetupField = iota
	FieldAPIKey
	FieldTimezone
	FieldLanguage
	FieldLayout
	FieldButtons
)

// OAuthLoginStartedMsg is sent when the OAuth login flow begins.
type OAuthLoginStartedMsg struct{}

// OAuthLoginResultMsg is sent when the OAuth login flow completes.
type OAuthLoginResultMsg struct {
	Success bool
	Error   string
}

// OAuthLogoutResultMsg is sent when the user logs out.
type OAuthLogoutResultMsg struct {
	Success bool
	Error   string
}

type SetupModel struct {
	// Connection panel
	authMethod AuthMethod
	endpoint   textinput.Model
	apiKey     textinput.Model
	connFocus  ConnectionField
	connButton int // 0 = Test/Login, 1 = Save
	testing    bool
	testResult string
	testError  string
	connSaved  bool
	connSaving bool

	// OAuth state
	oauthLoggingIn bool
	oauthLoggedIn  bool
	oauthError     string

	// Config panel
	timezones     []string
	timezoneIndex int
	languages     []string
	languageIndex int
	layouts       []string
	layoutIndex   int
	configFocus   ConfigField
	configSaved   bool
	configSaving  bool

	// Shared
	isFirstRun  bool
	welcome     WelcomeModel
	activePanel Panel
	spinner     spinner.Model
	width       int
	height      int

	// Track original values to detect changes
	originalTimezoneIndex int
	originalLanguageIndex int
	originalLayoutIndex   int
}

type APIKeyValidatedMsg struct {
	Valid bool
	Error string
}

type ConfigSavedMsg struct {
	Success bool
	Error   string
}

type ConnectionSavedMsg struct {
	Success bool
	Error   string
}

type PreferencesSavedMsg struct {
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
	endpointInput.Width = 40

	apiKeyInput := textinput.New()
	apiKeyInput.Placeholder = "Enter your API key"
	apiKeyInput.EchoMode = textinput.EchoPassword
	apiKeyInput.EchoCharacter = '*'
	apiKeyInput.Width = 40

	timezones := config.ListTimezones()
	languages := i18n.ListLanguages()
	layouts := []string{config.LayoutHorizontal, config.LayoutVertical}

	tzIndex := 0
	langIndex := 0
	layoutIndex := 0
	authMethod := AuthMethodOAuth // Default to OAuth

	// Check if we already have OAuth tokens
	oauthLoggedIn := cfg != nil && cfg.HasOAuthTokens()

	if cfg != nil && cfg.IsValid() {
		endpointInput.SetValue(cfg.Endpoint)
		apiKeyInput.SetValue(cfg.APIKey)

		if cfg.UseOAuth {
			authMethod = AuthMethodOAuth
		} else if cfg.APIKey != "" {
			authMethod = AuthMethodAPIKey
		}

		for i, tz := range timezones {
			if tz == cfg.Timezone {
				tzIndex = i
				break
			}
		}

		langIndex = i18n.LanguageIndex(cfg.Language)

		for i, layout := range layouts {
			if layout == cfg.Layout {
				layoutIndex = i
				break
			}
		}
	} else {
		endpointInput.SetValue(config.DefaultEndpoint)

		detectedTZ := config.DetectTimezone()
		for i, tz := range timezones {
			if tz == detectedTZ {
				tzIndex = i
				break
			}
		}

		detectedLang := i18n.DetectLanguage()
		langIndex = i18n.LanguageIndex(string(detectedLang))
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Spinner

	firstRun := cfg == nil || !cfg.IsValid()
	welcome := NewWelcomeModel()

	return SetupModel{
		authMethod:            authMethod,
		endpoint:              endpointInput,
		apiKey:                apiKeyInput,
		connFocus:             ConnFieldAuthMethod,
		connButton:            0,
		oauthLoggedIn:         oauthLoggedIn,
		isFirstRun:            firstRun,
		welcome:               welcome,
		timezones:             timezones,
		timezoneIndex:         tzIndex,
		languages:             languages,
		languageIndex:         langIndex,
		layouts:               layouts,
		layoutIndex:           layoutIndex,
		configFocus:           ConfigFieldTimezone,
		activePanel:           PanelConnection,
		spinner:               s,
		originalTimezoneIndex: tzIndex,
		originalLanguageIndex: langIndex,
		originalLayoutIndex:   layoutIndex,
	}
}

func (m SetupModel) Init() tea.Cmd {
	if m.isFirstRun {
		return tea.Batch(textinput.Blink, m.welcome.Init())
	}
	return textinput.Blink
}

func (m SetupModel) Update(msg tea.Msg) (SetupModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.testing || m.connSaving || m.configSaving || m.oauthLoggingIn {
			return m, nil
		}

		var cmd tea.Cmd
		var handled bool
		m, cmd, handled = m.handleKeyMsg(msg)
		if handled {
			return m, cmd
		}

	case OAuthLoginResultMsg:
		m.oauthLoggingIn = false
		if msg.Success {
			m.oauthLoggedIn = true
			m.oauthError = ""
			m.testResult = testResultSuccess
			if m.isFirstRun {
				// Auto-save and proceed to main screen
				m.connSaving = true
				return m, m.doSaveConnection()
			}
		} else {
			m.oauthError = msg.Error
			m.testResult = testResultError
			m.testError = msg.Error
		}
		return m, nil

	case OAuthLogoutResultMsg:
		if msg.Success {
			m.oauthLoggedIn = false
			m.testResult = ""
			m.testError = ""
			m.connSaved = false
			m.connButton = 0
		}
		return m, nil

	case welcomeTickMsg:
		if m.isFirstRun {
			var cmd tea.Cmd
			m.welcome, cmd = m.welcome.Update(msg)
			cmds = append(cmds, cmd)
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
	if m.activePanel == PanelConnection {
		switch m.connFocus {
		case ConnFieldEndpoint:
			m.endpoint, cmd = m.endpoint.Update(msg)
			cmds = append(cmds, cmd)
		case ConnFieldAPIKey:
			m.apiKey, cmd = m.apiKey.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m SetupModel) handleKeyMsg(msg tea.KeyMsg) (SetupModel, tea.Cmd, bool) {
	textInputFocused := m.activePanel == PanelConnection && (m.connFocus == ConnFieldEndpoint || m.connFocus == ConnFieldAPIKey)

	switch msg.String() {
	case "tab":
		return m.handleKeyTab(), nil, true
	case "down":
		return m.handleKeyDown(), nil, true
	case "j":
		if !textInputFocused {
			return m.handleKeyDown(), nil, true
		}
	case "up":
		return m.handleKeyUp(), nil, true
	case "k":
		if !textInputFocused {
			return m.handleKeyUp(), nil, true
		}
	case "left":
		if !textInputFocused {
			return m.handleKeyLeft(), nil, true
		}
	case "h":
		if !textInputFocused {
			return m.handleKeyLeft(), nil, true
		}
	case "right":
		if !textInputFocused {
			return m.handleKeyRight(), nil, true
		}
	case "l":
		if !textInputFocused {
			return m.handleKeyRight(), nil, true
		}
	case "enter":
		updated, cmd := m.handleKeyEnter()
		return updated, cmd, true
	}

	return m, nil, false
}

func (m SetupModel) handleKeyTab() SetupModel {
	if m.isFirstRun {
		return m // No tab switching during first-run wizard
	}
	if m.activePanel == PanelConnection {
		m.activePanel = PanelConfig
		m.endpoint.Blur()
		m.apiKey.Blur()
	} else {
		m.activePanel = PanelConnection
		m.updateConnectionFocus()
	}
	return m
}

func (m SetupModel) handleKeyDown() SetupModel {
	if m.activePanel == PanelConnection {
		m.connFocus++
		// Skip API key field when using OAuth
		if m.authMethod == AuthMethodOAuth && m.connFocus == ConnFieldAPIKey {
			m.connFocus++
		}
		if m.connFocus > ConnFieldButtons {
			m.connFocus = ConnFieldAuthMethod
		}
		m.updateConnectionFocus()
	} else {
		m.configFocus++
		if m.configFocus > ConfigFieldButton {
			m.configFocus = ConfigFieldTimezone
		}
	}
	return m
}

func (m SetupModel) handleKeyUp() SetupModel {
	if m.activePanel == PanelConnection {
		m.connFocus--
		// Skip API key field when using OAuth
		if m.authMethod == AuthMethodOAuth && m.connFocus == ConnFieldAPIKey {
			m.connFocus--
		}
		if m.connFocus < ConnFieldAuthMethod {
			m.connFocus = ConnFieldButtons
		}
		m.updateConnectionFocus()
	} else {
		m.configFocus--
		if m.configFocus < ConfigFieldTimezone {
			m.configFocus = ConfigFieldButton
		}
	}
	return m
}

func (m SetupModel) handleKeyLeft() SetupModel {
	if m.activePanel == PanelConnection {
		switch m.connFocus {
		case ConnFieldAuthMethod:
			if m.authMethod > AuthMethodOAuth {
				m.authMethod--
				m.resetAuthState()
			}
		case ConnFieldButtons:
			if m.connButton > 0 {
				m.connButton--
			}
		}
	} else {
		m.handleConfigLeft()
	}
	return m
}

func (m *SetupModel) handleConfigLeft() {
	switch m.configFocus {
	case ConfigFieldTimezone:
		if m.timezoneIndex > 0 {
			m.timezoneIndex--
		}
	case ConfigFieldLanguage:
		if m.languageIndex > 0 {
			m.languageIndex--
			i18n.SetLanguage(i18n.Language(m.languages[m.languageIndex]))
		}
	case ConfigFieldLayout:
		if m.layoutIndex > 0 {
			m.layoutIndex--
		}
	}
}

func (m SetupModel) handleKeyRight() SetupModel {
	if m.activePanel == PanelConnection {
		switch m.connFocus {
		case ConnFieldAuthMethod:
			if m.authMethod < AuthMethodAPIKey {
				m.authMethod++
				m.resetAuthState()
			}
		case ConnFieldButtons:
			if m.connButton < m.maxConnButton() {
				m.connButton++
			}
		}
	} else {
		m.handleConfigRight()
	}
	return m
}

func (m *SetupModel) resetAuthState() {
	m.testResult = ""
	m.testError = ""
	m.oauthError = ""
	m.connSaved = false
	m.connButton = 0
}

func (m *SetupModel) handleConfigRight() {
	switch m.configFocus {
	case ConfigFieldTimezone:
		if m.timezoneIndex < len(m.timezones)-1 {
			m.timezoneIndex++
		}
	case ConfigFieldLanguage:
		if m.languageIndex < len(m.languages)-1 {
			m.languageIndex++
			i18n.SetLanguage(i18n.Language(m.languages[m.languageIndex]))
		}
	case ConfigFieldLayout:
		if m.layoutIndex < len(m.layouts)-1 {
			m.layoutIndex++
		}
	}
}

func (m SetupModel) handleKeyEnter() (SetupModel, tea.Cmd) {
	if m.activePanel == PanelConnection {
		return m.handleConnectionEnter()
	}
	return m.handleConfigEnter()
}

func (m SetupModel) maxConnButton() int {
	if m.isFirstRun {
		return 0 // Just Login/Test
	}
	if m.authMethod == AuthMethodOAuth && m.oauthLoggedIn {
		return 2 // Login | Save | Logout
	}
	return 1 // Login/Test | Save
}

func (m SetupModel) handleConnectionEnter() (SetupModel, tea.Cmd) {
	if m.connFocus == ConnFieldButtons {
		if m.connButton == 0 {
			if m.authMethod == AuthMethodOAuth {
				// Start OAuth login flow
				m.oauthLoggingIn = true
				m.oauthError = ""
				m.testResult = ""
				m.testError = ""
				m.connSaved = false
				return m, tea.Batch(m.spinner.Tick, m.doOAuthLogin())
			}
			// Test API key connection
			m.testing = true
			m.testResult = ""
			m.testError = ""
			m.connSaved = false
			return m, tea.Batch(m.spinner.Tick, m.doTestConnection())
		}
		if m.connButton == 1 {
			// Save connection
			if m.testResult == testResultSuccess {
				m.connSaving = true
				return m, m.doSaveConnection()
			}
			return m, nil
		}
		if m.connButton == 2 {
			// Logout
			return m, m.doOAuthLogout()
		}
		return m, nil
	}
	// Move to next field on enter
	m.connFocus++
	// Skip API key field when using OAuth
	if m.authMethod == AuthMethodOAuth && m.connFocus == ConnFieldAPIKey {
		m.connFocus++
	}
	if m.connFocus > ConnFieldButtons {
		m.connFocus = ConnFieldButtons
	}
	m.updateConnectionFocus()
	return m, nil
}

func (m SetupModel) handleConfigEnter() (SetupModel, tea.Cmd) {
	if m.configFocus == ConfigFieldButton {
		// Save config
		m.configSaving = true
		return m, m.doSavePreferences()
	}
	// Move to next field on enter
	m.configFocus++
	if m.configFocus > ConfigFieldButton {
		m.configFocus = ConfigFieldButton
	}
	return m, nil
}

func (m *SetupModel) updateConnectionFocus() {
	m.endpoint.Blur()
	m.apiKey.Blur()

	switch m.connFocus {
	case ConnFieldEndpoint:
		m.endpoint.Focus()
	case ConnFieldAPIKey:
		if m.authMethod == AuthMethodAPIKey {
			m.apiKey.Focus()
		}
	}
}

func (m SetupModel) doOAuthLogout() tea.Cmd {
	return func() tea.Msg {
		if err := oauth.ClearTokens(); err != nil {
			return OAuthLogoutResultMsg{Success: false, Error: err.Error()}
		}
		return OAuthLogoutResultMsg{Success: true}
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

func (m SetupModel) doOAuthLogin() tea.Cmd {
	endpointVal := m.endpoint.Value()
	return func() tea.Msg {
		authBaseURL := oauth.DeriveAuthBaseURL(endpointVal)
		debug.Logger.Debug("OAuth config",
			"endpoint_input", endpointVal,
			"auth_base_url", authBaseURL,
			"token_url", authBaseURL+"/oauth/token",
		)

		// Get or register client_id
		clientID := oauth.LoadClientID()
		if clientID == "" {
			debug.Logger.Info("No cached client_id, registering new OAuth client")
			var err error
			clientID, err = oauth.RegisterClient(context.Background(), authBaseURL)
			if err != nil {
				return OAuthLoginResultMsg{Success: false, Error: "Could not register OAuth client: " + err.Error()}
			}
		}
		debug.Logger.Debug("Using OAuth client_id", "client_id", clientID)

		cfg := oauth.NewConfig(authBaseURL, clientID)

		state, err := oauth.GenerateState()
		if err != nil {
			return OAuthLoginResultMsg{Success: false, Error: err.Error()}
		}

		verifier := oauth2.GenerateVerifier()

		authURL := cfg.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

		debug.Logger.Debug("OAuth login starting",
			"auth_url", authURL,
			"state_length", len(state),
		)

		// Start callback server BEFORE opening browser
		codeCh := make(chan string, 1)
		errCh := make(chan error, 1)

		mux := http.NewServeMux()
		mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			gotState := r.URL.Query().Get("state")
			debug.Logger.Debug("OAuth callback received",
				"got_state", gotState,
				"expected_state", state,
				"full_url", r.URL.String(),
			)

			if errParam := r.URL.Query().Get("error"); errParam != "" {
				desc := r.URL.Query().Get("error_description")
				w.Header().Set("Content-Type", "text/html")
				_, _ = fmt.Fprintf(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#1a1a2e;color:#e0e0e0;"><div style="text-align:center;"><h1 style="color:#EF4444;">Login Failed</h1><p>%s: %s</p></div></body></html>`, errParam, desc)
				errCh <- fmt.Errorf("%s: %s", errParam, desc)
				return
			}

			if gotState != state {
				w.Header().Set("Content-Type", "text/html")
				_, _ = fmt.Fprintf(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#1a1a2e;color:#e0e0e0;"><div style="text-align:center;"><h1 style="color:#EF4444;">Login Failed</h1><p>State mismatch</p><p style="font-size:0.8em;color:#888;">Expected: %s</p><p style="font-size:0.8em;color:#888;">Got: %s</p></div></body></html>`, state, gotState)
				errCh <- fmt.Errorf("state mismatch: expected %q, got %q", state, gotState)
				return
			}

			code := r.URL.Query().Get("code")
			if code == "" {
				w.Header().Set("Content-Type", "text/html")
				_, _ = fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#1a1a2e;color:#e0e0e0;"><div style="text-align:center;"><h1 style="color:#EF4444;">Login Failed</h1><p>No code received</p></div></body></html>`)
				errCh <- fmt.Errorf("no code received")
				return
			}

			w.Header().Set("Content-Type", "text/html")
			_, _ = fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#1a1a2e;color:#e0e0e0;"><div style="text-align:center;"><h1 style="color:#7C3AED;">Login Successful</h1><p>You can close this window and return to your terminal.</p></div></body></html>`)
			codeCh <- code
		})

		// Use a listener so we know the server is ready before opening the browser
		lc := net.ListenConfig{}
		listener, err := lc.Listen(context.Background(), "tcp", oauth.CallbackPort)
		if err != nil {
			return OAuthLoginResultMsg{Success: false, Error: "Failed to start callback server: " + err.Error()}
		}

		srv := &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 30 * time.Second,
		}

		go func() {
			if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}()

		// Now open browser — server is guaranteed to be listening
		if err := openBrowser(authURL); err != nil {
			_ = srv.Close()
			return OAuthLoginResultMsg{Success: false, Error: "Failed to open browser: " + err.Error()}
		}

		// Wait with 5 minute timeout
		timer := time.NewTimer(5 * time.Minute)
		defer timer.Stop()
		defer func() { _ = srv.Close() }()

		select {
		case code := <-codeCh:
			ctx := context.Background()
			tok, err := oauth.ExchangeCode(ctx, cfg, code, verifier)
			if err != nil {
				return OAuthLoginResultMsg{Success: false, Error: "Token exchange failed: " + err.Error()}
			}
			if err := oauth.SaveOAuth2Token(tok); err != nil {
				return OAuthLoginResultMsg{Success: false, Error: "Failed to save tokens: " + err.Error()}
			}
			return OAuthLoginResultMsg{Success: true}
		case err := <-errCh:
			errMsg := err.Error()
			// If authorization failed (possibly stale client_id), clear it so next attempt re-registers
			if strings.Contains(errMsg, "invalid_client") || strings.Contains(errMsg, "client_not_found") {
				debug.Logger.Warn("OAuth client may be stale, clearing cached client_id")
				_ = oauth.ClearClientID()
			}
			return OAuthLoginResultMsg{Success: false, Error: errMsg}
		case <-timer.C:
			return OAuthLoginResultMsg{Success: false, Error: "Login timed out (5 minutes)"}
		}
	}
}

func openBrowser(url string) error {
	ctx := context.Background()
	switch runtime.GOOS {
	case "darwin":
		return exec.CommandContext(ctx, "open", url).Start()
	case "linux":
		return exec.CommandContext(ctx, "xdg-open", url).Start()
	case "windows":
		return exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func (m SetupModel) doSaveConnection() tea.Cmd {
	timezone := ""
	if m.timezoneIndex >= 0 && m.timezoneIndex < len(m.timezones) {
		timezone = m.timezones[m.timezoneIndex]
	}
	language := string(i18n.DefaultLanguage)
	if m.languageIndex >= 0 && m.languageIndex < len(m.languages) {
		language = m.languages[m.languageIndex]
	}
	layout := config.DefaultLayout
	if m.layoutIndex >= 0 && m.layoutIndex < len(m.layouts) {
		layout = m.layouts[m.layoutIndex]
	}

	useOAuth := m.authMethod == AuthMethodOAuth
	apiKeyVal := m.apiKey.Value()
	endpointVal := m.endpoint.Value()

	return func() tea.Msg {
		// Load existing config to preserve OAuth tokens
		cfg, err := config.Load()
		if err != nil {
			cfg = &config.Config{}
		}

		cfg.Endpoint = endpointVal
		cfg.APIKey = apiKeyVal
		cfg.Timezone = timezone
		cfg.Language = language
		cfg.Layout = layout
		cfg.UseOAuth = useOAuth

		if useOAuth {
			cfg.APIKey = "" // Clear API key when using OAuth
		}

		if err := config.Save(cfg); err != nil {
			return ConnectionSavedMsg{Success: false, Error: err.Error()}
		}

		return ConnectionSavedMsg{Success: true}
	}
}

func (m SetupModel) doSavePreferences() tea.Cmd {
	timezone := ""
	if m.timezoneIndex >= 0 && m.timezoneIndex < len(m.timezones) {
		timezone = m.timezones[m.timezoneIndex]
	}
	language := string(i18n.DefaultLanguage)
	if m.languageIndex >= 0 && m.languageIndex < len(m.languages) {
		language = m.languages[m.languageIndex]
	}
	layout := config.DefaultLayout
	if m.layoutIndex >= 0 && m.layoutIndex < len(m.layouts) {
		layout = m.layouts[m.layoutIndex]
	}

	return func() tea.Msg {
		// Load existing config to preserve connection settings
		existingCfg, err := config.Load()
		if err != nil {
			// If no existing config, use current values
			existingCfg = &config.Config{
				Endpoint: m.endpoint.Value(),
				APIKey:   m.apiKey.Value(),
			}
		}

		// Update only preferences, preserve everything else (including OAuth tokens)
		existingCfg.Timezone = timezone
		existingCfg.Language = language
		existingCfg.Layout = layout
		cfg := existingCfg

		if err := config.Save(cfg); err != nil {
			return PreferencesSavedMsg{Success: false, Error: err.Error()}
		}

		return PreferencesSavedMsg{Success: true}
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

func (m *SetupModel) HandleConnectionSaved(msg ConnectionSavedMsg) {
	m.connSaving = false
	m.connSaved = msg.Success
}

func (m *SetupModel) HandlePreferencesSaved(msg PreferencesSavedMsg) {
	m.configSaving = false
	m.configSaved = msg.Success
	if msg.Success {
		// Update original values after save
		m.originalTimezoneIndex = m.timezoneIndex
		m.originalLanguageIndex = m.languageIndex
		m.originalLayoutIndex = m.layoutIndex
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

func (m SetupModel) IsFirstRun() bool {
	return m.isFirstRun
}

func (m SetupModel) DoSaveConnection() tea.Cmd {
	return m.doSaveConnection()
}

func (m SetupModel) IsConnectionSaved() bool {
	return m.connSaved
}

func (m SetupModel) IsConfigSaved() bool {
	return m.configSaved
}

func (m SetupModel) View() string {
	if m.isFirstRun {
		return m.renderFirstRunView()
	}
	return m.renderFullSetupView()
}

func (m SetupModel) renderFirstRunView() string {
	panelWidth := 54

	var parts []string

	// Animated logo
	parts = append(parts, m.welcome.View())

	// Subtitle
	if sub := m.welcome.RenderSubtitle(); sub != "" {
		parts = append(parts, sub, "")
	}

	// Auth panel after animation
	if m.welcome.ShowPanel() {
		parts = append(parts,
			styles.TextDim.Render("Let's connect to your Rootly account."), "",
			m.renderFirstRunPanel(panelWidth), "",
			styles.HelpBar.Render("Configure preferences later with "+styles.HelpKey.Render("s")),
		)
	}

	content := lipgloss.JoinVertical(lipgloss.Center, parts...)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

func (m SetupModel) renderFirstRunPanel(panelWidth int) string {
	var b strings.Builder

	// Auth method selector
	authLabel := styles.InputLabel.Render("Authentication")
	b.WriteString(authLabel + "\n")
	var oauthOpt, apikeyOpt string
	if m.authMethod == AuthMethodOAuth {
		oauthOpt = styles.Primary.Bold(true).Render("● OAuth2")
		apikeyOpt = styles.TextDim.Render("○ API Key")
	} else {
		oauthOpt = styles.TextDim.Render("○ OAuth2")
		apikeyOpt = styles.Primary.Bold(true).Render("● API Key")
	}
	authDisplay := oauthOpt + "  " + apikeyOpt
	if m.connFocus == ConnFieldAuthMethod {
		b.WriteString(styles.InputFieldFocused.Render(authDisplay))
	} else {
		b.WriteString(styles.InputField.Render(authDisplay))
	}
	b.WriteString("\n\n")

	// Endpoint
	b.WriteString(styles.InputLabel.Render("API Endpoint") + "\n")
	if m.connFocus == ConnFieldEndpoint {
		b.WriteString(styles.InputFieldFocused.Render(m.endpoint.View()))
	} else {
		b.WriteString(styles.InputField.Render(m.endpoint.View()))
	}
	b.WriteString("\n\n")

	// API key field or OAuth hint
	if m.authMethod == AuthMethodAPIKey {
		b.WriteString(styles.InputLabel.Render("API Key") + "\n")
		if m.connFocus == ConnFieldAPIKey {
			b.WriteString(styles.InputFieldFocused.Render(m.apiKey.View()))
		} else {
			b.WriteString(styles.InputField.Render(m.apiKey.View()))
		}
		b.WriteString("\n\n")
	}

	// Status
	if m.testing || m.oauthLoggingIn {
		label := "Testing connection..."
		if m.oauthLoggingIn {
			label = "Waiting for browser login..."
		}
		b.WriteString(m.spinner.View() + " " + label + "\n\n")
	} else if m.testResult == testResultError {
		errMsg := m.testError
		if len(errMsg) > 40 {
			errMsg = errMsg[:37] + "..."
		}
		b.WriteString(styles.Error.Render("Error: "+errMsg) + "\n\n")
	} else if m.connSaving {
		b.WriteString(m.spinner.View() + " Saving...\n\n")
	}

	// Button
	actionLabel := "Login"
	if m.authMethod == AuthMethodAPIKey {
		actionLabel = "Connect"
	}
	if m.connFocus == ConnFieldButtons && m.connButton == 0 {
		b.WriteString(styles.ButtonFocused.Render(actionLabel))
	} else {
		b.WriteString(styles.Button.Render(actionLabel))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPurple).
		Padding(1, 2).
		Width(panelWidth).
		Render(b.String())
}

func (m SetupModel) renderFullSetupView() string {
	panelWidth := 52
	activeBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPurple).
		Padding(1, 2).
		Width(panelWidth)

	inactiveBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Padding(1, 2).
		Width(panelWidth)

	// Connection panel
	connPanel := m.renderConnectionPanel()
	if m.activePanel == PanelConnection {
		connPanel = activeBorder.Render(connPanel)
	} else {
		connPanel = inactiveBorder.Render(connPanel)
	}

	// Config panel
	configPanel := m.renderConfigPanel()
	if m.activePanel == PanelConfig {
		configPanel = activeBorder.Render(configPanel)
	} else {
		configPanel = inactiveBorder.Render(configPanel)
	}

	// Join panels horizontally
	panels := lipgloss.JoinHorizontal(lipgloss.Top, connPanel, "  ", configPanel)

	// Help text
	help := styles.HelpBar.Render(i18n.T("setup.help_panels"))

	// Combine all
	content := lipgloss.JoinVertical(lipgloss.Center, panels, "", help)

	// Center on screen
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	return content
}

func (m SetupModel) renderConnectionPanel() string {
	var b strings.Builder

	// Panel title
	title := styles.DialogTitle.Render(i18n.T("setup.connection_title"))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Auth method selector
	authLabel := styles.InputLabel.Render("Authentication")
	b.WriteString(authLabel)
	b.WriteString("\n")
	var oauthOpt, apikeyOpt string
	if m.authMethod == AuthMethodOAuth {
		oauthOpt = styles.Primary.Bold(true).Render("● OAuth2")
		apikeyOpt = styles.TextDim.Render("○ API Key")
	} else {
		oauthOpt = styles.TextDim.Render("○ OAuth2")
		apikeyOpt = styles.Primary.Bold(true).Render("● API Key")
	}
	authDisplay := oauthOpt + "  " + apikeyOpt
	if m.activePanel == PanelConnection && m.connFocus == ConnFieldAuthMethod {
		b.WriteString(styles.InputFieldFocused.Render(authDisplay))
	} else {
		b.WriteString(styles.InputField.Render(authDisplay))
	}
	b.WriteString("\n\n")

	// Endpoint field
	endpointLabel := styles.InputLabel.Render(i18n.T("setup.api_endpoint"))
	b.WriteString(endpointLabel)
	b.WriteString("\n")
	if m.activePanel == PanelConnection && m.connFocus == ConnFieldEndpoint {
		b.WriteString(styles.InputFieldFocused.Render(m.endpoint.View()))
	} else {
		b.WriteString(styles.InputField.Render(m.endpoint.View()))
	}
	b.WriteString("\n\n")

	if m.authMethod == AuthMethodAPIKey {
		// API Key field (only for API key auth)
		apiKeyLabel := styles.InputLabel.Render(i18n.T("setup.api_key"))
		b.WriteString(apiKeyLabel)
		b.WriteString("\n")
		if m.activePanel == PanelConnection && m.connFocus == ConnFieldAPIKey {
			b.WriteString(styles.InputFieldFocused.Render(m.apiKey.View()))
		} else {
			b.WriteString(styles.InputField.Render(m.apiKey.View()))
		}
		b.WriteString("\n\n")
	} else {
		// OAuth status
		if m.oauthLoggedIn {
			b.WriteString(styles.SuccessMsg.Render("Logged in via OAuth2"))
			b.WriteString("\n\n")
		} else {
			b.WriteString(styles.TextDim.Render("Opens browser for login"))
			b.WriteString("\n\n")
		}
	}

	// Status messages
	if m.testing || m.oauthLoggingIn {
		label := i18n.T("setup.testing_connection")
		if m.oauthLoggingIn {
			label = "Waiting for browser login..."
		}
		b.WriteString(m.spinner.View() + " " + label)
		b.WriteString("\n\n")
	} else if m.testResult == testResultSuccess {
		b.WriteString(styles.SuccessMsg.Render(i18n.T("setup.connection_success")))
		b.WriteString("\n\n")
	} else if m.testResult == testResultError {
		errMsg := i18n.T("common.error") + ": " + m.testError
		if len(errMsg) > 40 {
			errMsg = errMsg[:37] + "..."
		}
		b.WriteString(styles.Error.Render(errMsg))
		b.WriteString("\n\n")
	} else {
		b.WriteString("\n\n")
	}

	// Saved message
	if m.connSaved {
		b.WriteString(styles.SuccessMsg.Render(i18n.T("setup.connection_saved")))
		b.WriteString("\n\n")
	} else if m.connSaving {
		b.WriteString(m.spinner.View() + " " + i18n.T("common.saving"))
		b.WriteString("\n\n")
	} else {
		b.WriteString("\n\n")
	}

	// Buttons
	var actionBtn string
	actionLabel := i18n.T("setup.test")
	if m.authMethod == AuthMethodOAuth {
		actionLabel = "Login"
	}
	if m.activePanel == PanelConnection && m.connFocus == ConnFieldButtons && m.connButton == 0 {
		actionBtn = styles.ButtonFocused.Render(actionLabel)
	} else {
		actionBtn = styles.Button.Render(actionLabel)
	}

	buttons := actionBtn

	if !m.isFirstRun {
		var saveBtn string
		if m.activePanel == PanelConnection && m.connFocus == ConnFieldButtons && m.connButton == 1 {
			if m.testResult == testResultSuccess {
				saveBtn = styles.ButtonFocused.Render(i18n.T("setup.save"))
			} else {
				saveBtn = styles.ButtonDisabled.Render(i18n.T("setup.save"))
			}
		} else {
			if m.testResult == testResultSuccess {
				saveBtn = styles.Button.Render(i18n.T("setup.save"))
			} else {
				saveBtn = styles.ButtonDisabled.Render(i18n.T("setup.save"))
			}
		}
		buttons += " " + saveBtn

		// Show logout button when OAuth is active and logged in
		if m.authMethod == AuthMethodOAuth && m.oauthLoggedIn {
			var logoutBtn string
			if m.activePanel == PanelConnection && m.connFocus == ConnFieldButtons && m.connButton == 2 {
				logoutBtn = styles.ButtonFocused.Background(styles.ColorDanger).Render("Logout")
			} else {
				logoutBtn = styles.Button.Render("Logout")
			}
			buttons += " " + logoutBtn
		}
	}

	b.WriteString(buttons)

	return b.String()
}

func (m SetupModel) renderConfigPanel() string {
	var b strings.Builder

	// Panel title
	title := styles.DialogTitle.Render(i18n.T("setup.preferences_title"))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Timezone selector
	timezoneLabel := styles.InputLabel.Render(i18n.T("setup.timezone"))
	b.WriteString(timezoneLabel)
	b.WriteString("\n")
	selectedTZ := "UTC"
	if m.timezoneIndex >= 0 && m.timezoneIndex < len(m.timezones) {
		selectedTZ = m.timezones[m.timezoneIndex]
	}
	tzDisplay := fmt.Sprintf("◀ %s ▶", selectedTZ)
	if m.activePanel == PanelConfig && m.configFocus == ConfigFieldTimezone {
		b.WriteString(styles.InputFieldFocused.Render(tzDisplay))
	} else {
		b.WriteString(styles.InputField.Render(tzDisplay))
	}
	b.WriteString("\n\n")

	// Language selector
	languageLabel := styles.InputLabel.Render(i18n.T("setup.language"))
	b.WriteString(languageLabel)
	b.WriteString("\n")
	selectedLang := i18n.LanguageName(string(i18n.DefaultLanguage))
	if m.languageIndex >= 0 && m.languageIndex < len(m.languages) {
		selectedLang = i18n.LanguageName(m.languages[m.languageIndex])
	}
	langDisplay := fmt.Sprintf("◀ %s ▶", selectedLang)
	if m.activePanel == PanelConfig && m.configFocus == ConfigFieldLanguage {
		b.WriteString(styles.InputFieldFocused.Render(langDisplay))
	} else {
		b.WriteString(styles.InputField.Render(langDisplay))
	}
	b.WriteString("\n\n")

	// Layout selector
	layoutLabel := styles.InputLabel.Render(i18n.T("setup.layout"))
	b.WriteString(layoutLabel)
	b.WriteString("\n")
	selectedLayout := layoutDisplayName(config.DefaultLayout)
	if m.layoutIndex >= 0 && m.layoutIndex < len(m.layouts) {
		selectedLayout = layoutDisplayName(m.layouts[m.layoutIndex])
	}
	layoutDisplay := fmt.Sprintf("◀ %s ▶", selectedLayout)
	if m.activePanel == PanelConfig && m.configFocus == ConfigFieldLayout {
		b.WriteString(styles.InputFieldFocused.Render(layoutDisplay))
	} else {
		b.WriteString(styles.InputField.Render(layoutDisplay))
	}
	b.WriteString("\n\n")

	// Spacer to match connection panel height (test result area equivalent)
	b.WriteString("\n\n")

	// Saved message
	if m.configSaved {
		b.WriteString(styles.SuccessMsg.Render(i18n.T("setup.preferences_saved")))
		b.WriteString("\n\n")
	} else if m.configSaving {
		b.WriteString(m.spinner.View() + " " + i18n.T("common.saving"))
		b.WriteString("\n\n")
	} else {
		b.WriteString("\n\n")
	}

	// Save button
	var saveBtn string
	if m.activePanel == PanelConfig && m.configFocus == ConfigFieldButton {
		saveBtn = styles.ButtonFocused.Render(i18n.T("setup.save_preferences"))
	} else {
		saveBtn = styles.Button.Render(i18n.T("setup.save_preferences"))
	}
	b.WriteString(saveBtn)

	return b.String()
}

// layoutDisplayName returns a human-readable name for a layout value
func layoutDisplayName(layout string) string {
	switch layout {
	case config.LayoutHorizontal:
		return i18n.T("setup.layout_horizontal")
	case config.LayoutVertical:
		return i18n.T("setup.layout_vertical")
	default:
		return layout
	}
}

// Backward compatibility methods for tests
func (m SetupModel) FocusIndex() SetupField {
	// Map new structure to old SetupField for tests
	if m.activePanel == PanelConnection {
		switch m.connFocus {
		case ConnFieldAuthMethod:
			return FieldEndpoint // Map to first field for compat
		case ConnFieldEndpoint:
			return FieldEndpoint
		case ConnFieldAPIKey:
			return FieldAPIKey
		case ConnFieldButtons:
			return FieldButtons
		}
	} else {
		switch m.configFocus {
		case ConfigFieldTimezone:
			return FieldTimezone
		case ConfigFieldLanguage:
			return FieldLanguage
		case ConfigFieldLayout:
			return FieldLayout
		case ConfigFieldButton:
			return FieldButtons
		}
	}
	return FieldEndpoint
}

func (m SetupModel) ButtonIndex() int {
	if m.activePanel == PanelConnection {
		return m.connButton
	}
	return 0
}

func (m SetupModel) TimezoneIndex() int {
	return m.timezoneIndex
}

func (m SetupModel) LanguageIndex() int {
	return m.languageIndex
}

func (m SetupModel) LayoutIndex() int {
	return m.layoutIndex
}

func (m SetupModel) ActivePanel() Panel {
	return m.activePanel
}
