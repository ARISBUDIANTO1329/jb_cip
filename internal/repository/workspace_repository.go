package repository

import (
	"database/sql"
	"fmt"

	"github.com/jaybani/jb_cip/internal/domain"
)

type WorkspaceRepository struct {
	db *sql.DB
}

func NewWorkspaceRepository(db *sql.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

func (r *WorkspaceRepository) Create(workspace *domain.Workspace, userID string) error {
	query := `
		INSERT INTO core.workspaces (id, owner_id, name, slug, description, status)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, 'active')
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query, userID, workspace.Name, workspace.Slug, workspace.Description).
		Scan(&workspace.ID, &workspace.CreatedAt, &workspace.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	// Create owner membership
	memberQuery := `
		INSERT INTO core.workspace_members (workspace_id, user_id, role_id, status, joined_at)
		SELECT $1, $2, r.id, 'active', NOW()
		FROM core.roles r 
		WHERE r.workspace_id = $1 AND r.name = 'Owner'
		ON CONFLICT (workspace_id, user_id) DO NOTHING
	`
	_, err = r.db.Exec(memberQuery, workspace.ID, userID)
	if err != nil {
		return fmt.Errorf("failed to create workspace member: %w", err)
	}

	return nil
}

func (r *WorkspaceRepository) FindByID(id string) (*domain.Workspace, error) {
	query := `
		SELECT id, owner_id, name, slug, description, status, settings, created_at, updated_at
		FROM core.workspaces
		WHERE id = $1 AND deleted_at IS NULL
	`

	ws := &domain.Workspace{}
	var settings sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&ws.ID, &ws.OwnerID, &ws.Name, &ws.Slug, &ws.Description,
		&ws.Status, &settings, &ws.CreatedAt, &ws.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workspace not found")
	}
	if err != nil {
		return nil, err
	}

	if settings.Valid {
		ws.Settings = []byte(settings.String)
	}

	return ws, nil
}

func (r *WorkspaceRepository) List(userID string, limit, offset int) ([]*domain.Workspace, error) {
	query := `
		SELECT w.id, w.owner_id, w.name, w.slug, w.description, w.status, w.settings, w.created_at, w.updated_at
		FROM core.workspaces w
		JOIN core.workspace_members wm ON w.id = wm.workspace_id
		WHERE wm.user_id = $1 AND w.deleted_at IS NULL AND wm.deleted_at IS NULL AND wm.status = 'active'
		ORDER BY w.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		ws := &domain.Workspace{}
		var settings sql.NullString
		err := rows.Scan(
			&ws.ID, &ws.OwnerID, &ws.Name, &ws.Slug, &ws.Description,
			&ws.Status, &settings, &ws.CreatedAt, &ws.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if settings.Valid {
			ws.Settings = []byte(settings.String)
		}
		workspaces = append(workspaces, ws)
	}

	return workspaces, nil
}

func (r *WorkspaceRepository) Update(workspace *domain.Workspace) error {
	query := `
		UPDATE core.workspaces
		SET name = $1, description = $2, status = $3, settings = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, workspace.Name, workspace.Description, workspace.Status,
		workspace.Settings, workspace.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("workspace not found or already deleted")
	}

	return nil
}

func (r *WorkspaceRepository) Delete(id string) error {
	query := `UPDATE core.workspaces SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("workspace not found or already deleted")
	}

	// Soft delete members too
	_, _ = r.db.Exec(`UPDATE core.workspace_members SET deleted_at = NOW(), status = 'removed' 
		WHERE workspace_id = $1 AND deleted_at IS NULL`, id)

	return nil
}


func (r *WorkspaceRepository) FindBySlug(slug string) (*domain.Workspace, error) {
	query := `
		SELECT id, owner_id, name, slug, description, status, created_at, updated_at
		FROM core.workspaces
		WHERE slug = $1 AND deleted_at IS NULL
	`

	ws := &domain.Workspace{}
	err := r.db.QueryRow(query, slug).Scan(
		&ws.ID, &ws.OwnerID, &ws.Name, &ws.Slug, &ws.Description,
		&ws.Status, &ws.CreatedAt, &ws.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workspace not found")
	}
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (r *WorkspaceRepository) CountMembers(workspaceID string) (int, error) {
	query := `SELECT COUNT(*) FROM core.workspace_members 
		WHERE workspace_id = $1 AND deleted_at IS NULL AND status = 'active'`
	var count int
	err := r.db.QueryRow(query, workspaceID).Scan(&count)
	return count, err
}

func (r *WorkspaceRepository) GetUserWorkspace(userID string) (*domain.Workspace, error) {
	query := `
		SELECT w.id, w.owner_id, w.name, w.slug, w.description, w.status, w.created_at, w.updated_at
		FROM core.workspaces w
		JOIN core.workspace_members wm ON w.id = wm.workspace_id
		WHERE wm.user_id = $1 AND w.deleted_at IS NULL AND wm.deleted_at IS NULL AND wm.status = 'active'
		ORDER BY w.created_at ASC
		LIMIT 1
	`
	ws := &domain.Workspace{}
	err := r.db.QueryRow(query, userID).Scan(
		&ws.ID, &ws.OwnerID, &ws.Name, &ws.Slug, &ws.Description,
		&ws.Status, &ws.CreatedAt, &ws.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workspace not found")
	}
	if err != nil {
		return nil, err
	}
	return ws, nil
}

func (r *WorkspaceRepository) IsOwner(workspaceID, userID string) (bool, error) {

	var exists bool
	query := `SELECT EXISTS(
		SELECT 1 FROM core.workspaces 
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL
	)`
	err := r.db.QueryRow(query, workspaceID, userID).Scan(&exists)
	return exists, err
}

func (r *WorkspaceRepository) IsAdmin(workspaceID, userID string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM core.workspace_members wm
			JOIN core.roles r ON wm.role_id = r.id
			WHERE wm.workspace_id = $1 AND wm.user_id = $2 
			AND wm.deleted_at IS NULL AND wm.status = 'active'
			AND r.name IN ('Owner', 'Admin')
		)`
	err := r.db.QueryRow(query, workspaceID, userID).Scan(&exists)
	return exists, err
}

func (r *WorkspaceRepository) GetUserRole(workspaceID, userID string) (string, error) {
	var roleName string
	query := `
		SELECT r.name FROM core.workspace_members wm
		JOIN core.roles r ON wm.role_id = r.id
		WHERE wm.workspace_id = $1 AND wm.user_id = $2 
		AND wm.deleted_at IS NULL AND wm.status = 'active'
	`
	err := r.db.QueryRow(query, workspaceID, userID).Scan(&roleName)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("member not found")
	}
	return roleName, err
}