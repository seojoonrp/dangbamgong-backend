package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"dangbamgong-backend/internal/database"
	"dangbamgong-backend/internal/handler"
	"dangbamgong-backend/internal/repository"
	"dangbamgong-backend/internal/service"
)

type Server struct {
	port   int
	health *handler.HealthHandler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	db := database.New()

	healthRepo := repository.NewHealthRepository(db)
	healthSvc := service.NewHealthService(healthRepo)
	healthHandler := handler.NewHealthHandler(healthSvc)

	s := &Server{
		port:   port,
		health: healthHandler,
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
