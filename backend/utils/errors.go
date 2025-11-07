package utils

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound       ErrorType = "NOT_FOUND"
	ErrorTypeUnauthorized   ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden      ErrorType = "FORBIDDEN"
	ErrorTypeConflict       ErrorType = "CONFLICT"
	ErrorTypeInternal       ErrorType = "INTERNAL_ERROR"
	ErrorTypeBadRequest     ErrorType = "BAD_REQUEST"
	ErrorTypeTimeout        ErrorType = "TIMEOUT"
	ErrorTypeRateLimit      ErrorType = "RATE_LIMIT"
	ErrorTypeDatabase       ErrorType = "DATABASE_ERROR"
	ErrorTypeExternal       ErrorType = "EXTERNAL_SERVICE_ERROR"
)

// AppError represents a custom application error
type AppError struct {
	Type       ErrorType         `json:"type"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	StatusCode int               `json:"-"`
	Err        error             `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, message string, statusCode int) *AppError {
	return &AppError{
		Type:       errorType,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewAppErrorWithDetails creates a new application error with details
func NewAppErrorWithDetails(errorType ErrorType, message string, statusCode int, details map[string]string) *AppError {
	return &AppError{
		Type:       errorType,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
	}
}

// WrapError wraps an existing error as an AppError
func WrapError(err error, errorType ErrorType, message string, statusCode int) *AppError {
	return &AppError{
		Type:       errorType,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Predefined error constructors
func NewValidationError(message string, details map[string]string) *AppError {
	return NewAppErrorWithDetails(ErrorTypeValidation, message, http.StatusBadRequest, details)
}

func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrorTypeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "Unauthorized access"
	}
	return NewAppError(ErrorTypeUnauthorized, message, http.StatusUnauthorized)
}

func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "Access forbidden"
	}
	return NewAppError(ErrorTypeForbidden, message, http.StatusForbidden)
}

func NewConflictError(message string) *AppError {
	return NewAppError(ErrorTypeConflict, message, http.StatusConflict)
}

func NewInternalError(message string, err error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return WrapError(err, ErrorTypeInternal, message, http.StatusInternalServerError)
}

func NewBadRequestError(message string) *AppError {
	return NewAppError(ErrorTypeBadRequest, message, http.StatusBadRequest)
}

func NewDatabaseError(operation string, err error) *AppError {
	message := fmt.Sprintf("Database operation failed: %s", operation)
	return WrapError(err, ErrorTypeDatabase, message, http.StatusInternalServerError)
}

func NewTimeoutError(operation string) *AppError {
	message := fmt.Sprintf("Operation timed out: %s", operation)
	return NewAppError(ErrorTypeTimeout, message, http.StatusRequestTimeout)
}

func NewRateLimitError(message string) *AppError {
	if message == "" {
		message = "Rate limit exceeded"
	}
	return NewAppError(ErrorTypeRateLimit, message, http.StatusTooManyRequests)
}

func NewExternalServiceError(service string, err error) *AppError {
	message := fmt.Sprintf("External service error: %s", service)
	return WrapError(err, ErrorTypeExternal, message, http.StatusBadGateway)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from error
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    ErrorType         `json:"code"`
	Details map[string]string `json:"details,omitempty"`
	TraceID string            `json:"trace_id,omitempty"`
}

// ToErrorResponse converts AppError to ErrorResponse
func (e *AppError) ToErrorResponse(traceID string) ErrorResponse {
	return ErrorResponse{
		Error:   e.Message,
		Code:    e.Type,
		Details: e.Details,
		TraceID: traceID,
	}
}
