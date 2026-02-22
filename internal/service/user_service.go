package service

import (
	"context"
	"time"

	"dangbamgong-backend/internal/domain"
	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/model"
	"dangbamgong-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService interface {
	GetMe(ctx context.Context, userID string) (*dto.UserMeResponse, error)
	UpdateSettings(ctx context.Context, userID string, req dto.UpdateSettingsRequest) (*dto.UpdateSettingsResponse, error)
	GetBlocks(ctx context.Context, userID string) (*dto.BlockListResponse, error)
	Block(ctx context.Context, userID string, targetID string) error
	Unblock(ctx context.Context, userID string, targetID string) error
}

type userService struct {
	userRepo  repository.UserRepository
	blockRepo repository.BlockRepository
}

func NewUserService(ur repository.UserRepository, br repository.BlockRepository) UserService {
	return &userService{
		userRepo:  ur,
		blockRepo: br,
	}
}

func (s *userService) GetMe(ctx context.Context, userID string) (*dto.UserMeResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	user, err := s.userRepo.FindByID(ctx, oid)
	if err != nil || user == nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "user not found")
	}

	return &dto.UserMeResponse{
		ID:                   user.ID.Hex(),
		Tag:                  user.Tag,
		Nickname:             user.Nickname,
		IsInVoid:             user.IsInVoid,
		CurrentVoidStartedAt: user.CurrentVoidStartedAt,
		NotificationSettings: dto.NotificationSettings{
			VoidReminder:  user.NotificationSettings.VoidReminder,
			ReminderHours: user.NotificationSettings.ReminderHours,
			FriendNudge:   user.NotificationSettings.FriendNudge,
		},
	}, nil
}

func (s *userService) UpdateSettings(ctx context.Context, userID string, req dto.UpdateSettingsRequest) (*dto.UpdateSettingsResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	user, err := s.userRepo.FindByID(ctx, oid)
	if err != nil || user == nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "user not found")
	}

	settings := user.NotificationSettings
	if req.VoidReminder != nil {
		settings.VoidReminder = *req.VoidReminder
	}
	if req.ReminderHours != nil {
		settings.ReminderHours = *req.ReminderHours
	}
	if req.FriendNudge != nil {
		settings.FriendNudge = *req.FriendNudge
	}

	if err := s.userRepo.UpdateSettings(ctx, oid, settings); err != nil {
		return nil, domain.NewInternal("failed to update settings: " + err.Error())
	}

	return &dto.UpdateSettingsResponse{
		VoidReminder:  settings.VoidReminder,
		ReminderHours: settings.ReminderHours,
		FriendNudge:   settings.FriendNudge,
	}, nil
}

func (s *userService) GetBlocks(ctx context.Context, userID string) (*dto.BlockListResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	blocks, err := s.blockRepo.FindByUserID(ctx, oid)
	if err != nil {
		return nil, domain.NewInternal("failed to find blocks: " + err.Error())
	}

	items := make([]dto.BlockItem, 0, len(blocks))
	for _, b := range blocks {
		blocked, err := s.userRepo.FindByID(ctx, b.BlockedID)
		if err != nil || blocked == nil {
			continue
		}
		items = append(items, dto.BlockItem{
			UserID:    blocked.ID.Hex(),
			Nickname:  blocked.Nickname,
			Tag:       blocked.Tag,
			BlockedAt: b.CreatedAt,
		})
	}

	return &dto.BlockListResponse{Blocks: items}, nil
}

func (s *userService) Block(ctx context.Context, userID string, targetID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	targetOid, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return domain.NewNotFound(domain.ErrUserNotFound, "invalid target user id")
	}

	target, err := s.userRepo.FindByID(ctx, targetOid)
	if err != nil || target == nil {
		return domain.NewNotFound(domain.ErrUserNotFound, "user not found")
	}

	existing, err := s.blockRepo.FindOne(ctx, oid, targetOid)
	if err != nil {
		return domain.NewInternal("failed to check block: " + err.Error())
	}
	if existing != nil {
		return domain.NewConflict(domain.ErrAlreadyBlocked, "already blocked")
	}

	block := &model.Block{
		UserID:    oid,
		BlockedID: targetOid,
		CreatedAt: time.Now(),
	}

	if err := s.blockRepo.Create(ctx, block); err != nil {
		return domain.NewInternal("failed to create block: " + err.Error())
	}

	// TODO: 친구 관계가 있으면 함께 삭제

	return nil
}

func (s *userService) Unblock(ctx context.Context, userID string, targetID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	targetOid, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return domain.NewBadRequest(domain.ErrNotBlocked, "invalid target user id")
	}

	existing, err := s.blockRepo.FindOne(ctx, oid, targetOid)
	if err != nil {
		return domain.NewInternal("failed to check block: " + err.Error())
	}
	if existing == nil {
		return domain.NewBadRequest(domain.ErrNotBlocked, "not blocked")
	}

	if err := s.blockRepo.Delete(ctx, oid, targetOid); err != nil {
		return domain.NewInternal("failed to delete block: " + err.Error())
	}

	return nil
}
