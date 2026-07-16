package routes

import (
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/handler"
	"github.com/dhanuprys/momenu-backend-fiber/internal/middleware"
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/database"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// ─── Dependency Container ───────────────────────────────────────────────────

// deps holds all handler and middleware instances wired up at startup.
type deps struct {
	user            *handler.UserHandler
	theme           *handler.ThemeHandler
	project         *handler.ProjectHandler
	schedule        *handler.ScheduleHandler
	giftRegistry    *handler.GiftRegistryHandler
	disk            *handler.DiskHandler
	media           *handler.MediaHandler
	dressCode       *handler.DressCodeHandler
	liveStream      *handler.LiveStreamHandler
	rsvp            *handler.RSVPHandler
	guestbook       *handler.GuestbookHandler
	invitation      *handler.InvitationHandler
	admin           *handler.AdminHandler
	system          *handler.SystemHandler
	textOverride    *handler.TextOverrideHandler
	styleOverride   *handler.StyleOverrideHandler
	music           *handler.MusicHandler
	upload          *handler.UploadHandler
	analytics       *handler.AnalyticsHandler
	share           *handler.ShareHandler
	ownerMiddleware fiber.Handler
}

func initDependencies() *deps {
	// Repositories
	userRepo := repository.NewUserRepository(database.DB)
	themeRepo := repository.NewThemeRepository(database.DB)
	projectRepo := repository.NewProjectRepository(database.DB)
	scheduleRepo := repository.NewScheduleRepository(database.DB)
	giftRegistryRepo := repository.NewGiftRegistryRepository(database.DB)
	fileRepo := repository.NewFileRecordRepository(database.DB)
	mediaRepo := repository.NewMediaRepository(database.DB)
	dressCodeRepo := repository.NewDressCodeRepository(database.DB)
	liveStreamRepo := repository.NewLiveStreamRepository(database.DB)
	rsvpRepo := repository.NewRSVPRepository(database.DB)
	guestbookRepo := repository.NewGuestbookRepository(database.DB)
	musicRepo := repository.NewMusicRepository(database.DB)
	analyticsRepo := repository.NewAnalyticsRepository(database.DB)
	shareRepo := repository.NewShareRepository(database.DB)

	// Services
	userService := service.NewUserService(userRepo)
	themeService := service.NewThemeService(themeRepo)
	projectService := service.NewProjectService(projectRepo, themeRepo, userRepo)
	scheduleService := service.NewScheduleService(scheduleRepo)
	giftRegistryService := service.NewGiftRegistryService(giftRegistryRepo)
	quotaSvc := service.NewDiskQuotaService(fileRepo, projectRepo)
	mediaService := service.NewMediaService(mediaRepo, themeRepo)
	dressCodeService := service.NewDressCodeService(dressCodeRepo)
	liveStreamService := service.NewLiveStreamService(liveStreamRepo)
	rsvpService := service.NewRSVPService(rsvpRepo, projectRepo)
	guestbookService := service.NewGuestbookService(guestbookRepo, projectRepo)
	invitationService := service.NewInvitationService(projectRepo)
	musicService := service.NewMusicService(musicRepo)
	ipCheckerService := service.NewIPCheckerService()
	analyticsService := service.NewAnalyticsService(analyticsRepo, projectRepo, ipCheckerService)
	shareService := service.NewShareService(shareRepo, projectRepo, rsvpRepo, guestbookRepo, analyticsRepo)

	return &deps{
		user:            handler.NewUserHandler(userService),
		theme:           handler.NewThemeHandler(themeService),
		project:         handler.NewProjectHandler(projectService),
		schedule:        handler.NewScheduleHandler(scheduleService),
		giftRegistry:    handler.NewGiftRegistryHandler(giftRegistryService),
		disk:            handler.NewDiskHandler(fileRepo, projectRepo, quotaSvc, database.DB),
		media:           handler.NewMediaHandler(mediaService, fileRepo, quotaSvc),
		dressCode:       handler.NewDressCodeHandler(dressCodeService),
		liveStream:      handler.NewLiveStreamHandler(liveStreamService),
		rsvp:            handler.NewRSVPHandler(rsvpService, projectService),
		guestbook:       handler.NewGuestbookHandler(guestbookService, projectService, rsvpService),
		invitation:      handler.NewInvitationHandler(invitationService, rsvpService),
		admin:           handler.NewAdminHandler(),
		system:          handler.NewSystemHandler(),
		textOverride:    handler.NewTextOverrideHandler(),
		styleOverride:   handler.NewStyleOverrideHandler(),
		music:           handler.NewMusicHandler(musicService),
		upload:          handler.NewUploadHandler(projectRepo, fileRepo, quotaSvc),
		analytics:       handler.NewAnalyticsHandler(analyticsService),
		share:           handler.NewShareHandler(shareService),
		ownerMiddleware: middleware.ProjectOwner(projectRepo),
	}
}

// ─── Route Registration ────────────────────────────────────────────────────

func SetupRoutes(app *fiber.App) {
	d := initDependencies()

	api := app.Group("/api/v1")

	registerHealthCheck(api)
	registerAuthRoutes(api, d)
	registerPublicRoutes(api, d)
	registerProjectRoutes(api, d)
	registerInvitationRoutes(api, d)
	registerAdminRoutes(api, d)
}

// ─── Health Check ───────────────────────────────────────────────────────────

func registerHealthCheck(api fiber.Router) {
	api.Get("/health", func(c fiber.Ctx) error {
		return response.JSONSuccess(c, fiber.StatusOK, "Server is healthy", fiber.Map{
			"status": "up",
		}, nil)
	})
}

// ─── Auth & Identity ────────────────────────────────────────────────────────

func registerAuthRoutes(api fiber.Router, d *deps) {
	auth := api.Group("/auth")
	auth.Post("/register", d.user.Register)
	auth.Post("/login", d.user.Login)
	auth.Get("/google/login", d.user.GoogleLoginURL)
	auth.Get("/google/callback", d.user.GoogleCallback)
	auth.Post("/refresh", d.user.RefreshToken)
	auth.Get("/me", middleware.AuthRequired, d.user.Me)
	auth.Put("/me", middleware.AuthRequired, d.user.UpdateProfile)
}

// ─── Public Routes (no auth) ────────────────────────────────────────────────

func registerPublicRoutes(api fiber.Router, d *deps) {
	// Event Type Schema
	api.Get("/event-types/:type/schema", func(c fiber.Ctx) error {
		eventType := models.EventType(c.Params("type"))
		schema := models.GetFieldSchema(eventType)
		if schema == nil {
			return response.JSONError(c, fiber.StatusNotFound, "Event type not found", "EVENT_TYPE_NOT_FOUND")
		}
		return response.JSONSuccess(c, fiber.StatusOK, "Field schema retrieved", schema, nil)
	})

	// Themes
	themes := api.Group("/themes")
	themes.Get("/", d.theme.List)
	themes.Get("/:id", d.theme.Get)

	// Music
	music := api.Group("/music")
	music.Get("/categories", d.music.ListCategories)
	music.Get("/", d.music.ListMusics)

	// Public Resources (text/style overrides, share sessions)
	api.Get("/public/text-overrides/:slug", d.textOverride.GetBySlug)
	api.Get("/public/style-overrides/:slug", d.styleOverride.GetBySlug)
	api.Get("/public/share/:sessionId", d.share.GetSharedData)

	// Analytics (public visit tracking)
	inviteLimiter := newInviteLimiter()
	analytics := api.Group("/analytics")
	analytics.Post("/visit", inviteLimiter, d.analytics.RecordVisit)
}

// ─── Project Routes (authenticated) ────────────────────────────────────────

func registerProjectRoutes(api fiber.Router, d *deps) {
	// File Upload
	api.Post("/upload", middleware.AuthRequired, d.upload.Upload)

	// Project CRUD
	projects := api.Group("/projects", middleware.AuthRequired)
	projects.Post("/", d.project.Create)
	projects.Get("/", d.project.List)

	// Single Project (owner-scoped)
	project := projects.Group("/:projectId", d.ownerMiddleware, middleware.TrackProjectUpdate)
	project.Get("/", d.project.Get)
	project.Get("/version", d.project.GetVersion)
	project.Put("/", d.project.Update)
	project.Delete("/", d.project.Delete)
	project.Patch("/status", d.project.UpdateStatus)
	project.Get("/feature-toggle", d.project.GetFeatureToggle)
	project.Put("/feature-toggle", d.project.UpdateFeatureToggle)
	project.Get("/analytics", d.analytics.GetProjectAnalytics)
	project.Get("/disk-usage", d.disk.GetProjectDiskUsage)

	// ── Sub-resources ───────────────────────────────────────────

	// Share Sessions
	project.Post("/share", d.share.CreateSession)
	project.Get("/share", d.share.ListSessions)
	project.Delete("/share/:sessionId", d.share.RevokeSession)

	// Schedules
	project.Get("/schedules", d.schedule.List)
	project.Post("/schedules", d.schedule.Create)
	project.Put("/schedules/:scheduleId", d.schedule.Update)
	project.Delete("/schedules/:scheduleId", d.schedule.Delete)

	// Gift Registries
	project.Get("/gift-registries", d.giftRegistry.List)
	project.Post("/gift-registries", d.giftRegistry.Create)
	project.Put("/gift-registries/:registryId", d.giftRegistry.Update)
	project.Delete("/gift-registries/:registryId", d.giftRegistry.Delete)

	// Media Mappings
	project.Get("/media", d.media.List)
	project.Post("/media", d.media.Create)
	project.Put("/media/:mediaId", d.media.Update)
	project.Delete("/media/:mediaId", d.media.Delete)

	// Dress Codes
	project.Get("/dress-codes", d.dressCode.List)
	project.Post("/dress-codes", d.dressCode.Create)
	project.Put("/dress-codes/:dressCodeId", d.dressCode.Update)
	project.Delete("/dress-codes/:dressCodeId", d.dressCode.Delete)

	// Live Streams
	project.Get("/live-streams", d.liveStream.List)
	project.Post("/live-streams", d.liveStream.Create)
	project.Put("/live-streams/:streamId", d.liveStream.Update)
	project.Delete("/live-streams/:streamId", d.liveStream.Delete)

	// Text Overrides
	project.Get("/text-overrides", d.textOverride.List)
	project.Put("/text-overrides", d.textOverride.Upsert)
	project.Delete("/text-overrides/:slotKey", d.textOverride.Delete)

	// Style Overrides
	project.Get("/style-overrides", d.styleOverride.List)
	project.Put("/style-overrides", d.styleOverride.Upsert)
	project.Delete("/style-overrides/:slotKey", d.styleOverride.Delete)

	// RSVPs
	rsvps := project.Group("/rsvps")
	rsvps.Get("/export", d.rsvp.ExportXLSX)
	rsvps.Post("/bulk", d.rsvp.ImportXLSX)
	rsvps.Get("/stats", d.rsvp.Stats)
	rsvps.Get("/", d.rsvp.List)
	rsvps.Post("/", d.rsvp.OwnerUpsert)
	rsvps.Put("/:id", d.rsvp.Update)
	rsvps.Delete("/:id", d.rsvp.Delete)

	// Guestbook
	gb := project.Group("/guestbook")
	gb.Get("/", d.guestbook.List)
	gb.Delete("/:entryId", d.guestbook.Delete)
}

// ─── Public Invitation Routes ───────────────────────────────────────────────

func registerInvitationRoutes(api fiber.Router, d *deps) {
	inviteLimiter := newInviteLimiter()

	// Register the same handlers under both /invite and /public/invite prefixes
	for _, prefix := range []string{"/invite/:slug", "/public/invite/:slug"} {
		invite := api.Group(prefix)
		invite.Get("/og", d.invitation.GetOGMetadata)
		invite.Get("/", d.invitation.GetInvitation)
		invite.Get("/guest", d.rsvp.PublicGetGuest)
		invite.Post("/rsvp", inviteLimiter, d.rsvp.PublicUpsert)
		invite.Get("/guestbook", d.guestbook.PublicList)
		invite.Post("/guestbook", inviteLimiter, d.guestbook.PublicCreate)
	}
}

// ─── Admin Routes ───────────────────────────────────────────────────────────

func registerAdminRoutes(api fiber.Router, d *deps) {
	admin := api.Group("/admin", middleware.AuthRequired, middleware.AdminRequired)

	// User Management
	admin.Get("/users", d.admin.ListUsers)
	admin.Post("/users", d.admin.CreateUser)
	admin.Post("/users/:id/login-as", d.admin.LoginAsUser)
	admin.Delete("/users/:id", d.admin.DeleteUser)
	admin.Put("/users/:id/status", d.admin.UpdateUserStatus)

	// Project Management
	admin.Get("/projects", d.admin.ListProjects)
	admin.Delete("/projects/:id", d.admin.DeleteProject)
	admin.Patch("/projects/:projectId/disk-quota", d.disk.UpdateProjectDiskQuota)

	// Music Management
	admin.Post("/music/categories", d.admin.CreateMusicCategory)
	admin.Put("/music/categories/:id", d.admin.UpdateMusicCategory)
	admin.Delete("/music/categories/:id", d.admin.DeleteMusicCategory)
	admin.Post("/music", d.admin.CreateMusic)
	admin.Put("/music/:id", d.admin.UpdateMusic)
	admin.Delete("/music/:id", d.admin.DeleteMusic)

	// Theme Management
	admin.Get("/themes", d.admin.ListThemes)
	admin.Put("/themes/:id", d.admin.UpdateTheme)

	// System
	admin.Post("/upload", d.upload.AdminUpload)
	admin.Get("/disk-stats", d.disk.GetGlobalStats)
	admin.Get("/system/resources", d.system.GetResources)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func newInviteLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
	})
}
