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
		Format: "[\033[32m${time_rfc3339}\033[0m] ${status} | ${method} ${uri} | ${latency_human}\n",
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

	// User - all protected
	userGroup := e.Group("/users", middleware.JWTAuth())
	userGroup.GET("/me", s.user.GetMe)
	userGroup.PATCH("/me/settings", s.user.UpdateSettings)
	userGroup.GET("/blocks", s.user.GetBlocks)
	userGroup.POST("/:user_id/block", s.user.Block)
	userGroup.POST("/:user_id/unblock", s.user.Unblock)

	// Void - all protected
	voidGroup := e.Group("/void", middleware.JWTAuth())
	voidGroup.POST("/start", s.void.Start)
	voidGroup.POST("/end", s.void.End)
	voidGroup.POST("/cancel", s.void.Cancel)
	voidGroup.GET("/history", s.void.History)
	voidGroup.POST("/test", s.void.TestCreate)

	return e
}
