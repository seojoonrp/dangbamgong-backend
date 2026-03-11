package handler

import (
	"net/http"
	"strconv"

	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/middleware"
	"dangbamgong-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type NotificationHandler struct {
	service service.NotificationService
}

func NewNotificationHandler(s service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: s}
}

func (h *NotificationHandler) GetNotifications(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	resp, err := h.service.GetNotifications(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	notifID := c.Param("notification_id")

	if err := h.service.MarkAsRead(c.Request().Context(), userID, notifID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

func (h *NotificationHandler) GetUnreadCount(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.GetUnreadCount(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}
