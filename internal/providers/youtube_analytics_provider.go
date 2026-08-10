package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/domain"
)

type YouTubeAnalyticsProvider struct {
	googleProvider *GoogleProvider
}

type YouTubeVideoMetric struct {
	VideoID             string  `json:"video_id"`
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

// YouTubeAnalyticsReport represents v2 API response
type YouTubeAnalyticsReport struct {
	ColumnHeaders []struct {
		Name      string `json:"name"`
		ColumnType string `json:"columnType"`
		DataType   string `json:"dataType"`
	} `json:"columnHeaders"`
	Rows [][]interface{} `json:"rows"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewYouTubeAnalyticsProvider(gp *GoogleProvider) *YouTubeAnalyticsProvider {
	return &YouTubeAnalyticsProvider{googleProvider: gp}
}

// queryV2 executes a YouTube Analytics API v2 query and returns the report
func (p *YouTubeAnalyticsProvider) queryV2(accessToken, channelID string, dimensions, metrics, filters, startDate, endDate string, maxResults, startIndex int) (*YouTubeAnalyticsReport, error) {
	// Build query parameters
	params := fmt.Sprintf("ids=channel==%s&startDate=%s&endDate=%s&dimensions=%s&metrics=%s&maxResults=%d&startIndex=%d",
		channelID, startDate, endDate, dimensions, metrics, maxResults, startIndex)
	if filters != "" {
		params += "&filters=" + filters
	}

	url := "https://youtubeanalytics.googleapis.com/v2/reports?" + params
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status first
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("YouTube Analytics API v2 error (HTTP %d): %s", resp.StatusCode, string(body[:min(500, len(body))]))
	}

	var report YouTubeAnalyticsReport
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if report.Error != nil {
		return nil, fmt.Errorf("YouTube Analytics API v2 error (code %d): %s", report.Error.Code, report.Error.Message)
	}

	return &report, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ConvertRelativeDate converts special date values to ISO format
// 7daysAgo -> today minus 6 days (7 day range)
// today -> current date
func ConvertRelativeDate(dateStr string) string {
	if dateStr == "7daysAgo" {
		return time.Now().AddDate(0, 0, -6).Format("2006-01-02")
	}
	if dateStr == "today" {
		return time.Now().Format("2006-01-02")
	}
	return dateStr
}

// GetVideoMetricsBatch fetches YouTube Analytics for up to 50 videos in a single call.
// Uses API v2 with dimensions=video,date and filters=video==id1,id2,...
func (p *YouTubeAnalyticsProvider) GetVideoMetricsBatch(accessToken, channelID string, videoIDs []string, startDate, endDate string) ([]*YouTubeVideoMetric, error) {
	if len(videoIDs) == 0 {
		return nil, nil
	}
	if len(videoIDs) > 50 {
		return nil, fmt.Errorf("batch size exceeds 50 videos")
	}

	// Build filters: video==id1,id2,...
	filters := "video==" + joinStrings(videoIDs, ",")

	// Convert relative dates to ISO format
	startDate = ConvertRelativeDate(startDate)
	endDate = ConvertRelativeDate(endDate)

	// Metrics supported in v2 for video reports with youtube.readonly scope
	metrics := "views,estimatedMinutesWatched,averageViewDuration,averageViewPercentage,likes,comments,shares"

	report, err := p.queryV2(accessToken, channelID, "video,day", metrics, filters, startDate, endDate, 500, 0)
	if err != nil {
		return nil, err
	}

	var metricsList []*YouTubeVideoMetric

	// Map column indices
	headerMap := make(map[string]int)
	for i, h := range report.ColumnHeaders {
		headerMap[h.Name] = i
	}

	for _, row := range report.Rows {
		videoIdx := headerMap["video"]
		dayIdx := headerMap["day"]

		if videoIdx < 0 || dayIdx < 0 || videoIdx >= len(row) || dayIdx >= len(row) {
			continue
		}

		videoID, _ := row[videoIdx].(string)
		dateStr, _ := row[dayIdx].(string)

		metric := &YouTubeVideoMetric{
			VideoID: videoID,
			Date:    dateStr,
		}

		// Parse metrics safely with nil checks
		if idx, ok := headerMap["views"]; ok && idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				metric.Views = int64(v)
			}
		}
		if idx, ok := headerMap["estimatedMinutesWatched"]; ok && idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				metric.WatchTime = int64(v)
			}
		}
		if idx, ok := headerMap["averageViewDuration"]; ok && idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				metric.AverageViewDuration = v
			}
		}
		if idx, ok := headerMap["averageViewPercentage"]; ok && idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				metric.PercentageViewed = v
			}
		}
		if idx, ok := headerMap["impressions"]; ok && idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				metric.Impressions = int64(v)
			}
		}
		if idx, ok := headerMap["impressionsClickThroughRate"]; ok && idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				metric.ImpressionCTR = v
			}
		}
		if idx, ok := headerMap["likes"]; ok && idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				metric.Likes = int64(v)
			}
		}
		if idx, ok := headerMap["comments"]; ok && idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				metric.Comments = int64(v)
			}
		}
		if idx, ok := headerMap["shares"]; ok && idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				metric.Shares = int64(v)
			}
		}

		metricsList = append(metricsList, metric)
	}

	return metricsList, nil
}

// GetChannelMetrics fetches YouTube Analytics for a channel
func (p *YouTubeAnalyticsProvider) GetChannelMetrics(accessToken, channelID string, startDate, endDate string) ([]*YouTubeChannelMetric, error) {
	// Convert relative dates to ISO format
	startDate = ConvertRelativeDate(startDate)
	endDate = ConvertRelativeDate(endDate)

	// Metrics supported in v2 for channel reports
	metrics := "views,estimatedMinutesWatched,subscribersGained,subscribersLost"

	report, err := p.queryV2(accessToken, channelID, "day", metrics, "", startDate, endDate, 500, 0)
	if err != nil {
		return nil, err
	}

	var metricsList []*YouTubeChannelMetric

	// Map column indices
	headerMap := make(map[string]int)
	for i, h := range report.ColumnHeaders {
		headerMap[h.Name] = i
	}

	for _, row := range report.Rows {
		dayIdx := headerMap["day"]
		if dayIdx < 0 || dayIdx >= len(row) {
			continue
		}
		dateStr, _ := row[dayIdx].(string)

		metric := &YouTubeChannelMetric{
			Date: dateStr,
		}

		if idx, ok := headerMap["views"]; ok && idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				metric.Views = int64(v)
			}
		}
		if idx, ok := headerMap["estimatedMinutesWatched"]; ok && idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				metric.WatchTime = int64(v)
			}
		}
		if idx, ok := headerMap["subscribersGained"]; ok && idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				metric.Subscribers = int64(v)
			}
		}

		metricsList = append(metricsList, metric)
	}

	return metricsList, nil
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// ConvertToDailyMetrics converts batched YouTube Video Metrics to domain model.
// channelID is the internal analytics.channels.id. videoIDMap maps YouTube video_id to internal youtube_videos.id.
func (p *YouTubeAnalyticsProvider) ConvertToDailyMetrics(
	channelID uuid.UUID,
	videoIDMap map[string]uuid.UUID,
	metricsList []*YouTubeVideoMetric,
	syncJobID *uuid.UUID,
) []*domain.DailyMetric {
	var metrics []*domain.DailyMetric
	now := time.Now()

	for _, m := range metricsList {
		videoUUID, ok := videoIDMap[m.VideoID]
		if !ok {
			continue
		}
		metrics = append(metrics, &domain.DailyMetric{
			ID:                      uuid.New(),
			ChannelID:               channelID,
			VideoID:                 &videoUUID,
			Date:                    m.Date,
			MetricType:              "video",
			Views:                   m.Views,
			WatchTime:               m.WatchTime,
			AverageViewDuration:     m.AverageViewDuration,
			AveragePercentageViewed: m.PercentageViewed,
			Impressions:             m.Impressions,
			ImpressionCTR:           m.ImpressionCTR,
			Likes:                   m.Likes,
			Comments:                m.Comments,
			Shares:                  m.Shares,
			SyncJobID:               syncJobID,
			SyncedAt:                &now,
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
			ID:                 uuid.New(),
			ChannelID:          channelID,
			VideoID:            nil,
			Date:               m.Date,
			MetricType:         "channel",
			Views:              m.Views,
			WatchTime:          m.WatchTime,
			Subscribers:        m.Subscribers,
			ReturningViewers:   m.ReturningViewers,
			NewViewers:         m.NewViewers,
			SyncJobID:          syncJobID,
			SyncedAt:           &now,
		})
	}

	return metrics
}
