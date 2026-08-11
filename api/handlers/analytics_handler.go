package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/internal/service"
	"github.com/jaybani/jb_cip/pkg/errors"
)

type AnalyticsHandler struct {
	analyticsReadService *service.AnalyticsReadService
}

func NewAnalyticsHandler(analyticsReadService *service.AnalyticsReadService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsReadService: analyticsReadService}
}

// Summary returns aggregated channel analytics
func (h *AnalyticsHandler) Summary(c *fiber.Ctx) error {
	channelID := c.Query("channel_id")
	if channelID == "" {
		return helper.SendError(c, errors.New("ANALYTICS_001", "channel_id is required", 400))
	}

	workspaceID := c.Locals("workspace_id").(string)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	resp, err := h.analyticsReadService.GetSummary(channelID, workspaceID, startDate, endDate)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Analytics summary", resp, nil)
}

// Timeseries returns daily metric values
func (h *AnalyticsHandler) Timeseries(c *fiber.Ctx) error {
	channelID := c.Query("channel_id")
	if channelID == "" {
		return helper.SendError(c, errors.New("ANALYTICS_001", "channel_id is required", 400))
	}

	workspaceID := c.Locals("workspace_id").(string)
	metric := c.Query("metric", "views")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	resp, err := h.analyticsReadService.GetTimeseries(channelID, workspaceID, metric, startDate, endDate)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Analytics timeseries", resp, nil)
}

// TopVideos returns top videos by views
func (h *AnalyticsHandler) TopVideos(c *fiber.Ctx) error {
	channelID := c.Query("channel_id")
	if channelID == "" {
		return helper.SendError(c, errors.New("ANALYTICS_001", "channel_id is required", 400))
	}

	workspaceID := c.Locals("workspace_id").(string)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limit := 10
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	resp, err := h.analyticsReadService.GetTopVideos(channelID, workspaceID, startDate, endDate, limit)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Top videos", resp, nil)
}
