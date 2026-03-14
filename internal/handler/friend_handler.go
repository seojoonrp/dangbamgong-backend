package handler

import (
	"net/http"

	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/middleware"
	"dangbamgong-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type FriendHandler struct {
	service service.FriendService
}

func NewFriendHandler(s service.FriendService) *FriendHandler {
	return &FriendHandler{service: s}
}

// GetFriends godoc
// @Summary      친구 목록 조회
// @Description  친구 목록을 반환합니다. 각 친구의 공백 상태를 포함합니다.
// @Tags         Friends
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.Response[dto.FriendListResponse]
// @Router       /friends [get]
func (h *FriendHandler) GetFriends(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.GetFriends(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// RemoveFriend godoc
// @Summary      친구 삭제
// @Description  양방향 친구 관계를 삭제합니다
// @Tags         Friends
// @Produce      json
// @Security     BearerAuth
// @Param        user_id  path  string  true  "삭제할 친구 유저 ID"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.ErrorResponse  "NOT_FRIENDS"
// @Router       /friends/{user_id} [delete]
func (h *FriendHandler) RemoveFriend(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	targetID := c.Param("user_id")

	if err := h.service.RemoveFriend(c.Request().Context(), userID, targetID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

// GetRequests godoc
// @Summary      친구 요청 목록 조회
// @Description  받은(received) 또는 보낸(sent) 친구 요청 목록을 반환합니다
// @Tags         Friends
// @Produce      json
// @Security     BearerAuth
// @Param        type  query     string  true  "요청 타입"  Enums(received, sent)
// @Success      200   {object}  dto.Response[dto.ReceivedRequestsResponse]  "type=received"
// @Failure      400   {object}  dto.ErrorResponse  "INVALID_REQUEST_TYPE"
// @Router       /friends/requests [get]
func (h *FriendHandler) GetRequests(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	reqType := c.QueryParam("type")

	resp, err := h.service.GetRequests(c.Request().Context(), userID, reqType)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// SendRequest godoc
// @Summary      친구 요청 보내기
// @Description  유저에게 친구 요청을 보냅니다. 이미 친구이거나 차단 관계이면 실패합니다.
// @Tags         Friends
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.SendFriendRequestRequest  true  "받는 유저 ID"
// @Success      201   {object}  dto.Response[dto.SendFriendRequestResponse]
// @Failure      400   {object}  dto.ErrorResponse  "BLOCKED"
// @Failure      409   {object}  dto.ErrorResponse  "ALREADY_FRIENDS / REQUEST_ALREADY_SENT"
// @Router       /friends/requests [post]
func (h *FriendHandler) SendRequest(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req dto.SendFriendRequestRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	resp, err := h.service.SendRequest(c.Request().Context(), userID, req)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusCreated, resp)
}

// AcceptRequest godoc
// @Summary      친구 요청 수락
// @Description  받은 친구 요청을 수락합니다. 양방향 친구 관계가 생성됩니다.
// @Tags         Friends
// @Produce      json
// @Security     BearerAuth
// @Param        request_id  path  string  true  "요청 ID"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.ErrorResponse  "REQUEST_NOT_FOUND"
// @Failure      400  {object}  dto.ErrorResponse  "REQUEST_NOT_PENDING"
// @Router       /friends/requests/{request_id}/accept [post]
func (h *FriendHandler) AcceptRequest(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	requestID := c.Param("request_id")

	if err := h.service.AcceptRequest(c.Request().Context(), userID, requestID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

// RejectRequest godoc
// @Summary      친구 요청 거절
// @Description  받은 친구 요청을 거절합니다. 요청은 삭제되지 않고 상태만 변경됩니다.
// @Tags         Friends
// @Produce      json
// @Security     BearerAuth
// @Param        request_id  path  string  true  "요청 ID"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.ErrorResponse  "REQUEST_NOT_FOUND"
// @Failure      400  {object}  dto.ErrorResponse  "REQUEST_NOT_PENDING"
// @Router       /friends/requests/{request_id}/reject [post]
func (h *FriendHandler) RejectRequest(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	requestID := c.Param("request_id")

	if err := h.service.RejectRequest(c.Request().Context(), userID, requestID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

// DeleteRequest godoc
// @Summary      친구 요청 삭제
// @Description  보낸 친구 요청을 삭제합니다. 이미 수락된 요청은 삭제할 수 없습니다.
// @Tags         Friends
// @Produce      json
// @Security     BearerAuth
// @Param        request_id  path  string  true  "요청 ID"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.ErrorResponse  "REQUEST_NOT_FOUND"
// @Router       /friends/requests/{request_id} [delete]
func (h *FriendHandler) DeleteRequest(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	requestID := c.Param("request_id")

	if err := h.service.DeleteRequest(c.Request().Context(), userID, requestID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

// Nudge godoc
// @Summary      친구 찌르기
// @Description  공백 중인 친구에게 알림을 보냅니다. 친구가 공백 중이 아니면 실패합니다.
// @Tags         Friends
// @Produce      json
// @Security     BearerAuth
// @Param        user_id  path  string  true  "찌를 친구 유저 ID"
// @Success      200  {object}  dto.Response[any]
// @Failure      400  {object}  dto.ErrorResponse  "NOT_FRIENDS / FRIEND_NOT_IN_VOID"
// @Router       /friends/{user_id}/nudge [post]
func (h *FriendHandler) Nudge(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	targetID := c.Param("user_id")

	if err := h.service.Nudge(c.Request().Context(), userID, targetID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}
