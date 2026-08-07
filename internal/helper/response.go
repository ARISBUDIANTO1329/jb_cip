package helper

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jaybani/jb_cip/pkg/errors"
)

type ResponseData struct {
	Success    bool              `json:"success"`
	Message    string            `json:"message"`
	Data       interface{}       `json:"data,omitempty"`
	Error      *ResponseError    `json:"error,omitempty"`
	Meta       ResponseMeta      `json:"meta"`
	Pagination *PaginationMeta   `json:"pagination,omitempty"`
}

type ResponseError struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
	TraceID string   `json:"trace_id"`
	Hint    string   `json:"hint,omitempty"`
}

type ResponseMeta struct {
	Timestamp   string `json:"timestamp"`
	RequestID   string `json:"request_id"`
	APIVersion  string `json:"api_version"`
	WorkspaceID string `json:"workspace_id,omitempty"`
}

type PaginationMeta struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func getRequestID(c *fiber.Ctx) string {
	id := c.Locals("request_id")
	if id != nil {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return uuid.New().String()
}

func getAPIVersion(c *fiber.Ctx) string {
	v := c.Locals("api_version")
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "v1"
}

func SendSuccess(c *fiber.Ctx, message string, data interface{}, meta interface{}) error {
	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
		"meta": fiber.Map{
			"timestamp":   nowISO(),
			"request_id":  getRequestID(c),
			"api_version": getAPIVersion(c),
		},
	})
}

func SendSuccessWithPagination(c *fiber.Ctx, message string, data interface{}, pagination interface{}) error {
	return c.Status(200).JSON(fiber.Map{
		"success":    true,
		"message":    message,
		"data":       data,
		"pagination": pagination,
		"meta": fiber.Map{
			"timestamp":   nowISO(),
			"request_id":  getRequestID(c),
			"api_version": getAPIVersion(c),
		},
	})
}

func SendError(c *fiber.Ctx, appErr *errors.AppError) error {
	return c.Status(appErr.StatusCode).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    appErr.Code,
			"message": appErr.Message,
			"trace_id": uuid.New().String(),
		},
		"meta": fiber.Map{
			"timestamp":   nowISO(),
			"request_id":  getRequestID(c),
			"api_version": getAPIVersion(c),
		},
	})
}

func SendPaginatedError(c *fiber.Ctx, appErr *errors.AppError, pagination interface{}) error {
	return c.Status(appErr.StatusCode).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    appErr.Code,
			"message": appErr.Message,
			"trace_id": uuid.New().String(),
		},
		"pagination": pagination,
		"meta": fiber.Map{
			"timestamp":   nowISO(),
			"request_id":  getRequestID(c),
			"api_version": getAPIVersion(c),
		},
	})
}
