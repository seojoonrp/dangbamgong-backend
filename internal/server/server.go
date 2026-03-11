package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"dangbamgong-backend/internal/auth"
	"dangbamgong-backend/internal/database"
	"dangbamgong-backend/internal/handler"
	"dangbamgong-backend/internal/push"
	"dangbamgong-backend/internal/repository"
	"dangbamgong-backend/internal/service"
)

type Server struct {
	port         int
	health       *handler.HealthHandler
	auth         *handler.AuthHandler
	activity     *handler.ActivityHandler
	user         *handler.UserHandler
	void         *handler.VoidHandler
	friend       *handler.FriendHandler
	stat         *handler.StatHandler
	notification *handler.NotificationHandler
	device       *handler.DeviceHandler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	db := database.New()

	socialVerifier := auth.NewSocialVerifier()
	pushClient := push.NewAPNsClient()

	healthRepo := repository.NewHealthRepository(db)
	userRepo := repository.NewUserRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	blockRepo := repository.NewBlockRepository(db)
	friendshipRepo := repository.NewFriendshipRepository(db)
	friendRequestRepo := repository.NewFriendRequestRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	deviceTokenRepo := repository.NewDeviceTokenRepository(db)
	voidSessionRepo := repository.NewVoidSessionRepository(db)
	statRepo := repository.NewStatRepository(db)

	healthSvc := service.NewHealthService(healthRepo)
	authSvc := service.NewAuthService(userRepo, socialVerifier)
	activitySvc := service.NewActivityService(activityRepo)
	userSvc := service.NewUserService(userRepo, blockRepo, friendshipRepo, friendRequestRepo)
	notifSvc := service.NewNotificationService(notifRepo, deviceTokenRepo, userRepo, pushClient)
	reminderScheduler := service.NewVoidReminderScheduler(notifSvc, userRepo)
	voidSvc := service.NewVoidService(userRepo, voidSessionRepo, activityRepo, reminderScheduler)
	friendSvc := service.NewFriendService(userRepo, blockRepo, friendshipRepo, friendRequestRepo, notifSvc)
	statSvc := service.NewStatService(statRepo, voidSessionRepo)

	healthHandler := handler.NewHealthHandler(healthSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	activityHandler := handler.NewActivityHandler(activitySvc)
	userHandler := handler.NewUserHandler(userSvc)
	voidHandler := handler.NewVoidHandler(voidSvc)
	friendHandler := handler.NewFriendHandler(friendSvc)
	statHandler := handler.NewStatHandler(statSvc)
	notificationHandler := handler.NewNotificationHandler(notifSvc)
	deviceHandler := handler.NewDeviceHandler(deviceTokenRepo)

	reminderScheduler.RecoverAll(context.Background())

	s := &Server{
		port:         port,
		health:       healthHandler,
		auth:         authHandler,
		activity:     activityHandler,
		user:         userHandler,
		void:         voidHandler,
		friend:       friendHandler,
		stat:         statHandler,
		notification: notificationHandler,
		device:       deviceHandler,
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
