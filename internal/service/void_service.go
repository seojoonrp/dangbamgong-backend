package service

import (
	"context"
	"time"

	"dangbamgong-backend/internal/config"
	"dangbamgong-backend/internal/domain"
	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/model"
	"dangbamgong-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VoidService interface {
	Start(ctx context.Context, userID string) (*dto.VoidStartResponse, error)
	End(ctx context.Context, userID string, req dto.VoidEndRequest) (*dto.VoidEndResponse, error)
	Cancel(ctx context.Context, userID string) error
	History(ctx context.Context, userID string, targetDay string) (*dto.VoidHistoryResponse, error)
	TestCreate(ctx context.Context, userID string, req dto.TestVoidRequest) (*dto.VoidEndResponse, error)
}

type voidService struct {
	userRepo        repository.UserRepository
	voidSessionRepo repository.VoidSessionRepository
	activityRepo    repository.ActivityRepository
}

func NewVoidService(
	ur repository.UserRepository,
	vr repository.VoidSessionRepository,
	ar repository.ActivityRepository,
) VoidService {
	return &voidService{
		userRepo:        ur,
		voidSessionRepo: vr,
		activityRepo:    ar,
	}
}

func (s *voidService) Start(ctx context.Context, userID string) (*dto.VoidStartResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	user, err := s.userRepo.FindByID(ctx, oid)
	if err != nil || user == nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "user not found")
	}

	if user.IsInVoid {
		return nil, domain.NewConflict(domain.ErrAlreadyInVoid, "already in void")
	}

	now := time.Now()
	if err := s.userRepo.SetVoidState(ctx, oid, true, &now); err != nil {
		return nil, domain.NewInternal("failed to set void state: " + err.Error())
	}

	return &dto.VoidStartResponse{
		SessionID: "", // 세션은 종료 시 생성
		StartedAt: now,
		TargetDay: calcTargetDay(now),
	}, nil
}

func (s *voidService) End(ctx context.Context, userID string, req dto.VoidEndRequest) (*dto.VoidEndResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	user, err := s.userRepo.FindByID(ctx, oid)
	if err != nil || user == nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "user not found")
	}

	if !user.IsInVoid {
		return nil, domain.NewBadRequest(domain.ErrNotInVoid, "not in void")
	}

	if len(req.Activities) > 5 {
		return nil, domain.NewBadRequest(domain.ErrTooManyActivities, "activities must be 5 or fewer")
	}

	// 활동 존재 확인 및 usage 증가
	now := time.Now()
	for _, name := range req.Activities {
		activity, err := s.activityRepo.FindByUserIDAndName(ctx, oid, name)
		if err != nil {
			return nil, domain.NewInternal("failed to find activity: " + err.Error())
		}
		if activity == nil {
			return nil, domain.NewNotFound(domain.ErrActivityNotFound, "activity not found: "+name)
		}
		if err := s.activityRepo.IncrementUsage(ctx, activity.ID, now); err != nil {
			return nil, domain.NewInternal("failed to increment activity usage: " + err.Error())
		}
	}

	startedAt := *user.CurrentVoidStartedAt
	durationSec := int64(now.Sub(startedAt).Seconds())
	targetDay := calcTargetDay(startedAt)

	session := &model.VoidSession{
		UserID:      oid,
		StartedAt:   startedAt,
		EndedAt:     now,
		DurationSec: durationSec,
		TargetDay:   targetDay,
		Activities:  req.Activities,
		CreatedAt:   now,
	}

	if err := s.voidSessionRepo.Create(ctx, session); err != nil {
		return nil, domain.NewInternal("failed to create void session: " + err.Error())
	}

	if err := s.userRepo.SetVoidState(ctx, oid, false, nil); err != nil {
		return nil, domain.NewInternal("failed to reset void state: " + err.Error())
	}

	return &dto.VoidEndResponse{
		SessionID:   session.ID.Hex(),
		StartedAt:   startedAt,
		EndedAt:     now,
		DurationSec: durationSec,
		TargetDay:   targetDay,
		Activities:  req.Activities,
	}, nil
}

func (s *voidService) Cancel(ctx context.Context, userID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	user, err := s.userRepo.FindByID(ctx, oid)
	if err != nil || user == nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "user not found")
	}

	if !user.IsInVoid {
		return domain.NewBadRequest(domain.ErrNotInVoid, "not in void")
	}

	if err := s.userRepo.SetVoidState(ctx, oid, false, nil); err != nil {
		return domain.NewInternal("failed to reset void state: " + err.Error())
	}

	return nil
}

func (s *voidService) History(ctx context.Context, userID string, targetDay string) (*dto.VoidHistoryResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	sessions, err := s.voidSessionRepo.FindByUserIDAndTargetDay(ctx, oid, targetDay)
	if err != nil {
		return nil, domain.NewInternal("failed to find void sessions: " + err.Error())
	}

	items := make([]dto.VoidSession, len(sessions))
	var totalDuration int64
	for i, s := range sessions {
		items[i] = dto.VoidSession{
			SessionID:   s.ID.Hex(),
			StartedAt:   s.StartedAt,
			EndedAt:     s.EndedAt,
			DurationSec: s.DurationSec,
			Activities:  s.Activities,
		}
		totalDuration += s.DurationSec
	}

	return &dto.VoidHistoryResponse{
		TargetDay:        targetDay,
		Sessions:         items,
		TotalDurationSec: totalDuration,
	}, nil
}

func (s *voidService) TestCreate(ctx context.Context, userID string, req dto.TestVoidRequest) (*dto.VoidEndResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	durationSec := int64(req.EndedAt.Sub(req.StartedAt).Seconds())
	targetDay := calcTargetDay(req.StartedAt)

	session := &model.VoidSession{
		UserID:      oid,
		StartedAt:   req.StartedAt,
		EndedAt:     req.EndedAt,
		DurationSec: durationSec,
		TargetDay:   targetDay,
		Activities:  req.Activities,
		CreatedAt:   time.Now(),
	}

	if err := s.voidSessionRepo.Create(ctx, session); err != nil {
		return nil, domain.NewInternal("failed to create test void session: " + err.Error())
	}

	return &dto.VoidEndResponse{
		SessionID:   session.ID.Hex(),
		StartedAt:   req.StartedAt,
		EndedAt:     req.EndedAt,
		DurationSec: durationSec,
		TargetDay:   targetDay,
		Activities:  req.Activities,
	}, nil
}

func calcTargetDay(t time.Time) string {
	kst := t.In(config.KST)
	if kst.Hour() < config.DayStartHour {
		kst = kst.AddDate(0, 0, -1)
	}
	return kst.Format("2006-01-02")
}
