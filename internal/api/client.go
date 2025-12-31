package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	rootly "github.com/rootlyhq/rootly-go"
	"github.com/rootlyhq/rootly-tui/internal/config"
)

type Client struct {
	client   *rootly.ClientWithResponses
	endpoint string
}

type Incident struct {
	ID           string
	Title        string
	Summary      string
	Status       string
	Severity     string
	Kind         string
	CreatedAt    time.Time
	StartedAt    *time.Time
	DetectedAt   *time.Time
	AcknowledgedAt *time.Time
	MitigatedAt  *time.Time
	ResolvedAt   *time.Time
	Services     []string
	Environments []string
	Teams        []string
	SlackChannelURL string
	JiraIssueURL    string
}

type Alert struct {
	ID          string
	Summary     string
	Description string
	Status      string
	Source      string
	CreatedAt   time.Time
	StartedAt   *time.Time
	EndedAt     *time.Time
	ExternalURL string
	Services    []string
	Environments []string
	Groups      []string
	Labels      map[string]string
}

func NewClient(cfg *config.Config) (*Client, error) {
	endpoint := cfg.Endpoint
	if endpoint != "" && !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	client, err := rootly.NewClientWithResponses(endpoint,
		rootly.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
			req.Header.Set("Content-Type", "application/vnd.api+json")
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rootly client: %w", err)
	}

	return &Client{
		client:   client,
		endpoint: cfg.Endpoint,
	}, nil
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
	params := &rootly.ListIncidentsParams{
		PageSize: &pageSize,
	}

	resp, err := c.client.ListIncidentsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode())
	}

	var result struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				Title          string `json:"title"`
				Summary        string `json:"summary"`
				Status         string `json:"status"`
				Severity       *struct {
					Data *struct {
						Attributes *struct {
							Name string `json:"name"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"severity"`
				Kind           string  `json:"kind"`
				CreatedAt      string  `json:"created_at"`
				StartedAt      *string `json:"started_at"`
				DetectedAt     *string `json:"detected_at"`
				AcknowledgedAt *string `json:"acknowledged_at"`
				MitigatedAt    *string `json:"mitigated_at"`
				ResolvedAt     *string `json:"resolved_at"`
				SlackChannelURL *string `json:"slack_channel_url"`
				JiraIssueURL    *string `json:"jira_issue_url"`
				Services       *struct {
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
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	incidents := make([]Incident, 0, len(result.Data))
	for _, d := range result.Data {
		incident := Incident{
			ID:      d.ID,
			Title:   d.Attributes.Title,
			Summary: d.Attributes.Summary,
			Status:  d.Attributes.Status,
			Kind:    d.Attributes.Kind,
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

	return incidents, nil
}

func (c *Client) ListAlerts(ctx context.Context) ([]Alert, error) {
	pageSize := 50
	params := &rootly.ListAlertsParams{
		PageSize: &pageSize,
	}

	resp, err := c.client.ListAlertsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode())
	}

	var result struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				Summary     string  `json:"summary"`
				Description *string `json:"description"`
				Status      string  `json:"status"`
				Source      string  `json:"source"`
				CreatedAt   string  `json:"created_at"`
				StartedAt   *string `json:"started_at"`
				EndedAt     *string `json:"ended_at"`
				ExternalURL *string `json:"external_url"`
				Services    *struct {
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
				Labels *[]struct {
					Key   string      `json:"key"`
					Value interface{} `json:"value"`
				} `json:"labels"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	alerts := make([]Alert, 0, len(result.Data))
	for _, d := range result.Data {
		alert := Alert{
			ID:      d.ID,
			Summary: d.Attributes.Summary,
			Status:  d.Attributes.Status,
			Source:  d.Attributes.Source,
			Labels:  make(map[string]string),
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

		if d.Attributes.Services != nil {
			for _, s := range d.Attributes.Services.Data {
				alert.Services = append(alert.Services, s.Attributes.Name)
			}
		}
		if d.Attributes.Environments != nil {
			for _, e := range d.Attributes.Environments.Data {
				alert.Environments = append(alert.Environments, e.Attributes.Name)
			}
		}
		if d.Attributes.Groups != nil {
			for _, g := range d.Attributes.Groups.Data {
				alert.Groups = append(alert.Groups, g.Attributes.Name)
			}
		}
		if d.Attributes.Labels != nil {
			for _, l := range *d.Attributes.Labels {
				alert.Labels[l.Key] = fmt.Sprintf("%v", l.Value)
			}
		}

		alerts = append(alerts, alert)
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
