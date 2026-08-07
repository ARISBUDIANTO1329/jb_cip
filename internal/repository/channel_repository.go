package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/domain"
)

type ChannelRepository struct {
	db *sql.DB
}

func NewChannelRepository(db *sql.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) GetByID(id uuid.UUID) (*domain.YouTubeChannel, error) {
	query := `
		SELECT id, workspace_id, connection_id, platform_id, external_id, name, description,
			subscriber_count, view_count, video_count, status, created_at, updated_at
		FROM analytics.channels
		WHERE id = $1
	`
	channel := &domain.YouTubeChannel{}
	var description sql.NullString
	var createdAt, updatedAt sql.NullTime
	var connID sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&channel.ID, &channel.WorkspaceID, &connID, &channel.PlatformID, &channel.ExternalID,
		&channel.Name, &description, &channel.SubscriberCount, &channel.ViewCount,
		&channel.VideoCount, &channel.Status, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("channel not found")
	}
	if err != nil {
		return nil, err
	}
	if description.Valid {
		channel.Description = description.String
	}
	if connID.Valid {
		channel.ConnectionID = connID.String
	}
	if createdAt.Valid {
		channel.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		channel.UpdatedAt = updatedAt.Time
	}
	return channel, nil
}

func (r *ChannelRepository) Create(channel *domain.YouTubeChannel) error {
	query := `
		INSERT INTO analytics.channels (
			id, workspace_id, connection_id, platform_id, external_id, name, description,
			subscriber_count, view_count, video_count, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`
	_, err := r.db.Exec(query,
		channel.ID, channel.WorkspaceID, channel.ConnectionID, channel.PlatformID, channel.ExternalID,
		channel.Name, channel.Description, channel.SubscriberCount, channel.ViewCount,
		channel.VideoCount, channel.Status,
	)
	return err
}

func (r *ChannelRepository) UpdateSyncStats(channelID uuid.UUID, totalVideos int, lastSyncAt time.Time) error {
	query := `
		UPDATE analytics.channels SET
			video_count = $2,
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, channelID, totalVideos)
	return err
}
