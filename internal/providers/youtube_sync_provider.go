package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/domain"
	"golang.org/x/oauth2"
)

type YouTubeSyncProvider struct {
	googleProvider *GoogleProvider
}

type YouTubeVideoInfo struct {
	ID          string
	Title       string
	Description string
	PublishedAt time.Time
	Duration    int
	Thumbnail   string
	Privacy     string
	Views       int64
	Likes       int64
	Comments    int64
}

func NewYouTubeSyncProvider(gp *GoogleProvider) *YouTubeSyncProvider {
	return &YouTubeSyncProvider{googleProvider: gp}
}

func (p *YouTubeSyncProvider) GetSyncProviderName() string {
	return "youtube"
}

// GetChannelVideos fetches all videos for a YouTube channel using the playlist upload endpoint
func (p *YouTubeSyncProvider) GetChannelVideos(accessToken, channelExternalID string, since *time.Time) ([]*YouTubeVideoInfo, error) {
	// First, get upload playlist ID from channel details
	uploadPlaylistID, err := p.getUploadPlaylistID(accessToken, channelExternalID)
	if err != nil {
		return nil, err
	}

	// Fetch all videos from upload playlist
	videos, err := p.getVideosFromPlaylist(accessToken, uploadPlaylistID, since)
	if err != nil {
		return nil, err
	}

	return videos, nil
}

func (p *YouTubeSyncProvider) getUploadPlaylistID(accessToken, channelID string) (string, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?part=contentDetails&id=%s&access_token=%s", channelID, accessToken)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch channel: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			ContentDetails struct {
				RelatedPlaylists struct {
					Uploads string `json:"uploads"`
				} `json:"relatedPlaylists"`
			} `json:"contentDetails"`
		} `json:"items"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", fmt.Errorf("YouTube API error: %s", result.Error.Message)
	}
	if len(result.Items) == 0 {
		return "", fmt.Errorf("channel not found")
	}
	return result.Items[0].ContentDetails.RelatedPlaylists.Uploads, nil
}

func (p *YouTubeSyncProvider) getVideosFromPlaylist(accessToken, playlistID string, since *time.Time) ([]*YouTubeVideoInfo, error) {
	var videos []*YouTubeVideoInfo
	var pageToken string

	for {
		url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&playlistId=%s&maxResults=50&pageToken=%s&access_token=%s", playlistID, pageToken, accessToken)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		var result struct {
			Items []struct {
				Snippet struct {
					ResourceID struct {
						VideoID string `json:"videoId"`
					} `json:"resourceId"`
					Title       string `json:"title"`
					Description string `json:"description"`
					PublishedAt string `json:"publishedAt"`
					Thumbnails  struct {
						Default struct {
							URL string `json:"url"`
						} `json:"default"`
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
			Error         *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		if result.Error != nil {
			return nil, fmt.Errorf("YouTube API error: %s", result.Error.Message)
		}

		videoIDs := make([]string, 0, len(result.Items))
		for _, item := range result.Items {
			videoID := item.Snippet.ResourceID.VideoID
			publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)

			if since != nil && publishedAt.Before(*since) {
				continue
			}

			videoIDs = append(videoIDs, videoID)
			thumbnail := item.Snippet.Thumbnails.High.URL
			if thumbnail == "" {
				thumbnail = item.Snippet.Thumbnails.Medium.URL
			}
			if thumbnail == "" {
				thumbnail = item.Snippet.Thumbnails.Default.URL
			}

			videos = append(videos, &YouTubeVideoInfo{
				ID:          videoID,
				Title:       item.Snippet.Title,
				Description: item.Snippet.Description,
				PublishedAt: publishedAt,
				Thumbnail:   thumbnail,
			})
		}

		// Fetch video details (duration, statistics, privacy status) in batch
		if len(videoIDs) > 0 {
			details, err := p.getVideoDetails(accessToken, videoIDs)
			if err != nil {
				return nil, err
			}
			for _, v := range videos[len(videos)-len(videoIDs):] {
				if d, ok := details[v.ID]; ok {
					v.Duration = d.Duration
					v.Views = d.Views
					v.Likes = d.Likes
					v.Comments = d.Comments
					v.Privacy = d.Privacy
				}
			}
		}

		pageToken = result.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return videos, nil
}

func (p *YouTubeSyncProvider) getVideoDetails(accessToken string, videoIDs []string) (map[string]*YouTubeVideoInfo, error) {
	ids := strings.Join(videoIDs, ",")
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=contentDetails,statistics,status&id=%s&access_token=%s", ids, accessToken)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			ID         string `json:"id"`
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
			Statistics struct {
				ViewCount    string `json:"viewCount"`
				LikeCount    string `json:"likeCount"`
				CommentCount string `json:"commentCount"`
			} `json:"statistics"`
			Status struct {
				PrivacyStatus string `json:"privacyStatus"`
			} `json:"status"`
		} `json:"items"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, fmt.Errorf("YouTube API error: %s", result.Error.Message)
	}

	details := make(map[string]*YouTubeVideoInfo)
	for _, item := range result.Items {
		viewCount, _ := strconv.ParseInt(item.Statistics.ViewCount, 10, 64)
		likeCount, _ := strconv.ParseInt(item.Statistics.LikeCount, 10, 64)
		commentCount, _ := strconv.ParseInt(item.Statistics.CommentCount, 10, 64)
		details[item.ID] = &YouTubeVideoInfo{
			Duration: parseISODuration(item.ContentDetails.Duration),
			Views:    viewCount,
			Likes:    likeCount,
			Comments: commentCount,
			Privacy:  item.Status.PrivacyStatus,
		}
	}
	return details, nil
}

func parseISODuration(iso string) int {
	// PT4M13S -> 253
	duration := 0
	hours := 0
	minutes := 0
	seconds := 0
	fmt.Sscanf(iso, "PT%dH%dM%dS", &hours, &minutes, &seconds)
	if hours == 0 && minutes == 0 && seconds == 0 {
		fmt.Sscanf(iso, "PT%dM%dS", &minutes, &seconds)
	}
	if hours == 0 && minutes == 0 && seconds == 0 {
		fmt.Sscanf(iso, "PT%dS", &seconds)
	}
	if hours == 0 && minutes == 0 && seconds == 0 {
		fmt.Sscanf(iso, "PT%dH", &hours)
	}
	if hours == 0 && minutes == 0 && seconds == 0 {
		fmt.Sscanf(iso, "PT%dM", &minutes)
	}
	duration = hours*3600 + minutes*60 + seconds
	return duration
}

func (p *YouTubeSyncProvider) ConvertToDomainVideo(video *YouTubeVideoInfo, channelID uuid.UUID) *domain.YouTubeVideo {
	return &domain.YouTubeVideo{
		ID:            uuid.New(),
		ChannelID:     channelID,
		VideoID:       video.ID,
		Title:         video.Title,
		Description:   video.Description,
		PublishedAt:   &video.PublishedAt,
		Duration:      video.Duration,
		ThumbnailURL:  video.Thumbnail,
		PrivacyStatus: video.Privacy,
		ViewCount:     video.Views,
		LikeCount:     video.Likes,
		CommentCount:  video.Comments,
	}
}

func (p *YouTubeSyncProvider) BuildOAuthToken(accessToken, refreshToken string, expiresAt time.Time) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiresAt,
	}
}

func (p *YouTubeSyncProvider) ValidateVideoID(videoID string) bool {
	return videoID != "" && len(videoID) <= 20
}
