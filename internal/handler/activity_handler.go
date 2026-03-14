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

// List godoc
// @Summary      활동 목록 조회
// @Description  유저의 활동 목록을 사용 빈도순으로 반환합니다
// @Tags         Activities
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.Response[dto.ActivityListResponse]
// @Router       /activities [get]
func (h *ActivityHandler) List(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.List(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// Create godoc
// @Summary      활동 생성
// @Description  새로운 활동을 생성합니다. 이름은 1-10자, 중복 불가.
// @Tags         Activities
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateActivityRequest  true  "활동 이름"
// @Success      201   {object}  dto.Response[dto.CreateActivityResponse]
// @Failure      400   {object}  dto.ErrorResponse  "INVALID_ACTIVITY_NAME"
// @Failure      409   {object}  dto.ErrorResponse  "ACTIVITY_ALREADY_EXISTS"
// @Router       /activities [post]
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

// UpdateName godoc
// @Summary      활동 이름 수정
// @Description  기존 활동의 이름을 수정합니다
// @Tags         Activities
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        activity_id  path  string                     true  "활동 ID"
// @Param        body         body  dto.UpdateActivityRequest  true  "새 이름"
// @Success      200  {object}  dto.Response[any]
// @Failure      400  {object}  dto.ErrorResponse  "INVALID_ACTIVITY_NAME"
// @Failure      404  {object}  dto.ErrorResponse  "ACTIVITY_NOT_FOUND"
// @Failure      409  {object}  dto.ErrorResponse  "ACTIVITY_ALREADY_EXISTS"
// @Router       /activities/{activity_id} [patch]
func (h *ActivityHandler) UpdateName(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	activityID := c.Param("activity_id")

	var req dto.UpdateActivityRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.service.UpdateName(c.Request().Context(), userID, activityID, req.Name); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

// Delete godoc
// @Summary      활동 삭제
// @Description  활동을 삭제합니다
// @Tags         Activities
// @Produce      json
// @Security     BearerAuth
// @Param        activity_id  path  string  true  "활동 ID"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.ErrorResponse  "ACTIVITY_NOT_FOUND"
// @Router       /activities/{activity_id} [delete]
func (h *ActivityHandler) Delete(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	activityID := c.Param("activity_id")

	if err := h.service.Delete(c.Request().Context(), userID, activityID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}
