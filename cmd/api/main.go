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
	// Initialize Configuration
	config.Load()

	// Initialize Logger
	logger.InitLogger(config.AppConfig.Env)

	// Initialize Database Connection
	database.Connect()

	// Auto-migrate models
	err := database.DB.AutoMigrate(
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
	)
	if err != nil {
		logger.Fatal("Failed to run database migrations: " + err.Error())
	}
	logger.Info("Database migration completed successfully")

	// Synchronize Theme Registry
	seeder.SyncThemes(database.DB)

	// Synchronize Music Registry
	seeder.SyncMusicCategories(database.DB)
	seeder.SyncMusics(database.DB)

	// Initialize Workers
	worker.InitWorker()

	// Initialize Fiber App
	app := fiber.New(fiber.Config{
		AppName:   "Momenu Backend API",
		BodyLimit: 50 * 1024 * 1024, // 50 MB
	})

	// Global Middlewares
	app.Use(helmet.New(helmet.Config{
		CrossOriginResourcePolicy: "cross-origin",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // Adjust in production
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "authorization", "x-requested-with"},
		AllowMethods: []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
	}))
	app.Use(limiter.New(limiter.Config{
		Max: 100, // max requests per minute
	}))
	app.Use(fiblogger.New())
	app.Use(recover.New())

	// Ensure uploads directory exists
	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		logger.Fatal("Failed to create uploads directory: " + err.Error())
	}

	// Static Files - custom handler to ensure CORP headers are set
	app.Get("/uploads/*", func(c fiber.Ctx) error {
		requested := c.Params("*")
		logger.Info("Incoming request for upload: " + requested)

		// Set CORP header immediately for all responses (even 404s)
		c.Set("Cross-Origin-Resource-Policy", "cross-origin")

		// Prevent directory traversal
		cleaned := filepath.Clean(requested)
		if strings.Contains(cleaned, "..") {
			return c.SendStatus(fiber.StatusForbidden)
		}

		filePath := filepath.Join("./uploads", cleaned)

		data, err := os.ReadFile(filePath)
		if err != nil {
			logger.Error("File not found: " + filePath)
			return c.SendStatus(fiber.StatusNotFound)
		}

		// Detect content type
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = http.DetectContentType(data)
		}

		c.Set("Content-Type", contentType)
		c.Set("Cross-Origin-Resource-Policy", "cross-origin")
		return c.Send(data)
	})

	// Ensure static music directory exists
	if err := os.MkdirAll("./static/music", os.ModePerm); err != nil {
		logger.Fatal("Failed to create static music directory: " + err.Error())
	}

	app.Get("/static/*", func(c fiber.Ctx) error {
		requested := c.Params("*")
		c.Set("Cross-Origin-Resource-Policy", "cross-origin")
		
		cleaned := filepath.Clean(requested)
		if strings.Contains(cleaned, "..") {
			return c.SendStatus(fiber.StatusForbidden)
		}

		filePath := filepath.Join("./static", cleaned)
		data, err := os.ReadFile(filePath)
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}

		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = http.DetectContentType(data)
		}

		c.Set("Content-Type", contentType)
		return c.Send(data)
	})

	// Register Routes
	routes.SetupRoutes(app)

	// Start Server
	port := config.AppConfig.ServerPort
	logger.Info("Starting server on port " + port)
	if err := app.Listen(":" + port); err != nil {
		logger.Fatal("Error starting server: " + err.Error())
	}
}
