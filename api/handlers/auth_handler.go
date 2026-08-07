package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/internal/service"
	"github.com/jaybani/jb_cip/pkg/errors"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	body := c.Body()
	fmt.Printf("DEBUG: Content-Type: %s\n", c.Get("Content-Type"))
	fmt.Printf("DEBUG: Raw body: %s\n", string(body))

	var req service.LoginRequest
	
	// Try to parse as JSON first
	if err := json.Unmarshal(body, &req); err != nil {
		// If JSON fails, try form data
		req.Email = c.FormValue("email")
		req.Password = c.FormValue("password")
	}

	if req.Email == "" || req.Password == "" {
		return helper.SendError(c, errors.New("VALIDATION_001", "Email and password are required", 400))
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Login successful", resp, nil)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return helper.SendSuccess(c, "Logout successful", nil, nil)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	body := c.Body()
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		req.RefreshToken = c.FormValue("refresh_token")
	}

	if req.RefreshToken == "" {
		return helper.SendError(c, errors.New("VALIDATION_001", "Refresh token is required", 400))
	}

	resp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Token refreshed", resp, nil)
}

func (h *AuthHandler) Profile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return helper.SendError(c, errors.New("AUTH_001", "Unauthorized", 401))
	}

	email := c.Locals("email")
	role := c.Locals("role")

	data := map[string]interface{}{
		"user_id": userID,
		"email":   email,
		"role":    role,
	}

	return helper.SendSuccess(c, "Profile retrieved", data, nil)
}
