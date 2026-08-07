package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/internal/service"
	"github.com/jaybani/jb_cip/pkg/errors"
)

type WorkspaceHandler struct {
	workspaceService *service.WorkspaceService
}

func NewWorkspaceHandler(workspaceService *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{workspaceService: workspaceService}
}

func (h *WorkspaceHandler) Create(c *fiber.Ctx) error {
	var req service.CreateWorkspaceRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendError(c, errors.New("VALIDATION_001", "Invalid request body", 400))
	}

	if req.Name == "" || req.Slug == "" {
		return helper.SendError(c, errors.New("VALIDATION_001", "Name and slug are required", 400))
	}

	userID := c.Locals("user_id").(string)

	resp, err := h.workspaceService.CreateWorkspace(&req, userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	data := map[string]interface{}{
		"workspace": map[string]interface{}{
			"id":          resp.Workspace.ID,
			"name":        resp.Workspace.Name,
			"slug":        resp.Workspace.Slug,
			"description": resp.Workspace.Description,
			"status":      resp.Workspace.Status,
			"created_at":  resp.Workspace.CreatedAt,
			"updated_at":  resp.Workspace.UpdatedAt,
		},
		"role": resp.Role,
	}

	return helper.SendSuccess(c, "Workspace created", data, nil)
}

func (h *WorkspaceHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	ws, err := h.workspaceService.GetWorkspace(id, userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	data := map[string]interface{}{
		"id":          ws.ID,
		"owner_id":    ws.OwnerID,
		"name":        ws.Name,
		"slug":        ws.Slug,
		"description": ws.Description,
		"status":      ws.Status,
		"created_at":  ws.CreatedAt,
		"updated_at":  ws.UpdatedAt,
	}

	return helper.SendSuccess(c, "Workspace retrieved", data, nil)
}

func (h *WorkspaceHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	workspaces, err := h.workspaceService.ListWorkspaces(userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	var data []map[string]interface{}
	for _, ws := range workspaces {
		memberCount, _ := h.workspaceService.CountMembers(ws.ID)
		data = append(data, map[string]interface{}{
			"id":           ws.ID,
			"owner_id":     ws.OwnerID,
			"name":         ws.Name,
			"slug":         ws.Slug,
			"description":  ws.Description,
			"status":       ws.Status,
			"member_count": memberCount,
			"created_at":   ws.CreatedAt,
			"updated_at":   ws.UpdatedAt,
		})
	}

	pagination := map[string]interface{}{
		"page":        1,
		"per_page":    len(data),
		"total":       len(data),
		"total_pages": 1,
		"has_next":    false,
		"has_prev":    false,
	}

	return helper.SendSuccessWithPagination(c, "Workspaces retrieved", data, pagination)
}

func (h *WorkspaceHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	var req service.UpdateWorkspaceRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendError(c, errors.New("VALIDATION_001", "Invalid request body", 400))
	}

	if req.Name == "" {
		return helper.SendError(c, errors.New("VALIDATION_001", "Name is required", 400))
	}

	if err := h.workspaceService.UpdateWorkspace(id, userID, &req); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Workspace updated", nil, nil)
}

func (h *WorkspaceHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	if err := h.workspaceService.DeleteWorkspace(id, userID); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Workspace deleted", nil, nil)
}

func (h *WorkspaceHandler) InviteMember(c *fiber.Ctx) error {
	workspaceID := c.Params("id")
	userID := c.Locals("user_id").(string)

	var req service.InviteMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendError(c, errors.New("VALIDATION_001", "Invalid request body", 400))
	}

	if req.Email == "" {
		return helper.SendError(c, errors.New("VALIDATION_001", "Email is required", 400))
	}

	member, err := h.workspaceService.InviteMember(workspaceID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	data := map[string]interface{}{
		"id":           member.ID,
		"workspace_id": member.WorkspaceID,
		"user_id":      member.UserID,
		"status":       member.Status,
		"invited_at":   member.InvitedAt,
		"created_at":   member.CreatedAt,
	}

	return helper.SendSuccess(c, "Member invited", data, nil)
}

func (h *WorkspaceHandler) ListMembers(c *fiber.Ctx) error {
	workspaceID := c.Params("id")
	userID := c.Locals("user_id").(string)

	members, roles, err := h.workspaceService.ListMembers(workspaceID, userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	var data []map[string]interface{}
	for _, m := range members {
		parts := strings.SplitN(m.RoleID, "|", 2)
		userEmail := parts[0]
		userName := ""
		if len(parts) > 1 {
			userName = parts[1]
		}
		data = append(data, map[string]interface{}{
			"id":         m.ID,
			"user_id":    m.UserID,
			"email":      userEmail,
			"name":       userName,
			"status":     m.Status,
			"invited_at": m.InvitedAt,
			"joined_at":  m.JoinedAt,
			"created_at": m.CreatedAt,
		})
	}

	result := map[string]interface{}{
		"members": data,
		"roles":   roles,
	}

	return helper.SendSuccess(c, "Members retrieved", result, nil)
}

func (h *WorkspaceHandler) UpdateMemberRole(c *fiber.Ctx) error {
	workspaceID := c.Params("id")
	memberID := c.Params("member_id")
	userID := c.Locals("user_id").(string)

	var req service.UpdateMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendError(c, errors.New("VALIDATION_001", "Invalid request body", 400))
	}

	if req.Role == "" {
		return helper.SendError(c, errors.New("VALIDATION_001", "Role is required", 400))
	}

	if err := h.workspaceService.UpdateMemberRole(workspaceID, memberID, userID, req.Role); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Member role updated", nil, nil)
}

func (h *WorkspaceHandler) RemoveMember(c *fiber.Ctx) error {
	workspaceID := c.Params("id")
	memberID := c.Params("member_id")
	userID := c.Locals("user_id").(string)

	if err := h.workspaceService.RemoveMember(workspaceID, memberID, userID); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	return helper.SendSuccess(c, "Member removed", nil, nil)
}

func (h *WorkspaceHandler) SwitchWorkspace(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req service.SwitchWorkspaceRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendError(c, errors.New("VALIDATION_001", "Invalid request body", 400))
	}

	if req.WorkspaceID == "" {
		return helper.SendError(c, errors.New("VALIDATION_001", "workspace_id is required", 400))
	}

	role, err := h.workspaceService.SwitchWorkspace(userID, req.WorkspaceID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return helper.SendError(c, appErr)
		}
		return helper.SendError(c, errors.New("SYSTEM_001", "Internal server error", 500))
	}

	data := map[string]interface{}{
		"workspace_id": req.WorkspaceID,
		"role":         role,
	}

	return helper.SendSuccess(c, "Workspace switched", data, nil)
}
