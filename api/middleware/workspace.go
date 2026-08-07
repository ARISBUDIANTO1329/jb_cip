package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/pkg/errors"
)

func WorkspaceRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		workspaceID := c.Get("X-Workspace-ID")
		if workspaceID == "" {
			workspaceID = c.Query("workspace_id")
		}
		if workspaceID == "" {
			return helper.SendError(c, errors.New("WORKSPACE_001", "Workspace ID is required", 400))
		}
		c.Locals("workspace_id", workspaceID)
		return c.Next()
	}
}
