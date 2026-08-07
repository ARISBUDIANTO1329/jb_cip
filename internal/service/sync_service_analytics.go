package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/domain"
	"github.com/jaybani/jb_cip/internal/providers"
	"github.com/jaybani/jb_cip/pkg/errors"
)

// AnalyticsSyncResponse represents the response for analytics sync
type AnalyticsSyncResponse struct {
	JobID         string `json:"job_id"`
	Status        string `json:"status"`
	TotalVideos   int    `json:"total_videos"`
	TotalChannels int    `json:"total_channels"`
	TotalMetrics  int    `json:"total_metrics"`
	TotalSuccess  int    `json:"total_success"`
	TotalFailed   int    `json:"total_failed"`
	Duration      int    `json:"duration_seconds"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

// GetDailyMetrics gets daily metrics for a channel
func (s *SyncService) GetDailyMetrics(channelID string, startDate, endDate string, limit int) ([]*domain.DailyMetric, error) {
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		return nil, errors.New("SYNC_001", "Invalid channel ID", 400)
	}

	if _, err := s.channelRepo.GetByID(channelUUID); err != nil {
		return nil, errors.New("SYNC_002", "Channel not found", 404)
	}

	if limit <= 0 || limit > 1000 {
		limit = 365
	}

	var metrics []*domain.DailyMetric

	// If date range specified, use that, otherwise get latest
	if startDate != "" && endDate != "" {
		metrics, err = s.syncRepo.GetDailyMetricsByChannelAndDateRange(channelUUID, startDate, endDate, limit)
	} else {
		metrics, err = s.syncRepo.GetDailyMetricsByChannel(channelUUID, limit)
	}

	if err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to fetch daily metrics", 500)
	}

	return metrics, nil
}

// SyncAnalytics starts an async analytics sync job
func (s *SyncService) SyncAnalytics(userID, workspaceID, channelID string) (*AnalyticsSyncResponse, error) {
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		return nil, errors.New("SYNC_001", "Invalid channel ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("SYNC_001", "Invalid user ID", 400)
	}

	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("SYNC_001", "Invalid workspace ID", 400)
	}

	// Get channel
	channel, err := s.channelRepo.GetByID(channelUUID)
	if err != nil {
		return nil, errors.New("SYNC_002", "Channel not found", 404)
	}

	// Get Google connection and token
	conn, apiToken, err := s.getActiveConnection(userID)
	if err != nil {
		return nil, err
	}

	// Validate channel is connected to the same Google account
	if channel.ConnectionID == "" {
		return nil, errors.New("SYNC_005", "Channel is not connected to any Google account", 400)
	}
	if channel.ConnectionID != conn.ID {
		return nil, errors.New("SYNC_006", "Channel is not connected to the selected Google account", 400)
	}

	// Create job record
	job := &domain.SyncJob{
		ID:            uuid.New(),
		ChannelID:     channelUUID,
		UserID:        userUUID,
		WorkspaceID:   workspaceUUID,
		SyncType:      "analytics",
		Status:        "running",
		TotalVideos:   0,
		TotalSuccess:  0,
		TotalFailed:   0,
		DurationSeconds: 0,
		StartedAt:     &[]time.Time{time.Now()}[0],
		CreatedAt:     time.Now(),
	}
	if err := s.syncRepo.CreateSyncJob(job); err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to create sync job", 500)
	}

	// Run sync in goroutine
	go s.runAnalyticsSync(job, channel.ExternalID, conn, apiToken)

	return &AnalyticsSyncResponse{
		JobID:        job.ID.String(),
		Status:       "running",
		TotalVideos:  0,
		TotalChannels: 1,
		TotalMetrics:  0,
		TotalSuccess: 0,
		TotalFailed:  0,
		Duration:     0,
	}, nil
}

func (s *SyncService) runAnalyticsSync(job *domain.SyncJob, channelExternalID string, conn *domain.APIConnection, apiToken *domain.APIToken) {
	startTime := time.Now()

	// Decrypt access token
	accessToken, err := s.encryptor.Decrypt(apiToken.AccessTokenEncrypted)
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = "Failed to decrypt access token"
		job.DurationSeconds = int(time.Since(startTime).Seconds())
		job.CompletedAt = &[]time.Time{time.Now()}[0]
		s.syncRepo.UpdateSyncJob(job)
		return
	}

	// Create YouTube Analytics Provider
	ytAnalytics := providers.NewYouTubeAnalyticsProvider(s.provider.GoogleProvider)

	// Sync channel metrics
	channelMetrics, err := ytAnalytics.GetChannelMetrics(accessToken, channelExternalID, "7daysAgo", "today")
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = fmt.Sprintf("Failed to fetch channel metrics: %v", err)
		job.DurationSeconds = int(time.Since(startTime).Seconds())
		job.CompletedAt = &[]time.Time{time.Now()}[0]
		s.syncRepo.UpdateSyncJob(job)
		return
	}

	// Convert and save channel metrics
	channelMetricsDomain := ytAnalytics.ConvertToChannelMetrics(job.ChannelID, channelMetrics, &job.ID)
	channelSuccess := 0
	channelFailed := 0
	for _, m := range channelMetricsDomain {
		if err := s.syncRepo.UpsertDailyMetric(m); err != nil {
			channelFailed++
		} else {
			channelSuccess++
		}
	}

	// Get all videos for this channel
	videos, err := s.syncRepo.GetVideosByChannel(job.ChannelID, 500)
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = fmt.Sprintf("Failed to fetch videos: %v", err)
		job.DurationSeconds = int(time.Since(startTime).Seconds())
		job.CompletedAt = &[]time.Time{time.Now()}[0]
		s.syncRepo.UpdateSyncJob(job)
		return
	}

	videoSuccess := 0
	videoFailed := 0

	// Sync metrics for each video
	for _, video := range videos {
		videoMetrics, err := ytAnalytics.GetVideoMetrics(accessToken, video.VideoID, "7daysAgo", "today")
		if err != nil {
			videoFailed++
			continue
		}

		videoMetricsDomain := ytAnalytics.ConvertToDailyMetrics(video.ID, videoMetrics, &job.ID)
		for _, m := range videoMetricsDomain {
			if err := s.syncRepo.UpsertDailyMetric(m); err != nil {
				videoFailed++
			} else {
				videoSuccess++
			}
		}
	}

	job.Status = "completed"
	job.TotalSuccess = channelSuccess + videoSuccess
	job.TotalFailed = channelFailed + videoFailed
	job.DurationSeconds = int(time.Since(startTime).Seconds())
	job.CompletedAt = &[]time.Time{time.Now()}[0]
	s.syncRepo.UpdateSyncJob(job)
}
