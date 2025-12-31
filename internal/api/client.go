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
const DefaultCacheTTL = 30 * time.Second

type Client struct {
	client   *rootly.ClientWithResponses
	endpoint string
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
}

type Alert struct {
	ID           string
	ShortID      string
	Summary      string
	Description  string
	Status       string
	Source       string
	CreatedAt    time.Time
	StartedAt    *time.Time
	EndedAt      *time.Time
	ExternalURL  string
	Services     []string
	Environments []string
	Groups       []string
	Labels       map[string]string
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
			cache:    nil,
		}, nil
	}

	return &Client{
		client:   client,
		endpoint: cfg.Endpoint,
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
	pageSize := 1
	params := &rootly.ListIncidentsParams{
		PageSize: &pageSize,
	}
	resp, err := c.client.ListIncidentsWithResponse(ctx, params)
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

func (c *Client) ListIncidents(ctx context.Context) ([]Incident, error) {
	pageSize := 50

	// Build cache key with parameters
	cacheKey := NewCacheKey(CacheKeyPrefixIncidents).
		With("pageSize", pageSize).
		Build()

	// Check cache first
	if c.cache != nil {
		var cached []Incident
		if c.cache.GetTyped(cacheKey, &cached) {
			debug.Logger.Debug("Cache hit for incidents", "key", cacheKey)
			return cached, nil
		}
	}

	params := &rootly.ListIncidentsParams{
		PageSize: &pageSize,
	}

	debug.Logger.Debug("Fetching incidents", "pageSize", pageSize, "cache", "miss", "key", cacheKey)

	resp, err := c.client.ListIncidentsWithResponse(ctx, params)
	if err != nil {
		debug.Logger.Error("Failed to list incidents", "error", err)
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}

	debug.Logger.Debug("Incidents response",
		"status", resp.StatusCode(),
		"bodyLength", len(resp.Body),
	)
	debug.Logger.Debug("Incidents response body", "json", debug.PrettyJSON(resp.Body))

	if resp.StatusCode() != 200 {
		debug.Logger.Error("API error", "status", resp.StatusCode(), "body", debug.PrettyJSON(resp.Body))
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode())
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
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		debug.Logger.Error("Failed to parse incidents response",
			"error", err,
			"body", debug.PrettyJSON(resp.Body),
		)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	debug.Logger.Debug("Parsed incidents", "count", len(result.Data))

	incidents := make([]Incident, 0, len(result.Data))
	for _, d := range result.Data {
		incident := Incident{
			ID:      d.ID,
			Title:   d.Attributes.Title,
			Summary: d.Attributes.Summary,
			Status:  d.Attributes.Status,
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

	// Store in cache
	if c.cache != nil {
		c.cache.Set(cacheKey, incidents)
		debug.Logger.Debug("Cached incidents", "count", len(incidents), "key", cacheKey)
	}

	return incidents, nil
}

func (c *Client) ListAlerts(ctx context.Context) ([]Alert, error) {
	pageSize := 50

	// Build cache key with parameters
	cacheKey := NewCacheKey(CacheKeyPrefixAlerts).
		With("pageSize", pageSize).
		Build()

	// Check cache first
	if c.cache != nil {
		var cached []Alert
		if c.cache.GetTyped(cacheKey, &cached) {
			debug.Logger.Debug("Cache hit for alerts", "key", cacheKey)
			return cached, nil
		}
	}

	params := &rootly.ListAlertsParams{
		PageSize: &pageSize,
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
			Summary: d.Attributes.Summary,
			Status:  d.Attributes.Status,
			Labels:  make(map[string]string),
		}

		if d.Attributes.Source != nil {
			alert.Source = *d.Attributes.Source
		}

		if d.Attributes.ShortID != nil {
			alert.ShortID = *d.Attributes.ShortID
		}
		if d.Attributes.Description != nil {
			alert.Description = *d.Attributes.Description
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

	// Store in cache
	if c.cache != nil {
		c.cache.Set(cacheKey, alerts)
		debug.Logger.Debug("Cached alerts", "count", len(alerts), "key", cacheKey)
	}

	return alerts, nil
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
