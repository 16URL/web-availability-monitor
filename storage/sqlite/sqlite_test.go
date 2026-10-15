package sqlite

import (
	"path/filepath"
	"testing"
	"time"
)

func TestGetURL(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.db")
	storage, err := New(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	id, err := storage.SaveURL("example", "https://example.com")
	if err != nil {
		t.Fatal("failed to save URL:", err)
	}

	suite, err := storage.GetURL("example")
	if err != nil {
		t.Fatal("failed to get URL:", err)
	}
	if suite.ID != id {
		t.Fatalf("expected URL ID %d, got %d", id, suite.ID)
	}
	if suite.Alias != "example" {
		t.Fatalf("expected URL alias 'example', got '%s'", suite.Alias)
	}
	if suite.URL != "https://example.com" {
		t.Fatalf("expected URL 'https://example.com', got '%s'", suite.URL)
	}
}

func TestGetChecks(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.db")
	storage, err := New(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	urlID, err := storage.SaveURL("example", "https://example.com")
	if err != nil {
		t.Fatal("failed to save URL:", err)
	}

	_, err = storage.SaveCheck(urlID, 200, 123, time.Now())
	if err != nil {
		t.Fatal("failed to save check:", err)
	}

	checks, err := storage.GetChecks(urlID)
	if err != nil {
		t.Fatal("failed to get checks:", err)
	}
	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	if checks[0].URLID != urlID {
		t.Fatalf("expected check URL ID	%d, got %d", urlID, checks[0].URLID)
	}
	if checks[0].HTTPStatus != 200 {
		t.Fatalf("expected check HTTP status 200, got %d", checks[0].HTTPStatus)
	}
	if checks[0].ResponseTimeMs != 123 {
		t.Fatalf("expected check response time 123, got %d", checks[0].ResponseTimeMs)
	}
}

func TestGetChecksNoChecks(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.db")
	storage, err := New(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	urlID, err := storage.SaveURL("example", "https://example.com")
	if err != nil {
		t.Fatal("failed to save URL:", err)
	}

	checks, err := storage.GetChecks(urlID)
	if err != nil {
		t.Fatal("failed to get checks:", err)
	}
	if len(checks) != 0 {
		t.Fatalf("expected 0 checks, got %d", len(checks))
	}
}
func TestGetURLNotFound(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.db")
	storage, err := New(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	_, err = storage.GetURL("nonexistent")
	if err == nil {
		t.Fatal("expected error when getting nonexistent URL, got nil")
	}
}
