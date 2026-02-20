package handler

import (
	"net/http"

	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	service service.HealthService
}

func NewHealthHandler(s service.HealthService) *HealthHandler {
	return &HealthHandler{service: s}
}

func (h *HealthHandler) Health(c echo.Context) error {
	if err := h.service.Health(); err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, map[string]string{
		"message": "It's healthy",
	})
}

func (h *HealthHandler) HelloWorld(c echo.Context) error {
	return dto.Success(c, http.StatusOK, map[string]string{
		"message": "Hello World",
	})
}
