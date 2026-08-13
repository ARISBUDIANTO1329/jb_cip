package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/internal/service"
	"github.com/jaybani/jb_cip/pkg/errors"
)

type AuditHandler struct {
	auditService *service.AuditService
}

func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// AuditVideo runs audit rules against a single video
func (h *AuditHandler) AuditVideo(c *fiber.Ctx) error {
	videoID := c.Params("id")
	if videoID == "" {
		return helper.SendError(c, errors.New("AUDIT_001", "video id is required", 400))
	}

	workspaceID := c.Locals("workspace_id").(string)

	findings, err := h.auditService.AuditVideo(videoID, workspaceID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Video audit completed", findings, nil)
}

// AuditChannel runs audit rules against a channel and its top videos
func (h *AuditHandler) AuditChannel(c *fiber.Ctx) error {
	channelID := c.Params("id")
	if channelID == "" {
		return helper.SendError(c, errors.New("AUDIT_001", "channel id is required", 400))
	}

	workspaceID := c.Locals("workspace_id").(string)

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	findings, err := h.auditService.AuditChannel(channelID, workspaceID, limit)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Channel audit completed", findings, nil)
}
