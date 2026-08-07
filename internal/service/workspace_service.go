package service

import (
	"fmt"

	"github.com/jaybani/jb_cip/config"
	"github.com/jaybani/jb_cip/internal/domain"
	"github.com/jaybani/jb_cip/internal/repository"
	"github.com/jaybani/jb_cip/pkg/errors"
)

type WorkspaceService struct {
	wsRepo   *repository.WorkspaceRepository
	memberRepo *repository.MemberRepository
	cfg      *config.Config
}

func NewWorkspaceService(wsRepo *repository.WorkspaceRepository, memberRepo *repository.MemberRepository, cfg *config.Config) *WorkspaceService {
	return &WorkspaceService{
		wsRepo:   wsRepo,
		memberRepo: memberRepo,
		cfg:      cfg,
	}
}

type CreateWorkspaceRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
}

type CreateWorkspaceResponse struct {
	Workspace  *domain.Workspace `json:"workspace"`
	Role       string            `json:"role"`
}

type UpdateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

type InviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role,omitempty"`
}

type UpdateMemberRequest struct {
	Role string `json:"role"`
}

type SwitchWorkspaceRequest struct {
	WorkspaceID string `json:"workspace_id"`
}

func (s *WorkspaceService) CreateWorkspace(req *CreateWorkspaceRequest, userID string) (*CreateWorkspaceResponse, error) {
	// Check if user already owns a workspace with this slug
	existing, err := s.wsRepo.FindBySlug(req.Slug)
	if err == nil && existing != nil {
		return nil, errors.New("WORKSPACE_002", "Workspace with this slug already exists", 409)
	}

	ws := &domain.Workspace{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	}

	if err := s.wsRepo.Create(ws, userID); err != nil {
		return nil, errors.New("SYSTEM_001", fmt.Sprintf("Failed to create workspace: %v", err), 500)
	}

	return &CreateWorkspaceResponse{
		Workspace: ws,
		Role:      "owner",
	}, nil
}

func (s *WorkspaceService) GetWorkspace(id, userID string) (*domain.Workspace, error) {
	ws, err := s.wsRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("WORKSPACE_001", "Workspace not found", 404)
	}

	// Check if user is a member
	role, err := s.wsRepo.GetUserRole(id, userID)
	if err != nil {
		return nil, errors.New("AUTH_003", "Forbidden: not a member", 403)
	}

	_ = role // role info available if needed
	return ws, nil
}

func (s *WorkspaceService) ListWorkspaces(userID string) ([]*domain.Workspace, error) {
	workspaces, err := s.wsRepo.List(userID, 100, 0)
	if err != nil {
		return nil, errors.New("SYSTEM_001", fmt.Sprintf("Failed to list workspaces: %v", err), 500)
	}
	return workspaces, nil
}

