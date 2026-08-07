package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jaybani/jb_cip/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const YouTubeScope = "https://www.googleapis.com/auth/youtube.readonly"

type GoogleProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

func NewGoogleProvider(clientID, clientSecret, redirectURI string) *GoogleProvider {
	return &GoogleProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

func (p *GoogleProvider) GetOAuthURL(state string) string {
	config := &oauth2.Config{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		RedirectURL:  p.redirectURI,
		Scopes:      []string{YouTubeScope, "openid", "email", "profile"},
		Endpoint: google.Endpoint,
	}
	return config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (p *GoogleProvider) ExchangeCode(code string) (*oauth2.Token, string, error) {
	config := &oauth2.Config{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		RedirectURL:  p.redirectURI,
		Scopes:      []string{YouTubeScope, "openid", "email", "profile"},
		Endpoint: google.Endpoint,
	}

	token, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange code: %w", err)
	}

	userID, err := p.getUserID(token.AccessToken)
	if err != nil {
		return token, "", err
	}

	return token, userID, nil
}

func (p *GoogleProvider) getUserID(accessToken string) (string, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", err
	}
	return userInfo.ID, nil
}

func (p *GoogleProvider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	config := &oauth2.Config{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		RedirectURL:  p.redirectURI,
		Scopes:      []string{YouTubeScope, "openid", "email", "profile"},
		Endpoint: google.Endpoint,
	}

	token := &oauth2.Token{RefreshToken: refreshToken}
	tokenSource := config.TokenSource(oauth2.NoContext, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}

func (p *GoogleProvider) GetYouTubeChannels(accessToken string) ([]*domain.YouTubeChannel, error) {
	resp, err := http.Get("https://www.googleapis.com/youtube/v3/channels?part=snippet,contentDetails,statistics&mine=true&access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			ID         string `json:"id"`
			Snippet    struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"snippet"`
			Statistics struct {
				ViewCount       string `json:"viewCount"`
				SubscriberCount string `json:"subscriberCount"`
				VideoCount      string `json:"videoCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var channels []*domain.YouTubeChannel
	for _, item := range result.Items {
		var subscriberCount int64
		var viewCount int64
		var videoCount int64
		fmt.Sscanf(item.Statistics.SubscriberCount, "%d", &subscriberCount)
		fmt.Sscanf(item.Statistics.ViewCount, "%d", &viewCount)
		fmt.Sscanf(item.Statistics.VideoCount, "%d", &videoCount)

		channels = append(channels, &domain.YouTubeChannel{
			ExternalID:      item.ID,
			Name:            item.Snippet.Title,
			Description:     item.Snippet.Description,
			SubscriberCount: subscriberCount,
			ViewCount:       viewCount,
			VideoCount:      videoCount,
			Status:          "active",
		})
	}

	return channels, nil
}

func (p *GoogleProvider) TestConnection(accessToken string) (bool, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200, nil
}

func BuildRedirectURL(base string) string {
	return base + "/api/v1/integrations/google/callback"
}

func EnsureScopes(scopes []string) string {
	return strings.Join(scopes, " ")
}

func BuildAuthURL(clientID, redirectURL string, scopes []string) string {
	base := "https://accounts.google.com/o/oauth2/v2/auth"
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURL)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("access_type", "offline")
	params.Set("approval_prompt", "force")
	params.Set("response_type", "code")
	return base + "?" + params.Encode()
}

func ParseTokenResponse(resp *http.Response) (*oauth2.Token, error) {
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		TokenType:    tokenResp.TokenType,
	}, nil
}
