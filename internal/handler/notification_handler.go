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

// GetNotifications godoc
// @Summary      알림 목록 조회
// @Description  유저의 알림 목록을 페이지네이션으로 반환합니다
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query     int  false  "조회 개수 (기본 20)"
// @Param        offset  query     int  false  "오프셋"
// @Success      200  {object}  dto.Response[dto.NotificationListResponse]
// @Router       /notifications [get]
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

// MarkAsRead godoc
// @Summary      알림 읽음 처리
// @Description  특정 알림을 읽음 상태로 변경합니다
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Param        notification_id  path  string  true  "알림 ID"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.ErrorResponse  "NOTIFICATION_NOT_FOUND"
// @Router       /notifications/{notification_id}/read [patch]
func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	notifID := c.Param("notification_id")

	if err := h.service.MarkAsRead(c.Request().Context(), userID, notifID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

// GetUnreadCount godoc
// @Summary      읽지 않은 알림 수 조회
// @Description  읽지 않은 알림의 개수를 반환합니다
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.Response[dto.UnreadCountResponse]
// @Router       /notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.GetUnreadCount(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}
