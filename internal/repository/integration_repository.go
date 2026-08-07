package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jaybani/jb_cip/internal/domain"
)

type IntegrationRepository struct {
	db *sql.DB
}

func NewIntegrationRepository(db *sql.DB) *IntegrationRepository {
	return &IntegrationRepository{db: db}
}

type pqScanner struct {
	dest *[]string
}

func (s *pqScanner) Scan(src interface{}) error {
	if src == nil {
		*s.dest = []string{}
		return nil
	}
	switch v := src.(type) {
	case string:
		return parsePostgresArray(v, s.dest)
	case []byte:
		return parsePostgresArray(string(v), s.dest)
	case []string:
		*s.dest = v
		return nil
	default:
		return fmt.Errorf("cannot scan %T into []string", src)
	}
}

func pqArray(dest *[]string) *pqScanner {
	return &pqScanner{dest: dest}
}

func parsePostgresArray(s string, dest *[]string) error {
	s = strings.Trim(s, "{}")
	if s == "" {
		*dest = []string{}
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = strings.Trim(p, "\"")
	}
	*dest = result
	return nil
}

func (r *IntegrationRepository) GetConnection(userID, provider string) (*domain.APIConnection, error) {
	query := `
		SELECT id, user_id, workspace_id, provider, provider_user_id, status, scopes, created_at, updated_at
		FROM integration.api_connections
		WHERE user_id = $1 AND provider = $2 AND deleted_at IS NULL
	`
	conn := &domain.APIConnection{}
	if err := r.db.QueryRow(query, userID, provider).Scan(
		&conn.ID, &conn.UserID, &conn.WorkspaceID, &conn.Provider,
		&conn.ProviderUserID, &conn.Status, pqArray(&conn.Scopes),
		&conn.CreatedAt, &conn.UpdatedAt,
	); err == sql.ErrNoRows {
		return nil, fmt.Errorf("connection not found")
	} else if err != nil {
		return nil, err
	}
	return conn, nil
}

func (r *IntegrationRepository) CreateConnection(conn *domain.APIConnection) error {
	query := `
		INSERT INTO integration.api_connections (user_id, workspace_id, provider, provider_user_id, status, scopes)
		VALUES ($1, $2, $3, $4, 'authorized', $5)
		ON CONFLICT (user_id, provider) WHERE deleted_at IS NULL DO UPDATE
		SET status = 'authorized', provider_user_id = EXCLUDED.provider_user_id, scopes = EXCLUDED.scopes
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(query, conn.UserID, conn.WorkspaceID, conn.Provider,
		conn.ProviderUserID, pqArray(&conn.Scopes)).Scan(&conn.ID, &conn.CreatedAt, &conn.UpdatedAt)
}

func (r *IntegrationRepository) SaveToken(token *domain.APIToken) error {
	query := `
		INSERT INTO integration.api_tokens (connection_id, access_token_encrypted, refresh_token_encrypted,
			access_token_expires_at, refresh_token_expires_at, scope)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (connection_id) DO UPDATE
		SET access_token_encrypted = EXCLUDED.access_token_encrypted,
			refresh_token_encrypted = EXCLUDED.refresh_token_encrypted,
			access_token_expires_at = EXCLUDED.access_token_expires_at,
			refresh_token_expires_at = EXCLUDED.refresh_token_expires_at,
			scope = EXCLUDED.scope,
			updated_at = NOW()
	`
	_, err := r.db.Exec(query, token.ConnectionID, token.AccessTokenEncrypted,
		token.RefreshTokenEncrypted, token.AccessTokenExpiresAt,
		token.RefreshTokenExpiresAt, pqArray(&token.Scope))
	return err
}

func (r *IntegrationRepository) GetToken(connectionID string) (*domain.APIToken, error) {
	query := `
		SELECT id, connection_id, access_token_encrypted, refresh_token_encrypted,
			access_token_expires_at, refresh_token_expires_at, scope, created_at, updated_at
		FROM integration.api_tokens
		WHERE connection_id = $1
	`
	token := &domain.APIToken{}
	if err := r.db.QueryRow(query, connectionID).Scan(
		&token.ID, &token.ConnectionID, &token.AccessTokenEncrypted,
		&token.RefreshTokenEncrypted, &token.AccessTokenExpiresAt,
		&token.RefreshTokenExpiresAt, pqArray(&token.Scope), &token.CreatedAt, &token.UpdatedAt,
	); err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found")
	} else if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *IntegrationRepository) DeleteConnection(connectionID string) error {
	query := `UPDATE integration.api_connections SET deleted_at = NOW(), status = 'revoked' WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.Exec(query, connectionID)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("connection not found")
	}
	_, _ = r.db.Exec(`DELETE FROM integration.api_tokens WHERE connection_id = $1`, connectionID)
	return nil
}

func (r *IntegrationRepository) UpdateToken(token *domain.APIToken) error {
	return r.SaveToken(token)
}

func (r *IntegrationRepository) SaveChannels(channels []*domain.YouTubeChannel) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, ch := range channels {
		_, err := tx.Exec(`
			INSERT INTO analytics.channels (workspace_id, connection_id, platform_id, external_id, name, description,
				subscriber_count, view_count, video_count, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'active')
			ON CONFLICT (workspace_id, external_id) DO UPDATE
			SET name = EXCLUDED.name, description = EXCLUDED.description,
				subscriber_count = EXCLUDED.subscriber_count, view_count = EXCLUDED.view_count,
				video_count = EXCLUDED.video_count, connection_id = EXCLUDED.connection_id
		`, ch.WorkspaceID, ch.ConnectionID, ch.PlatformID, ch.ExternalID,
			ch.Name, ch.Description, ch.SubscriberCount, ch.ViewCount, ch.VideoCount)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *IntegrationRepository) GetChannels(workspaceID string) ([]*domain.YouTubeChannel, error) {
	query := `
		SELECT id, workspace_id, connection_id, platform_id, external_id, name, description,
			subscriber_count, view_count, video_count, status, created_at, updated_at
		FROM analytics.channels
		WHERE workspace_id = $1 AND deleted_at IS NULL AND status IN ('active', 'synced')
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*domain.YouTubeChannel
	for rows.Next() {
		ch := &domain.YouTubeChannel{}
		if err := rows.Scan(
			&ch.ID, &ch.WorkspaceID, &ch.ConnectionID, &ch.PlatformID,
			&ch.ExternalID, &ch.Name, &ch.Description, &ch.SubscriberCount,
			&ch.ViewCount, &ch.VideoCount, &ch.Status, &ch.CreatedAt, &ch.UpdatedAt,
		); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, nil
}
