package file

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/storage"
	"github.com/google/uuid"
)

type FileService interface {
	// Avatar uploads
	UploadAvatar(ctx context.Context, employeeID string, file io.Reader, filename string) (string, error)

	// Document uploads
	UploadDocument(ctx context.Context, employeeID string, file io.Reader, filename string, documentType string) (string, error)

	// Attendance proof uploads
	UploadAttendanceProof(ctx context.Context, employeeID string, date time.Time, file io.Reader, filename string, clockType string) (string, error)

	// Leave attachment uploads
	UploadLeaveAttachment(ctx context.Context, employeeID string, file io.Reader, filename string) (string, error)

	// Generic operations
	DeleteFile(ctx context.Context, path string) error
	GetFileURL(ctx context.Context, path string, expiry time.Duration) (string, error)
}

type fileServiceImpl struct {
	storage storage.FileStorage
}

func NewFileService(storage storage.FileStorage) FileService {
	return &fileServiceImpl{
		storage: storage,
	}
}

// UploadAvatar uploads employee avatar
func (s *fileServiceImpl) UploadAvatar(ctx context.Context, employeeID string, file io.Reader, filename string) (string, error) {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".jpg", ".jpeg", ".png"}

	isValid := false
	for _, allowed := range allowedExts {
		if ext == allowed {
			isValid = true
			break
		}
	}

	if !isValid {
		return "", fmt.Errorf("invalid file type: only jpg, jpeg, png allowed")
	}

	// Generate unique filename
	uniqueID := uuid.New().String()
	newFilename := fmt.Sprintf("%s-%s%s", employeeID, uniqueID, ext)
	path := filepath.Join("avatars", employeeID, newFilename)

	// Determine content type
	contentType := "image/jpeg"
	if ext == ".png" {
		contentType = "image/png"
	}

	// Upload
	uploadedPath, err := s.storage.Upload(ctx, file, path, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload avatar: %w", err)
	}

	return uploadedPath, nil
}

// UploadDocument uploads employee document
func (s *fileServiceImpl) UploadDocument(ctx context.Context, employeeID string, file io.Reader, filename string, documentType string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	// Generate unique filename
	uniqueID := uuid.New().String()
	newFilename := fmt.Sprintf("%s-%s%s", documentType, uniqueID, ext)
	path := filepath.Join("documents", employeeID, newFilename)

	// Upload with generic content type
	uploadedPath, err := s.storage.Upload(ctx, file, path, "application/octet-stream")
	if err != nil {
		return "", fmt.Errorf("failed to upload document: %w", err)
	}

	return uploadedPath, nil
}

// UploadAttendanceProof uploads attendance clock-in/out proof photo
func (s *fileServiceImpl) UploadAttendanceProof(ctx context.Context, employeeID string, date time.Time, file io.Reader, filename string, clockType string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	// Validate image format
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", fmt.Errorf("invalid file type: only jpg, jpeg, png allowed")
	}

	// Generate path: attendance/{date}/{employeeID}-{clockType}-{timestamp}.jpg
	dateStr := date.Format("2006-01-02")
	timestamp := time.Now().Unix()
	newFilename := fmt.Sprintf("%s-%s-%d%s", employeeID, clockType, timestamp, ext)
	path := filepath.Join("attendance", dateStr, newFilename)

	contentType := "image/jpeg"
	if ext == ".png" {
		contentType = "image/png"
	}

	uploadedPath, err := s.storage.Upload(ctx, file, path, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload attendance proof: %w", err)
	}

	return uploadedPath, nil
}

// UploadLeaveAttachment uploads leave request attachment
func (s *fileServiceImpl) UploadLeaveAttachment(ctx context.Context, employeeID string, file io.Reader, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	// Generate unique filename with timestamp
	uniqueID := uuid.New().String()
	timestamp := time.Now().Unix()
	newFilename := fmt.Sprintf("%s-%d%s", uniqueID, timestamp, ext)
	path := filepath.Join("leave", employeeID, newFilename)

	uploadedPath, err := s.storage.Upload(ctx, file, path, "application/octet-stream")
	if err != nil {
		return "", fmt.Errorf("failed to upload leave attachment: %w", err)
	}

	return uploadedPath, nil
}

// DeleteFile deletes a file
func (s *fileServiceImpl) DeleteFile(ctx context.Context, path string) error {
	return s.storage.Delete(ctx, path)
}

// GetFileURL generates URL to access file
func (s *fileServiceImpl) GetFileURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	return s.storage.GetURL(ctx, path, expiry)
}
