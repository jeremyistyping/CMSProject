package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AllowedImageExtensions contains allowed image file extensions
var AllowedImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

// MaxFileSize is the maximum allowed file size (10MB)
const MaxFileSize = 10 << 20 // 10 MB

// UploadConfig holds configuration for file uploads
type UploadConfig struct {
	UploadDir          string
	AllowedExtensions  []string
	MaxFileSize        int64
}

// DefaultUploadConfig returns default upload configuration
func DefaultUploadConfig() *UploadConfig {
	return &UploadConfig{
		UploadDir:         "./uploads/daily-updates",
		AllowedExtensions: AllowedImageExtensions,
		MaxFileSize:       MaxFileSize,
	}
}

// SaveUploadedFile saves an uploaded file and returns the file path
func SaveUploadedFile(file *multipart.FileHeader, config *UploadConfig) (string, error) {
	// Validate file size
	if file.Size > config.MaxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", config.MaxFileSize)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedExtension(ext, config.AllowedExtensions) {
		return "", fmt.Errorf("file extension %s is not allowed", ext)
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	timestamp := time.Now().Format("20060102-150405")
	uniqueID := uuid.New().String()[:8]
	newFilename := fmt.Sprintf("%s-%s%s", timestamp, uniqueID, ext)
	filePath := filepath.Join(config.UploadDir, newFilename)

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return relative path (for storing in database)
	return filePath, nil
}

// SaveMultipleFiles saves multiple uploaded files
func SaveMultipleFiles(files []*multipart.FileHeader, config *UploadConfig) ([]string, error) {
	var filePaths []string
	
	for _, file := range files {
		filePath, err := SaveUploadedFile(file, config)
		if err != nil {
			// If any file fails, we should still return the successfully uploaded files
			// You might want to implement cleanup logic here
			return filePaths, fmt.Errorf("failed to upload file %s: %w", file.Filename, err)
		}
		filePaths = append(filePaths, filePath)
	}
	
	return filePaths, nil
}

// DeleteFile deletes a file from the filesystem
func DeleteFile(filePath string) error {
	if filePath == "" {
		return nil
	}
	
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	return nil
}

// DeleteMultipleFiles deletes multiple files
func DeleteMultipleFiles(filePaths []string) error {
	for _, filePath := range filePaths {
		if err := DeleteFile(filePath); err != nil {
			return err
		}
	}
	return nil
}

// isAllowedExtension checks if the file extension is allowed
func isAllowedExtension(ext string, allowedExtensions []string) bool {
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

// GetPublicURL converts a file path to a public URL
func GetPublicURL(filePath string) string {
	// Remove the ./uploads prefix and replace with /uploads for web access
	publicPath := strings.Replace(filePath, "./uploads", "/uploads", 1)
	return publicPath
}

