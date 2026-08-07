package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/pkg/errors"
)

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)
		return c.Next()
	}
}

func APIVersion(version string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("api_version", version)
		return c.Next()
	}
}

func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := c.Response().StatusCode()
		if code == 0 || code == 200 {
			code = fiber.StatusInternalServerError
		}

		appErr, ok := err.(*errors.AppError)
		if !ok {
			if e, ok := err.(*fiber.Error); ok {
				appErr = errors.New("SYSTEM_001", e.Message, e.Code)
			} else {
				appErr = errors.New("SYSTEM_001", err.Error(), code)
			}
		}

		return helper.SendError(c, appErr)
	}
}

func NotFound() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helper.SendError(c, errors.New("NOT_FOUND", "Route not found", 404))
	}
}
