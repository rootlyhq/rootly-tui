package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rootlyhq/rootly-tui/internal/config"
	"github.com/rootlyhq/rootly-tui/internal/debug"
)

func TestGetIncidentPathEscaping(t *testing.T) {
	defer setupTestEnv(t)()

	var receivedRequestURI string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequestURI = r.RequestURI
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":         "safe_id",
				"attributes": map[string]interface{}{},
			},
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		APIKey:   "test-key",
		Endpoint: server.URL,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	tests := []struct {
		name          string
		id            string
		mustNotAccess string
		mustContain   string
	}{
		{
			name:          "path traversal attempt",
			id:            "../../admin/users",
			mustNotAccess: "/admin/users",
			mustContain:   "%2F",
		},
		{
			name:          "query injection attempt",
			id:            "123?admin=true&",
			mustNotAccess: "admin=true&include=",
			mustContain:   "%3F",
		},
		{
			name:          "slash injection",
			id:            "123/../../v1/users/me",
			mustNotAccess: "/v1/users/me",
			mustContain:   "%2F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _ = client.GetIncident(context.Background(), tt.id, time.Now())

			if strings.Contains(receivedRequestURI, tt.mustNotAccess) {
				t.Errorf("path traversal not prevented: URI %q contains %q", receivedRequestURI, tt.mustNotAccess)
			}
			if !strings.Contains(receivedRequestURI, tt.mustContain) {
				t.Errorf("expected URI to contain escaped sequence %q, got %q", tt.mustContain, receivedRequestURI)
			}
		})
	}
}

func TestGetAlertPathEscaping(t *testing.T) {
	defer setupTestEnv(t)()

	var receivedRequestURI string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequestURI = r.RequestURI
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":         "safe_id",
				"attributes": map[string]interface{}{},
			},
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		APIKey:   "test-key",
		Endpoint: server.URL,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	maliciousID := "../../admin/destroy"
	_, _ = client.GetAlert(context.Background(), maliciousID, time.Now())

	if strings.Contains(receivedRequestURI, "/admin/") {
		t.Errorf("path traversal not prevented: URI %q contains /admin/", receivedRequestURI)
	}
	if !strings.Contains(receivedRequestURI, "%2F") {
		t.Errorf("expected slashes to be escaped in URI: %q", receivedRequestURI)
	}
}

func TestResponseBodyNotInLogsWhenDebugDisabled(t *testing.T) {
	defer setupTestEnv(t)()

	debug.Disable()
	debug.ClearLogs()

	sensitiveEmail := "secret-user-42@internal.rootly.com"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]interface{}{
			"data": []map[string]interface{}{},
			"meta": map[string]interface{}{
				"current_page": 1,
				"total_pages":  1,
				"total_count":  0,
			},
			"sensitive_field": sensitiveEmail,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.Config{
		APIKey:   "test-key",
		Endpoint: server.URL,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	_, _ = client.ListIncidents(context.Background(), 1, "")

	logs := debug.GetLogs()
	for _, entry := range logs {
		if strings.Contains(entry, sensitiveEmail) {
			t.Errorf("sensitive data found in logs with debug disabled: %s", entry)
		}
	}
}
