// internal/middleware/error_handler.go
package middleware

import (
	"errors"
	"fmt"
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
		fmt.Printf("\n\033[33m[AppError] %d | %s | %s\033[0m\n",
			appErr.StatusCode, appErr.Code, appErr.Message)
		_ = dto.Fail(c, appErr.StatusCode, appErr.Code)
		return
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		fmt.Printf("\n\033[36m[ValidationError] %v\033[0m\n", validationErrors)
		_ = dto.Fail(c, http.StatusBadRequest, domain.ErrBadRequest)
		return
	}

	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) {
		code := mapHTTPStatusToErrorCode(echoErr.Code)
		fmt.Printf("\n\033[35m[HTTPError] %d | %v\033[0m\n",
			echoErr.Code, echoErr.Message)
		_ = dto.Fail(c, echoErr.Code, code)
		return
	}

	fmt.Printf("\n\033[31m[UnhandledError] %v\033[0m\n", err)
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
