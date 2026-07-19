package service

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CacheWarmerService interface {
	WarmActiveInvitations(concurrency int) (total int, success int, fail int, err error)
	CollectActiveURLs() ([]string, error)
}

type cacheWarmerService struct {
	db      *gorm.DB
	baseURL string
	logger  *zap.Logger
}

func NewCacheWarmerService(db *gorm.DB, baseURL string, logger *zap.Logger) CacheWarmerService {
	// Ensure baseURL doesn't have a trailing slash
	baseURL = strings.TrimRight(baseURL, "/")
	return &cacheWarmerService{
		db:      db,
		baseURL: baseURL,
		logger:  logger,
	}
}

func (s *cacheWarmerService) WarmActiveInvitations(concurrency int) (int, int, int, error) {
	urls, err := s.CollectActiveURLs()
	if err != nil {
		return 0, 0, 0, err
	}

	if len(urls) == 0 {
		return 0, 0, 0, nil
	}

	if concurrency <= 0 {
		concurrency = 5
	}

	s.logger.Info("Starting cache warm up", zap.Int("total_urls", len(urls)), zap.Int("concurrency", concurrency))

	var success, fail int
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, concurrency)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, u := range urls {
		wg.Add(1)
		sem <- struct{}{} // acquire

		go func(urlPath string) {
			defer wg.Done()
			defer func() { <-sem }() // release

			fullURL := s.baseURL + urlPath
			if !strings.HasPrefix(urlPath, "/") {
				fullURL = s.baseURL + "/" + urlPath
			}

			req, err := http.NewRequest(http.MethodGet, fullURL, nil)
			if err != nil {
				s.logger.Warn("Failed to create request for cache warm", zap.String("url", fullURL), zap.Error(err))
				mu.Lock()
				fail++
				mu.Unlock()
				return
			}

			// Add a custom user agent to distinguish this traffic
			req.Header.Set("User-Agent", "Momenu-Cache-Warmer/1.0")

			resp, err := client.Do(req)
			if err != nil {
				s.logger.Warn("Failed to warm URL", zap.String("url", fullURL), zap.Error(err))
				mu.Lock()
				fail++
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			// Discard the body to keep memory footprint low
			_, _ = io.Copy(io.Discard, resp.Body)

			mu.Lock()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				success++
			} else {
				fail++
				s.logger.Warn("Non-2xx response from URL", zap.String("url", fullURL), zap.Int("status", resp.StatusCode))
			}
			mu.Unlock()
		}(u)
	}

	wg.Wait()
	return len(urls), success, fail, nil
}

func (s *cacheWarmerService) CollectActiveURLs() ([]string, error) {
	now := time.Now()
	pastLimit := now.Add(-7 * 24 * time.Hour)
	futureLimit := now.Add(30 * 24 * time.Hour)

	var projectIDs []uuid.UUID

	// Step 1: Active project IDs
	err := s.db.Model(&models.Schedule{}).
		Select("DISTINCT schedules.project_id").
		Joins("JOIN projects ON projects.id = schedules.project_id AND projects.status = ? AND projects.deleted_at IS NULL", models.ProjectStatusPublished).
		Where("schedules.start_time <= ? AND COALESCE(schedules.end_time, schedules.start_time) >= ?", futureLimit, pastLimit).
		Pluck("project_id", &projectIDs).Error

	if err != nil {
		s.logger.Error("Failed to fetch active project IDs for cache warmer", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch project IDs: %w", err)
	}

	if len(projectIDs) == 0 {
		return []string{}, nil
	}

	urlSet := make(map[string]struct{})

	// 2a. media_mappings
	var mediaUrls []string
	s.db.Model(&models.MediaMapping{}).
		Where("project_id IN ?", projectIDs).
		Pluck("url", &mediaUrls)
	for _, u := range mediaUrls {
		if u != "" {
			urlSet[u] = struct{}{}
		}
	}

	// 2b. sharing_thumbnail
	var thumbUrls []string
	s.db.Model(&models.Project{}).
		Where("id IN ? AND sharing_thumbnail != ''", projectIDs).
		Pluck("sharing_thumbnail", &thumbUrls)
	for _, u := range thumbUrls {
		if u != "" {
			urlSet[u] = struct{}{}
		}
	}

	// 2c. gift_registries qr_code_image
	var qrUrls []string
	s.db.Model(&models.GiftRegistry{}).
		Where("project_id IN ? AND qr_code_image != ''", projectIDs).
		Pluck("qr_code_image", &qrUrls)
	for _, u := range qrUrls {
		if u != "" {
			urlSet[u] = struct{}{}
		}
	}

	var urls []string
	for u := range urlSet {
		urls = append(urls, u)
	}

	return urls, nil
}
