package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/domain"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/internal/providers"
	"github.com/jaybani/jb_cip/internal/repository"
	"github.com/jaybani/jb_cip/pkg/errors"
	"golang.org/x/oauth2"
)

type SyncService struct {
	syncRepo       *repository.SyncRepository
	integrationRepo *repository.IntegrationRepository
	channelRepo     *repository.ChannelRepository
	provider        *providers.YouTubeSyncProvider
	encryptor       *helper.TokenEncryptor
	cfg             interface{}
}

func NewSyncService(
	syncRepo *repository.SyncRepository,
	integrationRepo *repository.IntegrationRepository,
	channelRepo *repository.ChannelRepository,
	provider *providers.YouTubeSyncProvider,
	encryptor *helper.TokenEncryptor,
	cfg interface{},
) *SyncService {
	return &SyncService{
		syncRepo:        syncRepo,
		integrationRepo: integrationRepo,
		channelRepo:     channelRepo,
		provider:        provider,
		encryptor:       encryptor,
		cfg:             cfg,
	}
}

type SyncResponse struct {
	JobID         string `json:"job_id"`
	Status        string `json:"status"`
	TotalVideos   int    `json:"total_videos"`
	TotalSuccess  int    `json:"total_success"`
	TotalFailed   int    `json:"total_failed"`
	Duration      int    `json:"duration_seconds"`
	ErrorMessage  string `json:"error_message,omitempty"`
	SyncType      string `json:"sync_type"`
	ChannelID     string `json:"channel_id"`
}

type SyncStatusResponse struct {
	ChannelID       string     `json:"channel_id"`
	LastSyncAt      *time.Time `json:"last_sync_at,omitempty"`
	LastSyncStatus  string     `json:"last_sync_status,omitempty"`
	TotalVideos     int        `json:"total_videos"`
	TotalSynced     int        `json:"total_synced"`
	IsSyncing       bool       `json:"is_syncing"`
}

type SyncHistoryResponse struct {
	Jobs []*domain.SyncJob `json:"jobs"`
}

func (s *SyncService) StartSync(userID, workspaceID, channelID string, syncType string) (*SyncResponse, error) {
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

	syncMode := syncType
	var since *time.Time
	if syncMode == "incremental" {
		lastSync, _ := s.syncRepo.GetLastSyncJob(channelUUID)
		if lastSync != nil && lastSync.CompletedAt != nil {
			since = lastSync.CompletedAt
		} else {
			since, _ = s.syncRepo.GetLastVideoPublishedAt(channelUUID)
		}
	} else if syncMode == "" {
		syncMode = "manual"
	}

	job := &domain.SyncJob{
		ID:              uuid.New(),
		ChannelID:       channelUUID,
		UserID:          userUUID,
		WorkspaceID:     workspaceUUID,
		SyncType:        syncMode,
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

	go s.runSync(job, channel.ExternalID, conn, apiToken, since)

	return &SyncResponse{
		JobID:        job.ID.String(),
		Status:       "running",
		SyncType:     syncMode,
		ChannelID:    channelID,
		TotalVideos:  0,
		TotalSuccess: 0,
		TotalFailed:  0,
		Duration:     0,
	}, nil
}

func (s *SyncService) runSync(job *domain.SyncJob, channelExternalID string, conn *domain.APIConnection, apiToken *domain.APIToken, since *time.Time) {
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

	videos, err := s.provider.GetChannelVideos(accessToken, channelExternalID, since)
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = fmt.Sprintf("Failed to fetch videos: %v", err)
		job.DurationSeconds = int(time.Since(startTime).Seconds())
		job.CompletedAt = &[]time.Time{time.Now()}[0]
		s.syncRepo.UpdateSyncJob(job)
		return
	}

	job.TotalVideos = len(videos)
	successCount := 0
	failedCount := 0

	for _, video := range videos {
		domainVideo := s.provider.ConvertToDomainVideo(video, job.ChannelID)
		if err := s.syncRepo.UpsertVideo(domainVideo); err != nil {
			failedCount++
		} else {
			successCount++
		}
	}

	job.Status = "completed"
	job.TotalSuccess = successCount
	job.TotalFailed = failedCount
	job.DurationSeconds = int(time.Since(startTime).Seconds())
	job.CompletedAt = &[]time.Time{time.Now()}[0]
	s.syncRepo.UpdateSyncJob(job)

	_ = s.channelRepo.UpdateSyncStats(job.ChannelID, job.TotalVideos, time.Now())
}

