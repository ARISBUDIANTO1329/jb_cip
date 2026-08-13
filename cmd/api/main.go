package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jaybani/jb_cip/api/handlers"
	"github.com/jaybani/jb_cip/api/middleware"
	"github.com/jaybani/jb_cip/api/routes"
	"github.com/jaybani/jb_cip/config"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/internal/repository"
	"github.com/jaybani/jb_cip/internal/service"
	"github.com/jaybani/jb_cip/internal/providers"
	"github.com/jaybani/jb_cip/pkg/database"
	applogger "github.com/jaybani/jb_cip/pkg/logger"
	"github.com/jaybani/jb_cip/pkg/redis"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	applogger.Init(cfg.App.LogLevel, cfg.App.LogFormat)
	log := applogger.Get()

	db, err := database.Init(&cfg.Database)
	if err != nil {
		log.WithError(err).Warn("Database not available")
	} else {
		if err := database.RunMigrations(cfg); err != nil {
			log.WithError(err).Error("Auto-migration failed")
		} else {
			version, dirty, _ := database.MigrateVersion()
			log.Infof("Database migrated. Version: %d, Dirty: %v", version, dirty)
		}
	}
	defer database.Close()

	if _, err := redis.Init(&cfg.Redis); err != nil {
		log.WithError(err).Warn("Redis not available")
	}
	defer redis.Close()

	userRepo := repository.NewUserRepository(db)
	wsRepo := repository.NewWorkspaceRepository(db)
	memberRepo := repository.NewMemberRepository(db)
	integrationRepo := repository.NewIntegrationRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	syncRepo := repository.NewSyncRepository(db)

	authService := service.NewAuthService(userRepo, cfg)
	workspaceService := service.NewWorkspaceService(wsRepo, memberRepo, cfg)

	googleProvider := providers.NewGoogleProvider(
		cfg.Integrations.GoogleClientID,
		cfg.Integrations.GoogleClientSecret,
		cfg.Integrations.GoogleRedirectURI,
	)

	encryptor := helper.NewTokenEncryptor(cfg.JWT.Secret)
	integrationService := service.NewIntegrationService(integrationRepo, wsRepo, googleProvider, encryptor, cfg)

	youtubeSyncProvider := providers.NewYouTubeSyncProvider(googleProvider)
	syncService := service.NewSyncService(syncRepo, integrationRepo, channelRepo, youtubeSyncProvider, encryptor, cfg)

	authHandler := handlers.NewAuthHandler(authService)
	wsHandler := handlers.NewWorkspaceHandler(workspaceService)
	baseURL := fmt.Sprintf("http://%s:%s", cfg.App.Host, cfg.App.Port)
	intHandler := handlers.NewIntegrationHandler(integrationService, baseURL)
	syncHandler := handlers.NewSyncHandler(syncService)
	videoHandler := handlers.NewVideoHandler(channelRepo, syncRepo)
	analyticsReadService := service.NewAnalyticsReadService(channelRepo, syncRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsReadService)
		auditRepo := repository.NewAuditRepository(db)
	auditService := service.NewAuditService(auditRepo, channelRepo)
	auditHandler := handlers.NewAuditHandler(auditService)
	snapshotRepo := repository.NewAuditSnapshotRepository(db)
	snapshotService := service.NewAuditSnapshotService(snapshotRepo, auditService)
	snapshotHandler := handlers.NewAuditSnapshotHandler(snapshotService)

	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: middleware.ErrorHandler(),
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORS.AllowedOrigins,
		AllowMethods: cfg.CORS.AllowedMethods,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Request-ID",
	}))

	app.Use(middleware.RequestID())
	app.Use(middleware.APIVersion(cfg.App.APIVersion))

	app.Get("/health", handlers.HealthCheck(cfg))
	routes.SetupAPIRoutes(app, cfg, authHandler, wsHandler, intHandler, syncHandler, videoHandler, analyticsHandler, auditHandler, snapshotHandler)
	routes.SetupAnalyticsRoutes(app, cfg, syncHandler)

	app.Use(middleware.NotFound())

	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port)
		log.Infof("Server starting on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.WithError(err).Error("Error during shutdown")
	}
}
