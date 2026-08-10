package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/domain"
)

// UpsertDailyMetric inserts or updates a daily metric record.
// Uses partial unique indexes: video metrics use (channel_id, video_id, date, metric_type),
// channel metrics use (channel_id, date, metric_type) where video_id IS NULL.
func (r *SyncRepository) UpsertDailyMetric(metric *domain.DailyMetric) error {
	now := time.Now()

	if metric.VideoID == nil {
		// Channel metric: conflict on (channel_id, date, metric_type) WHERE video_id IS NULL
		query := `
			INSERT INTO analytics.daily_metrics (
				id, channel_id, video_id, date, metric_type,
				views, watch_time, average_view_duration, average_percentage_viewed,
				impressions, impression_ctr, likes, comments, shares,
				subscribers, returning_viewers, new_viewers,
				sync_job_id, synced_at, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
			ON CONFLICT (channel_id, date, metric_type)
				WHERE video_id IS NULL
			DO UPDATE SET
				views = EXCLUDED.views,
				watch_time = EXCLUDED.watch_time,
				average_view_duration = EXCLUDED.average_view_duration,
				average_percentage_viewed = EXCLUDED.average_percentage_viewed,
				impressions = EXCLUDED.impressions,
				impression_ctr = EXCLUDED.impression_ctr,
				likes = EXCLUDED.likes,
				comments = EXCLUDED.comments,
				shares = EXCLUDED.shares,
				subscribers = EXCLUDED.subscribers,
				returning_viewers = EXCLUDED.returning_viewers,
				new_viewers = EXCLUDED.new_viewers,
				sync_job_id = EXCLUDED.sync_job_id,
				synced_at = NOW(),
				updated_at = NOW()
		`
		_, err := r.db.Exec(query,
			metric.ID, metric.ChannelID, metric.VideoID, metric.Date, metric.MetricType,
			metric.Views, metric.WatchTime, metric.AverageViewDuration, metric.AveragePercentageViewed,
			metric.Impressions, metric.ImpressionCTR, metric.Likes, metric.Comments, metric.Shares,
			metric.Subscribers, metric.ReturningViewers, metric.NewViewers,
			metric.SyncJobID, now, now, now,
		)
		return err
	}

	// Video metric: conflict on (channel_id, video_id, date, metric_type) WHERE video_id IS NOT NULL
	query := `
		INSERT INTO analytics.daily_metrics (
			id, channel_id, video_id, date, metric_type,
			views, watch_time, average_view_duration, average_percentage_viewed,
			impressions, impression_ctr, likes, comments, shares,
			subscribers, returning_viewers, new_viewers,
			sync_job_id, synced_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		ON CONFLICT (channel_id, video_id, date, metric_type)
			WHERE video_id IS NOT NULL
		DO UPDATE SET
			views = EXCLUDED.views,
			watch_time = EXCLUDED.watch_time,
			average_view_duration = EXCLUDED.average_view_duration,
			average_percentage_viewed = EXCLUDED.average_percentage_viewed,
			impressions = EXCLUDED.impressions,
			impression_ctr = EXCLUDED.impression_ctr,
			likes = EXCLUDED.likes,
			comments = EXCLUDED.comments,
			shares = EXCLUDED.shares,
			subscribers = EXCLUDED.subscribers,
			returning_viewers = EXCLUDED.returning_viewers,
			new_viewers = EXCLUDED.new_viewers,
			sync_job_id = EXCLUDED.sync_job_id,
			synced_at = NOW(),
			updated_at = NOW()
	`
	_, err := r.db.Exec(query,
		metric.ID, metric.ChannelID, metric.VideoID, metric.Date, metric.MetricType,
		metric.Views, metric.WatchTime, metric.AverageViewDuration, metric.AveragePercentageViewed,
		metric.Impressions, metric.ImpressionCTR, metric.Likes, metric.Comments, metric.Shares,
		metric.Subscribers, metric.ReturningViewers, metric.NewViewers,
		metric.SyncJobID, now, now, now,
	)
	return err
}

