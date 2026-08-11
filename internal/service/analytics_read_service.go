package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/repository"
	"github.com/jaybani/jb_cip/pkg/errors"
)

const (
	dateLayout = "2006-01-02"
)

// AnalyticsReadService handles read-only analytics queries for the dashboard
type AnalyticsReadService struct {
	channelRepo *repository.ChannelRepository
	syncRepo    *repository.SyncRepository
}

func NewAnalyticsReadService(channelRepo *repository.ChannelRepository, syncRepo *repository.SyncRepository) *AnalyticsReadService {
	return &AnalyticsReadService{
		channelRepo: channelRepo,
		syncRepo:    syncRepo,
	}
}

// resolveChannel validates channel ownership and returns the channel
func (s *AnalyticsReadService) resolveChannel(channelID, workspaceID string) (uuid.UUID, error) {
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		return uuid.Nil, errors.New("ANALYTICS_001", "Invalid channel ID", 400)
	}

	channel, err := s.channelRepo.GetByID(channelUUID)
	if err != nil {
		return uuid.Nil, errors.New("ANALYTICS_002", "Channel not found", 404)
	}

	if channel.WorkspaceID != workspaceID {
		return uuid.Nil, errors.New("ANALYTICS_002", "Channel not found", 404)
	}

	return channelUUID, nil
}

// normalizeDateRange validates and returns a date range with defaults
func normalizeDateRange(startDate, endDate string) (string, string, error) {
	now := time.Now()
	if startDate == "" && endDate == "" {
		// Default: last 7 days (including today)
		start := now.AddDate(0, 0, -6)
		return start.Format(dateLayout), now.Format(dateLayout), nil
	}

	if startDate == "" || endDate == "" {
		return "", "", errors.New("ANALYTICS_003", "Both start_date and end_date are required", 400)
	}

	start, err := time.Parse(dateLayout, startDate)
	if err != nil {
		return "", "", errors.New("ANALYTICS_003", "Invalid start_date format (expected YYYY-MM-DD)", 400)
	}
	end, err := time.Parse(dateLayout, endDate)
	if err != nil {
		return "", "", errors.New("ANALYTICS_003", "Invalid end_date format (expected YYYY-MM-DD)", 400)
	}

	if start.After(end) {
		return "", "", errors.New("ANALYTICS_003", "start_date must be before or equal to end_date", 400)
	}

	return startDate, endDate, nil
}

// GetSummary returns aggregated channel analytics
func (s *AnalyticsReadService) GetSummary(channelID, workspaceID, startDate, endDate string) (*repository.ChannelAnalyticsSummary, error) {
	channelUUID, err := s.resolveChannel(channelID, workspaceID)
	if err != nil {
		return nil, err
	}

	start, end, err := normalizeDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	summary, err := s.syncRepo.GetChannelAnalyticsSummary(channelUUID, start, end)
	if err != nil {
		return nil, errors.New("SYSTEM_001", fmt.Sprintf("Failed to load analytics summary: %v", err), 500)
	}

	return summary, nil
}

// GetTimeseries returns daily metric values
func (s *AnalyticsReadService) GetTimeseries(channelID, workspaceID, metric, startDate, endDate string) ([]*repository.TimeseriesPoint, error) {
	channelUUID, err := s.resolveChannel(channelID, workspaceID)
	if err != nil {
		return nil, err
	}

	if metric == "" {
		metric = "views"
	}

	validMetrics := map[string]bool{
		"views": true, "watch_time": true, "likes": true,
		"comments": true, "shares": true, "subscribers": true,
	}
	if !validMetrics[metric] {
		return nil, errors.New("ANALYTICS_003", "Unsupported metric", 400)
	}

	start, end, err := normalizeDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	points, err := s.syncRepo.GetChannelTimeseries(channelUUID, metric, start, end)
	if err != nil {
		return nil, errors.New("SYSTEM_001", fmt.Sprintf("Failed to load timeseries: %v", err), 500)
	}

	return points, nil
}

// GetTopVideos returns top videos by views
func (s *AnalyticsReadService) GetTopVideos(channelID, workspaceID, startDate, endDate string, limit int) ([]*repository.TopVideoRow, error) {
	channelUUID, err := s.resolveChannel(channelID, workspaceID)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 50 {
		limit = 10
	}

	start, end, err := normalizeDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	videos, err := s.syncRepo.GetTopVideos(channelUUID, start, end, limit)
	if err != nil {
		return nil, errors.New("SYSTEM_001", fmt.Sprintf("Failed to load top videos: %v", err), 500)
	}

	return videos, nil
}
