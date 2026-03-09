package service

import (
	"context"
	"time"
	"unicode/utf8"

	"dangbamgong-backend/internal/domain"
	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/model"
	"dangbamgong-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivityService interface {
	List(ctx context.Context, userID string) (*dto.ActivityListResponse, error)
	Create(ctx context.Context, userID string, req dto.CreateActivityRequest) (*dto.CreateActivityResponse, error)
	UpdateName(ctx context.Context, userID string, activityID string, name string) error
	Delete(ctx context.Context, userID string, activityID string) error
}

type activityService struct {
	activityRepo repository.ActivityRepository
}

func NewActivityService(ar repository.ActivityRepository) ActivityService {
	return &activityService{activityRepo: ar}
}

func (s *activityService) List(ctx context.Context, userID string) (*dto.ActivityListResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id: "+err.Error())
	}

	activities, err := s.activityRepo.FindByUserID(ctx, oid)
	if err != nil {
		return nil, domain.NewInternal("failed to find activities: " + err.Error())
	}

	items := make([]dto.ActivityItem, len(activities))
	for i, a := range activities {
		items[i] = dto.ActivityItem{
			ID:         a.ID.Hex(),
			Name:       a.Name,
			UsageCount: a.UsageCount,
			LastUsedAt: a.LastUsedAt,
		}
	}

	return &dto.ActivityListResponse{Activities: items}, nil
}

func (s *activityService) Create(ctx context.Context, userID string, req dto.CreateActivityRequest) (*dto.CreateActivityResponse, error) {
	length := utf8.RuneCountInString(req.Name)
	if length < 1 || length > 10 {
		return nil, domain.NewBadRequest(domain.ErrInvalidActivityName, "activity name must be 1-10 characters")
	}

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id: "+err.Error())
	}

	existing, err := s.activityRepo.FindByUserIDAndName(ctx, oid, req.Name)
	if err != nil {
		return nil, domain.NewInternal("failed to check duplicate activity: " + err.Error())
	}
	if existing != nil {
		return nil, domain.NewConflict(domain.ErrActivityAlreadyExists, "activity already exists: "+req.Name)
	}

	now := time.Now()
	activity := &model.Activity{
		UserID:    oid,
		Name:      req.Name,
		CreatedAt: now,
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		return nil, domain.NewInternal("failed to create activity: " + err.Error())
	}

	return &dto.CreateActivityResponse{
		ID:         activity.ID.Hex(),
		Name:       activity.Name,
		UsageCount: 0,
		LastUsedAt: nil,
	}, nil
}

func (s *activityService) UpdateName(ctx context.Context, userID string, activityID string, name string) error {
	length := utf8.RuneCountInString(name)
	if length < 1 || length > 10 {
		return domain.NewBadRequest(domain.ErrInvalidActivityName, "activity name must be 1-10 characters")
	}

	userOid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id: "+err.Error())
	}

	actOid, err := primitive.ObjectIDFromHex(activityID)
	if err != nil {
		return domain.NewBadRequest(domain.ErrActivityNotFound, "invalid activity id: "+err.Error())
	}

	activity, err := s.activityRepo.FindByID(ctx, actOid)
	if err != nil {
		return domain.NewInternal("failed to find activity: " + err.Error())
	}
	if activity == nil {
		return domain.NewNotFound(domain.ErrActivityNotFound, "activity not found: "+activityID)
	}

	// 본인 소유 확인
	if activity.UserID.Hex() != userID {
		return domain.NewNotFound(domain.ErrActivityNotFound, "activity not found: "+activityID)
	}

	// 중복 이름 확인
	existing, err := s.activityRepo.FindByUserIDAndName(ctx, userOid, name)
	if err != nil {
		return domain.NewInternal("failed to check duplicate activity: " + err.Error())
	}
	if existing != nil {
		return domain.NewConflict(domain.ErrActivityAlreadyExists, "activity already exists: "+name)
	}

	if err := s.activityRepo.UpdateName(ctx, actOid, name); err != nil {
		return domain.NewInternal("failed to update activity name: " + err.Error())
	}

	return nil
}

func (s *activityService) Delete(ctx context.Context, userID string, activityID string) error {
	actOid, err := primitive.ObjectIDFromHex(activityID)
	if err != nil {
		return domain.NewBadRequest(domain.ErrActivityNotFound, "invalid activity id: "+err.Error())
	}

	activity, err := s.activityRepo.FindByID(ctx, actOid)
	if err != nil {
		return domain.NewInternal("failed to find activity: " + err.Error())
	}
	if activity == nil {
		return domain.NewNotFound(domain.ErrActivityNotFound, "activity not found: "+activityID)
	}

	// 본인 소유 확인
	if activity.UserID.Hex() != userID {
		return domain.NewNotFound(domain.ErrActivityNotFound, "activity not found: "+activityID)
	}

	if err := s.activityRepo.Delete(ctx, actOid); err != nil {
		return domain.NewInternal("failed to delete activity: " + err.Error())
	}

	return nil
}
