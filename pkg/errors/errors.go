package errors

import "fmt"

// AppError represents a standardized application error
type AppError struct {
	Code       string
	Message    string
	Details    []ErrorDetail
	Hint       string
	StatusCode int
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func New(code, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: status,
	}
}

func (e *AppError) WithHint(hint string) *AppError {
	e.Hint = hint
	return e
}

func (e *AppError) WithDetails(details []ErrorDetail) *AppError {
	e.Details = details
	return e
}

// Predefined errors
var (
	ErrInternalServer      = New("SYSTEM_001", "Internal server error", 500)
	ErrNotFound            = New("NOT_FOUND", "Resource not found", 404)
	ErrUnauthorized        = New("AUTH_001", "Unauthorized", 401)
	ErrForbidden           = New("AUTH_003", "Forbidden", 403)
	ErrBadRequest          = New("VALIDATION_001", "Bad request", 400)
	ErrUnprocessableEntity = New("VALIDATION_002", "Unprocessable entity", 422)
	ErrTooManyRequests     = New("RATE_LIMIT_001", "Too many requests", 429)
	ErrServiceUnavailable  = New("SYSTEM_002", "Service unavailable", 503)
)
