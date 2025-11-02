package storage

import (
	"context"
	"io"
	"time"
)

type FileStorage interface {
	// Upload uploads a file and returns the file path/key
	Upload(ctx context.Context, file io.Reader, path string, contentType string) (string, error)

	// Download retrieves a file
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes a file
	Delete(ctx context.Context, path string) error

	// GetURL generates a presigned/public URL
	GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)

	// Exists checks if file exists
	Exists(ctx context.Context, path string) (bool, error)
}

type UploadOptions struct {
	ContentType string
	MaxSize     int64
	AllowedExts []string
}
