package api

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPersistentCache(t *testing.T) {
	// Use a unique temp directory for this test
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cache, err := NewPersistentCache(30 * time.Second)
	if err != nil {
		t.Fatalf("NewPersistentCache() error = %v", err)
	}
	defer cache.Close()

	// Check that the database file was created
	dbPath := filepath.Join(tmpDir, ".rootly-tui", "cache.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected cache.db to exist")
	}
}

func TestPersistentCacheSetGet(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cache, err := NewPersistentCache(30 * time.Second)
	if err != nil {
		t.Fatalf("NewPersistentCache() error = %v", err)
	}
	defer cache.Close()

	// Test Set and Get
	testValue := []string{"item1", "item2", "item3"}
	cache.Set("test-key", testValue)

	var result []string
	if !cache.GetTyped("test-key", &result) {
		t.Error("expected GetTyped to return true")
	}

	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}
	if result[0] != "item1" {
		t.Errorf("expected 'item1', got '%s'", result[0])
	}
}

func TestPersistentCacheExpiry(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Use a very short TTL
	cache, err := NewPersistentCache(50 * time.Millisecond)
	if err != nil {
		t.Fatalf("NewPersistentCache() error = %v", err)
	}
	defer cache.Close()

	cache.Set("expiring-key", "test-value")

	// Should be available immediately
	var result string
	if !cache.GetTyped("expiring-key", &result) {
		t.Error("expected GetTyped to return true before expiry")
	}

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	if cache.GetTyped("expiring-key", &result) {
		t.Error("expected GetTyped to return false after expiry")
	}
}

func TestPersistentCacheDelete(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cache, err := NewPersistentCache(30 * time.Second)
	if err != nil {
		t.Fatalf("NewPersistentCache() error = %v", err)
	}
	defer cache.Close()

	cache.Set("to-delete", "value")

	var result string
	if !cache.GetTyped("to-delete", &result) {
		t.Error("expected value to exist before delete")
	}

	cache.Delete("to-delete")

	if cache.GetTyped("to-delete", &result) {
		t.Error("expected value to be deleted")
	}
}

func TestPersistentCacheClear(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cache, err := NewPersistentCache(30 * time.Second)
	if err != nil {
		t.Fatalf("NewPersistentCache() error = %v", err)
	}
	defer cache.Close()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	cache.Clear()

	var result string
	if cache.GetTyped("key1", &result) {
		t.Error("expected key1 to be cleared")
	}
	if cache.GetTyped("key2", &result) {
		t.Error("expected key2 to be cleared")
	}
}

func TestPersistentCachePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create cache and set a value
	cache1, err := NewPersistentCache(30 * time.Second)
	if err != nil {
		t.Fatalf("NewPersistentCache() error = %v", err)
	}
	cache1.Set("persistent-key", "persistent-value")
	cache1.Close()

	// Create a new cache instance - should see the same value
	cache2, err := NewPersistentCache(30 * time.Second)
	if err != nil {
		t.Fatalf("NewPersistentCache() second instance error = %v", err)
	}
	defer cache2.Close()

	var result string
	if !cache2.GetTyped("persistent-key", &result) {
		t.Error("expected persistent-key to persist across cache instances")
	}
	if result != "persistent-value" {
		t.Errorf("expected 'persistent-value', got '%s'", result)
	}
}

func TestPersistentCacheWithIncidentStruct(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cache, err := NewPersistentCache(30 * time.Second)
	if err != nil {
		t.Fatalf("NewPersistentCache() error = %v", err)
	}
	defer cache.Close()

	incidents := []Incident{
		{
			ID:           "inc_001",
			SequentialID: "INC-1",
			Title:        "Test Incident",
			Status:       "in_progress",
			Severity:     "critical",
			Services:     []string{"api", "web"},
		},
		{
			ID:           "inc_002",
			SequentialID: "INC-2",
			Title:        "Another Incident",
			Status:       "resolved",
			Severity:     "low",
		},
	}

	cache.Set("incidents:pageSize=50", incidents)

	var result []Incident
	if !cache.GetTyped("incidents:pageSize=50", &result) {
		t.Fatal("expected GetTyped to return true for incidents")
	}

	if len(result) != 2 {
		t.Errorf("expected 2 incidents, got %d", len(result))
	}
	if result[0].Title != "Test Incident" {
		t.Errorf("expected 'Test Incident', got '%s'", result[0].Title)
	}
	if len(result[0].Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(result[0].Services))
	}
}
