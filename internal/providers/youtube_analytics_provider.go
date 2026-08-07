package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/domain"
)

type YouTubeAnalyticsProvider struct {
	googleProvider *GoogleProvider
}

type YouTubeVideoMetric struct {
	Date                string  `json:"date"`
	Views               int64   `json:"views"`
	WatchTime           int64   `json:"watchTime"`
	AverageViewDuration float64 `json:"averageViewDuration"`
	PercentageViewed    float64 `json:"percentageViewed"`
	Impressions         int64   `json:"impressions"`
	ImpressionCTR       float64 `json:"impressionCTR"`
	Likes               int64   `json:"likes"`
	Comments            int64   `json:"comments"`
	Shares              int64   `json:"shares"`
}

type YouTubeChannelMetric struct {
	Date             string  `json:"date"`
	Subscribers      int64   `json:"subscribers"`
	Views            int64   `json:"views"`
	WatchTime        int64   `json:"watchTime"`
	ReturningViewers int64   `json:"returningViewers"`
	NewViewers       int64   `json:"newViewers"`
}

func NewYouTubeAnalyticsProvider(gp *GoogleProvider) *YouTubeAnalyticsProvider {
	return &YouTubeAnalyticsProvider{googleProvider: gp}
}

// GetVideoMetrics fetches YouTube Analytics for a specific video
func (p *YouTubeAnalyticsProvider) GetVideoMetrics(accessToken, videoID string, startDate, endDate string) ([]*YouTubeVideoMetric, error) {
	metrics := "views,watchTime,averageViewDuration,percentageViewed,impressions,impressionCTR,likes,comments,shares"

	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/analytics/v1/query?ids=channel==%s&metrics=%s&dimensions=date&start-date=%s&end-date=%s&filters=video==%s&access_token=%s",
		videoID, metrics, startDate, endDate, videoID, accessToken,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video metrics: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Rows [][]string `json:"rows"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, fmt.Errorf("YouTube Analytics API error: %s", result.Error.Message)
	}

	var metricsList []*YouTubeVideoMetric
	for _, row := range result.Rows {
		if len(row) < 10 {
			continue
		}

		views, _ := strconv.ParseInt(row[1], 10, 64)
		watchTime, _ := strconv.ParseInt(row[2], 10, 64)
		avgDuration, _ := strconv.ParseFloat(row[3], 64)
		pctViewed, _ := strconv.ParseFloat(row[4], 64)
		impressions, _ := strconv.ParseInt(row[5], 10, 64)
		ctr, _ := strconv.ParseFloat(row[6], 64)
		likes, _ := strconv.ParseInt(row[7], 10, 64)
		comments, _ := strconv.ParseInt(row[8], 10, 64)
		shares, _ := strconv.ParseInt(row[9], 10, 64)

		metricsList = append(metricsList, &YouTubeVideoMetric{
			Date:                row[0],
			Views:               views,
			WatchTime:           watchTime,
			AverageViewDuration: avgDuration,
			PercentageViewed:    pctViewed,
			Impressions:         impressions,
			ImpressionCTR:       ctr,
			Likes:               likes,
			Comments:            comments,
			Shares:              shares,
		})
	}

	return metricsList, nil
}

// GetChannelMetrics fetches YouTube Analytics for a channel
func (p *YouTubeAnalyticsProvider) GetChannelMetrics(accessToken, channelID string, startDate, endDate string) ([]*YouTubeChannelMetric, error) {
	metrics := "subscribers,views,watchTime,returningViewerCount,newViewerCount"

	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/analytics/v1/query?ids=channel==%s&metrics=%s&dimensions=date&start-date=%s&end-date=%s&access_token=%s",
		channelID, metrics, startDate, endDate, accessToken,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel metrics: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Rows [][]string `json:"rows"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, fmt.Errorf("YouTube Analytics API error: %s", result.Error.Message)
	}

	var metricsList []*YouTubeChannelMetric
	for _, row := range result.Rows {
		if len(row) < 6 {
			continue
		}

		subscribers, _ := strconv.ParseInt(row[1], 10, 64)
		views, _ := strconv.ParseInt(row[2], 10, 64)
		watchTime, _ := strconv.ParseInt(row[3], 10, 64)
		returningViewers, _ := strconv.ParseInt(row[4], 10, 64)
		newViewers, _ := strconv.ParseInt(row[5], 10, 64)

		metricsList = append(metricsList, &YouTubeChannelMetric{
			Date:             row[0],
			Subscribers:      subscribers,
			Views:            views,
			WatchTime:        watchTime,
			ReturningViewers: returningViewers,
			NewViewers:       newViewers,
		})
	}

	return metricsList, nil
}

// ConvertToDailyMetrics converts YouTube Video Metrics to domain model
func (p *YouTubeAnalyticsProvider) ConvertToDailyMetrics(
	videoID uuid.UUID,
	metricsList []*YouTubeVideoMetric,
	syncJobID *uuid.UUID,
) []*domain.DailyMetric {
	var metrics []*domain.DailyMetric
	now := time.Now()

	for _, m := range metricsList {
		metrics = append(metrics, &domain.DailyMetric{
			ID:                    uuid.New(),
			ChannelID:             videoID,
			VideoID:               &videoID,
			Date:                  m.Date,
			MetricType:            "video",
			Views:                 m.Views,
			WatchTime:             m.WatchTime,
			AverageViewDuration:   m.AverageViewDuration,
			AveragePercentageViewed: m.PercentageViewed,
			Impressions:           m.Impressions,
			ImpressionCTR:         m.ImpressionCTR,
			Likes:                 m.Likes,
			Comments:              m.Comments,
			Shares:                m.Shares,
			SyncJobID:             syncJobID,
			SyncedAt:              &now,
		})
	}

	return metrics
}

// ConvertToChannelMetrics converts YouTube Channel Metrics to domain model
func (p *YouTubeAnalyticsProvider) ConvertToChannelMetrics(
	channelID uuid.UUID,
	metricsList []*YouTubeChannelMetric,
	syncJobID *uuid.UUID,
) []*domain.DailyMetric {
	var metrics []*domain.DailyMetric
	now := time.Now()

	for _, m := range metricsList {
		metrics = append(metrics, &domain.DailyMetric{
			ID:                    uuid.New(),
			ChannelID:             channelID,
			VideoID:               nil,
			Date:                  m.Date,
			MetricType:            "channel",
			Views:                 m.Views,
			WatchTime:             m.WatchTime,
			Subscribers:           m.Subscribers,
			ReturningViewers:      m.ReturningViewers,
			NewViewers:            m.NewViewers,
			SyncJobID:             syncJobID,
			SyncedAt:              &now,
		})
	}

	return metrics
}
