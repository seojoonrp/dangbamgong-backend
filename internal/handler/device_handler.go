package handler

import (
	"net/http"

	"dangbamgong-backend/internal/domain"
	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/middleware"
	"dangbamgong-backend/internal/repository"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeviceHandler struct {
	deviceTokenRepo repository.DeviceTokenRepository
}

func NewDeviceHandler(dr repository.DeviceTokenRepository) *DeviceHandler {
	return &DeviceHandler{deviceTokenRepo: dr}
}

// RegisterToken godoc
// @Summary      디바이스 토큰 등록
// @Description  APNs 푸시 알림용 디바이스 토큰을 등록(또는 갱신)합니다
// @Tags         Devices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.RegisterDeviceRequest  true  "디바이스 토큰"
// @Success      200   {object}  dto.Response[any]
// @Failure      400   {object}  dto.ErrorResponse
// @Router       /devices/token [put]
func (h *DeviceHandler) RegisterToken(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req dto.RegisterDeviceRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if req.Token == "" {
		return domain.NewBadRequest(domain.ErrBadRequest, "token is required")
	}

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	if err := h.deviceTokenRepo.Upsert(c.Request().Context(), oid, req.Token); err != nil {
		return domain.NewInternal("failed to register device token: " + err.Error())
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}

// DeleteToken godoc
// @Summary      디바이스 토큰 삭제
// @Description  등록된 디바이스 토큰을 삭제합니다
// @Tags         Devices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.RegisterDeviceRequest  true  "삭제할 토큰"
// @Success      200   {object}  dto.Response[any]
// @Failure      400   {object}  dto.ErrorResponse
// @Router       /devices/token [delete]
func (h *DeviceHandler) DeleteToken(c echo.Context) error {
	var req dto.RegisterDeviceRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if req.Token == "" {
		return domain.NewBadRequest(domain.ErrBadRequest, "token is required")
	}

	if err := h.deviceTokenRepo.DeleteByToken(c.Request().Context(), req.Token); err != nil {
		return domain.NewInternal("failed to delete device token: " + err.Error())
	}

	return dto.SuccessEmpty(c, http.StatusOK)
}
