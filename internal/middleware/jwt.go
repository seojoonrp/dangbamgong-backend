package middleware

import (
	"strings"

	"dangbamgong-backend/internal/auth"
	"dangbamgong-backend/internal/domain"

	"github.com/labstack/echo/v4"
)

const ContextKeyUserID = "user_id"

func JWTAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				return domain.NewUnauthorized(domain.ErrUnauthorized, "missing or invalid authorization header")
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims, err := auth.ParseToken(tokenStr)
			if err != nil {
				return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid or expired token: "+err.Error())
			}

			c.Set(ContextKeyUserID, claims.UserID)
			return next(c)
		}
	}
}
