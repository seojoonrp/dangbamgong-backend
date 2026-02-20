package domain

import "net/http"

type ErrorCode string

// Common
const (
	ErrBadRequest         ErrorCode = "BAD_REQUEST"
	ErrNotFound           ErrorCode = "NOT_FOUND"
	ErrInternalServer     ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrForbidden          ErrorCode = "FORBIDDEN"
	ErrConflict           ErrorCode = "CONFLICT"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// Auth
const (
	ErrInvalidToken       ErrorCode = "INVALID_TOKEN"
	ErrInvalidNickname    ErrorCode = "INVALID_NICKNAME"
	ErrNicknameAlreadySet ErrorCode = "NICKNAME_ALREADY_SET"
)

// User
const (
	ErrUserNotFound   ErrorCode = "USER_NOT_FOUND"
	ErrAlreadyBlocked ErrorCode = "ALREADY_BLOCKED"
	ErrNotBlocked     ErrorCode = "NOT_BLOCKED"
)

// Void
const (
	ErrAlreadyInVoid     ErrorCode = "ALREADY_IN_VOID"
	ErrNotInVoid         ErrorCode = "NOT_IN_VOID"
	ErrTooManyActivities ErrorCode = "TOO_MANY_ACTIVITIES"
)

// Activity
const (
	ErrInvalidActivityName   ErrorCode = "INVALID_ACTIVITY_NAME"
	ErrActivityAlreadyExists ErrorCode = "ACTIVITY_ALREADY_EXISTS"
	ErrActivityNotFound      ErrorCode = "ACTIVITY_NOT_FOUND"
)

// Friend
const (
	ErrAlreadyFriends     ErrorCode = "ALREADY_FRIENDS"
	ErrRequestAlreadySent ErrorCode = "REQUEST_ALREADY_SENT"
	ErrBlocked            ErrorCode = "BLOCKED"
	ErrRequestNotFound    ErrorCode = "REQUEST_NOT_FOUND"
	ErrRequestNotPending  ErrorCode = "REQUEST_NOT_PENDING"
	ErrNotFriends         ErrorCode = "NOT_FRIENDS"
	ErrFriendNotInVoid    ErrorCode = "FRIEND_NOT_IN_VOID"
	ErrInvalidRequestType ErrorCode = "INVALID_REQUEST_TYPE"
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
