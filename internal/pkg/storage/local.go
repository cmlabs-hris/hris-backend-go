package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalStorage struct {
	basePath string
	baseURL  string // e.g., "http://localhost:8080/uploads"
}

func NewLocalStorage(basePath, baseURL string) (*LocalStorage, error) {
	// Create base directory if not exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}, nil
}

func (s *LocalStorage) Upload(ctx context.Context, file io.Reader, path string, contentType string) (string, error) {
	// Sanitize path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(s.basePath, cleanPath)

	// Ensure file is within basePath
	if !strings.HasPrefix(fullPath, s.basePath) {
		return "", fmt.Errorf("invalid file path: %s", path)
	}

	// Create directory structure
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, file); err != nil {
		// Cleanup on error
		os.Remove(fullPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return cleanPath, nil
}

func (s *LocalStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(s.basePath, cleanPath)

	// Security check
	if !strings.HasPrefix(fullPath, s.basePath) {
		return nil, fmt.Errorf("invalid file path: %s", path)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(s.basePath, cleanPath)

	// Security check
	if !filepath.HasPrefix(fullPath, s.basePath) {
		return fmt.Errorf("invalid file path: %s", path)
	}

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *LocalStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	// For local storage, return static URL
	// In production with auth, you might generate signed tokens
	cleanPath := filepath.Clean(path)
	return fmt.Sprintf("%s/%s", s.baseURL, cleanPath), nil
}

func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(s.basePath, cleanPath)

	// Security check
	if !filepath.HasPrefix(fullPath, s.basePath) {
		return false, fmt.Errorf("invalid file path: %s", path)
	}

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
