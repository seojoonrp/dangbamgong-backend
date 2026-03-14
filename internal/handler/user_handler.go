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

// Search godoc
// @Summary      유저 검색
// @Description  태그 접두사로 유저를 검색합니다. 나를 차단한 유저는 제외됩니다.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        tag  query     string  true  "태그 접두사"
// @Success      200  {object}  dto.Response[dto.UserSearchResponse]
// @Failure      400  {object}  dto.ErrorResponse
// @Router       /users/search [get]
func (h *UserHandler) Search(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	tag := c.QueryParam("tag")

	resp, err := h.service.Search(c.Request().Context(), userID, tag)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// GetMe godoc
// @Summary      내 정보 조회
// @Description  현재 로그인한 유저의 프로필과 알림 설정을 반환합니다
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.Response[dto.UserMeResponse]
// @Router       /users/me [get]
func (h *UserHandler) GetMe(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.GetMe(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// UpdateSettings godoc
// @Summary      알림 설정 변경
// @Description  알림 설정을 부분 업데이트합니다. 전달된 필드만 변경됩니다.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.UpdateSettingsRequest  true  "변경할 설정"
// @Success      200   {object}  dto.Response[dto.UpdateSettingsResponse]
// @Router       /users/me/settings [patch]
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

// GetBlocks godoc
// @Summary      차단 목록 조회
// @Description  내가 차단한 유저 목록을 반환합니다
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.Response[dto.BlockListResponse]
// @Router       /users/blocks [get]
func (h *UserHandler) GetBlocks(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.GetBlocks(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// Block godoc
// @Summary      유저 차단
// @Description  유저를 차단합니다. 기존 친구 관계 및 친구 요청이 모두 삭제됩니다.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        user_id  path  string  true  "차단할 유저 ID"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.ErrorResponse  "USER_NOT_FOUND"
// @Failure      409  {object}  dto.ErrorResponse  "ALREADY_BLOCKED"
// @Router       /users/{user_id}/block [post]
func (h *UserHandler) Block(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	targetID := c.Param("user_id")

	if err := h.service.Block(c.Request().Context(), userID, targetID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

// ChangeNickname godoc
// @Summary      닉네임 변경
// @Description  기존 닉네임을 변경합니다. 최초 설정은 /auth/nickname을 사용하세요.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.ChangeNicknameRequest  true  "새 닉네임 (3-15자)"
// @Success      200   {object}  dto.Response[dto.ChangeNicknameResponse]
// @Failure      400   {object}  dto.ErrorResponse  "INVALID_NICKNAME"
// @Router       /users/me/nickname [patch]
func (h *UserHandler) ChangeNickname(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req dto.ChangeNicknameRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	resp, err := h.service.ChangeNickname(c.Request().Context(), userID, req)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// Unblock godoc
// @Summary      유저 차단 해제
// @Description  차단을 해제합니다
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        user_id  path  string  true  "차단 해제할 유저 ID"
// @Success      200  {object}  dto.Response[any]
// @Failure      400  {object}  dto.ErrorResponse  "NOT_BLOCKED"
// @Router       /users/{user_id}/unblock [post]
func (h *UserHandler) Unblock(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	targetID := c.Param("user_id")

	if err := h.service.Unblock(c.Request().Context(), userID, targetID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}
