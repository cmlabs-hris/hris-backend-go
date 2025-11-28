package file

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // Import for PNG decoding support
	"io"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/storage"
	"github.com/google/uuid"
	"golang.org/x/image/draw"
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

	// UploadCompanyLogo uploads a company logo
	UploadCompanyLogo(ctx context.Context, companyUsername string, file io.Reader, filename string) (string, error)

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
// Compresses image to target size between 50KB - 150KB
func (s *fileServiceImpl) UploadAttendanceProof(ctx context.Context, employeeID string, date time.Time, file io.Reader, filename string, clockType string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	// Validate image format
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", fmt.Errorf("invalid file type: only jpg, jpeg, png allowed")
	}

	// Read the entire file into memory
	buffer, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	// Compress image to target size (50KB - 150KB)
	compressed, err := compressImage(buffer, 150*1024, 50*1024)
	if err != nil {
		return "", fmt.Errorf("failed to compress image: %w", err)
	}

	// Generate path: attendance/{date}/{employeeID}-{clockType}-{timestamp}.jpg
	// Always output as JPEG after compression for consistency
	dateStr := date.Format("2006-01-02")
	timestamp := time.Now().Unix()
	newFilename := fmt.Sprintf("%s-%s-%d.jpg", employeeID, clockType, timestamp)
	path := filepath.Join("attendance", dateStr, newFilename)

	// Upload compressed image
	uploadedPath, err := s.storage.Upload(ctx, bytes.NewReader(compressed), path, "image/jpeg")
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

// UploadCompanyLogo uploads a company logo
func (s *fileServiceImpl) UploadCompanyLogo(ctx context.Context, companyUsername string, file io.Reader, filename string) (string, error) {
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
	newFilename := fmt.Sprintf("%s-%s%s", companyUsername, uniqueID, ext)
	path := filepath.Join("logos", companyUsername, newFilename)

	// Determine content type
	contentType := "image/jpeg"
	if ext == ".png" {
		contentType = "image/png"
	}

	// Upload
	uploadedPath, err := s.storage.Upload(ctx, file, path, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload company logo: %w", err)
	}

	return uploadedPath, nil
}

// ==================== HELPER FUNCTIONS ====================

// compressImage compresses an image to target size range using Go standard library
// maxSize: maximum allowed size (e.g., 150KB)
// minSize: minimum target size (e.g., 50KB)
func compressImage(buffer []byte, maxSize int, minSize int) ([]byte, error) {
	// Check if compression is needed
	if len(buffer) <= maxSize && len(buffer) >= minSize {
		return buffer, nil
	}

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(buffer))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get original dimensions
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Start with quality 85 and reduce progressively
	quality := 85
	var compressed []byte
	currentImg := img

	// Try compression with decreasing quality first
	for quality >= 50 {
		buf := new(bytes.Buffer)
		err = jpeg.Encode(buf, currentImg, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, fmt.Errorf("failed to encode JPEG: %w", err)
		}

		compressed = buf.Bytes()

		// Check if we've reached target size
		if len(compressed) <= maxSize && len(compressed) >= minSize {
			return compressed, nil
		}

		// If still too large, reduce quality
		if len(compressed) > maxSize {
			quality -= 5
			continue
		}

		// If too small but quality already low, accept it
		if len(compressed) < minSize && quality <= 60 {
			return compressed, nil
		}

		break
	}

	// If still too large after quality reduction, try resizing
	if len(compressed) > maxSize {
		// Calculate resize ratio to target 100KB (middle of range)
		targetSize := 100 * 1024
		ratio := math.Sqrt(float64(targetSize) / float64(len(compressed)))
		newWidth := int(float64(originalWidth) * ratio)
		newHeight := int(float64(originalHeight) * ratio)

		// Ensure minimum dimensions
		if newWidth < 600 {
			newWidth = 600
		}
		if newHeight < 400 {
			newHeight = 400
		}

		// Resize the image
		resized := resizeImage(img, newWidth, newHeight)

		// Encode with quality 70
		buf := new(bytes.Buffer)
		err = jpeg.Encode(buf, resized, &jpeg.Options{Quality: 70})
		if err != nil {
			return nil, fmt.Errorf("failed to encode resized image: %w", err)
		}

		compressed = buf.Bytes()
	}

	// Log format info (optional, for debugging)
	_ = format // PNG will be converted to JPEG

	return compressed, nil
}

// resizeImage resizes an image to the specified dimensions using high-quality interpolation
func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	// Use CatmullRom for high-quality downscaling
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}
