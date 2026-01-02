package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	rootly "github.com/rootlyhq/rootly-go"

	"github.com/rootlyhq/rootly-tui/internal/config"
	"github.com/rootlyhq/rootly-tui/internal/debug"
)

// DefaultCacheTTL is the default cache duration
const DefaultCacheTTL = 5 * time.Minute

type Client struct {
	client   *rootly.ClientWithResponses
	endpoint string
	apiKey   string
	cache    *PersistentCache
}

type Incident struct {
	ID              string
	SequentialID    string
	Title           string
	Summary         string
	Status          string
	Severity        string
	Kind            string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	StartedAt       *time.Time
	DetectedAt      *time.Time
	AcknowledgedAt  *time.Time
	MitigatedAt     *time.Time
	ResolvedAt      *time.Time
	Services        []string
	Environments    []string
	Teams           []string
	SlackChannelURL string
	JiraIssueURL    string
	// Detail fields (populated by GetIncident)
	URL              string
	ShortURL         string
	Causes           []string
	IncidentTypes    []string
	Functionalities  []string
	Roles            []IncidentRole
	CommanderName    string
	CommunicatorName string
	CreatedByName    string
	CreatedByEmail   string
	DetailLoaded     bool
}

type IncidentRole struct {
	Name      string
	UserName  string
	UserEmail string
}

type Alert struct {
	ID           string
	ShortID      string
	Summary      string
	Description  string
	Status       string
	Source       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	StartedAt    *time.Time
	EndedAt      *time.Time
	ExternalURL  string
	Services     []string
	Environments []string
	Groups       []string
	Labels       map[string]string
	// Detail fields (populated by GetAlert)
	Responders   []string
	Urgency      string
	DetailLoaded bool
}

// PaginationInfo contains pagination state
type PaginationInfo struct {
	CurrentPage int
	HasNext     bool
	HasPrev     bool
}

// IncidentsResult contains incidents and pagination info
type IncidentsResult struct {
	Incidents  []Incident
	Pagination PaginationInfo
}

// AlertsResult contains alerts and pagination info
type AlertsResult struct {
	Alerts     []Alert
	Pagination PaginationInfo
}

func NewClient(cfg *config.Config) (*Client, error) {
	endpoint := cfg.Endpoint
	if endpoint != "" && !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	debug.Logger.Debug("Creating API client", "endpoint", endpoint)

	client, err := rootly.NewClientWithResponses(endpoint,
		rootly.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
			req.Header.Set("Content-Type", "application/vnd.api+json")
			debug.Logger.Debug("API request",
				"method", req.Method,
				"url", req.URL.String(),
			)
			return nil
		}),
	)
	if err != nil {
		debug.Logger.Error("Failed to create client", "error", err)
		return nil, fmt.Errorf("failed to create rootly client: %w", err)
	}

	cache, err := NewPersistentCache(DefaultCacheTTL)
	if err != nil {
		debug.Logger.Warn("Failed to create persistent cache, using in-memory", "error", err)
		// Fall back to in-memory cache
		return &Client{
			client:   client,
			endpoint: cfg.Endpoint,
			apiKey:   cfg.APIKey,
			cache:    nil,
		}, nil
	}

	return &Client{
		client:   client,
		endpoint: cfg.Endpoint,
		apiKey:   cfg.APIKey,
		cache:    cache,
	}, nil
}

// ClearCache clears all cached data
func (c *Client) ClearCache() {
	if c.cache != nil {
		c.cache.Clear()
		debug.Logger.Debug("Cache cleared")
	}
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	if c.cache != nil {
		return c.cache.Close()
	}
	return nil
}

func (c *Client) ValidateAPIKey(ctx context.Context) error {
	// Use /v1/users/me endpoint to validate the API key
	resp, err := c.client.GetCurrentUserWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	if resp.StatusCode() == 401 {
		return fmt.Errorf("invalid API key")
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("API returned status %d", resp.StatusCode())
	}
	return nil
}

