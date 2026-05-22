package service

import (
	"context"
	"log"
	"time"

	"dangbamgong-backend/internal/config"
	"dangbamgong-backend/internal/repository"

	"github.com/robfig/cron/v3"
)

type DayResetScheduler struct {
	cron              *cron.Cron
	userRepo          repository.UserRepository
	statSvc           StatService
	notifSvc          NotificationService
	reminderScheduler *VoidReminderScheduler
}

func NewDayResetScheduler(
	userRepo repository.UserRepository,
	statSvc StatService,
	notifSvc NotificationService,
	reminderScheduler *VoidReminderScheduler,
) *DayResetScheduler {
	return &DayResetScheduler{
		userRepo:          userRepo,
		statSvc:           statSvc,
		notifSvc:          notifSvc,
		reminderScheduler: reminderScheduler,
	}
}

func (s *DayResetScheduler) Start() {
	s.cron = cron.New(cron.WithLocation(config.KST))

	_, err := s.cron.AddFunc("0 16 * * *", s.execute)
	if err != nil {
		log.Printf("[DAY_RESET] failed to register cron job: %v\n", err)
		return
	}

	s.cron.Start()
	log.Println("[DAY_RESET] scheduler started (0 16 * * * KST)")
}

func (s *DayResetScheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
		log.Println("[DAY_RESET] scheduler stopped")
	}
}

func (s *DayResetScheduler) execute() {
	log.Println("[DAY_RESET] starting daily reset...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. 취소 대상 유저 조회 (알림 발송용)
	usersInVoid, err := s.userRepo.FindUsersInVoid(ctx)
	if err != nil {
		log.Printf("[DAY_RESET] failed to find users in void: %v\n", err)
	}

	// 2. 모든 공백 상태 일괄 취소
	cancelled, err := s.userRepo.CancelAllVoidStates(ctx)
	if err != nil {
		log.Printf("[DAY_RESET] failed to cancel void states: %v\n", err)
	} else {
		log.Printf("[DAY_RESET] cancelled %d void sessions\n", cancelled)
	}

	// 3. 리마인더 타이머 전체 해제
	reminderCount := s.reminderScheduler.CancelAll()
	log.Printf("[DAY_RESET] cleared %d reminder timers\n", reminderCount)

	// 4. 취소된 유저들에게 푸시 알림 전송 (각 호출은 내부에서 비동기로 처리됨)
	// TODO: 유저가 많아지면 goroutine이 유저 수만큼 생긴다. expo_push의 배치 전송으로 묶는 게 좋다.
	for _, user := range usersInVoid {
		s.notifSvc.SendVoidAutoCancel(user.ID)
	}

	// 5. 전날 스탯 확정
	prevDay := config.CalcTargetDay(time.Now().In(config.KST).Add(-1 * time.Hour))
	if err := s.statSvc.FinalizeDailyStats(ctx, prevDay); err != nil {
		log.Printf("[DAY_RESET] failed to finalize stats for %s: %v\n", prevDay, err)
	} else {
		log.Printf("[DAY_RESET] finalized stats for %s\n", prevDay)
	}

	log.Println("[DAY_RESET] daily reset complete")
}
