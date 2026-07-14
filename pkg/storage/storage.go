package storage

import (
	"bytes"
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
	ErrQuotaExceeded   = errors.New("quota exceeded")
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

type FileRecordInfo struct {
	URL          string
	FilePath     string
	OriginalName string
	ContentType  string
	Size         int64
	MediaType    string
}

type QuotaLimitReader struct {
	R         io.Reader
	Remaining int64
	ReadBytes int64
}

func (q *QuotaLimitReader) Read(p []byte) (n int, err error) {
	n, err = q.R.Read(p)
	q.ReadBytes += int64(n)
	if q.Remaining > 0 && q.ReadBytes > q.Remaining {
		return n, ErrQuotaExceeded
	}
	return n, err
}

func SaveFile(file *multipart.FileHeader, subdir string, mediaType string) (*FileRecordInfo, error) {
	// 1. Validate file size based on type
	if mediaType == "image" && file.Size > MaxImageSize {
		return nil, ErrFileTooLarge
	} else if mediaType == "video" && file.Size > MaxVideoSize {
		return nil, ErrFileTooLarge
	}

	// 2. Validate MIME type
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	
	// Reset file pointer
	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	contentType := http.DetectContentType(buffer)
	
	if mediaType == "image" && !allowedImageTypes[contentType] {
		return nil, ErrInvalidFileType
	} else if mediaType == "video" && !allowedVideoTypes[contentType] {
		return nil, ErrInvalidFileType
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
		return nil, err
	}

	// 5. Save file
	dstPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	// 6. Return public URL and metadata
	publicURL := fmt.Sprintf("/uploads/%s/%s", subdir, filename)
	
	return &FileRecordInfo{
		URL:          publicURL,
		FilePath:     dstPath,
		OriginalName: file.Filename,
		ContentType:  contentType,
		Size:         file.Size,
		MediaType:    mediaType,
	}, nil
}

func StreamFile(src io.Reader, originalFilename string, subdir string, mediaType string, maxAllowedSize int64) (*FileRecordInfo, error) {
	// 1. Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	n, err := io.ReadFull(src, buffer)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}

	contentType := http.DetectContentType(buffer[:n])
	
	if mediaType == "image" && !allowedImageTypes[contentType] {
		return nil, ErrInvalidFileType
	} else if mediaType == "video" && !allowedVideoTypes[contentType] {
		return nil, ErrInvalidFileType
	} else if mediaType == "audio" {
		// Just a basic check or allow any audio based on frontend
		if !strings.HasPrefix(contentType, "audio/") {
			// fallback check
		}
	}

	// 2. Generate unique filename
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" {
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

	// 3. Create directory if not exists
	uploadDir := filepath.Join(".", "uploads", subdir)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, err
	}

	// 4. Save file via MultiReader (buffer + rest of stream)
	dstPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	fullStream := io.MultiReader(bytes.NewReader(buffer[:n]), src)
	
	limitReader := &QuotaLimitReader{
		R:         fullStream,
		Remaining: maxAllowedSize,
	}

	written, err := io.Copy(dst, limitReader)
	if err != nil {
		os.Remove(dstPath)
		if err == ErrQuotaExceeded {
			return nil, ErrFileTooLarge
		}
		return nil, err
	}

	// 5. Return public URL and metadata
	publicURL := fmt.Sprintf("/uploads/%s/%s", subdir, filename)
	
	return &FileRecordInfo{
		URL:          publicURL,
		FilePath:     dstPath,
		OriginalName: originalFilename,
		ContentType:  contentType,
		Size:         written,
		MediaType:    mediaType,
	}, nil
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
