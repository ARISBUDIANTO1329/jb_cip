package repository

import (
	"fmt"

	"github.com/google/uuid"
)

// ChannelAnalyticsSummary holds aggregated channel metrics
type ChannelAnalyticsSummary struct {
	ChannelID   uuid.UUID `json:"channel_id"`
	Views       int64     `json:"views"`
	WatchTime   int64     `json:"watch_time"`
	Likes       int64     `json:"likes"`
	Comments    int64     `json:"comments"`
	Shares      int64     `json:"shares"`
	Subscribers int64     `json:"subscribers"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
}

// TimeseriesPoint holds one data point for a date
type TimeseriesPoint struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
}

// TopVideoRow holds aggregated video metrics joined with video metadata
type TopVideoRow struct {
	VideoID        string    `json:"video_id"`
	InternalVideoID uuid.UUID `json:"internal_video_id"`
	Title          string    `json:"title"`
	ThumbnailURL   string    `json:"thumbnail_url"`
	Views          int64     `json:"views"`
	Likes          int64     `json:"likes"`
	Comments       int64     `json:"comments"`
	WatchTime      int64     `json:"watch_time"`
}

// GetChannelAnalyticsSummary aggregates channel metrics for a date range
func (r *SyncRepository) GetChannelAnalyticsSummary(channelID uuid.UUID, startDate, endDate string) (*ChannelAnalyticsSummary, error) {
	query := `
		SELECT
			COALESCE(SUM(views), 0),
			COALESCE(SUM(watch_time), 0),
			COALESCE(SUM(likes), 0),
			COALESCE(SUM(comments), 0),
			COALESCE(SUM(shares), 0),
			COALESCE(SUM(subscribers), 0)
		FROM analytics.daily_metrics
		WHERE channel_id = $1
			AND metric_type = 'channel'
			AND video_id IS NULL
			AND date >= $2
			AND date <= $3
	`
	summary := &ChannelAnalyticsSummary{
		ChannelID: channelID,
		StartDate: startDate,
		EndDate:   endDate,
	}
	err := r.db.QueryRow(query, channelID, startDate, endDate).Scan(
		&summary.Views, &summary.WatchTime, &summary.Likes,
		&summary.Comments, &summary.Shares, &summary.Subscribers,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate channel analytics: %w", err)
	}
	return summary, nil
}

// GetChannelTimeseries returns daily aggregated values for a given metric
func (r *SyncRepository) GetChannelTimeseries(channelID uuid.UUID, metric, startDate, endDate string) ([]*TimeseriesPoint, error) {
	var column string
	switch metric {
	case "views":
		column = "views"
	case "watch_time":
		column = "watch_time"
	case "likes":
		column = "likes"
	case "comments":
		column = "comments"
	case "shares":
		column = "shares"
	case "subscribers":
		column = "subscribers"
	default:
		return nil, fmt.Errorf("unsupported metric: %s", metric)
	}

	query := fmt.Sprintf(`
		SELECT TO_CHAR(date, 'YYYY-MM-DD') AS date, COALESCE(SUM(%s), 0)
		FROM analytics.daily_metrics
		WHERE channel_id = $1
			AND metric_type = 'channel'
			AND video_id IS NULL
			AND date >= $2
			AND date <= $3
		GROUP BY date
		ORDER BY date ASC
	`, column)

	rows, err := r.db.Query(query, channelID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query timeseries: %w", err)
	}
	defer rows.Close()

	var points []*TimeseriesPoint
	for rows.Next() {
		point := &TimeseriesPoint{}
		if err := rows.Scan(&point.Date, &point.Value); err != nil {
			return nil, fmt.Errorf("failed to scan timeseries: %w", err)
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return points, nil
}

// GetTopVideos returns top videos by views within a date range
func (r *SyncRepository) GetTopVideos(channelID uuid.UUID, startDate, endDate string, limit int) ([]*TopVideoRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	query := `
		SELECT
			v.video_id,
			v.id AS internal_video_id,
			COALESCE(v.title, ''),
			COALESCE(v.thumbnail_url, ''),
			COALESCE(SUM(dm.views), 0),
			COALESCE(SUM(dm.likes), 0),
			COALESCE(SUM(dm.comments), 0),
			COALESCE(SUM(dm.watch_time), 0)
		FROM analytics.daily_metrics dm
		JOIN analytics.youtube_videos v ON v.id = dm.video_id
		WHERE dm.channel_id = $1
			AND dm.metric_type = 'video'
			AND dm.video_id IS NOT NULL
			AND dm.date >= $2
			AND dm.date <= $3
		GROUP BY v.video_id, v.id, v.title, v.thumbnail_url
		ORDER BY SUM(dm.views) DESC
		LIMIT $4
	`

	rows, err := r.db.Query(query, channelID, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top videos: %w", err)
	}
	defer rows.Close()

	var videos []*TopVideoRow
	for rows.Next() {
		row := &TopVideoRow{}
		if err := rows.Scan(
			&row.VideoID,
			&row.InternalVideoID,
			&row.Title,
			&row.ThumbnailURL,
			&row.Views,
			&row.Likes,
			&row.Comments,
			&row.WatchTime,
		); err != nil {
			return nil, fmt.Errorf("failed to scan top video: %w", err)
		}
		videos = append(videos, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return videos, nil
}
