package handler

import (
	"net/http"

	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/middleware"
	"dangbamgong-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type VoidHandler struct {
	service service.VoidService
}

func NewVoidHandler(s service.VoidService) *VoidHandler {
	return &VoidHandler{service: s}
}

func (h *VoidHandler) Start(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.Start(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *VoidHandler) End(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req dto.VoidEndRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	resp, err := h.service.End(c.Request().Context(), userID, req)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *VoidHandler) Cancel(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	if err := h.service.Cancel(c.Request().Context(), userID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

func (h *VoidHandler) TestCreate(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req dto.TestVoidRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	resp, err := h.service.TestCreate(c.Request().Context(), userID, req)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusCreated, resp)
}

func (h *VoidHandler) History(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	targetDay := c.QueryParam("target_day")

	resp, err := h.service.History(c.Request().Context(), userID, targetDay)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}