func (s *SyncService) getActiveConnection(userID string) (*domain.APIConnection, *domain.APIToken, error) {
	conn, err := s.integrationRepo.GetConnection(userID, "google")
	if err != nil {
		return nil, nil, errors.New("INTEGRATION_002", "No Google connection found", 404)
	}
	if conn.Status != "active" && conn.Status != "authorized" {
		return nil, nil, errors.New("INTEGRATION_003", "Google connection is not active", 400)
	}
	apiToken, err := s.integrationRepo.GetToken(conn.ID)
	if err != nil {
		return nil, nil, errors.New("INTEGRATION_004", "No token found for connection", 404)
	}
	return conn, apiToken, nil
}

func (s *SyncService) GetSyncStatus(channelID string) (*SyncStatusResponse, error) {
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		return nil, errors.New("SYNC_001", "Invalid channel ID", 400)
	}

	channel, err := s.channelRepo.GetByID(channelUUID)
	if err != nil {
		return nil, errors.New("SYNC_002", "Channel not found", 404)
	}

	lastJob, err := s.syncRepo.GetLastSyncJob(channelUUID)
	if err != nil && err.Error() != "no sync job found" {
		return nil, err
	}

	videoCount, _ := s.syncRepo.GetVideoCount(channelUUID)
	isSyncing := lastJob != nil && lastJob.Status == "running"

	status := "never_synced"
	var lastSyncAt *time.Time
	if lastJob != nil {
		status = lastJob.Status
		if lastJob.CompletedAt != nil {
			lastSyncAt = lastJob.CompletedAt
		} else if lastJob.StartedAt != nil {
			lastSyncAt = lastJob.StartedAt
		}
	}

	return &SyncStatusResponse{
		ChannelID:      channelID,
		LastSyncAt:     lastSyncAt,
		LastSyncStatus: status,
		TotalVideos:    int(channel.VideoCount),
		TotalSynced:    videoCount,
		IsSyncing:      isSyncing,
	}, nil
}

func (s *SyncService) GetSyncHistory(channelID string, limit int) (*SyncHistoryResponse, error) {
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		return nil, errors.New("SYNC_001", "Invalid channel ID", 400)
	}

	if _, err := s.channelRepo.GetByID(channelUUID); err != nil {
		return nil, errors.New("SYNC_002", "Channel not found", 404)
	}

	jobs, err := s.syncRepo.GetSyncJobsByChannel(channelUUID, limit)
	if err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to fetch sync history", 500)
	}

	return &SyncHistoryResponse{Jobs: jobs}, nil
}

func (s *SyncService) RetrySync(userID, workspaceID, jobID string) (*SyncResponse, error) {
	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		return nil, errors.New("SYNC_001", "Invalid job ID", 400)
	}

	job, err := s.syncRepo.GetSyncJobByID(jobUUID)
	if err != nil {
		return nil, errors.New("SYNC_003", "Sync job not found", 404)
	}

	if job.Status != "failed" {
		return nil, errors.New("SYNC_004", "Only failed jobs can be retried", 400)
	}

	return s.StartSync(userID, workspaceID, job.ChannelID.String(), job.SyncType)
}

func (s *SyncService) GetSyncResult(jobID string) (*SyncResponse, error) {
	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		return nil, errors.New("SYNC_001", "Invalid job ID", 400)
	}

	job, err := s.syncRepo.GetSyncJobByID(jobUUID)
	if err != nil {
		return nil, errors.New("SYNC_003", "Sync job not found", 404)
	}

	return &SyncResponse{
		JobID:        job.ID.String(),
		Status:       job.Status,
		TotalVideos:  job.TotalVideos,
		TotalSuccess: job.TotalSuccess,
		TotalFailed:  job.TotalFailed,
		Duration:     job.DurationSeconds,
		ErrorMessage: job.ErrorMessage,
		SyncType:     job.SyncType,
		ChannelID:    job.ChannelID.String(),
	}, nil
}

func (s *SyncService) BuildOAuthToken(accessToken, refreshToken string, expiresAt time.Time) *oauth2.Token {
	return s.provider.BuildOAuthToken(accessToken, refreshToken, expiresAt)
}
