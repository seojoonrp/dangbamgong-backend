package domain

import "net/http"

type ErrorCode string

const (
	ErrBadRequest          ErrorCode = "BAD_REQUEST"
	ErrNotFound            ErrorCode = "NOT_FOUND"
	ErrInternalServer      ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrForbidden           ErrorCode = "FORBIDDEN"
	ErrConflict            ErrorCode = "CONFLICT"
	ErrUserAlreadyExists   ErrorCode = "USER_ALREADY_EXISTS"
	ErrServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
)

type AppError struct {
	StatusCode int
	Code       ErrorCode
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}

func NewBadRequest(code ErrorCode, message string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       code,
		Message:    message,
	}
}

func NewNotFound(code ErrorCode, message string) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Code:       code,
		Message:    message,
	}
}

func NewUnauthorized(code ErrorCode, message string) *AppError {
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		Code:       code,
		Message:    message,
	}
}

func NewForbidden(code ErrorCode, message string) *AppError {
	return &AppError{
		StatusCode: http.StatusForbidden,
		Code:       code,
		Message:    message,
	}
}

func NewConflict(code ErrorCode, message string) *AppError {
	return &AppError{
		StatusCode: http.StatusConflict,
		Code:       code,
		Message:    message,
	}
}

func NewInternal(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       ErrInternalServer,
		Message:    message,
	}
}
