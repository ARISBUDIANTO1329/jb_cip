package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App         AppConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Security    SecurityConfig
	CORS        CORSConfig
	RateLimit   RateLimitConfig
	Integrations IntegrationsConfig
}

type AppConfig struct {
	Env            string
	Name           string
	Version        string
	APIVersion     string
	Host           string
	Port           string
	LogLevel       string
	LogFormat      string
	RequestTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	MigrationPath   string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type SecurityConfig struct {
	BcryptCost    int
	EncryptionKey string
}

type CORSConfig struct {
	AllowedOrigins string
	AllowedMethods string
}

type RateLimitConfig struct {
	Enabled  bool
	Requests int
}

type IntegrationsConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	requestTimeout, err := parseDuration("REQUEST_TIMEOUT", "30s")
	if err != nil {
		return nil, err
	}

	maxOpenConns, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdleConns, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "10"))
	connMaxLifetime, err := parseDuration("DB_CONN_MAX_LIFETIME", "1h")
	if err != nil {
		return nil, err
	}

	accessTokenExpiry, err := parseDuration("JWT_ACCESS_TOKEN_EXPIRY", "15m")
	if err != nil {
		return nil, err
	}

	refreshTokenExpiry, err := parseDuration("JWT_REFRESH_TOKEN_EXPIRY", "168h")
	if err != nil {
		return nil, err
	}

	bcryptCost, _ := strconv.Atoi(getEnv("BCRYPT_COST", "12"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	rateLimitEnabled, _ := strconv.ParseBool(getEnv("RATE_LIMIT_ENABLED", "false"))
	rateLimitRequests, _ := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "100"))

	return &Config{
		App: AppConfig{
			Env:            getEnv("APP_ENV", "development"),
			Name:           getEnv("APP_NAME", "CIP"),
			Version:        getEnv("APP_VERSION", "1.0.0"),
			APIVersion:     getEnv("API_VERSION", "v1"),
			Host:           getEnv("APP_HOST", "0.0.0.0"),
			Port:           getEnv("APP_PORT", "8080"),
			LogLevel:       getEnv("LOG_LEVEL", "info"),
			LogFormat:      getEnv("LOG_FORMAT", "json"),
			RequestTimeout: requestTimeout,
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			Name:            getEnv("DB_NAME", "cip_dev"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
			MigrationPath:   getEnv("MIGRATION_PATH", "migrations"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "change-me"),
			AccessTokenExpiry:  accessTokenExpiry,
			RefreshTokenExpiry: refreshTokenExpiry,
		},
		Security: SecurityConfig{
			BcryptCost:    bcryptCost,
			EncryptionKey: getEnv("ENCRYPTION_KEY", "change-me-in-production-32-bytes-long"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
			AllowedMethods: getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,PATCH,OPTIONS"),
		},
		RateLimit: RateLimitConfig{
			Enabled:  rateLimitEnabled,
			Requests: rateLimitRequests,
		},
		Integrations: IntegrationsConfig{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirectURI:  getEnv("GOOGLE_REDIRECT_URI", "http://localhost:8082/api/v1/integrations/google/callback"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(key, defaultValue string) (time.Duration, error) {
	value := getEnv(key, defaultValue)
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}
	return duration, nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
