package service

import (
	"github.com/jaybani/jb_cip/config"
	"github.com/jaybani/jb_cip/internal/domain"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/internal/repository"
	"github.com/jaybani/jb_cip/pkg/errors"
)

type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User         *domain.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("AUTH_001", "Invalid email or password", 401)
	}

	if user.Status != "active" {
		return nil, errors.New("AUTH_002", "Account is not active", 401)
	}

	if !helper.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("AUTH_001", "Invalid email or password", 401)
	}

	_ = s.userRepo.UpdateLastLogin(user.ID)

	accessToken, err := helper.GenerateAccessToken(
		user.ID, user.Email, "owner",
		s.cfg.JWT.Secret, s.cfg.JWT.AccessTokenExpiry,
	)
	if err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to generate token", 500)
	}

	refreshToken, err := helper.GenerateRefreshToken(
		user.ID, user.Email,
		s.cfg.JWT.Secret, s.cfg.JWT.RefreshTokenExpiry,
	)
	if err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to generate refresh token", 500)
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.JWT.AccessTokenExpiry.Seconds()),
	}, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*LoginResponse, error) {
	claims, err := helper.ValidateToken(refreshToken, s.cfg.JWT.Secret)
	if err != nil {
		return nil, errors.New("AUTH_003", "Invalid refresh token", 401)
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("AUTH_003", "Invalid token type", 401)
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("AUTH_001", "User not found", 401)
	}

	if user.Status != "active" {
		return nil, errors.New("AUTH_002", "Account is not active", 401)
	}

	accessToken, err := helper.GenerateAccessToken(
		user.ID, user.Email, "owner",
		s.cfg.JWT.Secret, s.cfg.JWT.AccessTokenExpiry,
	)
	if err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to generate token", 500)
	}

	newRefreshToken, err := helper.GenerateRefreshToken(
		user.ID, user.Email,
		s.cfg.JWT.Secret, s.cfg.JWT.RefreshTokenExpiry,
	)
	if err != nil {
		return nil, errors.New("SYSTEM_001", "Failed to generate refresh token", 500)
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.cfg.JWT.AccessTokenExpiry.Seconds()),
	}, nil
}

func (s *AuthService) ValidateAccessToken(token string) (*helper.TokenClaims, error) {
	claims, err := helper.ValidateToken(token, s.cfg.JWT.Secret)
	if err != nil {
		return nil, errors.New("AUTH_001", "Invalid or expired token", 401)
	}

	if claims.TokenType != "access" {
		return nil, errors.New("AUTH_001", "Invalid token type", 401)
	}

	return claims, nil
}
