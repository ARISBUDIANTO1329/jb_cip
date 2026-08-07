package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jaybani/jb_cip/api/handlers"
	"github.com/jaybani/jb_cip/api/middleware"
	"github.com/jaybani/jb_cip/config"
)

func SetupAPIRoutes(app *fiber.App, cfg *config.Config, authHandler *handlers.AuthHandler, wsHandler *handlers.WorkspaceHandler, intHandler *handlers.IntegrationHandler, syncHandler *handlers.SyncHandler) {
	api := app.Group("/api/" + cfg.App.APIVersion)

	healthHandler := handlers.HealthCheck(cfg)
	api.Get("/health", healthHandler)

	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/logout", authHandler.Logout)
	auth.Post("/refresh", authHandler.Refresh)

	// Google OAuth callback - PUBLIC (identitas dari state, tanpa JWT)
	api.Get("/integrations/google/callback", intHandler.GoogleCallback)

	protected := api.Group("")
	protected.Use(middleware.AuthRequired(cfg))
	protected.Get("/auth/profile", authHandler.Profile)

	ws := protected.Group("/workspaces")
	ws.Post("", wsHandler.Create)
	ws.Get("", wsHandler.List)
	ws.Get("/:id", wsHandler.Get)
	ws.Put("/:id", wsHandler.Update)
	ws.Delete("/:id", wsHandler.Delete)
	ws.Post("/switch", wsHandler.SwitchWorkspace)
	ws.Get("/:id/members", wsHandler.ListMembers)
	ws.Post("/:id/members", wsHandler.InviteMember)
	ws.Put("/:id/members/:member_id", wsHandler.UpdateMemberRole)
	ws.Delete("/:id/members/:member_id", wsHandler.RemoveMember)

	// Integration routes
	integrations := protected.Group("/integrations")
	integrations.Get("/google/login", intHandler.GoogleLogin)
	integrations.Post("/google/disconnect", intHandler.Disconnect)
	integrations.Post("/google/test", intHandler.TestConnection)
	integrations.Get("/youtube/channels", intHandler.GetYouTubeChannels)

	// YouTube Sync routes (require workspace)
	yt := protected.Group("/youtube")
	yt.Use(middleware.WorkspaceRequired())
	yt.Post("/sync", syncHandler.Sync)
	yt.Get("/sync/status", syncHandler.SyncStatus)
	yt.Get("/sync/history", syncHandler.SyncHistory)
	yt.Post("/sync/retry", syncHandler.RetrySync)
	yt.Get("/sync/result", syncHandler.SyncResult)
}
