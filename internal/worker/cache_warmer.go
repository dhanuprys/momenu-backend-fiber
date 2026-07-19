package worker

import (
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/config"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var cacheWarmerTicker *time.Ticker

// InitCacheWarmer starts the background cache warmer worker.
func InitCacheWarmer(db *gorm.DB, logger *zap.Logger) {
	baseURL := config.AppConfig.CacheWarmerBaseURL
	if baseURL == "" {
		logger.Warn("CACHE_WARMER_BASE_URL not set, cache warmer worker disabled")
		return
	}

	svc := service.NewCacheWarmerService(db, baseURL, logger)
	cacheWarmerTicker = time.NewTicker(2 * time.Hour)

	go func() {
		// Run once immediately on startup
		runCacheWarmer(svc, logger)
		for range cacheWarmerTicker.C {
			runCacheWarmer(svc, logger)
		}
	}()

	logger.Info("Cache warmer worker started (every 2 hours)")
}

// StopCacheWarmer stops the cache warmer worker.
func StopCacheWarmer() {
	if cacheWarmerTicker != nil {
		cacheWarmerTicker.Stop()
	}
}

func runCacheWarmer(svc service.CacheWarmerService, logger *zap.Logger) {
	total, ok, fail, err := svc.WarmActiveInvitations(5)
	if err != nil {
		logger.Error("Cache warmer failed", zap.Error(err))
		return
	}
	logger.Info("Cache warmer completed",
		zap.Int("total", total), zap.Int("ok", ok), zap.Int("fail", fail))
}
