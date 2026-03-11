package handler

import (
	"net/http"

	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/middleware"
	"dangbamgong-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type StatHandler struct {
	service service.StatService
}

func NewStatHandler(s service.StatService) *StatHandler {
	return &StatHandler{service: s}
}

func (h *StatHandler) GetHomeStat(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.GetHomeStat(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *StatHandler) GetMyVoidStat(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.GetMyVoidStat(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *StatHandler) GetDailyStat(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	targetDay := c.QueryParam("target_day")

	resp, err := h.service.GetDailyStat(c.Request().Context(), userID, targetDay)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}