func (c *Client) ListIncidents(ctx context.Context, page int, sort string) (*IncidentsResult, error) {
	pageSize := 25

	// Build cache key with parameters including sort
	cacheKeyBuilder := NewCacheKey(CacheKeyPrefixIncidents).
		With("page", page).
		With("pageSize", pageSize)
	if sort != "" {
		cacheKeyBuilder = cacheKeyBuilder.With("sort", sort)
	}
	cacheKey := cacheKeyBuilder.Build()

	// Check cache first
	if c.cache != nil {
		var cached IncidentsResult
		if c.cache.GetTyped(cacheKey, &cached) {
			debug.Logger.Debug("Cache hit for incidents", "key", cacheKey)
			return &cached, nil
		}
	}

	// Build URL with query parameters
	baseURL := c.endpoint
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	url := fmt.Sprintf("%s/v1/incidents?page[number]=%d&page[size]=%d", baseURL, page, pageSize)
	if sort != "" {
		url += fmt.Sprintf("&sort=%s", sort)
	}

	debug.Logger.Debug("Fetching incidents", "page", page, "pageSize", pageSize, "sort", sort, "cache", "miss", "key", cacheKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.Logger.Error("Failed to list incidents", "error", err)
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	debug.Logger.Debug("Incidents response",
		"status", httpResp.StatusCode,
		"bodyLength", len(body),
	)
	debug.Logger.Debug("Incidents response body", "json", debug.PrettyJSON(body))

	if httpResp.StatusCode == 403 {
		debug.Logger.Error("API forbidden", "status", httpResp.StatusCode)
		return nil, fmt.Errorf("access denied: API key lacks 'read incidents' permission")
	}
	if httpResp.StatusCode != 200 {
		debug.Logger.Error("API error", "status", httpResp.StatusCode, "body", debug.PrettyJSON(body))
		return nil, fmt.Errorf("API returned status %d", httpResp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				SequentialID *int   `json:"sequential_id"`
				Title        string `json:"title"`
				Summary      string `json:"summary"`
				Status       string `json:"status"`
				Severity     *struct {
					Data *struct {
						Attributes *struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"severity"`
				Kind            string  `json:"kind"`
				CreatedAt       string  `json:"created_at"`
				StartedAt       *string `json:"started_at"`
				DetectedAt      *string `json:"detected_at"`
				AcknowledgedAt  *string `json:"acknowledged_at"`
				MitigatedAt     *string `json:"mitigated_at"`
				ResolvedAt      *string `json:"resolved_at"`
				SlackChannelURL *string `json:"slack_channel_url"`
				JiraIssueURL    *string `json:"jira_issue_url"`
				Services        *struct {
					Data []struct {
						Attributes struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"services"`
				Environments *struct {
					Data []struct {
						Attributes struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"environments"`
				Groups *struct {
					Data []struct {
						Attributes struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"groups"`
			} `json:"attributes"`
		} `json:"data"`
		Links struct {
			Next *string `json:"next"`
			Prev *string `json:"prev"`
		} `json:"links"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		debug.Logger.Error("Failed to parse incidents response",
			"error", err,
			"body", debug.PrettyJSON(body),
		)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	debug.Logger.Debug("Parsed incidents", "count", len(result.Data))

	incidents := make([]Incident, 0, len(result.Data))
	for _, d := range result.Data {
		incident := Incident{
			ID:      d.ID,
			Title:   strings.TrimSpace(d.Attributes.Title),
			Summary: strings.TrimSpace(d.Attributes.Summary),
			Status:  strings.TrimSpace(d.Attributes.Status),
			Kind:    d.Attributes.Kind,
		}

		if d.Attributes.SequentialID != nil {
			incident.SequentialID = fmt.Sprintf("INC-%d", *d.Attributes.SequentialID)
		}

		if d.Attributes.Severity != nil && d.Attributes.Severity.Data != nil && d.Attributes.Severity.Data.Attributes != nil {
			incident.Severity = d.Attributes.Severity.Data.Attributes.Name
		}

		if t, err := time.Parse(time.RFC3339, d.Attributes.CreatedAt); err == nil {
			incident.CreatedAt = t
		}
		incident.StartedAt = parseTimePtr(d.Attributes.StartedAt)
		incident.DetectedAt = parseTimePtr(d.Attributes.DetectedAt)
		incident.AcknowledgedAt = parseTimePtr(d.Attributes.AcknowledgedAt)
		incident.MitigatedAt = parseTimePtr(d.Attributes.MitigatedAt)
		incident.ResolvedAt = parseTimePtr(d.Attributes.ResolvedAt)

		if d.Attributes.SlackChannelURL != nil {
			incident.SlackChannelURL = *d.Attributes.SlackChannelURL
		}
		if d.Attributes.JiraIssueURL != nil {
			incident.JiraIssueURL = *d.Attributes.JiraIssueURL
		}

		if d.Attributes.Services != nil {
			for _, s := range d.Attributes.Services.Data {
				incident.Services = append(incident.Services, s.Attributes.Name)
			}
		}
		if d.Attributes.Environments != nil {
			for _, e := range d.Attributes.Environments.Data {
				incident.Environments = append(incident.Environments, e.Attributes.Name)
			}
		}
		if d.Attributes.Groups != nil {
			for _, g := range d.Attributes.Groups.Data {
				incident.Teams = append(incident.Teams, g.Attributes.Name)
			}
		}

		incidents = append(incidents, incident)
	}

	// Build result with pagination info
	incidentsResult := &IncidentsResult{
		Incidents: incidents,
		Pagination: PaginationInfo{
			CurrentPage: page,
			HasNext:     result.Links.Next != nil && *result.Links.Next != "",
			HasPrev:     result.Links.Prev != nil && *result.Links.Prev != "",
		},
	}

	// Store in cache
	if c.cache != nil {
		c.cache.Set(cacheKey, incidentsResult)
		debug.Logger.Debug("Cached incidents", "count", len(incidents), "key", cacheKey)
	}

	return incidentsResult, nil
}

func (c *Client) ListAlerts(ctx context.Context, page int) (*AlertsResult, error) {
	pageSize := 25

	// Build cache key with parameters
	cacheKey := NewCacheKey(CacheKeyPrefixAlerts).
		With("page", page).
		With("pageSize", pageSize).
		Build()

	// Check cache first
	if c.cache != nil {
		var cached AlertsResult
		if c.cache.GetTyped(cacheKey, &cached) {
			debug.Logger.Debug("Cache hit for alerts", "key", cacheKey)
			return &cached, nil
		}
	}

	params := &rootly.ListAlertsParams{
		PageNumber: &page,
		PageSize:   &pageSize,
	}

	debug.Logger.Debug("Fetching alerts", "pageSize", pageSize, "cache", "miss", "key", cacheKey)

	// Use raw ListAlerts to bypass SDK's broken parsing (labels.value is interface{}, not string)
	httpResp, err := c.client.ListAlerts(ctx, params)
	if err != nil {
		debug.Logger.Error("Failed to list alerts", "error", err)
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		debug.Logger.Error("Failed to read alerts response", "error", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	debug.Logger.Debug("Alerts response",
		"status", httpResp.StatusCode,
		"bodyLength", len(body),
	)
	debug.Logger.Debug("Alerts response body", "json", debug.PrettyJSON(body))

	if httpResp.StatusCode == 403 {
		debug.Logger.Error("API forbidden", "status", httpResp.StatusCode)
		return nil, fmt.Errorf("access denied: API key lacks 'read alerts' permission")
	}
	if httpResp.StatusCode != 200 {
		debug.Logger.Error("API error", "status", httpResp.StatusCode, "body", debug.PrettyJSON(body))
		return nil, fmt.Errorf("API returned status %d", httpResp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				ShortID     *string `json:"short_id"`
				Summary     string  `json:"summary"`
				Description *string `json:"description"`
				Status      string  `json:"status"`
				Source      *string `json:"source"`
				CreatedAt   string  `json:"created_at"`
				StartedAt   *string `json:"started_at"`
				EndedAt     *string `json:"ended_at"`
				ExternalURL *string `json:"external_url"`
				// Direct arrays (not nested data structures like incidents)
				Services []struct {
					Name string `json:"name"`
				} `json:"services"`
				Environments []struct {
					Name string `json:"name"`
				} `json:"environments"`
				Groups []struct {
					Name string `json:"name"`
				} `json:"groups"`
				Labels []struct {
					Key   string      `json:"key"`
					Value interface{} `json:"value"`
				} `json:"labels"`
			} `json:"attributes"`
		} `json:"data"`
		Links struct {
			Next *string `json:"next"`
			Prev *string `json:"prev"`
		} `json:"links"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		debug.Logger.Error("Failed to parse alerts response",
			"error", err,
			"body", debug.PrettyJSON(body),
		)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	debug.Logger.Debug("Parsed alerts", "count", len(result.Data))

	alerts := make([]Alert, 0, len(result.Data))
	for _, d := range result.Data {
		alert := Alert{
			ID:      d.ID,
			Summary: strings.TrimSpace(d.Attributes.Summary),
			Status:  strings.TrimSpace(d.Attributes.Status),
			Labels:  make(map[string]string),
		}

		if d.Attributes.Source != nil {
			alert.Source = *d.Attributes.Source
		}

		if d.Attributes.ShortID != nil {
			alert.ShortID = strings.TrimSpace(*d.Attributes.ShortID)
		}
		if d.Attributes.Description != nil {
			alert.Description = strings.TrimSpace(*d.Attributes.Description)
		}
		if d.Attributes.ExternalURL != nil {
			alert.ExternalURL = *d.Attributes.ExternalURL
		}

		if t, err := time.Parse(time.RFC3339, d.Attributes.CreatedAt); err == nil {
			alert.CreatedAt = t
		}
		alert.StartedAt = parseTimePtr(d.Attributes.StartedAt)
		alert.EndedAt = parseTimePtr(d.Attributes.EndedAt)

		for _, s := range d.Attributes.Services {
			alert.Services = append(alert.Services, s.Name)
		}
		for _, e := range d.Attributes.Environments {
			alert.Environments = append(alert.Environments, e.Name)
		}
		for _, g := range d.Attributes.Groups {
			alert.Groups = append(alert.Groups, g.Name)
		}
		for _, l := range d.Attributes.Labels {
			alert.Labels[l.Key] = fmt.Sprintf("%v", l.Value)
		}

		alerts = append(alerts, alert)
	}

	// Build result with pagination info
	alertsResult := &AlertsResult{
		Alerts: alerts,
		Pagination: PaginationInfo{
			CurrentPage: page,
			HasNext:     result.Links.Next != nil && *result.Links.Next != "",
			HasPrev:     result.Links.Prev != nil && *result.Links.Prev != "",
		},
	}

	// Store in cache
	if c.cache != nil {
		c.cache.Set(cacheKey, alertsResult)
		debug.Logger.Debug("Cached alerts", "count", len(alerts), "key", cacheKey)
	}

	return alertsResult, nil
}

func parseTimePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	return &t
}

// GetIncident fetches detailed incident data by ID
// updatedAt is used for cache invalidation - cache key includes it so changes invalidate the cache
//
//nolint:gocyclo // complexity from parsing deeply nested API response with many optional fields
func (c *Client) GetIncident(ctx context.Context, id string, updatedAt time.Time) (*Incident, error) {
	// Build cache key with updated_at for smart invalidation
	cacheKey := NewCacheKey(CacheKeyPrefixIncidentDetail).
		With("id", id).
		With("updated_at", updatedAt.UTC().Format(time.RFC3339)).
		Build()

	// Check cache first
	if c.cache != nil {
		var cached Incident
		if c.cache.GetTyped(cacheKey, &cached) {
			debug.Logger.Debug("Cache hit for incident detail", "key", cacheKey)
			return &cached, nil
		}
	}

	debug.Logger.Debug("Fetching incident detail", "id", id, "cache", "miss")

	// Build URL - endpoint may already have scheme
	baseURL := c.endpoint
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	url := fmt.Sprintf("%s/v1/incidents/%s?include=roles,causes,incident_types,functionalities,services,environments,groups,user", baseURL, id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.Logger.Error("Failed to fetch incident", "error", err)
		return nil, fmt.Errorf("failed to fetch incident: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	debug.Logger.Debug("Incident detail response",
		"status", httpResp.StatusCode,
		"bodyLength", len(body),
	)
	debug.Logger.Debug("Incident detail response body", "json", debug.PrettyJSON(body))

	if httpResp.StatusCode == 403 {
		debug.Logger.Error("API forbidden", "status", httpResp.StatusCode)
		return nil, fmt.Errorf("access denied: API key lacks 'read incidents' permission")
	}
	if httpResp.StatusCode != 200 {
		debug.Logger.Error("API error", "status", httpResp.StatusCode, "body", debug.PrettyJSON(body))
		return nil, fmt.Errorf("API returned status %d", httpResp.StatusCode)
	}

	var result struct {
		Data struct {
			ID         string `json:"id"`
			Attributes struct {
				SequentialID *int   `json:"sequential_id"`
				Title        string `json:"title"`
				Summary      string `json:"summary"`
				Status       string `json:"status"`
				Severity     *struct {
					Data *struct {
						Attributes *struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"severity"`
				Kind            string  `json:"kind"`
				URL             *string `json:"url"`
				ShortURL        *string `json:"short_url"`
				CreatedAt       string  `json:"created_at"`
				UpdatedAt       string  `json:"updated_at"`
				StartedAt       *string `json:"started_at"`
				DetectedAt      *string `json:"detected_at"`
				AcknowledgedAt  *string `json:"acknowledged_at"`
				MitigatedAt     *string `json:"mitigated_at"`
				ResolvedAt      *string `json:"resolved_at"`
				SlackChannelURL *string `json:"slack_channel_url"`
				JiraIssueURL    *string `json:"jira_issue_url"`
				Services        *struct {
					Data []struct {
						Attributes struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"services"`
				Environments *struct {
					Data []struct {
						Attributes struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"environments"`
				Groups *struct {
					Data []struct {
						Attributes struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"groups"`
				Causes *struct {
					Data []struct {
						Attributes struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"causes"`
				IncidentTypes *struct {
					Data []struct {
						Attributes struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"incident_types"`
				Functionalities *struct {
					Data []struct {
						Attributes struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"functionalities"`
				// User who created the incident
				User *struct {
					Data *struct {
						Attributes struct {
							Name  string `json:"name"`
							Email string `json:"email"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"user"`
			} `json:"attributes"`
		} `json:"data"`
		Included []struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Attributes struct {
				// For groups, services, environments
				Name string `json:"name"`
				// For incident_role_assignments
				IncidentRole *struct {
					Data *struct {
						Attributes struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"incident_role"`
				User *struct {
					Data *struct {
						Attributes struct {
							Name  string `json:"name"`
							Email string `json:"email"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"user"`
			} `json:"attributes"`
		} `json:"included"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		debug.Logger.Error("Failed to parse incident detail response",
			"error", err,
			"body", debug.PrettyJSON(body),
		)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	d := result.Data
	incident := &Incident{
		ID:           d.ID,
		Title:        strings.TrimSpace(d.Attributes.Title),
		Summary:      strings.TrimSpace(d.Attributes.Summary),
		Status:       strings.TrimSpace(d.Attributes.Status),
		Kind:         d.Attributes.Kind,
		DetailLoaded: true,
	}

	if d.Attributes.SequentialID != nil {
		incident.SequentialID = fmt.Sprintf("INC-%d", *d.Attributes.SequentialID)
	}

	if d.Attributes.Severity != nil && d.Attributes.Severity.Data != nil && d.Attributes.Severity.Data.Attributes != nil {
		incident.Severity = d.Attributes.Severity.Data.Attributes.Name
	}

	if d.Attributes.URL != nil {
		incident.URL = *d.Attributes.URL
	}
	if d.Attributes.ShortURL != nil {
		incident.ShortURL = *d.Attributes.ShortURL
	}

	// Parse creator (user who created the incident)
	if d.Attributes.User != nil && d.Attributes.User.Data != nil {
		incident.CreatedByName = d.Attributes.User.Data.Attributes.Name
		incident.CreatedByEmail = d.Attributes.User.Data.Attributes.Email
	}

	if t, err := time.Parse(time.RFC3339, d.Attributes.CreatedAt); err == nil {
		incident.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339, d.Attributes.UpdatedAt); err == nil {
		incident.UpdatedAt = t
	}
	incident.StartedAt = parseTimePtr(d.Attributes.StartedAt)
	incident.DetectedAt = parseTimePtr(d.Attributes.DetectedAt)
	incident.AcknowledgedAt = parseTimePtr(d.Attributes.AcknowledgedAt)
	incident.MitigatedAt = parseTimePtr(d.Attributes.MitigatedAt)
	incident.ResolvedAt = parseTimePtr(d.Attributes.ResolvedAt)

	if d.Attributes.SlackChannelURL != nil {
		incident.SlackChannelURL = *d.Attributes.SlackChannelURL
	}
	if d.Attributes.JiraIssueURL != nil {
		incident.JiraIssueURL = *d.Attributes.JiraIssueURL
	}

	if d.Attributes.Services != nil {
		for _, s := range d.Attributes.Services.Data {
			incident.Services = append(incident.Services, s.Attributes.Name)
		}
	}
	if d.Attributes.Environments != nil {
		for _, e := range d.Attributes.Environments.Data {
			incident.Environments = append(incident.Environments, e.Attributes.Name)
		}
	}
	if d.Attributes.Groups != nil {
		for _, g := range d.Attributes.Groups.Data {
			incident.Teams = append(incident.Teams, g.Attributes.Name)
		}
	}
	// Parse included array (JSON:API format) for roles, groups, services, environments
	for _, inc := range result.Included {
		switch inc.Type {
		case "incident_role_assignments":
			var roleName, userName, userEmail string
			if inc.Attributes.IncidentRole != nil && inc.Attributes.IncidentRole.Data != nil {
				roleName = inc.Attributes.IncidentRole.Data.Attributes.Name
			}
			if inc.Attributes.User != nil && inc.Attributes.User.Data != nil {
				userName = inc.Attributes.User.Data.Attributes.Name
				userEmail = inc.Attributes.User.Data.Attributes.Email
			}
			if roleName != "" {
				role := IncidentRole{Name: roleName, UserName: userName, UserEmail: userEmail}
				incident.Roles = append(incident.Roles, role)
				// Extract commander and communicator
				if strings.EqualFold(roleName, "commander") && userName != "" {
					incident.CommanderName = userName
				}
				if strings.Contains(strings.ToLower(roleName), "communications") && userName != "" {
					incident.CommunicatorName = userName
				}
			}
		case "groups":
			if inc.Attributes.Name != "" {
				incident.Teams = append(incident.Teams, inc.Attributes.Name)
			}
		case "services":
			if inc.Attributes.Name != "" {
				incident.Services = append(incident.Services, inc.Attributes.Name)
			}
		case "environments":
			if inc.Attributes.Name != "" {
				incident.Environments = append(incident.Environments, inc.Attributes.Name)
			}
		}
	}
	if d.Attributes.Causes != nil {
		for _, cause := range d.Attributes.Causes.Data {
			incident.Causes = append(incident.Causes, cause.Attributes.Name)
		}
	}
	if d.Attributes.IncidentTypes != nil {
		for _, it := range d.Attributes.IncidentTypes.Data {
			incident.IncidentTypes = append(incident.IncidentTypes, it.Attributes.Name)
		}
	}
	if d.Attributes.Functionalities != nil {
		for _, f := range d.Attributes.Functionalities.Data {
			incident.Functionalities = append(incident.Functionalities, f.Attributes.Name)
		}
	}

	// Store in cache
	if c.cache != nil {
		c.cache.Set(cacheKey, incident)
		debug.Logger.Debug("Cached incident detail", "id", incident.ID, "key", cacheKey)
	}

	debug.Logger.Debug("Parsed incident detail", "id", incident.ID, "title", incident.Title)
	return incident, nil
}

// GetAlert fetches detailed alert data by ID
// updatedAt is used for cache invalidation - cache key includes it so changes invalidate the cache
func (c *Client) GetAlert(ctx context.Context, id string, updatedAt time.Time) (*Alert, error) {
	// Build cache key with updated_at for smart invalidation
	cacheKey := NewCacheKey(CacheKeyPrefixAlertDetail).
		With("id", id).
		With("updated_at", updatedAt.UTC().Format(time.RFC3339)).
		Build()

	// Check cache first
	if c.cache != nil {
		var cached Alert
		if c.cache.GetTyped(cacheKey, &cached) {
			debug.Logger.Debug("Cache hit for alert detail", "key", cacheKey)
			return &cached, nil
		}
	}

	debug.Logger.Debug("Fetching alert detail", "id", id, "cache", "miss")

	// Build URL - endpoint may already have scheme
	baseURL := c.endpoint
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	url := fmt.Sprintf("%s/v1/alerts/%s?include=services,environments,groups,responders,alert_urgency", baseURL, id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.Logger.Error("Failed to fetch alert", "error", err)
		return nil, fmt.Errorf("failed to fetch alert: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	debug.Logger.Debug("Alert detail response",
		"status", httpResp.StatusCode,
		"bodyLength", len(body),
	)
	debug.Logger.Debug("Alert detail response body", "json", debug.PrettyJSON(body))

	if httpResp.StatusCode == 403 {
		debug.Logger.Error("API forbidden", "status", httpResp.StatusCode)
		return nil, fmt.Errorf("access denied: API key lacks 'read alerts' permission")
	}
	if httpResp.StatusCode != 200 {
		debug.Logger.Error("API error", "status", httpResp.StatusCode, "body", debug.PrettyJSON(body))
		return nil, fmt.Errorf("API returned status %d", httpResp.StatusCode)
	}

	var result struct {
		Data struct {
			ID         string `json:"id"`
			Attributes struct {
				ShortID     *string `json:"short_id"`
				Summary     string  `json:"summary"`
				Description *string `json:"description"`
				Status      string  `json:"status"`
				Source      *string `json:"source"`
				ExternalURL *string `json:"external_url"`
				CreatedAt   string  `json:"created_at"`
				UpdatedAt   string  `json:"updated_at"`
				StartedAt   *string `json:"started_at"`
				EndedAt     *string `json:"ended_at"`
				Labels      []struct {
					Key   string      `json:"key"`
					Value interface{} `json:"value"`
				} `json:"labels"`
				// Direct arrays (same format as list API)
				Services []struct {
					Name string `json:"name"`
				} `json:"services"`
				Environments []struct {
					Name string `json:"name"`
				} `json:"environments"`
				Groups []struct {
					Name string `json:"name"`
				} `json:"groups"`
				Responders []struct {
					ID         interface{} `json:"id"`
					Attributes struct {
						User *struct {
							Data *struct {
								Attributes struct {
									Name string `json:"name"`
								} `json:"attributes"`
							} `json:"data"`
						} `json:"user"`
					} `json:"attributes"`
				} `json:"responders"`
				AlertUrgency *struct {
					Data *struct {
						Attributes struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"alert_urgency"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		debug.Logger.Error("Failed to parse alert detail response",
			"error", err,
			"body", debug.PrettyJSON(body),
		)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	d := result.Data
	alert := &Alert{
		ID:           d.ID,
		Summary:      strings.TrimSpace(d.Attributes.Summary),
		Status:       strings.TrimSpace(d.Attributes.Status),
		Labels:       make(map[string]string),
		DetailLoaded: true,
	}

	if d.Attributes.ShortID != nil {
		alert.ShortID = strings.TrimSpace(*d.Attributes.ShortID)
	}
	if d.Attributes.Source != nil {
		alert.Source = *d.Attributes.Source
	}
	if d.Attributes.Description != nil {
		alert.Description = strings.TrimSpace(*d.Attributes.Description)
	}
	if d.Attributes.ExternalURL != nil {
		alert.ExternalURL = *d.Attributes.ExternalURL
	}

	if t, err := time.Parse(time.RFC3339, d.Attributes.CreatedAt); err == nil {
		alert.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339, d.Attributes.UpdatedAt); err == nil {
		alert.UpdatedAt = t
	}
	alert.StartedAt = parseTimePtr(d.Attributes.StartedAt)
	alert.EndedAt = parseTimePtr(d.Attributes.EndedAt)

	for _, l := range d.Attributes.Labels {
		alert.Labels[l.Key] = fmt.Sprintf("%v", l.Value)
	}

	for _, s := range d.Attributes.Services {
		alert.Services = append(alert.Services, s.Name)
	}
	for _, e := range d.Attributes.Environments {
		alert.Environments = append(alert.Environments, e.Name)
	}
	for _, g := range d.Attributes.Groups {
		alert.Groups = append(alert.Groups, g.Name)
	}

	for _, r := range d.Attributes.Responders {
		if r.Attributes.User != nil && r.Attributes.User.Data != nil {
			alert.Responders = append(alert.Responders, r.Attributes.User.Data.Attributes.Name)
		}
	}

	if d.Attributes.AlertUrgency != nil && d.Attributes.AlertUrgency.Data != nil {
		alert.Urgency = d.Attributes.AlertUrgency.Data.Attributes.Name
	}

	// Store in cache
	if c.cache != nil {
		c.cache.Set(cacheKey, alert)
		debug.Logger.Debug("Cached alert detail", "id", alert.ID, "key", cacheKey)
	}

	debug.Logger.Debug("Parsed alert detail", "id", alert.ID, "summary", alert.Summary)
	return alert, nil
}
