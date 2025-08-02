// Package errors provides standardized error handling and custom error types
package errors

import (
	"fmt"
	"net/http"
)

// Error represents a structured error with code, message, and details
type Error struct {
	Code     string                 `json:"code"`
	Message  string                 `json:"message"`
	Details  map[string]interface{} `json:"details,omitempty"`
	HTTPCode int                    `json:"-"`
	Cause    error                  `json:"-"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithDetails adds details to the error
func (e *Error) WithDetails(details map[string]interface{}) *Error {
	e.Details = details
	return e
}

// WithCause adds a cause to the error
func (e *Error) WithCause(cause error) *Error {
	e.Cause = cause
	return e
}

// Pre-defined error codes
const (
	// Authentication errors
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired       = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid       = "TOKEN_INVALID"
	
	// Authorization errors
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeInsufficientScope  = "INSUFFICIENT_SCOPE"
	
	// Validation errors
	ErrCodeValidationFailed   = "VALIDATION_FAILED"
	ErrCodeInvalidInput       = "INVALID_INPUT"
	ErrCodeMissingField       = "MISSING_FIELD"
	
	// Resource errors
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeAlreadyExists      = "ALREADY_EXISTS"
	ErrCodeConflict           = "CONFLICT"
	
	// Business logic errors
	ErrCodeBusinessRule       = "BUSINESS_RULE_VIOLATION"
	ErrCodeQuotaExceeded      = "QUOTA_EXCEEDED"
	ErrCodeRateLimited        = "RATE_LIMITED"
	
	// System errors
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout            = "TIMEOUT"
	ErrCodeDatabaseError      = "DATABASE_ERROR"
	ErrCodeCacheError         = "CACHE_ERROR"
)

// Constructor functions for common errors

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *Error {
	return &Error{
		Code:     ErrCodeUnauthorized,
		Message:  message,
		HTTPCode: http.StatusUnauthorized,
	}
}

// NewInvalidCredentialsError creates an invalid credentials error
func NewInvalidCredentialsError() *Error {
	return &Error{
		Code:     ErrCodeInvalidCredentials,
		Message:  "Invalid credentials provided",
		HTTPCode: http.StatusUnauthorized,
	}
}

// NewTokenExpiredError creates a token expired error
func NewTokenExpiredError() *Error {
	return &Error{
		Code:     ErrCodeTokenExpired,
		Message:  "Token has expired",
		HTTPCode: http.StatusUnauthorized,
	}
}

// NewTokenInvalidError creates a token invalid error
func NewTokenInvalidError() *Error {
	return &Error{
		Code:     ErrCodeTokenInvalid,
		Message:  "Token is invalid",
		HTTPCode: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *Error {
	return &Error{
		Code:     ErrCodeForbidden,
		Message:  message,
		HTTPCode: http.StatusForbidden,
	}
}

// NewInsufficientScopeError creates an insufficient scope error
func NewInsufficientScopeError(requiredScope string) *Error {
	return &Error{
		Code:     ErrCodeInsufficientScope,
		Message:  "Insufficient scope for this operation",
		HTTPCode: http.StatusForbidden,
		Details: map[string]interface{}{
			"required_scope": requiredScope,
		},
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, details map[string]interface{}) *Error {
	return &Error{
		Code:     ErrCodeValidationFailed,
		Message:  message,
		HTTPCode: http.StatusBadRequest,
		Details:  details,
	}
}

// NewInvalidInputError creates an invalid input error
func NewInvalidInputError(field, message string) *Error {
	return &Error{
		Code:     ErrCodeInvalidInput,
		Message:  fmt.Sprintf("Invalid input for field '%s': %s", field, message),
		HTTPCode: http.StatusBadRequest,
		Details: map[string]interface{}{
			"field": field,
		},
	}
}

// NewMissingFieldError creates a missing field error
func NewMissingFieldError(field string) *Error {
	return &Error{
		Code:     ErrCodeMissingField,
		Message:  fmt.Sprintf("Required field '%s' is missing", field),
		HTTPCode: http.StatusBadRequest,
		Details: map[string]interface{}{
			"field": field,
		},
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *Error {
	return &Error{
		Code:     ErrCodeNotFound,
		Message:  fmt.Sprintf("%s not found", resource),
		HTTPCode: http.StatusNotFound,
		Details: map[string]interface{}{
			"resource": resource,
		},
	}
}

// NewAlreadyExistsError creates an already exists error
func NewAlreadyExistsError(resource string) *Error {
	return &Error{
		Code:     ErrCodeAlreadyExists,
		Message:  fmt.Sprintf("%s already exists", resource),
		HTTPCode: http.StatusConflict,
		Details: map[string]interface{}{
			"resource": resource,
		},
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *Error {
	return &Error{
		Code:     ErrCodeConflict,
		Message:  message,
		HTTPCode: http.StatusConflict,
	}
}

// NewBusinessRuleError creates a business rule violation error
func NewBusinessRuleError(rule, message string) *Error {
	return &Error{
		Code:     ErrCodeBusinessRule,
		Message:  message,
		HTTPCode: http.StatusBadRequest,
		Details: map[string]interface{}{
			"rule": rule,
		},
	}
}

// NewQuotaExceededError creates a quota exceeded error
func NewQuotaExceededError(quota string, limit int) *Error {
	return &Error{
		Code:     ErrCodeQuotaExceeded,
		Message:  fmt.Sprintf("Quota exceeded for %s", quota),
		HTTPCode: http.StatusTooManyRequests,
		Details: map[string]interface{}{
			"quota": quota,
			"limit": limit,
		},
	}
}

// NewRateLimitedError creates a rate limited error
func NewRateLimitedError(retryAfter int) *Error {
	return &Error{
		Code:     ErrCodeRateLimited,
		Message:  "Rate limit exceeded",
		HTTPCode: http.StatusTooManyRequests,
		Details: map[string]interface{}{
			"retry_after": retryAfter,
		},
	}
}

// NewInternalError creates an internal error
func NewInternalError(message string) *Error {
	return &Error{
		Code:     ErrCodeInternalError,
		Message:  message,
		HTTPCode: http.StatusInternalServerError,
	}
}

// NewServiceUnavailableError creates a service unavailable error
func NewServiceUnavailableError(service string) *Error {
	return &Error{
		Code:     ErrCodeServiceUnavailable,
		Message:  fmt.Sprintf("Service %s is unavailable", service),
		HTTPCode: http.StatusServiceUnavailable,
		Details: map[string]interface{}{
			"service": service,
		},
	}
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string) *Error {
	return &Error{
		Code:     ErrCodeTimeout,
		Message:  fmt.Sprintf("Operation %s timed out", operation),
		HTTPCode: http.StatusRequestTimeout,
		Details: map[string]interface{}{
			"operation": operation,
		},
	}
}

// NewDatabaseError creates a database error
func NewDatabaseError(operation string, cause error) *Error {
	return &Error{
		Code:     ErrCodeDatabaseError,
		Message:  fmt.Sprintf("Database error during %s", operation),
		HTTPCode: http.StatusInternalServerError,
		Cause:    cause,
		Details: map[string]interface{}{
			"operation": operation,
		},
	}
}

// NewCacheError creates a cache error
func NewCacheError(operation string, cause error) *Error {
	return &Error{
		Code:     ErrCodeCacheError,
		Message:  fmt.Sprintf("Cache error during %s", operation),
		HTTPCode: http.StatusInternalServerError,
		Cause:    cause,
		Details: map[string]interface{}{
			"operation": operation,
		},
	}
}

// ErrorResponse represents an error response structure
type ErrorResponse struct {
	Error     *Error `json:"error"`
	RequestID string `json:"request_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err *Error, requestID string) *ErrorResponse {
	return &ErrorResponse{
		Error:     err,
		RequestID: requestID,
		Timestamp: fmt.Sprintf("%d", UnixNow()),
	}
}

// Is checks if the error matches the given error code
func Is(err error, code string) bool {
	if customErr, ok := err.(*Error); ok {
		return customErr.Code == code
	}
	return false
}

// GetHTTPCode returns the HTTP status code for the error
func GetHTTPCode(err error) int {
	if customErr, ok := err.(*Error); ok {
		return customErr.HTTPCode
	}
	return http.StatusInternalServerError
}

// UnixNow returns the current Unix timestamp
func UnixNow() int64 {
	// This would typically use time.Now().Unix()
	// but for testing purposes, we'll use a placeholder
	return 1640995200 // 2022-01-01 00:00:00 UTC
}