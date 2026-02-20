package handler

import (
	"net/http"

	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/middleware"
	"dangbamgong-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type ActivityHandler struct {
	service service.ActivityService
}

func NewActivityHandler(s service.ActivityService) *ActivityHandler {
	return &ActivityHandler{service: s}
}

func (h *ActivityHandler) List(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.List(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *ActivityHandler) Create(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req dto.CreateActivityRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	resp, err := h.service.Create(c.Request().Context(), userID, req)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusCreated, resp)
}

func (h *ActivityHandler) Delete(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	activityID := c.Param("activity_id")

	if err := h.service.Delete(c.Request().Context(), userID, activityID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}