func (s *WorkspaceService) UpdateWorkspace(id, userID string, req *UpdateWorkspaceRequest) error {
	// Check ownership/admin
	isOwner, err := s.wsRepo.IsOwner(id, userID)
	if err != nil {
		return errors.New("WORKSPACE_001", "Workspace not found", 404)
	}
	if !isOwner {
		isAdmin, err := s.wsRepo.IsAdmin(id, userID)
		if err != nil || !isAdmin {
			return errors.New("AUTH_003", "Forbidden: insufficient permissions", 403)
		}
	}

	ws := &domain.Workspace{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := s.wsRepo.Update(ws); err != nil {
		return errors.New("SYSTEM_001", fmt.Sprintf("Failed to update workspace: %v", err), 500)
	}

	return nil
}

func (s *WorkspaceService) DeleteWorkspace(id, userID string) error {
	// Check ownership
	isOwner, err := s.wsRepo.IsOwner(id, userID)
	if err != nil {
		return errors.New("WORKSPACE_001", "Workspace not found", 404)
	}
	if !isOwner {
		return errors.New("AUTH_003", "Forbidden: only owner can delete workspace", 403)
	}

	if err := s.wsRepo.Delete(id); err != nil {
		return errors.New("SYSTEM_001", fmt.Sprintf("Failed to delete workspace: %v", err), 500)
	}

	return nil
}

func (s *WorkspaceService) InviteMember(workspaceID, inviterID string, req *InviteMemberRequest) (*domain.WorkspaceMember, error) {
	// Check ownership/admin
	isOwner, err := s.wsRepo.IsOwner(workspaceID, inviterID)
	if err != nil {
		return nil, errors.New("WORKSPACE_001", "Workspace not found", 404)
	}
	if !isOwner {
		isAdmin, err := s.wsRepo.IsAdmin(workspaceID, inviterID)
		if err != nil || !isAdmin {
			return nil, errors.New("AUTH_003", "Forbidden: insufficient permissions", 403)
		}
	}

	// Find user by email
	user, err := s.memberRepo.FindUserByEmail(req.Email)
	if err != nil {
		return nil, errors.New("USER_001", "User not found", 404)
	}

	member, err := s.memberRepo.Invite(workspaceID, user.ID, inviterID)
	if err != nil {
		return nil, errors.New("MEMBER_001", fmt.Sprintf("Failed to invite member: %v", err), 400)
	}

	return member, nil
}

func (s *WorkspaceService) ListMembers(workspaceID, userID string) ([]*domain.WorkspaceMember, []string, error) {
	// Check if user is a member
	_, err := s.wsRepo.GetUserRole(workspaceID, userID)
	if err != nil {
		return nil, nil, errors.New("AUTH_003", "Forbidden: not a member", 403)
	}

	members, err := s.memberRepo.List(workspaceID)
	if err != nil {
		return nil, nil, errors.New("SYSTEM_001", fmt.Sprintf("Failed to list members: %v", err), 500)
	}

	roles, err := s.memberRepo.GetRoles(workspaceID)
	if err != nil {
		return nil, nil, errors.New("SYSTEM_001", "Failed to get roles", 500)
	}

	return members, roles, nil
}

func (s *WorkspaceService) UpdateMemberRole(workspaceID, memberID, requesterID, roleName string) error {
	// Check ownership/admin
	isOwner, err := s.wsRepo.IsOwner(workspaceID, requesterID)
	if err != nil {
		return errors.New("WORKSPACE_001", "Workspace not found", 404)
	}
	if !isOwner {
		isAdmin, err := s.wsRepo.IsAdmin(workspaceID, requesterID)
		if err != nil || !isAdmin {
			return errors.New("AUTH_003", "Forbidden: insufficient permissions", 403)
		}
	}

	if err := s.memberRepo.UpdateRole(memberID, roleName); err != nil {
		return errors.New("MEMBER_001", fmt.Sprintf("Failed to update role: %v", err), 400)
	}

	return nil
}

func (s *WorkspaceService) RemoveMember(workspaceID, memberID, requesterID string) error {
	// Check ownership/admin
	isOwner, err := s.wsRepo.IsOwner(workspaceID, requesterID)
	if err != nil {
		return errors.New("WORKSPACE_001", "Workspace not found", 404)
	}
	if !isOwner {
		isAdmin, err := s.wsRepo.IsAdmin(workspaceID, requesterID)
		if err != nil || !isAdmin {
			return errors.New("AUTH_003", "Forbidden: insufficient permissions", 403)
		}
	}

	if err := s.memberRepo.Remove(memberID, requesterID); err != nil {
		return errors.New("MEMBER_001", fmt.Sprintf("Failed to remove member: %v", err), 400)
	}

	return nil
}


func (s *WorkspaceService) CountMembers(workspaceID string) (int, error) {
	return s.wsRepo.CountMembers(workspaceID)
}

func (s *WorkspaceService) SwitchWorkspace(userID, workspaceID string) (string, error) {
	// Verify user is a member
	role, err := s.wsRepo.GetUserRole(workspaceID, userID)
	if err != nil {
		return "", errors.New("AUTH_003", "Forbidden: not a member", 403)
	}

	return role, nil
}