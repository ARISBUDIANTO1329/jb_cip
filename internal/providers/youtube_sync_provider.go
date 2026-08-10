package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/domain"
	"golang.org/x/oauth2"
)

type YouTubeSyncProvider struct {
	GoogleProvider *GoogleProvider
}

func NewYouTubeSyncProvider(gp *GoogleProvider) *YouTubeSyncProvider {
	return &YouTubeSyncProvider{GoogleProvider: gp}
}

// BuildOAuthToken creates an OAuth token from access and refresh tokens
func (p *YouTubeSyncProvider) BuildOAuthToken(accessToken string, refreshToken string, expiry time.Time) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}
}

// GetChannelVideos fetches all videos for a channel with pagination and statistics.
// It loops over every page returned by YouTube search.list (maxResults=50) until
// nextPageToken is empty, then enriches each batch with view/like/comment counts
// from videos.list. The `since` filter maps to publishedAfter for incremental syncs.
// Note: YouTube's search.list pagination with order=date can return duplicate video_ids
// across pages. This function deduplicates in-memory before returning.
func (p *YouTubeSyncProvider) GetChannelVideos(accessToken string, channelID string, since *time.Time) ([]*domain.YouTubeVideo, error) {
	var videos []*domain.YouTubeVideo
	nextPageToken := ""
	seen := make(map[string]bool)

	for {
		reqURL := fmt.Sprintf(
			"https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=%s&maxResults=50&order=date&type=video&access_token=%s",
			channelID, accessToken,
		)
		if nextPageToken != "" {
			reqURL += "&pageToken=" + nextPageToken
		}
		if since != nil {
			reqURL += "&publishedAfter=" + since.UTC().Format(time.RFC3339)
		}

		resp, err := http.Get(reqURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch videos page: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			resp.Body.Close()
			return nil, fmt.Errorf("youtube search.list returned status %d: %s", resp.StatusCode, string(body))
		}

		var result struct {
			Items []struct {
				ID struct {
					VideoID string `json:"videoId"`
				} `json:"id"`
				Snippet struct {
					Title       string `json:"title"`
					Description string `json:"description"`
					PublishedAt string `json:"publishedAt"`
					Thumbnails  struct {
						Medium struct {
							URL string `json:"url"`
						} `json:"medium"`
						High struct {
							URL string `json:"url"`
						} `json:"high"`
					} `json:"thumbnails"`
				} `json:"snippet"`
			} `json:"items"`
			NextPageToken string `json:"nextPageToken"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode search response: %w", err)
		}
		resp.Body.Close()

		var pageVideoIDs []string
		for _, item := range result.Items {
			// Deduplicate: skip if we've already processed this video_id
			if seen[item.ID.VideoID] {
				continue
			}
			seen[item.ID.VideoID] = true

			publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
			thumb := item.Snippet.Thumbnails.High.URL
			if thumb == "" {
				thumb = item.Snippet.Thumbnails.Medium.URL
			}
			videos = append(videos, &domain.YouTubeVideo{
				VideoID:       item.ID.VideoID,
				Title:         item.Snippet.Title,
				Description:   item.Snippet.Description,
				PublishedAt:   &publishedAt,
				ThumbnailURL:  thumb,
				PrivacyStatus: "public",
			})
			pageVideoIDs = append(pageVideoIDs, item.ID.VideoID)
		}

		if len(pageVideoIDs) > 0 {
			if err := p.fetchVideoStatistics(accessToken, pageVideoIDs, videos); err != nil {
				return nil, err
			}
		}

		if result.NextPageToken == "" {
			break
		}
		nextPageToken = result.NextPageToken
	}

	return videos, nil
}

// fetchVideoStatistics enriches videos with view/like/comment counts using the
// videos.list endpoint (max 50 IDs per call).
func (p *YouTubeSyncProvider) fetchVideoStatistics(accessToken string, videoIDs []string, videos []*domain.YouTubeVideo) error {
	idSet := make(map[string]bool, len(videoIDs))
	for _, id := range videoIDs {
		idSet[id] = true
	}
	byID := make(map[string]*domain.YouTubeVideo, len(videos))
	for _, v := range videos {
		if idSet[v.VideoID] {
			byID[v.VideoID] = v
		}
	}

	for i := 0; i < len(videoIDs); i += 50 {
		end := i + 50
		if end > len(videoIDs) {
			end = len(videoIDs)
		}
		batch := videoIDs[i:end]

		reqURL := fmt.Sprintf(
			"https://www.googleapis.com/youtube/v3/videos?part=statistics&id=%s&maxResults=50&access_token=%s",
			strings.Join(batch, ","), accessToken,
		)

		resp, err := http.Get(reqURL)
		if err != nil {
			return fmt.Errorf("failed to fetch video statistics: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			resp.Body.Close()
			return fmt.Errorf("youtube videos.list returned status %d: %s", resp.StatusCode, string(body))
		}

		var result struct {
			Items []struct {
				ID         string `json:"id"`
				Statistics struct {
					ViewCount    string `json:"viewCount"`
					LikeCount    string `json:"likeCount"`
					CommentCount string `json:"commentCount"`
				} `json:"statistics"`
			} `json:"items"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to decode statistics response: %w", err)
		}
		resp.Body.Close()

		for _, item := range result.Items {
			if v, ok := byID[item.ID]; ok {
				v.ViewCount = parseInt64(item.Statistics.ViewCount)
				v.LikeCount = parseInt64(item.Statistics.LikeCount)
				v.CommentCount = parseInt64(item.Statistics.CommentCount)
			}
		}
	}

	return nil
}

func parseInt64(s string) int64 {
	var n int64
	fmt.Sscanf(s, "%d", &n)
	return n
}

// ConvertToDomainVideo attaches the channel UUID and ensures a fresh row ID.
func (p *YouTubeSyncProvider) ConvertToDomainVideo(video *domain.YouTubeVideo, channelID uuid.UUID) *domain.YouTubeVideo {
	video.ChannelID = channelID
	if video.ID == uuid.Nil {
		video.ID = uuid.New()
	}
	return video
}

// GetAccessToken returns the access token
func (p *YouTubeSyncProvider) GetAccessToken(accessToken string) (string, error) {
	return accessToken, nil
}

// GetChannelID gets the channel ID from YouTube
func (p *YouTubeSyncProvider) GetChannelID(accessToken string) (string, error) {
	channels, err := p.GoogleProvider.GetYouTubeChannels(accessToken)
	if err != nil {
		return "", err
	}
	if len(channels) > 0 {
		return channels[0].ExternalID, nil
	}
	return "", nil
}

// GetChannelDetails gets detailed channel information
func (p *YouTubeSyncProvider) GetChannelDetails(accessToken string, channelID string) (*domain.YouTubeChannel, error) {
	channels, err := p.GoogleProvider.GetYouTubeChannels(accessToken)
	if err != nil {
		return nil, err
	}
	for _, ch := range channels {
		if ch.ExternalID == channelID {
			return ch, nil
		}
	}
	return nil, nil
}

// ListChannels lists all channels for a user
func (p *YouTubeSyncProvider) ListChannels(accessToken string) ([]*domain.YouTubeChannel, error) {
	return p.GoogleProvider.GetYouTubeChannels(accessToken)
}
