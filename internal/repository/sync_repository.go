package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/domain"
)

type SyncRepository struct {
	db *sql.DB
}

func NewSyncRepository(db *sql.DB) *SyncRepository {
	return &SyncRepository{db: db}
}

// CreateSyncJob creates a new sync job
func (r *SyncRepository) CreateSyncJob(job *domain.SyncJob) error {
	query := `
		INSERT INTO analytics.sync_jobs (
			id, channel_id, user_id, workspace_id, sync_type, status,
			total_videos, total_success, total_failed, duration_seconds,
			error_message, started_at, completed_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.Exec(query,
		job.ID, job.ChannelID, job.UserID, job.WorkspaceID,
		job.SyncType, job.Status, job.TotalVideos, job.TotalSuccess,
		job.TotalFailed, job.DurationSeconds, job.ErrorMessage,
		job.StartedAt, job.CompletedAt, job.CreatedAt,
	)
	return err
}

// UpdateSyncJob updates a sync job status
func (r *SyncRepository) UpdateSyncJob(job *domain.SyncJob) error {
	query := `
		UPDATE analytics.sync_jobs SET
			status = $2,
			total_videos = $3,
			total_success = $4,
			total_failed = $5,
			duration_seconds = $6,
			error_message = $7,
			completed_at = $8
		WHERE id = $1
	`
	_, err := r.db.Exec(query,
		job.ID, job.Status, job.TotalVideos, job.TotalSuccess,
		job.TotalFailed, job.DurationSeconds, job.ErrorMessage,
		job.CompletedAt,
	)
	return err
}

// GetSyncJobByID gets a sync job by ID
func (r *SyncRepository) GetSyncJobByID(id uuid.UUID) (*domain.SyncJob, error) {
	query := `
		SELECT id, channel_id, user_id, workspace_id, sync_type, status,
			total_videos, total_success, total_failed, duration_seconds,
			error_message, started_at, completed_at, created_at
		FROM analytics.sync_jobs
		WHERE id = $1
	`
	job := &domain.SyncJob{}
	var startedAt, completedAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&job.ID, &job.ChannelID, &job.UserID, &job.WorkspaceID,
		&job.SyncType, &job.Status, &job.TotalVideos, &job.TotalSuccess,
		&job.TotalFailed, &job.DurationSeconds, &job.ErrorMessage,
		&startedAt, &completedAt, &job.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("sync job not found")
	}
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}
	return job, nil
}

// GetSyncJobsByChannel gets sync history for a channel
func (r *SyncRepository) GetSyncJobsByChannel(channelID uuid.UUID, limit int) ([]*domain.SyncJob, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := `
		SELECT id, channel_id, user_id, workspace_id, sync_type, status,
			total_videos, total_success, total_failed, duration_seconds,
			error_message, started_at, completed_at, created_at
		FROM analytics.sync_jobs
		WHERE channel_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(query, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*domain.SyncJob
	for rows.Next() {
		job := &domain.SyncJob{}
		var startedAt, completedAt sql.NullTime
		if err := rows.Scan(
			&job.ID, &job.ChannelID, &job.UserID, &job.WorkspaceID,
			&job.SyncType, &job.Status, &job.TotalVideos, &job.TotalSuccess,
			&job.TotalFailed, &job.DurationSeconds, &job.ErrorMessage,
			&startedAt, &completedAt, &job.CreatedAt,
		); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

// GetLastSyncJob gets the latest sync job for a channel
func (r *SyncRepository) GetLastSyncJob(channelID uuid.UUID) (*domain.SyncJob, error) {
	query := `
		SELECT id, channel_id, user_id, workspace_id, sync_type, status,
			total_videos, total_success, total_failed, duration_seconds,
			error_message, started_at, completed_at, created_at
		FROM analytics.sync_jobs
		WHERE channel_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	job := &domain.SyncJob{}
	var startedAt, completedAt sql.NullTime
	err := r.db.QueryRow(query, channelID).Scan(
		&job.ID, &job.ChannelID, &job.UserID, &job.WorkspaceID,
		&job.SyncType, &job.Status, &job.TotalVideos, &job.TotalSuccess,
		&job.TotalFailed, &job.DurationSeconds, &job.ErrorMessage,
		&startedAt, &completedAt, &job.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no sync job found")
	}
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}
	return job, nil
}

// UpsertVideo inserts or updates a video
func (r *SyncRepository) UpsertVideo(video *domain.YouTubeVideo) error {
	query := `
		INSERT INTO analytics.youtube_videos (
			id, channel_id, video_id, title, description, published_at,
			duration, thumbnail_url, privacy_status, view_count,
			like_count, comment_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		ON CONFLICT (channel_id, video_id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			published_at = EXCLUDED.published_at,
			duration = EXCLUDED.duration,
			thumbnail_url = EXCLUDED.thumbnail_url,
			privacy_status = EXCLUDED.privacy_status,
			view_count = EXCLUDED.view_count,
			like_count = EXCLUDED.like_count,
			comment_count = EXCLUDED.comment_count,
			updated_at = NOW()
	`
	_, err := r.db.Exec(query,
		video.ID, video.ChannelID, video.VideoID, video.Title,
		video.Description, video.PublishedAt, video.Duration,
		video.ThumbnailURL, video.PrivacyStatus, video.ViewCount,
		video.LikeCount, video.CommentCount,
	)
	return err
}

// GetVideosByChannel gets all videos for a channel
func (r *SyncRepository) GetVideosByChannel(channelID uuid.UUID, limit int) ([]*domain.YouTubeVideo, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := `
		SELECT id, channel_id, video_id, title, description, published_at,
			duration, thumbnail_url, privacy_status, view_count,
			like_count, comment_count, created_at, updated_at
		FROM analytics.youtube_videos
		WHERE channel_id = $1
		ORDER BY published_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(query, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*domain.YouTubeVideo
	for rows.Next() {
		video := &domain.YouTubeVideo{}
		var publishedAt sql.NullTime
		if err := rows.Scan(
			&video.ID, &video.ChannelID, &video.VideoID, &video.Title,
			&video.Description, &publishedAt, &video.Duration,
			&video.ThumbnailURL, &video.PrivacyStatus, &video.ViewCount,
			&video.LikeCount, &video.CommentCount, &video.CreatedAt, &video.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if publishedAt.Valid {
			video.PublishedAt = &publishedAt.Time
		}
		videos = append(videos, video)
	}
	return videos, rows.Err()
}

// GetVideoCount gets total videos for a channel
func (r *SyncRepository) GetVideoCount(channelID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM analytics.youtube_videos WHERE channel_id = $1`
	err := r.db.QueryRow(query, channelID).Scan(&count)
	return count, err
}

// GetLastVideoPublishedAt gets the most recent video published_at for a channel
func (r *SyncRepository) GetLastVideoPublishedAt(channelID uuid.UUID) (*time.Time, error) {
	var publishedAt sql.NullTime
	query := `
		SELECT MAX(published_at) FROM analytics.youtube_videos
		WHERE channel_id = $1
	`
	if err := r.db.QueryRow(query, channelID).Scan(&publishedAt); err != nil {
		return nil, err
	}
	if !publishedAt.Valid {
		return nil, nil
	}
	return &publishedAt.Time, nil
}

// GetVideoByID gets video by video_id
func (r *SyncRepository) GetVideoByID(channelID uuid.UUID, videoID string) (*domain.YouTubeVideo, error) {
	query := `
		SELECT id, channel_id, video_id, title, description, published_at,
			duration, thumbnail_url, privacy_status, view_count,
			like_count, comment_count, created_at, updated_at
		FROM analytics.youtube_videos
		WHERE channel_id = $1 AND video_id = $2
	`
	video := &domain.YouTubeVideo{}
	var publishedAt sql.NullTime
	err := r.db.QueryRow(query, channelID, videoID).Scan(
		&video.ID, &video.ChannelID, &video.VideoID, &video.Title,
		&video.Description, &publishedAt, &video.Duration,
		&video.ThumbnailURL, &video.PrivacyStatus, &video.ViewCount,
		&video.LikeCount, &video.CommentCount, &video.CreatedAt, &video.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("video not found")
	}
	if err != nil {
		return nil, err
	}
	if publishedAt.Valid {
		video.PublishedAt = &publishedAt.Time
	}
	return video, nil
}

// GetVideosByChannelWithPagination gets videos with pagination and total count
func (r *SyncRepository) GetVideosByChannelWithPagination(channelID uuid.UUID, limit, offset int) ([]*domain.YouTubeVideo, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM analytics.youtube_videos WHERE channel_id = $1`
	if err := r.db.QueryRow(countQuery, channelID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count videos: %w", err)
	}

	query := `
		SELECT id, channel_id, video_id, title, description, published_at,
			duration, thumbnail_url, privacy_status, view_count,
			like_count, comment_count, created_at, updated_at
		FROM analytics.youtube_videos
		WHERE channel_id = $1
		ORDER BY published_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(query, channelID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query videos: %w", err)
	}
	defer rows.Close()

	var videos []*domain.YouTubeVideo
	for rows.Next() {
		video := &domain.YouTubeVideo{}
		var publishedAt sql.NullTime
		if err := rows.Scan(
			&video.ID, &video.ChannelID, &video.VideoID, &video.Title,
			&video.Description, &publishedAt, &video.Duration,
			&video.ThumbnailURL, &video.PrivacyStatus, &video.ViewCount,
			&video.LikeCount, &video.CommentCount, &video.CreatedAt, &video.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan video: %w", err)
		}
		if publishedAt.Valid {
			video.PublishedAt = &publishedAt.Time
		}
		videos = append(videos, video)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return videos, total, nil
}
