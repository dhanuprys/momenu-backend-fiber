package routes

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/handler"
	"github.com/dhanuprys/momenu-backend-fiber/internal/middleware"
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/database"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"time"
)

func SetupRoutes(app *fiber.App) {
	// Initialize Dependencies
	userRepo := repository.NewUserRepository(database.DB)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	themeRepo := repository.NewThemeRepository(database.DB)
	themeService := service.NewThemeService(themeRepo)
	themeHandler := handler.NewThemeHandler(themeService)

	projectRepo := repository.NewProjectRepository(database.DB)
	projectService := service.NewProjectService(projectRepo, themeRepo, userRepo)
	projectHandler := handler.NewProjectHandler(projectService)

	scheduleRepo := repository.NewScheduleRepository(database.DB)
	scheduleService := service.NewScheduleService(scheduleRepo)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)

	giftRegistryRepo := repository.NewGiftRegistryRepository(database.DB)
	giftRegistryService := service.NewGiftRegistryService(giftRegistryRepo)
	giftRegistryHandler := handler.NewGiftRegistryHandler(giftRegistryService)

	mediaRepo := repository.NewMediaRepository(database.DB)
	mediaService := service.NewMediaService(mediaRepo, themeRepo)
	mediaHandler := handler.NewMediaHandler(mediaService)

	dressCodeRepo := repository.NewDressCodeRepository(database.DB)
	dressCodeService := service.NewDressCodeService(dressCodeRepo)
	dressCodeHandler := handler.NewDressCodeHandler(dressCodeService)

	liveStreamRepo := repository.NewLiveStreamRepository(database.DB)
	liveStreamService := service.NewLiveStreamService(liveStreamRepo)
	liveStreamHandler := handler.NewLiveStreamHandler(liveStreamService)

	rsvpRepo := repository.NewRSVPRepository(database.DB)
	rsvpService := service.NewRSVPService(rsvpRepo, projectRepo)
	rsvpHandler := handler.NewRSVPHandler(rsvpService, projectService)

	guestbookRepo := repository.NewGuestbookRepository(database.DB)
	guestbookService := service.NewGuestbookService(guestbookRepo, projectRepo)
	guestbookHandler := handler.NewGuestbookHandler(guestbookService, projectService, rsvpService)

	invitationService := service.NewInvitationService(projectRepo)
	invitationHandler := handler.NewInvitationHandler(invitationService, rsvpService)
	adminHandler := handler.NewAdminHandler()

	textOverrideHandler := handler.NewTextOverrideHandler()
	styleOverrideHandler := handler.NewStyleOverrideHandler()

	musicRepo := repository.NewMusicRepository(database.DB)
	musicService := service.NewMusicService(musicRepo)
	musicHandler := handler.NewMusicHandler(musicService)

	uploadHandler := handler.NewUploadHandler(projectRepo)

	analyticsRepo := repository.NewAnalyticsRepository(database.DB)
	analyticsService := service.NewAnalyticsService(analyticsRepo, projectRepo)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)

	ownerMiddleware := middleware.ProjectOwner(projectRepo)

	// API Versioning Group
	api := app.Group("/api/v1")

	// Health Check
	api.Get("/health", func(c fiber.Ctx) error {
		return response.JSONSuccess(c, fiber.StatusOK, "Server is healthy", fiber.Map{
			"status": "up",
		}, nil)
	})

	// Auth & Identity Routes
	auth := api.Group("/auth")
	auth.Post("/register", userHandler.Register)
	auth.Post("/login", userHandler.Login)
	auth.Get("/google/login", userHandler.GoogleLoginURL)
	auth.Get("/google/callback", userHandler.GoogleCallback)
	auth.Post("/refresh", userHandler.RefreshToken)
	auth.Get("/me", middleware.AuthRequired, userHandler.Me)

	// File Upload (Authenticated)
	api.Post("/upload", middleware.AuthRequired, uploadHandler.Upload)

	// Event Type Schema
	api.Get("/event-types/:type/schema", func(c fiber.Ctx) error {
		eventType := models.EventType(c.Params("type"))
		schema := models.GetFieldSchema(eventType)
		if schema == nil {
			return response.JSONError(c, fiber.StatusNotFound, "Event type not found", "EVENT_TYPE_NOT_FOUND")
		}
		return response.JSONSuccess(c, fiber.StatusOK, "Field schema retrieved", schema, nil)
	})

	// Themes (Public)
	themes := api.Group("/themes")
	themes.Get("/", themeHandler.List)
	themes.Get("/:id", themeHandler.Get)

	// Music (Public)
	music := api.Group("/music")
	music.Get("/categories", musicHandler.ListCategories)
	music.Get("/", musicHandler.ListMusics)

	// Projects (Authenticated)
	projects := api.Group("/projects", middleware.AuthRequired)
	projects.Post("/", projectHandler.Create)
	projects.Get("/", projectHandler.List)

	// Single Project (Authenticated + Owner)
	project := projects.Group("/:projectId", ownerMiddleware, middleware.TrackProjectUpdate)
	project.Get("/", projectHandler.Get)
	project.Get("/version", projectHandler.GetVersion)
	project.Put("/", projectHandler.Update)
	project.Delete("/", projectHandler.Delete)
	project.Patch("/status", projectHandler.UpdateStatus)
	project.Get("/feature-toggle", projectHandler.GetFeatureToggle)
	project.Put("/feature-toggle", projectHandler.UpdateFeatureToggle)
	project.Get("/analytics", analyticsHandler.GetProjectAnalytics)

	// Sub-resource: Schedules
	project.Get("/schedules", scheduleHandler.List)
	project.Post("/schedules", scheduleHandler.Create)
	project.Put("/schedules/:scheduleId", scheduleHandler.Update)
	project.Delete("/schedules/:scheduleId", scheduleHandler.Delete)

	// Sub-resource: Gift Registries
	project.Get("/gift-registries", giftRegistryHandler.List)
	project.Post("/gift-registries", giftRegistryHandler.Create)
	project.Put("/gift-registries/:registryId", giftRegistryHandler.Update)
	project.Delete("/gift-registries/:registryId", giftRegistryHandler.Delete)

	// Sub-resource: Media Mappings
	project.Get("/media", mediaHandler.List)
	project.Post("/media", mediaHandler.Create)
	project.Put("/media/:mediaId", mediaHandler.Update)
	project.Delete("/media/:mediaId", mediaHandler.Delete)

	// Sub-resource: Dress Codes
	project.Get("/dress-codes", dressCodeHandler.List)
	project.Post("/dress-codes", dressCodeHandler.Create)
	project.Put("/dress-codes/:dressCodeId", dressCodeHandler.Update)
	project.Delete("/dress-codes/:dressCodeId", dressCodeHandler.Delete)

	// Sub-resource: Live Streams
	project.Get("/live-streams", liveStreamHandler.List)
	project.Post("/live-streams", liveStreamHandler.Create)
	project.Put("/live-streams/:streamId", liveStreamHandler.Update)
	project.Delete("/live-streams/:streamId", liveStreamHandler.Delete)

	// Sub-resource: Text Overrides
	project.Get("/text-overrides", textOverrideHandler.List)
	project.Put("/text-overrides", textOverrideHandler.Upsert)
	project.Delete("/text-overrides/:slotKey", textOverrideHandler.Delete)

	// Sub-resource: Style Overrides
	project.Get("/style-overrides", styleOverrideHandler.List)
	project.Put("/style-overrides", styleOverrideHandler.Upsert)
	project.Delete("/style-overrides/:slotKey", styleOverrideHandler.Delete)

	// Owner: RSVPs
	rsvpsGroup := project.Group("/rsvps")
	rsvpsGroup.Get("/export", rsvpHandler.ExportXLSX)
	rsvpsGroup.Post("/bulk", rsvpHandler.ImportXLSX)
	rsvpsGroup.Get("/stats", rsvpHandler.Stats)
	rsvpsGroup.Get("/", rsvpHandler.List)
	rsvpsGroup.Post("/", rsvpHandler.OwnerUpsert)
	rsvpsGroup.Delete("/:id", rsvpHandler.Delete)

	// Owner: Guestbook
	gbGroup := project.Group("/guestbook")
	gbGroup.Get("/", guestbookHandler.List)
	gbGroup.Delete("/:entryId", guestbookHandler.Delete)

	// Public Invitation Routes
	inviteLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
	})

	invite := api.Group("/invite/:slug")
	invite.Get("/og", invitationHandler.GetOGMetadata) // New public get OG metadata
	invite.Get("/", invitationHandler.GetInvitation)
	invite.Get("/guest", rsvpHandler.PublicGetGuest) // New public get guest
	invite.Post("/rsvp", inviteLimiter, rsvpHandler.PublicUpsert)
	invite.Get("/guestbook", guestbookHandler.PublicList)
	invite.Post("/guestbook", inviteLimiter, guestbookHandler.PublicCreate)

	// Public Invitation Routes (New Prefix)
	publicInvite := api.Group("/public/invite/:slug")
	publicInvite.Get("/og", invitationHandler.GetOGMetadata)
	publicInvite.Get("/", invitationHandler.GetInvitation)
	publicInvite.Get("/guest", rsvpHandler.PublicGetGuest)
	publicInvite.Post("/rsvp", inviteLimiter, rsvpHandler.PublicUpsert)
	publicInvite.Get("/guestbook", guestbookHandler.PublicList)
	publicInvite.Post("/guestbook", inviteLimiter, guestbookHandler.PublicCreate)

	// Public Analytics
	analyticsPublic := api.Group("/analytics")
	analyticsPublic.Post("/visit", inviteLimiter, analyticsHandler.RecordVisit)

	// Admin Routes
	admin := api.Group("/admin", middleware.AuthRequired, middleware.AdminRequired)
	admin.Get("/users", adminHandler.ListUsers)
	admin.Post("/users", adminHandler.CreateUser)
	admin.Put("/users/:id/status", adminHandler.UpdateUserStatus)
	admin.Get("/projects", adminHandler.ListProjects)
	admin.Delete("/projects/:id", adminHandler.DeleteProject)
	admin.Post("/music/categories", adminHandler.CreateMusicCategory)
	admin.Put("/music/categories/:id", adminHandler.UpdateMusicCategory)
	admin.Delete("/music/categories/:id", adminHandler.DeleteMusicCategory)
	admin.Post("/music", adminHandler.CreateMusic)
	admin.Delete("/music/:id", adminHandler.DeleteMusic)
	admin.Get("/themes", adminHandler.ListThemes)
	admin.Put("/themes/:id", adminHandler.UpdateTheme)
	admin.Post("/upload", uploadHandler.AdminUpload)
}
