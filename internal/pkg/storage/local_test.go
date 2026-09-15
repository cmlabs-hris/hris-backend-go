package storage

import (
	"context"
	"strings"
	"testing"
)

func TestLocalStorageUploadReturnsForwardSlashPath(t *testing.T) {
	s, err := NewLocalStorage(t.TempDir(), "http://localhost:8080/uploads")
	if err != nil {
		t.Fatalf("new local storage: %v", err)
	}

	path, err := s.Upload(context.Background(), strings.NewReader("logo"), "logos/acme/acme.png", "image/png")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if want := "logos/acme/acme.png"; path != want {
		t.Errorf("stored path = %q, want %q", path, want)
	}
	if strings.ContainsRune(path, '\\') {
		t.Errorf("stored path should not contain backslashes: %q", path)
	}
}

func TestLocalStorageGetURLFromRelativePath(t *testing.T) {
	s, err := NewLocalStorage(t.TempDir(), "http://localhost:8080/uploads")
	if err != nil {
		t.Fatalf("new local storage: %v", err)
	}

	got, err := s.GetURL(context.Background(), "logos\\acme\\acme.png", 0)
	if err != nil {
		t.Fatalf("get url: %v", err)
	}
	want := "http://localhost:8080/uploads/logos/acme/acme.png"
	if got != want {
		t.Errorf("GetURL(relative) = %q, want %q", got, want)
	}
}

func TestLocalStorageGetURLOnAbsolutePathDoesNotDoublePrefix(t *testing.T) {
	s, err := NewLocalStorage(t.TempDir(), "http://localhost:8080/uploads")
	if err != nil {
		t.Fatalf("new local storage: %v", err)
	}

	// Legacy row that stored the full URL using Windows separators.
	got, err := s.GetURL(context.Background(), "http://localhost:8080/uploads\\logos\\acme\\acme.png", 0)
	if err != nil {
		t.Fatalf("get url: %v", err)
	}
	want := "http://localhost:8080/uploads/logos/acme/acme.png"
	if got != want {
		t.Errorf("GetURL(absolute) = %q, want %q", got, want)
	}
	if strings.Count(got, "/uploads/") != 1 {
		t.Errorf("URL has unexpected /uploads/ segments: %q", got)
	}
}

func TestLocalStorageGetURLOnEmptyPath(t *testing.T) {
	s, err := NewLocalStorage(t.TempDir(), "http://localhost:8080/uploads")
	if err != nil {
		t.Fatalf("new local storage: %v", err)
	}
	got, err := s.GetURL(context.Background(), "", 0)
	if err != nil {
		t.Fatalf("get url: %v", err)
	}
	if got != "" {
		t.Errorf("GetURL(empty) = %q, want empty", got)
	}
}