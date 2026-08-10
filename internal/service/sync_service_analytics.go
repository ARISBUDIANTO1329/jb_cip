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

	channel, err := s.channelRepo.GetByID(channelUUID)
	if err != nil {
		return nil, errors.New("SYNC_002", "Channel not found", 404)
	}

	conn, apiToken, err := s.getActiveConnection(userID)
	if err != nil {
		return nil, err
	}

	if channel.ConnectionID == "" {
		return nil, errors.New("SYNC_005", "Channel is not connected to any Google account", 400)
	}
	if channel.ConnectionID != conn.ID {
		return nil, errors.New("SYNC_006", "Channel is not connected to the selected Google account", 400)
	}

	job := &domain.SyncJob{
		ID:              uuid.New(),
		ChannelID:       channelUUID,
		UserID:          userUUID,
		WorkspaceID:     workspaceUUID,
		SyncType:        "analytics",
		Status:          "running",
		TotalVideos:     0,
		TotalSuccess:    0,
		TotalFailed:     0,
		DurationSeconds: 0,
		StartedAt:       &[]time.Time{time.Now()}[0],
		CreatedAt:       time.Now(),
	}
	if err := s.syncRepo.CreateSyncJob(job); err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to create sync job", 500)
	}

	go s.runAnalyticsSync(job, channel.ExternalID, conn, apiToken)

	return &AnalyticsSyncResponse{
		JobID:         job.ID.String(),
		Status:        "running",
		TotalVideos:   0,
		TotalChannels: 1,
		TotalMetrics:  0,
		TotalSuccess:  0,
		TotalFailed:   0,
		Duration:      0,
	}, nil
}

func (s *SyncService) runAnalyticsSync(job *domain.SyncJob, channelExternalID string, conn *domain.APIConnection, apiToken *domain.APIToken) {
	startTime := time.Now()

	accessToken, err := s.encryptor.Decrypt(apiToken.AccessTokenEncrypted)
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = "Failed to decrypt access token"
		job.DurationSeconds = int(time.Since(startTime).Seconds())
		job.CompletedAt = &[]time.Time{time.Now()}[0]
		s.syncRepo.UpdateSyncJob(job)
		return
	}

	// Refresh token if expired or about to expire (within 5 minutes)
	if apiToken.AccessTokenExpiresAt != nil && apiToken.AccessTokenExpiresAt.Before(time.Now().Add(5*time.Minute)) {
		refreshToken, err := s.encryptor.Decrypt(apiToken.RefreshTokenEncrypted)
		if err != nil {
			job.Status = "failed"
			job.ErrorMessage = "Failed to decrypt refresh token"
			job.DurationSeconds = int(time.Since(startTime).Seconds())
			job.CompletedAt = &[]time.Time{time.Now()}[0]
			s.syncRepo.UpdateSyncJob(job)
			return
		}

		newToken, err := s.provider.GoogleProvider.RefreshToken(refreshToken)
		if err != nil {
			job.Status = "failed"
			job.ErrorMessage = "Failed to refresh token"
			job.DurationSeconds = int(time.Since(startTime).Seconds())
			job.CompletedAt = &[]time.Time{time.Now()}[0]
			s.syncRepo.UpdateSyncJob(job)
			return
		}

		accessToken = newToken.AccessToken
		encryptedAccess, err := s.encryptor.Encrypt(newToken.AccessToken)
		if err != nil {
			job.Status = "failed"
			job.ErrorMessage = "Failed to encrypt new access token"
			job.DurationSeconds = int(time.Since(startTime).Seconds())
			job.CompletedAt = &[]time.Time{time.Now()}[0]
			s.syncRepo.UpdateSyncJob(job)
			return
		}

		apiToken.AccessTokenEncrypted = encryptedAccess
		apiToken.AccessTokenExpiresAt = &newToken.Expiry
		if err := s.integrationRepo.UpdateToken(apiToken); err != nil {
			job.Status = "failed"
			job.ErrorMessage = "Failed to update token in database"
			job.DurationSeconds = int(time.Since(startTime).Seconds())
			job.CompletedAt = &[]time.Time{time.Now()}[0]
			s.syncRepo.UpdateSyncJob(job)
			return
		}
	}

	ytAnalytics := providers.NewYouTubeAnalyticsProvider(s.provider.GoogleProvider)

	// Channel metrics (always 1 call)
	channelMetrics, err := ytAnalytics.GetChannelMetrics(accessToken, channelExternalID, "7daysAgo", "today")
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = fmt.Sprintf("Failed to fetch channel metrics: %v", err)
		job.DurationSeconds = int(time.Since(startTime).Seconds())
		job.CompletedAt = &[]time.Time{time.Now()}[0]
		s.syncRepo.UpdateSyncJob(job)
		return
	}

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

	// Get videos for this channel
	videos, err := s.syncRepo.GetVideosByChannel(job.ChannelID, 500)
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = fmt.Sprintf("Failed to fetch videos: %v", err)
		job.DurationSeconds = int(time.Since(startTime).Seconds())
		job.CompletedAt = &[]time.Time{time.Now()}[0]
		s.syncRepo.UpdateSyncJob(job)
		return
	}

	job.TotalVideos = len(videos)
	videoIDMap := make(map[string]uuid.UUID, len(videos))
	youtubeVideoIDs := make([]string, 0, len(videos))
	for _, v := range videos {
		videoIDMap[v.VideoID] = v.ID
		youtubeVideoIDs = append(youtubeVideoIDs, v.VideoID)
	}

	videoSuccess := 0
	videoFailed := 0
	batchFailed := 0

	// Batch video metrics: max 50 IDs per request
	for i := 0; i < len(youtubeVideoIDs); i += 50 {
		end := i + 50
		if end > len(youtubeVideoIDs) {
			end = len(youtubeVideoIDs)
		}
		batch := youtubeVideoIDs[i:end]

		batchMetrics, err := ytAnalytics.GetVideoMetricsBatch(accessToken, channelExternalID, batch, "7daysAgo", "today")
		if err != nil {
			batchFailed++
			videoFailed += len(batch)
			job.ErrorMessage = fmt.Sprintf("Batch %d failed (%d videos): %v", batchFailed, len(batch), err)
			// Continue to next batch rather than failing the entire job
			continue
		}

		domainMetrics := ytAnalytics.ConvertToDailyMetrics(job.ChannelID, videoIDMap, batchMetrics, &job.ID)
		for _, m := range domainMetrics {
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
