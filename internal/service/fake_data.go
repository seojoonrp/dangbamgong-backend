package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"dangbamgong-backend/internal/config"
	"dangbamgong-backend/internal/model"
	"dangbamgong-backend/internal/repository"
)

const seedUserCount = 30

type FakeDataService struct {
	userRepo        repository.UserRepository
	voidSessionRepo repository.VoidSessionRepository
	statSvc         StatService
}

func NewFakeDataService(
	userRepo repository.UserRepository,
	voidSessionRepo repository.VoidSessionRepository,
	statSvc StatService,
) *FakeDataService {
	return &FakeDataService{
		userRepo:        userRepo,
		voidSessionRepo: voidSessionRepo,
		statSvc:         statSvc,
	}
}

func (s *FakeDataService) EnsureSeedUsers(ctx context.Context) ([]model.User, error) {
	users := make([]model.User, 0, seedUserCount)

	for i := 1; i <= seedUserCount; i++ {
		socialID := fmt.Sprintf("seed_%03d", i)
		existing, _ := s.userRepo.FindBySocial(ctx, model.ProviderTest, socialID)
		if existing != nil {
			users = append(users, *existing)
			continue
		}

		tag := generateSeedTag(10)
		now := time.Now()
		user := &model.User{
			SocialProvider: model.ProviderTest,
			SocialID:       socialID,
			Nickname:       fmt.Sprintf("시드유저%02d", i),
			Tag:            tag,
			NotificationSettings: model.NotificationSettings{
				VoidReminder:  false,
				ReminderHours: 0,
				FriendRequest: false,
				FriendNudge:   false,
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create seed user %d: %w", i, err)
		}
		users = append(users, *user)
	}

	log.Printf("[FAKE_DATA] ensured %d seed users\n", len(users))
	return users, nil
}

func (s *FakeDataService) GenerateSessionsForDay(ctx context.Context, users []model.User, targetDay string) error {
	// Check if sessions already exist for this day (idempotency)
	existing, err := s.voidSessionRepo.FindByTargetDay(ctx, targetDay)
	if err != nil {
		return fmt.Errorf("failed to check existing sessions for %s: %w", targetDay, err)
	}

	seedUserIDs := make(map[string]bool, len(users))
	for _, u := range users {
		seedUserIDs[u.ID.Hex()] = true
	}

	for _, sess := range existing {
		if seedUserIDs[sess.UserID.Hex()] {
			log.Printf("[FAKE_DATA] sessions already exist for %s, skipping\n", targetDay)
			return nil
		}
	}

	// Parse targetDay to get base time (22:00 KST)
	dayTime, err := time.ParseInLocation("2006-01-02", targetDay, config.KST)
	if err != nil {
		return fmt.Errorf("failed to parse target day %s: %w", targetDay, err)
	}
	baseTime := dayTime.Add(22 * time.Hour) // 22:00 KST on targetDay

	created := 0
	for _, user := range users {
		// 70% chance to create a session
		if rand.Float64() > 0.7 {
			continue
		}

		// Random start: 22:00 ~ 02:00 KST (4 hour window)
		offsetSec := rand.Intn(4 * 3600)
		startedAt := baseTime.Add(time.Duration(offsetSec) * time.Second)

		// Random duration: 1~5 hours
		durationSec := int64(rand.Intn(4*3600) + 3600)
		endedAt := startedAt.Add(time.Duration(durationSec) * time.Second)

		// Clamp to 04:00 KST next day
		maxEnd := dayTime.AddDate(0, 0, 1).Add(4 * time.Hour)
		if endedAt.After(maxEnd) {
			endedAt = maxEnd
			durationSec = int64(endedAt.Sub(startedAt).Seconds())
		}

		if durationSec < 60 {
			continue
		}

		session := &model.VoidSession{
			UserID:      user.ID,
			StartedAt:   startedAt,
			EndedAt:     endedAt,
			DurationSec: durationSec,
			TargetDay:   config.CalcTargetDay(startedAt),
			Activities:  []string{},
			CreatedAt:   endedAt,
		}

		if err := s.voidSessionRepo.Create(ctx, session); err != nil {
			log.Printf("[FAKE_DATA] failed to create session for user %s on %s: %v\n", user.ID.Hex(), targetDay, err)
			continue
		}
		created++
	}

	log.Printf("[FAKE_DATA] created %d sessions for %s\n", created, targetDay)

	if err := s.statSvc.FinalizeDailyStats(ctx, targetDay); err != nil {
		log.Printf("[FAKE_DATA] failed to finalize stats for %s: %v\n", targetDay, err)
	}

	return nil
}

func (s *FakeDataService) Backfill(ctx context.Context, fromDate, toDate string) error {
	from, err := time.ParseInLocation("2006-01-02", fromDate, config.KST)
	if err != nil {
		return fmt.Errorf("invalid fromDate: %w", err)
	}
	to, err := time.ParseInLocation("2006-01-02", toDate, config.KST)
	if err != nil {
		return fmt.Errorf("invalid toDate: %w", err)
	}

	users, err := s.EnsureSeedUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure seed users: %w", err)
	}

	days := 0
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		targetDay := d.Format("2006-01-02")
		if err := s.GenerateSessionsForDay(ctx, users, targetDay); err != nil {
			log.Printf("[FAKE_DATA] backfill error for %s: %v\n", targetDay, err)
			continue
		}
		days++
	}

	log.Printf("[FAKE_DATA] backfill complete: %d days processed\n", days)
	return nil
}

func generateSeedTag(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
