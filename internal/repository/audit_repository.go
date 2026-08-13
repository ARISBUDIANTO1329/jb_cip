package repository

import (
	"database/sql"
	"time"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

type VideoAuditData struct {
	VideoID          string
	Title            string
	Description      string
	Duration         int
	ThumbnailURL     string
	ViewCount        int64
	LikeCount        int64
	CommentCount     int64
	PublishedAt      time.Time
	AvgCTR           float64
	AvgAVD           float64
	AvgRetention     float64
	TotalImpressions int64
	TotalViews       int64
	TotalShares      int64
}

type ChannelAuditData struct {
	ChannelID           string
	VideoCount          int
	TotalViews          int64
	AvgViewsPerVideo    float64
	LastUploadDate      time.Time
	DaysSinceLastUpload int
}

func (r *AuditRepository) GetVideoAuditData(videoID string, workspaceID string) (*VideoAuditData, error) {
	query := `
		SELECT
			v.video_id,
			COALESCE(v.title, '') as title,
			COALESCE(v.description, '') as description,
			COALESCE(v.duration, 0) as duration,
			COALESCE(v.thumbnail_url, '') as thumbnail_url,
			COALESCE(v.view_count, 0) as view_count,
			COALESCE(v.like_count, 0) as like_count,
			COALESCE(v.comment_count, 0) as comment_count,
			v.published_at,
			COALESCE(AVG(dm.impression_ctr), 0) as avg_ctr,
			COALESCE(AVG(dm.average_view_duration), 0) as avg_avd,
			COALESCE(AVG(dm.average_percentage_viewed), 0) as avg_retention,
			COALESCE(SUM(dm.impressions), 0) as total_impressions,
			COALESCE(SUM(dm.views), 0) as total_views,
			COALESCE(SUM(dm.shares), 0) as total_shares
		FROM analytics.youtube_videos v
		INNER JOIN analytics.channels c ON v.channel_id = c.id
		LEFT JOIN analytics.daily_metrics dm ON v.id = dm.video_id AND dm.metric_type = 'video'
		WHERE v.video_id = $1 AND c.workspace_id = $2
		GROUP BY v.id, v.video_id, v.title, v.description, v.duration, v.thumbnail_url,
		         v.view_count, v.like_count, v.comment_count, v.published_at
	`

	var data VideoAuditData
	err := r.db.QueryRow(query, videoID, workspaceID).Scan(
		&data.VideoID,
		&data.Title,
		&data.Description,
		&data.Duration,
		&data.ThumbnailURL,
		&data.ViewCount,
		&data.LikeCount,
		&data.CommentCount,
		&data.PublishedAt,
		&data.AvgCTR,
		&data.AvgAVD,
		&data.AvgRetention,
		&data.TotalImpressions,
		&data.TotalViews,
		&data.TotalShares,
	)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *AuditRepository) GetChannelAuditData(channelID string, workspaceID string) (*ChannelAuditData, error) {
	query := `
		SELECT
			c.external_id,
			COUNT(v.id)::int as video_count,
			COALESCE(SUM(v.view_count), 0) as total_views,
			COALESCE(AVG(v.view_count), 0) as avg_views_per_video,
			COALESCE(MAX(v.published_at), NOW()) as last_upload_date,
			COALESCE(EXTRACT(DAY FROM NOW() - MAX(v.published_at))::int, 0) as days_since_last_upload
		FROM analytics.channels c
		LEFT JOIN analytics.youtube_videos v ON c.id = v.channel_id
		WHERE c.id = $1 AND c.workspace_id = $2
		GROUP BY c.id, c.external_id
	`

	var data ChannelAuditData
	err := r.db.QueryRow(query, channelID, workspaceID).Scan(
		&data.ChannelID,
		&data.VideoCount,
		&data.TotalViews,
		&data.AvgViewsPerVideo,
		&data.LastUploadDate,
		&data.DaysSinceLastUpload,
	)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *AuditRepository) GetChannelVideosAuditData(channelID string, workspaceID string, limit int) ([]VideoAuditData, error) {
	query := `
		SELECT
			v.video_id,
			COALESCE(v.title, '') as title,
			COALESCE(v.description, '') as description,
			COALESCE(v.duration, 0) as duration,
			COALESCE(v.thumbnail_url, '') as thumbnail_url,
			COALESCE(v.view_count, 0) as view_count,
			COALESCE(v.like_count, 0) as like_count,
			COALESCE(v.comment_count, 0) as comment_count,
			v.published_at,
			COALESCE(AVG(dm.impression_ctr), 0) as avg_ctr,
			COALESCE(AVG(dm.average_view_duration), 0) as avg_avd,
			COALESCE(AVG(dm.average_percentage_viewed), 0) as avg_retention,
			COALESCE(SUM(dm.impressions), 0) as total_impressions,
			COALESCE(SUM(dm.views), 0) as total_views,
			COALESCE(SUM(dm.shares), 0) as total_shares
		FROM analytics.youtube_videos v
		INNER JOIN analytics.channels c ON v.channel_id = c.id
		LEFT JOIN analytics.daily_metrics dm ON v.id = dm.video_id AND dm.metric_type = 'video'
		WHERE c.id = $1 AND c.workspace_id = $2
		GROUP BY v.id, v.video_id, v.title, v.description, v.duration, v.thumbnail_url,
		         v.view_count, v.like_count, v.comment_count, v.published_at
		ORDER BY v.published_at DESC
		LIMIT $3
	`

	rows, err := r.db.Query(query, channelID, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []VideoAuditData
	for rows.Next() {
		var data VideoAuditData
		err := rows.Scan(
			&data.VideoID,
			&data.Title,
			&data.Description,
			&data.Duration,
			&data.ThumbnailURL,
			&data.ViewCount,
			&data.LikeCount,
			&data.CommentCount,
			&data.PublishedAt,
			&data.AvgCTR,
			&data.AvgAVD,
			&data.AvgRetention,
			&data.TotalImpressions,
			&data.TotalViews,
			&data.TotalShares,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, data)
	}

	return videos, nil
}
