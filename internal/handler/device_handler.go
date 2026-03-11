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
