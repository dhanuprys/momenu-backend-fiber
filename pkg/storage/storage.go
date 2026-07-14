package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxImageSize = 10 * 1024 * 1024 // 10 MB
	MaxVideoSize = 50 * 1024 * 1024 // 50 MB
)

var (
	ErrInvalidFileType = errors.New("invalid file type")
	ErrFileTooLarge    = errors.New("file is too large")
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

var allowedVideoTypes = map[string]bool{
	"video/mp4":  true,
	"video/webm": true,
}

func SaveFile(file *multipart.FileHeader, subdir string, mediaType string) (string, error) {
	// 1. Validate file size based on type
	if mediaType == "image" && file.Size > MaxImageSize {
		return "", ErrFileTooLarge
	} else if mediaType == "video" && file.Size > MaxVideoSize {
		return "", ErrFileTooLarge
	}

	// 2. Validate MIME type
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	
	// Reset file pointer
	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer)
	
	if mediaType == "image" && !allowedImageTypes[contentType] {
		return "", ErrInvalidFileType
	} else if mediaType == "video" && !allowedVideoTypes[contentType] {
		return "", ErrInvalidFileType
	}

	// 3. Generate unique filename
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		// Fallback based on content type
		if contentType == "image/jpeg" {
			ext = ".jpg"
		} else if contentType == "image/png" {
			ext = ".png"
		} else if contentType == "image/webp" {
			ext = ".webp"
		} else if contentType == "image/gif" {
			ext = ".gif"
		} else if contentType == "video/mp4" {
			ext = ".mp4"
		} else if contentType == "video/webm" {
			ext = ".webm"
		}
	}
	
	shortID := strings.Split(uuid.New().String(), "-")[0]
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), shortID, ext)

	// 4. Create directory if not exists
	uploadDir := filepath.Join(".", "uploads", subdir)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", err
	}

	// 5. Save file
	dstPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	// 6. Return public URL
	publicURL := fmt.Sprintf("/uploads/%s/%s", subdir, filename)
	return publicURL, nil
}

func DeleteFile(fileURL string) error {
	// Only delete files from local storage (starts with /uploads/)
	if !strings.HasPrefix(fileURL, "/uploads/") {
		return nil
	}

	// Remove leading slash to get relative path
	filePath := filepath.Join(".", strings.TrimPrefix(fileURL, "/"))
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}
	
	return os.Remove(filePath)
}
