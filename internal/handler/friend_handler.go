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

func (h *FriendHandler) Search(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	tag := c.QueryParam("tag")

	resp, err := h.service.Search(c.Request().Context(), userID, tag)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *FriendHandler) GetFriends(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.GetFriends(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *FriendHandler) RemoveFriend(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	targetID := c.Param("user_id")

	if err := h.service.RemoveFriend(c.Request().Context(), userID, targetID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

func (h *FriendHandler) GetRequests(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	reqType := c.QueryParam("type")

	resp, err := h.service.GetRequests(c.Request().Context(), userID, reqType)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

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

func (h *FriendHandler) AcceptRequest(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	requestID := c.Param("request_id")

	if err := h.service.AcceptRequest(c.Request().Context(), userID, requestID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

func (h *FriendHandler) RejectRequest(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	requestID := c.Param("request_id")

	if err := h.service.RejectRequest(c.Request().Context(), userID, requestID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

func (h *FriendHandler) Nudge(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	targetID := c.Param("user_id")

	if err := h.service.Nudge(c.Request().Context(), userID, targetID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}
