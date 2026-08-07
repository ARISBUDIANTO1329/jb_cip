package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/internal/service"
	"github.com/jaybani/jb_cip/pkg/errors"
)

type SyncHandler struct {
	syncService *service.SyncService
}

func NewSyncHandler(syncService *service.SyncService) *SyncHandler {
	return &SyncHandler{syncService: syncService}
}

type SyncRequest struct {
	ChannelID string `json:"channel_id"`
	SyncType  string `json:"sync_type,omitempty"`
}

type SyncRetryRequest struct {
	JobID string `json:"job_id"`
}

func (h *SyncHandler) Sync(c *fiber.Ctx) error {
	var req SyncRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendError(c, errors.New("SYNC_001", "Invalid request body", 400))
	}

	if req.ChannelID == "" {
		return helper.SendError(c, errors.New("SYNC_001", "channel_id is required", 400))
	}

	userID := c.Locals("user_id").(string)
	workspaceID := c.Locals("workspace_id").(string)

	resp, err := h.syncService.StartSync(userID, workspaceID, req.ChannelID, req.SyncType)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Sync started", resp, nil)
}

func (h *SyncHandler) SyncStatus(c *fiber.Ctx) error {
	channelID := c.Query("channel_id")
	if channelID == "" {
		return helper.SendError(c, errors.New("SYNC_001", "channel_id is required", 400))
	}

	resp, err := h.syncService.GetSyncStatus(channelID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Sync status", resp, nil)
}

func (h *SyncHandler) SyncHistory(c *fiber.Ctx) error {
	channelID := c.Query("channel_id")
	if channelID == "" {
		return helper.SendError(c, errors.New("SYNC_001", "channel_id is required", 400))
	}

	limit := c.QueryInt("limit", 20)
	if limit > 100 {
		limit = 100
	}

	resp, err := h.syncService.GetSyncHistory(channelID, limit)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Sync history", resp, nil)
}

func (h *SyncHandler) RetrySync(c *fiber.Ctx) error {
	var req SyncRetryRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendError(c, errors.New("SYNC_001", "Invalid request body", 400))
	}

	if req.JobID == "" {
		return helper.SendError(c, errors.New("SYNC_001", "job_id is required", 400))
	}

	userID := c.Locals("user_id").(string)
	workspaceID := c.Locals("workspace_id").(string)

	resp, err := h.syncService.RetrySync(userID, workspaceID, req.JobID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Sync retry started", resp, nil)
}

func (h *SyncHandler) SyncResult(c *fiber.Ctx) error {
	jobID := c.Query("job_id")
	if jobID == "" {
		return helper.SendError(c, errors.New("SYNC_001", "job_id is required", 400))
	}

	resp, err := h.syncService.GetSyncResult(jobID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Sync result", resp, nil)
}
