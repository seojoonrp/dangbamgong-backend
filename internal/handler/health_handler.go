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

// Health godoc
// @Summary      헬스 체크
// @Description  DB 연결 상태를 확인합니다
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /health [get]
func (h *HealthHandler) Health(c echo.Context) error {
	if err := h.service.Health(); err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, map[string]string{
		"message": "It's healthy",
	})
}

// HelloWorld godoc
// @Summary      Hello World
// @Description  서버 동작 확인용 엔드포인트
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       / [get]
func (h *HealthHandler) HelloWorld(c echo.Context) error {
	return dto.Success(c, http.StatusOK, map[string]string{
		"message": "Hello World",
	})
}
