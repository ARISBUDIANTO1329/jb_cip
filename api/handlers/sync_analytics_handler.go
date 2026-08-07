package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/pkg/errors"
)

type SyncAnalyticsRequest struct {
	ChannelID string `json:"channel_id"`
}

type AnalyticsStatusResponse struct {
	JobID         string `json:"job_id"`
	Status        string `json:"status"`
	TotalVideos   int    `json:"total_videos"`
	TotalChannels int    `json:"total_channels"`
	TotalMetrics  int    `json:"total_metrics"`
	TotalSuccess  int    `json:"total_success"`
	TotalFailed   int    `json:"total_failed"`
	Duration      int    `json:"duration_seconds"`
}

func (h *SyncHandler) SyncAnalytics(c *fiber.Ctx) error {
	var req SyncAnalyticsRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendError(c, errors.New("SYNC_001", "Invalid request body", 400))
	}

	if req.ChannelID == "" {
		return helper.SendError(c, errors.New("SYNC_001", "channel_id is required", 400))
	}

	userID := c.Locals("user_id").(string)
	workspaceID := c.Locals("workspace_id").(string)

	resp, err := h.syncService.SyncAnalytics(userID, workspaceID, req.ChannelID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Analytics sync started", resp, nil)
}

func (h *SyncHandler) GetDailyMetrics(c *fiber.Ctx) error {
	channelID := c.Query("channel_id")
	if channelID == "" {
		return helper.SendError(c, errors.New("SYNC_001", "channel_id is required", 400))
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limit := c.QueryInt("limit", 365)
	if limit > 1000 {
		limit = 1000
	}

	metrics, err := h.syncService.GetDailyMetrics(channelID, startDate, endDate, limit)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Daily metrics", metrics, nil)
}
