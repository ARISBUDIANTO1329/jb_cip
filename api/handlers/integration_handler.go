package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/internal/service"
	"github.com/jaybani/jb_cip/pkg/errors"
)

type IntegrationHandler struct {
	integrationService *service.IntegrationService
	baseURL            string
}

func NewIntegrationHandler(integrationService *service.IntegrationService, baseURL string) *IntegrationHandler {
	return &IntegrationHandler{
		integrationService: integrationService,
		baseURL:            baseURL,
	}
}

func (h *IntegrationHandler) GoogleLogin(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	wsRepo := h.integrationService.GetWorkspaceRepo()
	workspace, err := wsRepo.GetUserWorkspace(userID)
	if err != nil {
		return helper.SendError(c, errors.New("WORKSPACE_001", "No workspace found", 404))
	}

	resp, err := h.integrationService.GoogleLogin(userID, workspace.ID, h.baseURL)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Google OAuth URL generated", resp, nil)
}

func (h *IntegrationHandler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return helper.SendError(c, errors.New("VALIDATION_001", "Authorization code is required", 400))
	}

	state := c.Query("state")
	userID, workspaceID, err := parseState(state)
	if err != nil {
		return helper.SendError(c, errors.New("VALIDATION_002", "Invalid state parameter", 400))
	}

	resp, err := h.integrationService.GoogleCallback(code, userID, workspaceID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Google OAuth connected", resp, nil)
}

func (h *IntegrationHandler) Disconnect(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.integrationService.Disconnect(userID); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Google account disconnected", nil, nil)
}

func (h *IntegrationHandler) TestConnection(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	resp, err := h.integrationService.TestConnection(userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Connection test completed", resp, nil)
}

func (h *IntegrationHandler) GetYouTubeChannels(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	channels, err := h.integrationService.GetYouTubeChannels(userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "YouTube channels retrieved", channels, nil)
}

func parseState(state string) (string, string, error) {
	parts := strings.Split(state, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid state format")
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		return "", "", fmt.Errorf("invalid user id in state")
	}
	if _, err := uuid.Parse(parts[1]); err != nil {
		return "", "", fmt.Errorf("invalid workspace id in state")
	}
	return parts[0], parts[1], nil
}
