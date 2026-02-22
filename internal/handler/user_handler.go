package handler

import (
	"net/http"

	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/middleware"
	"dangbamgong-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) GetMe(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.GetMe(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *UserHandler) UpdateSettings(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req dto.UpdateSettingsRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	resp, err := h.service.UpdateSettings(c.Request().Context(), userID, req)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *UserHandler) GetBlocks(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.GetBlocks(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *UserHandler) Block(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	targetID := c.Param("user_id")

	if err := h.service.Block(c.Request().Context(), userID, targetID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

func (h *UserHandler) Unblock(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	targetID := c.Param("user_id")

	if err := h.service.Unblock(c.Request().Context(), userID, targetID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}
