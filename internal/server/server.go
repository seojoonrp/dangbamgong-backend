package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"dangbamgong-backend/internal/auth"
	"dangbamgong-backend/internal/database"
	"dangbamgong-backend/internal/handler"
	"dangbamgong-backend/internal/repository"
	"dangbamgong-backend/internal/service"
)

type Server struct {
	port     int
	health   *handler.HealthHandler
	auth     *handler.AuthHandler
	activity *handler.ActivityHandler
	user     *handler.UserHandler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	db := database.New()

	// Health
	healthRepo := repository.NewHealthRepository(db)
	healthSvc := service.NewHealthService(healthRepo)
	healthHandler := handler.NewHealthHandler(healthSvc)

	// Auth
	userRepo := repository.NewUserRepository(db)
	socialVerifier := auth.NewSocialVerifier()
	authSvc := service.NewAuthService(userRepo, socialVerifier)
	authHandler := handler.NewAuthHandler(authSvc)

	// Activity
	activityRepo := repository.NewActivityRepository(db)
	activitySvc := service.NewActivityService(activityRepo)
	activityHandler := handler.NewActivityHandler(activitySvc)

	// User
	blockRepo := repository.NewBlockRepository(db)
	userSvc := service.NewUserService(userRepo, blockRepo)
	userHandler := handler.NewUserHandler(userSvc)

	s := &Server{
		port:     port,
		health:   healthHandler,
		auth:     authHandler,
		activity: activityHandler,
		user:     userHandler,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
