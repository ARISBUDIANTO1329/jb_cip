package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/internal/repository"
	"github.com/jaybani/jb_cip/pkg/errors"
)

type VideoHandler struct {
	channelRepo *repository.ChannelRepository
	syncRepo    *repository.SyncRepository
}

func NewVideoHandler(channelRepo *repository.ChannelRepository, syncRepo *repository.SyncRepository) *VideoHandler {
	return &VideoHandler{
		channelRepo: channelRepo,
		syncRepo:    syncRepo,
	}
}

func (h *VideoHandler) ListVideos(c *fiber.Ctx) error {
	channelIDStr := c.Params("channel_id")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		return helper.SendError(c, errors.New("VIDEO_001", "Invalid channel ID", 400))
	}

	workspaceID := c.Locals("workspace_id").(string)

	channel, err := h.channelRepo.GetByID(channelID)
	if err != nil {
		return helper.SendError(c, errors.New("VIDEO_002", "Channel not found", 404))
	}

	if channel.WorkspaceID != workspaceID {
		return helper.SendError(c, errors.New("VIDEO_003", "Channel not found in workspace", 404))
	}

	limit := 20
	offset := 0
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if v := c.Query("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	if limit > 100 {
		limit = 100
	}

	videos, total, err := h.syncRepo.GetVideosByChannelWithPagination(channelID, limit, offset)
	if err != nil {
		return helper.SendError(c, errors.New("SYSTEM_001", "Failed to fetch videos", 500))
	}

	return helper.SendSuccessWithPagination(c, "Videos retrieved", videos, fiber.Map{
		"limit":  limit,
		"offset": offset,
		"total":  total,
	})
}

func (h *VideoHandler) GetVideo(c *fiber.Ctx) error {
	channelIDStr := c.Params("channel_id")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		return helper.SendError(c, errors.New("VIDEO_001", "Invalid channel ID", 400))
	}

	videoIDStr := c.Params("video_id")
	if videoIDStr == "" {
		return helper.SendError(c, errors.New("VIDEO_004", "Video ID is required", 400))
	}

	workspaceID := c.Locals("workspace_id").(string)

	channel, err := h.channelRepo.GetByID(channelID)
	if err != nil {
		return helper.SendError(c, errors.New("VIDEO_002", "Channel not found", 404))
	}

	if channel.WorkspaceID != workspaceID {
		return helper.SendError(c, errors.New("VIDEO_003", "Channel not found in workspace", 404))
	}

	video, err := h.syncRepo.GetVideoByID(channelID, videoIDStr)
	if err != nil {
		return helper.SendError(c, errors.New("VIDEO_005", "Video not found", 404))
	}

	return helper.SendSuccess(c, "Video retrieved", video, nil)
}
