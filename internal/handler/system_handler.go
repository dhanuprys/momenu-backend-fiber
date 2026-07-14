package handler

import (
	"runtime"
	"syscall"
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type SystemHandler struct {
	startTime time.Time
}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		startTime: time.Now(),
	}
}

type SystemResources struct {
	UptimeSeconds    float64 `json:"uptime_seconds"`
	Goroutines       int     `json:"goroutines"`
	MemoryAllocMB    float64 `json:"memory_alloc_mb"`
	MemoryTotalMB    float64 `json:"memory_total_mb"`
	MemorySysMB      float64 `json:"memory_sys_mb"`
	NumCPU           int     `json:"num_cpu"`
	GoVersion        string  `json:"go_version"`
	OS               string  `json:"os"`
	Arch             string  `json:"arch"`
	DiskTotalGB      float64 `json:"disk_total_gb"`
	DiskFreeGB       float64 `json:"disk_free_gb"`
	DiskUsedGB       float64 `json:"disk_used_gb"`
	DiskUsagePercent float64 `json:"disk_usage_percent"`
}

func (h *SystemHandler) GetResources(c fiber.Ctx) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	
	var diskTotal, diskFree, diskUsed, diskPercent float64
	if err == nil {
		diskTotal = float64(stat.Blocks) * float64(stat.Bsize) / 1024 / 1024 / 1024
		diskFree = float64(stat.Bavail) * float64(stat.Bsize) / 1024 / 1024 / 1024
		diskUsed = diskTotal - diskFree
		if diskTotal > 0 {
			diskPercent = (diskUsed / diskTotal) * 100
		}
	}

	res := SystemResources{
		UptimeSeconds:    time.Since(h.startTime).Seconds(),
		Goroutines:       runtime.NumGoroutine(),
		MemoryAllocMB:    float64(m.Alloc) / 1024 / 1024,
		MemoryTotalMB:    float64(m.TotalAlloc) / 1024 / 1024,
		MemorySysMB:      float64(m.Sys) / 1024 / 1024,
		NumCPU:           runtime.NumCPU(),
		GoVersion:        runtime.Version(),
		OS:               runtime.GOOS,
		Arch:             runtime.GOARCH,
		DiskTotalGB:      diskTotal,
		DiskFreeGB:       diskFree,
		DiskUsedGB:       diskUsed,
		DiskUsagePercent: diskPercent,
	}

	return response.JSONSuccess(c, fiber.StatusOK, "System resources retrieved successfully", res, nil)
}
