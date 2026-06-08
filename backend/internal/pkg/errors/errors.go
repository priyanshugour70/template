package errors

import "fmt"

type AppError struct {
	Code    string
	Message string
	Details map[string]interface{}
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

func New(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func NewWithDetails(code, message string, details map[string]interface{}, err error) *AppError {
	return &AppError{Code: code, Message: message, Details: details, Err: err}
}

const (
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeInvalidToken       = "INVALID_TOKEN"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeValidationError    = "VALIDATION_ERROR"
	CodeInvalidInput       = "INVALID_INPUT"
	CodeNotFound           = "NOT_FOUND"
	CodeAlreadyExists      = "ALREADY_EXISTS"
	CodeInternalError      = "INTERNAL_ERROR"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeTimeout            = "TIMEOUT"
	CodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	CodeBadRequest         = "BAD_REQUEST"

	CodeValidation = CodeValidationError
	CodeInternal   = CodeInternalError
	CodeConflict   = CodeAlreadyExists
)
