package debug

import (
	"os"
	"testing"
)

func TestSetLogFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/test-debug.log"

	err := SetLogFile(logPath)
	if err != nil {
		t.Fatalf("SetLogFile failed: %v", err)
	}

	Logger.Info("test entry for permissions check")

	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("failed to stat log file: %v", err)
	}

	perm := info.Mode().Perm()
	if perm&0044 != 0 {
		t.Errorf("log file is readable by group/other: %o (expected 0600)", perm)
	}
	if perm != 0600 {
		t.Errorf("log file permissions = %o, want 0600", perm)
	}

	Disable()
	fileOutput = nil
}

func TestResponseBodyNotLoggedWhenDisabled(t *testing.T) {
	ClearLogs()
	Disable()

	sensitiveBody := []byte(`{"email":"user@example.com","phone":"+1234567890"}`)

	if Enabled {
		Logger.Debug("Response body", "json", PrettyJSON(sensitiveBody))
	}

	logs := GetLogs()
	for _, entry := range logs {
		if contains(entry, "user@example.com") || contains(entry, "+1234567890") {
			t.Errorf("sensitive data found in logs when debug disabled: %s", entry)
		}
	}
}

func TestResponseBodyLoggedWhenEnabled(t *testing.T) {
	ClearLogs()
	Enable()
	defer Disable()

	body := []byte(`{"status":"ok"}`)

	if Enabled {
		Logger.Debug("Response body", "json", PrettyJSON(body))
	}

	logs := GetLogs()
	found := false
	for _, entry := range logs {
		if contains(entry, "status") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected response body in logs when debug is enabled")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
