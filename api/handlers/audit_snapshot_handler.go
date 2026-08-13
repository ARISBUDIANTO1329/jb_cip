package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/internal/service"
	"github.com/jaybani/jb_cip/pkg/errors"
)

type AuditSnapshotHandler struct {
	snapshotService *service.AuditSnapshotService
}

func NewAuditSnapshotHandler(snapshotService *service.AuditSnapshotService) *AuditSnapshotHandler {
	return &AuditSnapshotHandler{snapshotService: snapshotService}
}

func (h *AuditSnapshotHandler) CreateSnapshot(c *fiber.Ctx) error {
	channelID := c.Params("id")
	if channelID == "" {
		return helper.SendError(c, errors.New("SNAP_001", "channel id is required", 400))
	}

	workspaceID := c.Locals("workspace_id").(string)

	err := h.snapshotService.CreateSnapshot(channelID, workspaceID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Snapshot created", nil, nil)
}

func (h *AuditSnapshotHandler) Compare(c *fiber.Ctx) error {
	channelID := c.Params("id")
	if channelID == "" {
		return helper.SendError(c, errors.New("SNAP_001", "channel id is required", 400))
	}

	result, err := h.snapshotService.CompareSnapshots(channelID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Snapshot comparison", result, nil)
}
