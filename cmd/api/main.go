package main

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhanuprys/momenu-backend-fiber/internal/config"
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/routes"
	"github.com/dhanuprys/momenu-backend-fiber/internal/seeder"
	"github.com/dhanuprys/momenu-backend-fiber/internal/worker"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/database"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	fiblogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	// ── Bootstrap ────────────────────────────────────────────────
	config.Load()
	logger.InitLogger(config.AppConfig.Env)
	database.Connect()

	// ── Database Migrations ─────────────────────────────────────
	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Theme{},
		&models.Project{},
		&models.FeatureToggle{},
		&models.Schedule{},
		&models.GiftRegistry{},
		&models.MediaMapping{},
		&models.DressCode{},
		&models.LiveStream{},
		&models.RSVP{},
		&models.Guestbook{},
		&models.MusicCategory{},
		&models.Music{},
		&models.ProjectVisit{},
		&models.TextOverride{},
		&models.StyleOverride{},
		&models.FileRecord{},
		&models.ProjectShareSession{},
	); err != nil {
		logger.Fatal("Failed to run database migrations: " + err.Error())
	}

	// ── Functional & Partial Indexes ────────────────────────────
	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_rsvp_project_name_lower ON rsvps (project_id, LOWER(name));").Error; err != nil {
		logger.Fatal("Failed to create idx_rsvp_project_name_lower: " + err.Error())
	}
	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_file_unoptimized ON file_records (created_at) WHERE is_optimized = false AND media_type = 'image';").Error; err != nil {
		logger.Fatal("Failed to create idx_file_unoptimized: " + err.Error())
	}

	logger.Info("Database migration completed successfully")

	// ── Seeders ─────────────────────────────────────────────────
	seeder.SyncThemes(database.DB)
	seeder.SyncMusicCategories(database.DB)
	seeder.SyncMusics(database.DB)

	// ── Background Workers ──────────────────────────────────────
	worker.InitWorker()

	// ── Fiber App ───────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:           "Momenu Backend API",
		BodyLimit:         50 * 1024 * 1024, // 50 MB
		StreamRequestBody: true,
	})

	// ── Global Middleware ───────────────────────────────────────
	app.Use(helmet.New(helmet.Config{
		CrossOriginResourcePolicy: "cross-origin",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // Adjust in production
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "authorization", "x-requested-with"},
		AllowMethods: []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
	}))
	app.Use(limiter.New(limiter.Config{
		Max: 500, // max requests per minute
	}))
	app.Use(fiblogger.New())
	app.Use(recover.New())

	// ── Static Files ────────────────────────────────────────────
	setupStaticFiles(app)

	// ── API Routes ──────────────────────────────────────────────
	routes.SetupRoutes(app)

	// ── Start Server ────────────────────────────────────────────
	port := config.AppConfig.ServerPort
	logger.Info("Starting server on port " + port)
	if err := app.Listen(":" + port); err != nil {
		logger.Fatal("Error starting server: " + err.Error())
	}
}

// setupStaticFiles registers the /uploads and /static file-serving routes
// with Cross-Origin-Resource-Policy headers and directory-traversal protection.
func setupStaticFiles(app *fiber.App) {
	// Ensure directories exist
	for _, dir := range []string{"./uploads", "./static/music"} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.Fatal("Failed to create directory " + dir + ": " + err.Error())
		}
	}

	// Serve uploaded files
	app.Get("/uploads/*", func(c fiber.Ctx) error {
		return serveStaticFile(c, "./uploads")
	})

	// Serve static assets (music, etc.)
	app.Get("/static/*", func(c fiber.Ctx) error {
		return serveStaticFile(c, "./static")
	})
}

// serveStaticFile is a shared helper for serving files from a base directory
// with CORP headers and directory-traversal protection.
func serveStaticFile(c fiber.Ctx, baseDir string) error {
	requested := c.Params("*")
	c.Set("Cross-Origin-Resource-Policy", "cross-origin")

	// Prevent directory traversal
	cleaned := filepath.Clean(requested)
	if strings.Contains(cleaned, "..") {
		return c.SendStatus(fiber.StatusForbidden)
	}

	filePath := filepath.Join(baseDir, cleaned)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	// Detect content type
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	c.Set("Content-Type", contentType)
	return c.Send(data)
}
