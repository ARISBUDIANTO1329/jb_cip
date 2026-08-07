package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// GetChannelVideos fetches all videos for a channel
func (p *YouTubeSyncProvider) GetChannelVideos(accessToken string, channelID string, since *time.Time) ([]*domain.YouTubeVideo, error) {
	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=%s&maxResults=50&order=date&type=video&fields=items(id(videoId),snippet(title,description,publishedAt)),nextPageToken&access_token=%s",
		channelID, accessToken,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch videos: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				PublishedAt string `json:"publishedAt"`
			} `json:"snippet"`
		} `json:"items"`
		NextPageToken string `json:"nextPageToken"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var videos []*domain.YouTubeVideo
	for _, item := range result.Items {
		publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		videos = append(videos, &domain.YouTubeVideo{
			VideoID:      item.ID.VideoID,
			Title:        item.Snippet.Title,
			Description:  item.Snippet.Description,
			PublishedAt:  &publishedAt,
			PrivacyStatus: "public",
		})
	}

	return videos, nil
}

// ConvertToDomainVideo converts API video data to domain model
func (p *YouTubeSyncProvider) ConvertToDomainVideo(video *domain.YouTubeVideo, channelID uuid.UUID) *domain.YouTubeVideo {
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
