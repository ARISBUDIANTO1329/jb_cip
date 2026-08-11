package service

import (
	"fmt"
	"time"

	"github.com/jaybani/jb_cip/config"
	"github.com/jaybani/jb_cip/internal/domain"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/internal/providers"
	"github.com/jaybani/jb_cip/internal/repository"
	"github.com/jaybani/jb_cip/pkg/errors"
	"golang.org/x/oauth2"
)

type IntegrationService struct {
	integrationRepo *repository.IntegrationRepository
	wsRepo          *repository.WorkspaceRepository
	provider        *providers.GoogleProvider
	encryptor       *helper.TokenEncryptor
	cfg             *config.Config
}

func NewIntegrationService(
	integrationRepo *repository.IntegrationRepository,
	wsRepo *repository.WorkspaceRepository,
	provider *providers.GoogleProvider,
	encryptor *helper.TokenEncryptor,
	cfg *config.Config,
) *IntegrationService {
	return &IntegrationService{
		integrationRepo: integrationRepo,
		wsRepo:          wsRepo,
		provider:        provider,
		encryptor:       encryptor,
		cfg:             cfg,
	}
}

type GoogleLoginResponse struct {
	AuthURL string `json:"auth_url"`
}

type CallbackResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Provider     string `json:"provider"`
	Status       string `json:"status"`
}

type TestConnectionResponse struct {
	Connected bool   `json:"connected"`
	Message   string `json:"message"`
}

