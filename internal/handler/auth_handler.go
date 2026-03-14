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

// Login godoc
// @Summary      소셜 로그인
// @Description  Google/Kakao/Apple 소셜 로그인을 처리합니다. 신규 유저인 경우 자동 가입됩니다.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "소셜 로그인 정보"
// @Success      200   {object}  dto.Response[dto.LoginResponse]
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	resp, err := h.service.Login(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// TestLogin godoc
// @Summary      테스트 로그인
// @Description  개발용 테스트 로그인. socialId로 직접 로그인합니다.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.TestLoginRequest  true  "테스트 로그인 정보"
// @Success      200   {object}  dto.Response[dto.LoginResponse]
// @Router       /auth/login/test [post]
func (h *AuthHandler) TestLogin(c echo.Context) error {
	var req dto.TestLoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	resp, err := h.service.TestLogin(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// SetNickname godoc
// @Summary      닉네임 설정
// @Description  신규 가입 후 최초 닉네임을 설정합니다. 이미 설정된 경우 실패합니다.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.SetNicknameRequest  true  "닉네임 (3-15자)"
// @Success      200   {object}  dto.Response[dto.SetNicknameResponse]
// @Failure      400   {object}  dto.ErrorResponse  "INVALID_NICKNAME"
// @Failure      409   {object}  dto.ErrorResponse  "NICKNAME_ALREADY_SET"
// @Router       /auth/nickname [post]
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

// Withdraw godoc
// @Summary      회원 탈퇴
// @Description  계정을 삭제합니다. (미완성)
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  dto.Response[any]
// @Router       /auth/withdraw [delete]
func (h *AuthHandler) Withdraw(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	if err := h.service.Withdraw(c.Request().Context(), userID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}