// GetDailyMetricsByChannel gets daily metrics for a channel
func (r *SyncRepository) GetDailyMetricsByChannel(channelID uuid.UUID, limit int) ([]*domain.DailyMetric, error) {
	if limit <= 0 || limit > 1000 {
		limit = 365
	}
	query := `
		SELECT id, channel_id, video_id, date, metric_type,
			views, watch_time, average_view_duration, average_percentage_viewed,
			impressions, impression_ctr, likes, comments, shares,
			subscribers, returning_viewers, new_viewers,
			sync_job_id, synced_at, created_at, updated_at
		FROM analytics.daily_metrics
		WHERE channel_id = $1
		ORDER BY date DESC
		LIMIT $2
	`
	rows, err := r.db.Query(query, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*domain.DailyMetric
	for rows.Next() {
		metric := &domain.DailyMetric{}
		var videoIDStr, syncJobIDStr sql.NullString
		var syncedAt sql.NullTime
		if err := rows.Scan(
			&metric.ID, &metric.ChannelID, &videoIDStr, &metric.Date, &metric.MetricType,
			&metric.Views, &metric.WatchTime, &metric.AverageViewDuration, &metric.AveragePercentageViewed,
			&metric.Impressions, &metric.ImpressionCTR, &metric.Likes, &metric.Comments, &metric.Shares,
			&metric.Subscribers, &metric.ReturningViewers, &metric.NewViewers,
			&syncJobIDStr, &syncedAt,
		); err != nil {
			return nil, err
		}
		if videoIDStr.Valid {
			v, _ := uuid.Parse(videoIDStr.String)
			metric.VideoID = &v
		}
		if syncJobIDStr.Valid {
			j, _ := uuid.Parse(syncJobIDStr.String)
			metric.SyncJobID = &j
		}
		if syncedAt.Valid {
			metric.SyncedAt = &syncedAt.Time
		}
		metrics = append(metrics, metric)
	}
	return metrics, rows.Err()
}

// GetDailyMetricsByChannelAndDateRange gets daily metrics for a channel within date range
func (r *SyncRepository) GetDailyMetricsByChannelAndDateRange(channelID uuid.UUID, startDate, endDate string, limit int) ([]*domain.DailyMetric, error) {
	if limit <= 0 || limit > 1000 {
		limit = 365
	}
	query := `
		SELECT id, channel_id, video_id, date, metric_type,
			views, watch_time, average_view_duration, average_percentage_viewed,
			impressions, impression_ctr, likes, comments, shares,
			subscribers, returning_viewers, new_viewers,
			sync_job_id, synced_at, created_at, updated_at
		FROM analytics.daily_metrics
		WHERE channel_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date DESC
		LIMIT $4
	`
	rows, err := r.db.Query(query, channelID, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*domain.DailyMetric
	for rows.Next() {
		metric := &domain.DailyMetric{}
		var videoIDStr, syncJobIDStr sql.NullString
		var syncedAt sql.NullTime
		if err := rows.Scan(
			&metric.ID, &metric.ChannelID, &videoIDStr, &metric.Date, &metric.MetricType,
			&metric.Views, &metric.WatchTime, &metric.AverageViewDuration, &metric.AveragePercentageViewed,
			&metric.Impressions, &metric.ImpressionCTR, &metric.Likes, &metric.Comments, &metric.Shares,
			&metric.Subscribers, &metric.ReturningViewers, &metric.NewViewers,
			&syncJobIDStr, &syncedAt,
		); err != nil {
			return nil, err
		}
		if videoIDStr.Valid {
			v, _ := uuid.Parse(videoIDStr.String)
			metric.VideoID = &v
		}
		if syncJobIDStr.Valid {
			j, _ := uuid.Parse(syncJobIDStr.String)
			metric.SyncJobID = &j
		}
		if syncedAt.Valid {
			metric.SyncedAt = &syncedAt.Time
		}
		metrics = append(metrics, metric)
	}
	return metrics, rows.Err()
}
