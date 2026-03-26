package service

import (
	"context"
	"log"
	"time"

	"dangbamgong-backend/internal/config"

	"github.com/robfig/cron/v3"
)

type FakeDataScheduler struct {
	cron    *cron.Cron
	service *FakeDataService
}

func NewFakeDataScheduler(service *FakeDataService) *FakeDataScheduler {
	return &FakeDataScheduler{service: service}
}

func (s *FakeDataScheduler) Start() {
	s.cron = cron.New(cron.WithLocation(config.KST))

	_, err := s.cron.AddFunc("5 16 * * *", s.execute)
	if err != nil {
		log.Printf("[FAKE_DATA] failed to register cron job: %v\n", err)
		return
	}

	s.cron.Start()
	log.Println("[FAKE_DATA] scheduler started (5 16 * * * KST)")
}

func (s *FakeDataScheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
		log.Println("[FAKE_DATA] scheduler stopped")
	}
}

func (s *FakeDataScheduler) RunOnStartup() {
	go func() {
		ctx := context.Background()

		now := time.Now().In(config.KST)
		yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

		log.Println("[FAKE_DATA] starting backfill on startup...")
		if err := s.service.Backfill(ctx, "2026-03-20", yesterday); err != nil {
			log.Printf("[FAKE_DATA] backfill failed: %v\n", err)
		}
		log.Println("[FAKE_DATA] startup backfill complete")
	}()
}

func (s *FakeDataScheduler) execute() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	log.Println("[FAKE_DATA] running daily generation...")

	users, err := s.service.EnsureSeedUsers(ctx)
	if err != nil {
		log.Printf("[FAKE_DATA] failed to ensure seed users: %v\n", err)
		return
	}

	// Generate for yesterday's target day
	now := time.Now().In(config.KST)
	targetDay := now.AddDate(0, 0, -1).Format("2006-01-02")

	if err := s.service.GenerateSessionsForDay(ctx, users, targetDay); err != nil {
		log.Printf("[FAKE_DATA] daily generation failed for %s: %v\n", targetDay, err)
	}

	log.Printf("[FAKE_DATA] daily generation complete for %s\n", targetDay)
}
