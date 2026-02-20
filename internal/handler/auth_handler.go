package handler

import (
	"net/http"

	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/middleware"
	"dangbamgong-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	resp, err := h.service.Login(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *AuthHandler) TestLogin(c echo.Context) error {
	var req dto.TestLoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	resp, err := h.service.TestLogin(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *AuthHandler) SetNickname(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req dto.SetNicknameRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	resp, err := h.service.SetNickname(c.Request().Context(), userID, req)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

func (h *AuthHandler) Withdraw(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	if err := h.service.Withdraw(c.Request().Context(), userID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}
