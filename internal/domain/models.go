package domain
import (
	"time"

	"github.com/google/uuid"
)


type User struct {
	ID              string     `json:"id" db:"id"`
	Email           string     `json:"email" db:"email"`
	PasswordHash    string     `json:"-" db:"password_hash"`
	Name            string     `json:"name" db:"name"`
	AvatarURL       string     `json:"avatar_url,omitempty" db:"avatar_url"`
	Status          string     `json:"status" db:"status"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time `json:"-" db:"deleted_at"`
}

type Workspace struct {
	ID          string     `json:"id" db:"id"`
	OwnerID     string     `json:"owner_id" db:"owner_id"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	Description string     `json:"description,omitempty" db:"description"`
	Status      string     `json:"status" db:"status"`
	Settings    []byte     `json:"settings,omitempty" db:"settings"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
}

type WorkspaceMember struct {
	ID          string     `json:"id" db:"id"`
	WorkspaceID string     `json:"workspace_id" db:"workspace_id"`
	UserID      string     `json:"user_id" db:"user_id"`
	RoleID      string     `json:"role_id" db:"role_id"`
	InvitedBy   *string    `json:"invited_by,omitempty" db:"invited_by"`
	InvitedAt   time.Time  `json:"invited_at" db:"invited_at"`
	JoinedAt    *time.Time `json:"joined_at,omitempty" db:"joined_at"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
}

type Role struct {
	ID          string     `json:"id" db:"id"`
	WorkspaceID string     `json:"workspace_id" db:"workspace_id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description,omitempty" db:"description"`
	Permissions []string   `json:"permissions" db:"permissions"`
	IsSystem    bool       `json:"is_system" db:"is_system"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
}

type Permission struct {
	ID          string    `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	Category    string    `json:"category" db:"category"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type APIConnection struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	WorkspaceID  string    `json:"workspace_id" db:"workspace_id"`
	Provider     string    `json:"provider" db:"provider"`
	ProviderUserID string  `json:"provider_user_id,omitempty" db:"provider_user_id"`
	Status       string    `json:"status" db:"status"`
	Scopes       []string  `json:"scopes,omitempty" db:"scopes"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"-" db:"deleted_at"`
}

type APIToken struct {
	ID                    string     `json:"-" db:"id"`
	ConnectionID          string     `json:"-" db:"connection_id"`
	AccessTokenEncrypted  string     `json:"-" db:"access_token_encrypted"`
	RefreshTokenEncrypted string     `json:"-" db:"refresh_token_encrypted"`
	AccessTokenExpiresAt  *time.Time `json:"expires_at,omitempty" db:"access_token_expires_at"`
	RefreshTokenExpiresAt *time.Time `json:"refresh_expires_at,omitempty" db:"refresh_token_expires_at"`
	Scope                 []string   `json:"scope,omitempty" db:"scope"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

type YouTubeChannel struct {
	ID           string     `json:"id" db:"id"`
	WorkspaceID  string     `json:"workspace_id" db:"workspace_id"`
	ConnectionID string     `json:"-" db:"connection_id"`
	PlatformID   string     `json:"platform_id" db:"platform_id"`
	ExternalID   string     `json:"external_id" db:"external_id"`
	Name         string     `json:"name" db:"name"`
	Description  string     `json:"description,omitempty" db:"description"`
	SubscriberCount int64   `json:"subscriber_count" db:"subscriber_count"`
	ViewCount    int64      `json:"view_count" db:"view_count"`
	VideoCount   int64      `json:"video_count" db:"video_count"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"-" db:"deleted_at"`
}

// YouTubeVideo represents a synced YouTube video
type YouTubeVideo struct {
	ID             uuid.UUID  `json:"id"`
	ChannelID      uuid.UUID  `json:"channel_id"`
	VideoID        string     `json:"video_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	PublishedAt    *time.Time `json:"published_at"`
	Duration       int        `json:"duration"`
	ThumbnailURL   string     `json:"thumbnail_url"`
	PrivacyStatus  string     `json:"privacy_status"`
	ViewCount      int64      `json:"view_count"`
	LikeCount      int64      `json:"like_count"`
	CommentCount   int64      `json:"comment_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Sync Job model
type SyncJob struct {
	ID              uuid.UUID  `json:"id"`
	ChannelID       uuid.UUID  `json:"channel_id"`
	UserID          uuid.UUID  `json:"user_id"`
	WorkspaceID     uuid.UUID  `json:"workspace_id"`
	SyncType        string     `json:"sync_type"`
	Status          string     `json:"status"`
	TotalVideos     int        `json:"total_videos"`
	TotalSuccess    int        `json:"total_success"`
	TotalFailed     int        `json:"total_failed"`
	DurationSeconds int        `json:"duration_seconds"`
	ErrorMessage    string     `json:"error_message"`
	StartedAt       *time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// DailyMetric represents a daily analytics metric for YouTube channel or video
type DailyMetric struct {
	ID                    uuid.UUID  `json:"id"`
	ChannelID             uuid.UUID  `json:"channel_id"`
	VideoID               *uuid.UUID `json:"video_id,omitempty"`
	Date                  string     `json:"date"`
	MetricType            string     `json:"metric_type"` // 'channel' or 'video'
	Views                 int64      `json:"views"`
	WatchTime             int64      `json:"watch_time"`
	AverageViewDuration   float64    `json:"average_view_duration"`
	AveragePercentageViewed float64  `json:"average_percentage_viewed"`
	Impressions           int64      `json:"impressions"`
	ImpressionCTR         float64    `json:"impression_ctr"`
	Likes                 int64      `json:"likes"`
	Comments              int64      `json:"comments"`
	Shares                int64      `json:"shares"`
	Subscribers           int64      `json:"subscribers"`
	ReturningViewers      int64      `json:"returning_viewers"`
	NewViewers            int64      `json:"new_viewers"`
	SyncJobID             *uuid.UUID `json:"sync_job_id,omitempty"`
	SyncedAt              *time.Time `json:"synced_at,omitempty"`
}