type ChannelResponse struct {
	ID              string `json:"id"`
	ExternalID      string `json:"external_id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	SubscriberCount int64  `json:"subscriber_count"`
	ViewCount       int64  `json:"view_count"`
	VideoCount      int64  `json:"video_count"`
	Status          string `json:"status"`
}

func (s *IntegrationService) GetWorkspaceRepo() *repository.WorkspaceRepository {
	return s.wsRepo
}

func (s *IntegrationService) GoogleLogin(userID, workspaceID, baseURL string) (*GoogleLoginResponse, error) {
	authURL := s.provider.GetOAuthURL(fmt.Sprintf("%s:%s", userID, workspaceID))
	return &GoogleLoginResponse{
		AuthURL: authURL,
	}, nil
}

func (s *IntegrationService) GoogleCallback(code, userID, workspaceID string) (*CallbackResponse, error) {
	token, googleUserID, err := s.provider.ExchangeCode(code)
	if err != nil {
		return nil, errors.New("INTEGRATION_001", fmt.Sprintf("Failed to exchange code: %v", err), 500)
	}

	// Save connection
	conn := &domain.APIConnection{
		UserID:         userID,
		WorkspaceID:    workspaceID,
		Provider:       "google",
		ProviderUserID: googleUserID,
		Status:         "active",
	}
	if err := s.integrationRepo.CreateConnection(conn); err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to save connection", 500)
	}

	// Encrypt and save tokens
	encAccess, err := s.encryptor.Encrypt(token.AccessToken)
	if err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to encrypt token", 500)
	}

	apiToken := &domain.APIToken{
		ConnectionID:          conn.ID,
		AccessTokenEncrypted:  encAccess,
		AccessTokenExpiresAt:  &token.Expiry,
	}

	if token.RefreshToken != "" {
		encRefresh, err := s.encryptor.Encrypt(token.RefreshToken)
		if err != nil {
			return nil, errors.New("SYSTEM_001", "Failed to encrypt refresh token", 500)
		}
		apiToken.RefreshTokenEncrypted = encRefresh
		refreshExpiry := time.Now().Add(30 * 24 * time.Hour)
		apiToken.RefreshTokenExpiresAt = &refreshExpiry
	}

	if err := s.integrationRepo.SaveToken(apiToken); err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to save token", 500)
	}

	expiresIn := token.Expiry.Sub(time.Now())

	return &CallbackResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    int64(expiresIn.Seconds()),
		Provider:     "google",
		Status:       "connected",
	}, nil
}

func (s *IntegrationService) Disconnect(userID string) error {
	// Find and delete all connections for user
	conn, err := s.integrationRepo.GetConnection(userID, "google")
	if err != nil {
		return nil
	}

	if err := s.integrationRepo.DeleteConnection(conn.ID); err != nil {
		return errors.New("SYSTEM_001", "Failed to disconnect", 500)
	}

	return nil
}

func (s *IntegrationService) TestConnection(userID string) (*TestConnectionResponse, error) {
	conn, err := s.integrationRepo.GetConnection(userID, "google")
	if err != nil {
		return &TestConnectionResponse{
			Connected: false,
			Message:   "No Google connection found. Please connect first.",
		}, nil
	}

	token, err := s.integrationRepo.GetToken(conn.ID)
	if err != nil {
		return &TestConnectionResponse{
			Connected: false,
			Message:   "Token not found",
		}, nil
	}

	// Decrypt access token
	accessToken, err := s.encryptor.Decrypt(token.AccessTokenEncrypted)
	if err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to decrypt token", 500)
	}

	// Check if expired
	if token.AccessTokenExpiresAt != nil && time.Now().After(*token.AccessTokenExpiresAt) {
		// Try refresh
		refreshToken, err := s.encryptor.Decrypt(token.RefreshTokenEncrypted)
		if err != nil {
			return &TestConnectionResponse{Connected: false, Message: "Token expired and refresh failed"}, nil
		}

		newToken, err := s.provider.RefreshToken(refreshToken)
		if err != nil {
			return &TestConnectionResponse{Connected: false, Message: "Token refresh failed"}, nil
		}

		// Save new token
		encAccess, _ := s.encryptor.Encrypt(newToken.AccessToken)
		apiToken := &domain.APIToken{
			ConnectionID:         conn.ID,
			AccessTokenEncrypted: encAccess,
			AccessTokenExpiresAt: &newToken.Expiry,
		}
		if newToken.RefreshToken != "" {
			encRefresh, _ := s.encryptor.Encrypt(newToken.RefreshToken)
			apiToken.RefreshTokenEncrypted = encRefresh
			refreshExpiry := time.Now().Add(30 * 24 * time.Hour)
			apiToken.RefreshTokenExpiresAt = &refreshExpiry
		}
		_ = s.integrationRepo.UpdateToken(apiToken)
		accessToken = newToken.AccessToken
	}

	connected, err := s.provider.TestConnection(accessToken)
	if err != nil {
		return &TestConnectionResponse{Connected: false, Message: fmt.Sprintf("Connection test failed: %v", err)}, nil
	}

	return &TestConnectionResponse{
		Connected: connected,
		Message:   "Connection successful",
	}, nil
}

func (s *IntegrationService) GetYouTubeChannels(userID string) ([]*ChannelResponse, error) {
	conn, err := s.integrationRepo.GetConnection(userID, "google")
	if err != nil {
		return nil, errors.New("INTEGRATION_002", "No Google connection found", 404)
	}

	token, err := s.integrationRepo.GetToken(conn.ID)
	if err != nil {
		return nil, errors.New("INTEGRATION_002", "No token found", 404)
	}

	// Decrypt and refresh if needed
	accessToken, err := s.encryptor.Decrypt(token.AccessTokenEncrypted)
	if err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to decrypt token", 500)
	}

	if token.AccessTokenExpiresAt != nil && time.Now().After(*token.AccessTokenExpiresAt) {
		refreshToken, _ := s.encryptor.Decrypt(token.RefreshTokenEncrypted)
		newToken, err := s.provider.RefreshToken(refreshToken)
		if err != nil {
			return nil, errors.New("INTEGRATION_003", "Token refresh failed", 401)
		}

		encAccess, _ := s.encryptor.Encrypt(newToken.AccessToken)
		apiToken := &domain.APIToken{
			ConnectionID:         conn.ID,
			AccessTokenEncrypted: encAccess,
			AccessTokenExpiresAt: &newToken.Expiry,
		}
		if newToken.RefreshToken != "" {
			encRefresh, _ := s.encryptor.Encrypt(newToken.RefreshToken)
			apiToken.RefreshTokenEncrypted = encRefresh
			refreshExpiry := time.Now().Add(30 * 24 * time.Hour)
			apiToken.RefreshTokenExpiresAt = &refreshExpiry
		}
		_ = s.integrationRepo.UpdateToken(apiToken)
		accessToken = newToken.AccessToken
	}

	channels, err := s.provider.GetYouTubeChannels(accessToken)
	if err != nil {
		return nil, errors.New("INTEGRATION_004", fmt.Sprintf("Failed to fetch channels: %v", err), 500)
	}

	// Resolve YouTube platform UUID dari analytics.platforms (07_DATABASE_DESIGN.md 3.5)
	platformID, err := s.integrationRepo.GetPlatformID("youtube")
	if err != nil {
		return nil, errors.New("INTEGRATION_005", fmt.Sprintf("YouTube platform not found: %v", err), 500)
	}

	// Save channels to DB
	for _, ch := range channels {
		ch.WorkspaceID = conn.WorkspaceID
		ch.ConnectionID = conn.ID
		ch.PlatformID = platformID
	}
	if err := s.integrationRepo.SaveChannels(channels); err != nil {
		return nil, errors.New("INTEGRATION_005", fmt.Sprintf("Failed to save channels: %v", err), 500)
	}

	// Fetch channels from DB to get internal UUID (analytics.channels.id)
	dbChannels, err := s.integrationRepo.GetChannels(conn.WorkspaceID)
	if err != nil {
		return nil, errors.New("INTEGRATION_005", fmt.Sprintf("Failed to load channels: %v", err), 500)
	}

	var result []*ChannelResponse
	for _, ch := range dbChannels {
		result = append(result, &ChannelResponse{
			ID:              ch.ID,
			ExternalID:      ch.ExternalID,
			Name:            ch.Name,
			Description:     ch.Description,
			SubscriberCount: ch.SubscriberCount,
			ViewCount:       ch.ViewCount,
			VideoCount:      ch.VideoCount,
			Status:          ch.Status,
		})
	}

	return result, nil
}

type RefreshTokenResult struct {
	Token *oauth2.Token `json:"-"`
}