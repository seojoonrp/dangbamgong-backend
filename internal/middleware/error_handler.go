// internal/middleware/error_handler.go
package middleware

import (
	"errors"
	"log"
	"net/http"

	"dangbamgong-backend/internal/domain"
	"dangbamgong-backend/internal/dto"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		log.Printf("[AppError] status=%d code=%s message=%s", appErr.StatusCode, appErr.Code, appErr.Message)
		_ = dto.Fail(c, appErr.StatusCode, appErr.Code)
		return
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		log.Printf("[ValidationError] %v", validationErrors)
		_ = dto.Fail(c, http.StatusBadRequest, domain.ErrBadRequest)
		return
	}

	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) {
		code := mapHTTPStatusToErrorCode(echoErr.Code)
		log.Printf("[HTTPError] status=%d message=%v", echoErr.Code, echoErr.Message)
		_ = dto.Fail(c, echoErr.Code, code)
		return
	}

	log.Printf("[UnhandledError] %v", err)
	_ = dto.Fail(c, http.StatusInternalServerError, domain.ErrInternalServer)
}

func mapHTTPStatusToErrorCode(status int) domain.ErrorCode {
	switch status {
	case http.StatusBadRequest:
		return domain.ErrBadRequest
	case http.StatusNotFound:
		return domain.ErrNotFound
	case http.StatusUnauthorized:
		return domain.ErrUnauthorized
	case http.StatusForbidden:
		return domain.ErrForbidden
	case http.StatusConflict:
		return domain.ErrConflict
	default:
		return domain.ErrInternalServer
	}
}
