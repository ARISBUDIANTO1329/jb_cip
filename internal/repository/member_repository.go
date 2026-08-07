package repository

import (
	"database/sql"
	"fmt"

	"github.com/jaybani/jb_cip/internal/domain"
)

type MemberRepository struct {
	db *sql.DB
}

func NewMemberRepository(db *sql.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) FindByID(id string) (*domain.WorkspaceMember, error) {
	query := `
		SELECT m.id, m.workspace_id, m.user_id, m.role_id, m.invited_by,
		       m.invited_at, m.joined_at, m.status, m.created_at, m.updated_at
		FROM core.workspace_members m
		WHERE m.id = $1 AND m.deleted_at IS NULL
	`

	m := &domain.WorkspaceMember{}
	err := r.db.QueryRow(query, id).Scan(
		&m.ID, &m.WorkspaceID, &m.UserID, &m.RoleID, &m.InvitedBy,
		&m.InvitedAt, &m.JoinedAt, &m.Status, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("member not found")
	}
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (r *MemberRepository) Invite(workspaceID, userID, inviterID string) (*domain.WorkspaceMember, error) {
	// Get the default role 'Member' for this workspace
	var roleID string
	err := r.db.QueryRow(
		`SELECT id FROM core.roles WHERE workspace_id = $1 AND name = 'Member' AND is_system = true`,
		workspaceID,
	).Scan(&roleID)
	if err != nil {
		// Default Member role not found, create it
		err = r.db.QueryRow(
			`INSERT INTO core.roles (workspace_id, name, description, permissions, is_system)
			VALUES ($1, 'Member', 'Edit workspace content', ARRAY['workspace:read','channel:read'], true)
			RETURNING id`,
			workspaceID,
		).Scan(&roleID)
		if err != nil {
			return nil, fmt.Errorf("failed to get/create Member role: %w", err)
		}
	}

	query := `
		INSERT INTO core.workspace_members (workspace_id, user_id, role_id, invited_by, status)
		VALUES ($1, $2, $3, $4, 'invited')
		RETURNING id, workspace_id, user_id, role_id, invited_by, invited_at, joined_at, status, created_at, updated_at
	`

	m := &domain.WorkspaceMember{}
	var invitedBy sql.NullString
	err = r.db.QueryRow(query, workspaceID, userID, roleID, inviterID).Scan(
		&m.ID, &m.WorkspaceID, &m.UserID, &m.RoleID, &invitedBy,
		&m.InvitedAt, &m.JoinedAt, &m.Status, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user already a member")
		}
		return nil, fmt.Errorf("failed to invite member: %w", err)
	}

	if invitedBy.Valid {
		*m.InvitedBy = invitedBy.String
	}

	return m, nil
}

func (r *MemberRepository) List(workspaceID string) ([]*domain.WorkspaceMember, error) {
	query := `
		SELECT m.id, m.workspace_id, m.user_id, r.name as role_name, m.invited_by,
		       m.invited_at, m.joined_at, m.status, m.created_at, m.updated_at,
		       u.email, u.name as user_name
		FROM core.workspace_members m
		JOIN core.roles r ON m.role_id = r.id
		JOIN core.users u ON m.user_id = u.id
		WHERE m.workspace_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.created_at ASC
	`

	rows, err := r.db.Query(query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.WorkspaceMember
	for rows.Next() {
		m := &domain.WorkspaceMember{}
		var invitedBy sql.NullString
		var email, userName string

		err := rows.Scan(
			&m.ID, &m.WorkspaceID, &m.UserID, &m.RoleID, &invitedBy,
			&m.InvitedAt, &m.JoinedAt, &m.Status, &m.CreatedAt, &m.UpdatedAt,
			&email, &userName,
		)
		if err != nil {
			return nil, err
		}
		if invitedBy.Valid {
			*m.InvitedBy = invitedBy.String
		}

		// Store email/name in role_id temporarily for response
		m.RoleID = email + "|" + userName

		members = append(members, m)
	}

	return members, nil
}

func (r *MemberRepository) UpdateRole(memberID, roleName string) error {
	updateQuery := `
		UPDATE core.workspace_members m
		SET role_id = (SELECT id FROM core.roles r WHERE r.workspace_id = 
			(SELECT workspace_id FROM core.workspace_members WHERE id = $1) AND r.name = $2)
		WHERE m.id = $1 AND m.deleted_at IS NULL
	`

	result, err := r.db.Exec(updateQuery, memberID, roleName)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

func (r *MemberRepository) Remove(memberID, userID string) error {
	query := `
		UPDATE core.workspace_members 
		SET deleted_at = NOW(), status = 'removed' 
		WHERE id = $1 AND user_id != $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, memberID, userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("member not found or cannot remove self")
	}

	return nil
}

func (r *MemberRepository) FindUserByEmail(email string) (*domain.User, error) {
	query := `SELECT id, email, name, status FROM core.users WHERE email = $1 AND deleted_at IS NULL`
	u := &domain.User{}
	err := r.db.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.Name, &u.Status)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *MemberRepository) GetRoles(workspaceID string) ([]string, error) {
	query := `
		SELECT DISTINCT r.name 
		FROM core.roles r 
		WHERE r.workspace_id = $1 AND r.is_system = true
		ORDER BY r.name
	`

	rows, err := r.db.Query(query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		roles = append(roles, name)
	}

	return roles, nil
}