package server

import (
	"net/http"

	"dangbamgong-backend/internal/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	e.HTTPErrorHandler = middleware.ErrorHandler

	e.Use(echoMiddleware.LoggerWithConfig(echoMiddleware.LoggerConfig{
		Format: "[\033[32m${time_rfc3339}\033[0m] ${status} | ${method} ${uri} | ${latency_human}",
	}))
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	e.GET("/", s.health.HelloWorld)
	e.GET("/health", s.health.Health)

	// Auth - public
	authGroup := e.Group("/auth")
	authGroup.POST("/login", s.auth.Login)
	authGroup.POST("/login/test", s.auth.TestLogin)

	// Auth - protected
	authProtected := authGroup.Group("", middleware.JWTAuth())
	authProtected.POST("/nickname", s.auth.SetNickname)
	authProtected.DELETE("/withdraw", s.auth.Withdraw)

	// Activity - all protected
	activityGroup := e.Group("/activities", middleware.JWTAuth())
	activityGroup.GET("", s.activity.List)
	activityGroup.POST("", s.activity.Create)
	activityGroup.DELETE("/:activity_id", s.activity.Delete)

	return e
}
