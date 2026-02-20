// internal/dto/response.go
package dto

import (
	"dangbamgong-backend/internal/domain"

	"github.com/labstack/echo/v4"
)

type Response[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool             `json:"success"`
	Code    domain.ErrorCode `json:"code"`
}

func Success[T any](c echo.Context, statusCode int, data T) error {
	return c.JSON(statusCode, Response[T]{
		Success: true,
		Data:    data,
	})
}

func SuccessEmpty(c echo.Context, statusCode int) error {
	return c.JSON(statusCode, Response[any]{
		Success: true,
	})
}

func Fail(c echo.Context, statusCode int, code domain.ErrorCode) error {
	return c.JSON(statusCode, ErrorResponse{
		Success: false,
		Code:    code,
	})
}
