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

// Start godoc
// @Summary      공백 시작
// @Description  공백(밤의 공백) 세션을 시작합니다. 이미 공백 중이면 실패합니다.
// @Tags         Void
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.Response[dto.VoidStartResponse]
// @Failure      409  {object}  dto.ErrorResponse  "ALREADY_IN_VOID"
// @Router       /void/start [post]
func (h *VoidHandler) Start(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.Start(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// End godoc
// @Summary      공백 종료
// @Description  진행 중인 공백 세션을 종료하고 활동을 기록합니다. 활동은 최대 5개.
// @Tags         Void
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.VoidEndRequest  true  "종료 시 기록할 활동 목록"
// @Success      200   {object}  dto.Response[dto.VoidEndResponse]
// @Failure      400   {object}  dto.ErrorResponse  "NOT_IN_VOID / TOO_MANY_ACTIVITIES / ACTIVITY_NOT_FOUND"
// @Router       /void/end [post]
func (h *VoidHandler) End(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req dto.VoidEndRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	resp, err := h.service.End(c.Request().Context(), userID, req)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// Cancel godoc
// @Summary      공백 취소
// @Description  진행 중인 공백을 취소합니다. 세션 기록이 남지 않습니다.
// @Tags         Void
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.Response[any]
// @Failure      400  {object}  dto.ErrorResponse  "NOT_IN_VOID"
// @Router       /void/cancel [post]
func (h *VoidHandler) Cancel(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	if err := h.service.Cancel(c.Request().Context(), userID); err != nil {
		return err
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

// TestCreate godoc
// @Summary      테스트 공백 생성
// @Description  개발용. 임의의 시작/종료 시각으로 공백 세션을 생성합니다.
// @Tags         Void
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.TestVoidRequest  true  "테스트 공백 데이터"
// @Success      201   {object}  dto.Response[dto.VoidEndResponse]
// @Router       /void/test [post]
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

// History godoc
// @Summary      공백 히스토리 조회
// @Description  특정 날짜의 공백 세션 목록과 총 시간을 반환합니다. 날짜 기준은 KST 16:00.
// @Tags         Void
// @Produce      json
// @Security     BearerAuth
// @Param        target_day  query     string  true  "조회할 날짜 (YYYY-MM-DD)"
// @Success      200  {object}  dto.Response[dto.VoidHistoryResponse]
// @Router       /void/history [get]
func (h *VoidHandler) History(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	targetDay := c.QueryParam("target_day")

	resp, err := h.service.History(c.Request().Context(), userID, targetDay)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}
