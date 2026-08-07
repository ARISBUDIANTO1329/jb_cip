package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jaybani/jb_cip/config"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/pkg/errors"
)

func AuthRequired(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.SendError(c, errors.New("AUTH_001", "Authorization header is required", 401))
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return helper.SendError(c, errors.New("AUTH_001", "Invalid authorization format", 401))
		}

		token := parts[1]

		claims, err := helper.ValidateToken(token, cfg.JWT.Secret)
		if err != nil {
			return helper.SendError(c, errors.New("AUTH_001", "Invalid or expired token", 401))
		}

		if claims.TokenType != "access" {
			return helper.SendError(c, errors.New("AUTH_001", "Invalid token type", 401))
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role")
		if userRole == nil {
			return helper.SendError(c, errors.New("AUTH_003", "Forbidden", 403))
		}

		roleStr, ok := userRole.(string)
		if !ok {
			return helper.SendError(c, errors.New("AUTH_003", "Forbidden", 403))
		}

		for _, role := range roles {
			if roleStr == role {
				return c.Next()
			}
		}

		return helper.SendError(c, errors.New("AUTH_003", "Insufficient permissions", 403))
	}
}
