package app

import "github.com/rootlyhq/rootly-tui/internal/api"

// IncidentsLoadedMsg is sent when incidents are loaded from the API
type IncidentsLoadedMsg struct {
	Incidents []api.Incident
	Err       error
}

// AlertsLoadedMsg is sent when alerts are loaded from the API
type AlertsLoadedMsg struct {
	Alerts []api.Alert
	Err    error
}

// ErrorMsg represents a generic error
type ErrorMsg struct {
	Err error
}
