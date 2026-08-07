package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jaybani/jb_cip/api/handlers"
	"github.com/jaybani/jb_cip/api/middleware"
	"github.com/jaybani/jb_cip/config"
)

func SetupAnalyticsRoutes(app *fiber.App, cfg *config.Config, syncHandler *handlers.SyncHandler) {
	api := app.Group("/api/" + cfg.App.APIVersion)

	protected := api.Group("")
	protected.Use(middleware.AuthRequired(cfg))

	// YouTube Analytics routes
	ytAnalytics := protected.Group("/youtube/analytics")
	ytAnalytics.Post("/sync", syncHandler.SyncAnalytics)
	ytAnalytics.Get("/status", syncHandler.GetDailyMetrics)
	ytAnalytics.Get("/history", syncHandler.SyncHistory) // Reuse existing history
}
