package handler

import (
	"net/http"

	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/middleware"
	"dangbamgong-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type StatHandler struct {
	service service.StatService
}

func NewStatHandler(s service.StatService) *StatHandler {
	return &StatHandler{service: s}
}

// GetHomeStat godoc
// @Summary      홈 통계 조회
// @Description  현재 공백 중인 유저 수, 오늘 잠에 든 유저 수, 내 랭킹/총 공백 시간을 반환합니다
// @Tags         Stats
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.Response[dto.HomeStatResponse]
// @Router       /stats/home [get]
func (h *StatHandler) GetHomeStat(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.GetHomeStat(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// GetMyVoidStat godoc
// @Summary      내 공백 통계 조회
// @Description  전체 공백 총 시간, 평균 시간, 최대 시간을 반환합니다
// @Tags         Stats
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.Response[dto.MyVoidStatResponse]
// @Router       /stats/me [get]
func (h *StatHandler) GetMyVoidStat(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	resp, err := h.service.GetMyVoidStat(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}

// GetDailyStat godoc
// @Summary      일별 통계 조회
// @Description  특정 날짜의 20분 단위 버킷 통계와 내 공백 세션 목록을 반환합니다
// @Tags         Stats
// @Produce      json
// @Security     BearerAuth
// @Param        target_day  query     string  true  "조회할 날짜 (YYYY-MM-DD)"
// @Success      200  {object}  dto.Response[dto.DailyStatResponse]
// @Router       /stats/daily [get]
func (h *StatHandler) GetDailyStat(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)
	targetDay := c.QueryParam("target_day")

	resp, err := h.service.GetDailyStat(c.Request().Context(), userID, targetDay)
	if err != nil {
		return err
	}

	return dto.Success(c, http.StatusOK, resp)
}
