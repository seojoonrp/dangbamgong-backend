package service

import (
	"context"
	"regexp"
	"time"
	"unicode/utf8"

	"dangbamgong-backend/internal/domain"
	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/model"
	"dangbamgong-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService interface {
	GetMe(ctx context.Context, userID string) (*dto.UserMeResponse, error)
	UpdateSettings(ctx context.Context, userID string, req dto.UpdateSettingsRequest) (*dto.UpdateSettingsResponse, error)
	Search(ctx context.Context, userID string, tagPrefix string) (*dto.UserSearchResponse, error)
	GetBlocks(ctx context.Context, userID string) (*dto.BlockListResponse, error)
	Block(ctx context.Context, userID string, targetID string) error
	Unblock(ctx context.Context, userID string, targetID string) error
	ChangeNickname(ctx context.Context, userID string, req dto.ChangeNicknameRequest) (*dto.ChangeNicknameResponse, error)
}

type userService struct {
	userRepo          repository.UserRepository
	blockRepo         repository.BlockRepository
	friendshipRepo    repository.FriendshipRepository
	friendRequestRepo repository.FriendRequestRepository
}

func NewUserService(
	ur repository.UserRepository,
	br repository.BlockRepository,
	fr repository.FriendshipRepository,
	frr repository.FriendRequestRepository,
) UserService {
	return &userService{
		userRepo:          ur,
		blockRepo:         br,
		friendshipRepo:    fr,
		friendRequestRepo: frr,
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
			FriendRequest: user.NotificationSettings.FriendRequest,
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
		if *req.ReminderHours < 0 || *req.ReminderHours > 24 {
			return nil, domain.NewBadRequest(domain.ErrBadRequest, "reminder hours must be between 0 and 24")
		}
		settings.ReminderHours = *req.ReminderHours
	}
	if req.FriendRequest != nil {
		settings.FriendRequest = *req.FriendRequest
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
		FriendRequest: settings.FriendRequest,
		FriendNudge:   settings.FriendNudge,
	}, nil
}

func (s *userService) Search(ctx context.Context, userID string, tagPrefix string) (*dto.UserSearchResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	if tagPrefix == "" {
		return nil, domain.NewBadRequest(domain.ErrBadRequest, "tag prefix is required")
	}

	// 내가 차단한 유저
	myBlocks, err := s.blockRepo.FindByUserID(ctx, oid)
	if err != nil {
		return nil, domain.NewInternal("failed to find blocks: " + err.Error())
	}

	// 나를 차단한 유저
	blockedMe, err := s.blockRepo.FindByBlockedID(ctx, oid)
	if err != nil {
		return nil, domain.NewInternal("failed to find blocks: " + err.Error())
	}

	// 나를 차단한 유저만 검색에서 제외 (내가 차단한 유저는 표시)
	excludeIDs := []primitive.ObjectID{oid}
	for _, b := range blockedMe {
		excludeIDs = append(excludeIDs, b.UserID)
	}

	myBlockedSet := make(map[primitive.ObjectID]bool, len(myBlocks))
	for _, b := range myBlocks {
		myBlockedSet[b.BlockedID] = true
	}

	sanitized := regexp.QuoteMeta(tagPrefix)
	users, err := s.userRepo.SearchByTagPrefix(ctx, sanitized, excludeIDs, 20)
	if err != nil {
		return nil, domain.NewInternal("failed to search users: " + err.Error())
	}

	items := make([]dto.UserSearchItem, len(users))
	for i, u := range users {
		items[i] = dto.UserSearchItem{
			UserID:    u.ID.Hex(),
			Nickname:  u.Nickname,
			Tag:       u.Tag,
			IsBlocked: myBlockedSet[u.ID],
		}
	}

	return &dto.UserSearchResponse{Users: items}, nil
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

	if oid == targetOid {
		return domain.NewBadRequest(domain.ErrBadRequest, "cannot block yourself")
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

	if err := s.friendshipRepo.DeleteByUserPair(ctx, oid, targetOid); err != nil {
		return domain.NewInternal("failed to delete friendship: " + err.Error())
	}

	if err := s.friendRequestRepo.DeleteByUserPair(ctx, oid, targetOid); err != nil {
		return domain.NewInternal("failed to delete friend requests: " + err.Error())
	}

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

func (s *userService) ChangeNickname(ctx context.Context, userID string, req dto.ChangeNicknameRequest) (*dto.ChangeNicknameResponse, error) {
	length := utf8.RuneCountInString(req.Nickname)
	if length < 3 || length > 15 {
		return nil, domain.NewBadRequest(domain.ErrInvalidNickname, "nickname must be 3-15 characters")
	}

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	user, err := s.userRepo.FindByID(ctx, oid)
	if err != nil || user == nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "user not found")
	}

	if user.Nickname == "" {
		return nil, domain.NewBadRequest(domain.ErrInvalidNickname, "nickname is not set yet")
	}

	if err := s.userRepo.UpdateNickname(ctx, oid, req.Nickname); err != nil {
		return nil, domain.NewInternal("failed to update nickname: " + err.Error())
	}

	return &dto.ChangeNicknameResponse{Nickname: req.Nickname}, nil
}
