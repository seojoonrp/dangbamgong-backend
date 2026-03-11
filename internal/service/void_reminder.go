package service

import (
	"context"
	"log"
	"sync"
	"time"

	"dangbamgong-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VoidReminderScheduler struct {
	mu       sync.Mutex
	timers   map[string]*time.Timer
	notifSvc NotificationService
	userRepo repository.UserRepository
}

func NewVoidReminderScheduler(notifSvc NotificationService, userRepo repository.UserRepository) *VoidReminderScheduler {
	return &VoidReminderScheduler{
		timers:   make(map[string]*time.Timer),
		notifSvc: notifSvc,
		userRepo: userRepo,
	}
}

func (s *VoidReminderScheduler) Schedule(userID primitive.ObjectID, startedAt time.Time, reminderHours int) {
	if reminderHours <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := userID.Hex()

	if existing, ok := s.timers[key]; ok {
		existing.Stop()
	}

	deadline := startedAt.Add(time.Duration(reminderHours) * time.Hour)
	remaining := time.Until(deadline)

	if remaining <= 0 {
		go s.fire(userID)
		return
	}

	s.timers[key] = time.AfterFunc(remaining, func() {
		s.fire(userID)
	})

	log.Printf("[REMINDER] scheduled for user %s in %v\n", key, remaining)
}

func (s *VoidReminderScheduler) Cancel(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if timer, ok := s.timers[userID]; ok {
		timer.Stop()
		delete(s.timers, userID)
		log.Printf("[REMINDER] cancelled for user %s\n", userID)
	}
}

func (s *VoidReminderScheduler) fire(userID primitive.ObjectID) {
	s.mu.Lock()
	delete(s.timers, userID.Hex())
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.notifSvc.SendVoidReminder(ctx, userID); err != nil {
		log.Printf("[REMINDER] failed to send void reminder for %s: %v\n", userID.Hex(), err)
	}
}

func (s *VoidReminderScheduler) RecoverAll(ctx context.Context) {
	users, err := s.userRepo.FindUsersInVoid(ctx)
	if err != nil {
		log.Printf("[REMINDER] failed to recover void reminders: %v\n", err)
		return
	}

	for _, user := range users {
		if !user.NotificationSettings.VoidReminder || user.CurrentVoidStartedAt == nil {
			continue
		}
		s.Schedule(user.ID, *user.CurrentVoidStartedAt, user.NotificationSettings.ReminderHours)
	}

	log.Printf("[REMINDER] recovered %d void reminders\n", len(users))
}
