package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"
	"unicode/utf8"

	"dangbamgong-backend/internal/auth"
	"dangbamgong-backend/internal/domain"
	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/model"
	"dangbamgong-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	TestLogin(ctx context.Context, req dto.TestLoginRequest) (*dto.LoginResponse, error)
	SetNickname(ctx context.Context, userID string, req dto.SetNicknameRequest) (*dto.SetNicknameResponse, error)
	Withdraw(ctx context.Context, userID string) error
}

type authService struct {
	userRepo          repository.UserRepository
	activityRepo      repository.ActivityRepository
	friendshipRepo    repository.FriendshipRepository
	friendRequestRepo repository.FriendRequestRepository
	voidSessionRepo   repository.VoidSessionRepository
	blockRepo         repository.BlockRepository
	deviceTokenRepo   repository.DeviceTokenRepository
	notifRepo         repository.NotificationRepository
	socialVerifier    auth.SocialVerifier
}

func NewAuthService(
	ur repository.UserRepository,
	ar repository.ActivityRepository,
	fr repository.FriendshipRepository,
	frr repository.FriendRequestRepository,
	vsr repository.VoidSessionRepository,
	br repository.BlockRepository,
	dtr repository.DeviceTokenRepository,
	nr repository.NotificationRepository,
	socialVerifier auth.SocialVerifier,
) AuthService {
	return &authService{
		userRepo:          ur,
		activityRepo:      ar,
		friendshipRepo:    fr,
		friendRequestRepo: frr,
		voidSessionRepo:   vsr,
		blockRepo:         br,
		deviceTokenRepo:   dtr,
		notifRepo:         nr,
		socialVerifier:    socialVerifier,
	}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	result, err := s.socialVerifier.Verify(ctx, req.Provider, req.IDToken)
	if err != nil {
		return nil, err
	}

	var appleRefresh string
	if req.AppleRefreshToken != nil {
		appleRefresh = *req.AppleRefreshToken
	}

	return s.findOrCreateAndGenerateToken(ctx, model.SocialProvider(req.Provider), result.SocialID, appleRefresh)
}

func (s *authService) TestLogin(ctx context.Context, req dto.TestLoginRequest) (*dto.LoginResponse, error) {
	return s.findOrCreateAndGenerateToken(ctx, model.ProviderTest, req.SocialID, "")
}

func (s *authService) findOrCreateAndGenerateToken(
	ctx context.Context,
	provider model.SocialProvider,
	socialID string,
	appleAuthCode string,
) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindBySocial(ctx, provider, socialID)
	if err != nil {
		return nil, domain.NewInternal("failed to find user: " + err.Error())
	}

	isNewUser := false
	if user == nil {
		isNewUser = true

		var appleRefreshToken string
		if provider == model.ProviderApple && appleAuthCode != "" {
			if rt, err := auth.ExchangeAppleAuthCode(ctx, appleAuthCode); err == nil {
				appleRefreshToken = rt
			}
		}

		now := time.Now()
		user = &model.User{
			SocialProvider: provider,
			SocialID:       socialID,
			NotificationSettings: model.NotificationSettings{
				VoidReminder:  true,
				ReminderHours: 1,
				FriendRequest: true,
				FriendNudge:   true,
			},
			AppleRefreshToken: appleRefreshToken,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := s.createUserWithUniqueTag(ctx, user); err != nil {
			return nil, err
		}
		s.createDefaultActivities(ctx, user.ID)
	} else if provider == model.ProviderApple && appleAuthCode != "" {
		// 기존 Apple 사용자 재로그인 시 refresh token 갱신
		if rt, err := auth.ExchangeAppleAuthCode(ctx, appleAuthCode); err == nil {
			_ = s.userRepo.UpdateAppleRefreshToken(ctx, user.ID, rt)
		}
	}

	token, err := auth.GenerateToken(user.ID.Hex())
	if err != nil {
		return nil, domain.NewInternal("failed to generate token: " + err.Error())
	}

	return &dto.LoginResponse{
		AccessToken: token,
		IsNewUser:   isNewUser,
	}, nil
}

func (s *authService) SetNickname(ctx context.Context, userID string, req dto.SetNicknameRequest) (*dto.SetNicknameResponse, error) {
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

	if user.Nickname != "" {
		return nil, domain.NewConflict(domain.ErrNicknameAlreadySet, "nickname is already set")
	}

	if err := s.userRepo.UpdateNickname(ctx, oid, req.Nickname); err != nil {
		return nil, domain.NewInternal("failed to update nickname: " + err.Error())
	}

	return &dto.SetNicknameResponse{Nickname: req.Nickname}, nil
}

func (s *authService) Withdraw(ctx context.Context, userID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	user, err := s.userRepo.FindByID(ctx, oid)
	if err != nil || user == nil {
		return domain.NewUnauthorized(domain.ErrUnauthorized, "user not found")
	}

	// Apple 계정인 경우 refresh token revoke
	if user.SocialProvider == model.ProviderApple && user.AppleRefreshToken != "" {
		if err := auth.RevokeAppleToken(ctx, user.AppleRefreshToken); err != nil {
			return domain.NewInternal("failed to revoke apple token: " + err.Error())
		}
	}

	// 연관 데이터 hard delete
	deleteOps := []struct {
		name string
		fn   func() error
	}{
		{"activities", func() error { return s.activityRepo.DeleteByUserID(ctx, oid) }},
		{"friendships", func() error { return s.friendshipRepo.DeleteByUserID(ctx, oid) }},
		{"friend_requests", func() error { return s.friendRequestRepo.DeleteByUserID(ctx, oid) }},
		{"void_sessions", func() error { return s.voidSessionRepo.DeleteByUserID(ctx, oid) }},
		{"blocks", func() error { return s.blockRepo.DeleteByUserID(ctx, oid) }},
		{"device_tokens", func() error { return s.deviceTokenRepo.DeleteByUserID(ctx, oid) }},
		{"notifications", func() error { return s.notifRepo.DeleteByUserID(ctx, oid) }},
	}

	for _, op := range deleteOps {
		if err := op.fn(); err != nil {
			return domain.NewInternal("failed to delete " + op.name + ": " + err.Error())
		}
	}

	// 마지막으로 유저 삭제
	if err := s.userRepo.DeleteByID(ctx, oid); err != nil {
		return domain.NewInternal("failed to delete user: " + err.Error())
	}

	return nil
}

func (s *authService) createUserWithUniqueTag(ctx context.Context, user *model.User) error {
	const maxRetries = 5
	for i := 0; i < maxRetries; i++ {
		tag, err := generateTag(10)
		if err != nil {
			return domain.NewInternal("failed to generate tag: " + err.Error())
		}
		user.Tag = tag
		err = s.userRepo.Create(ctx, user)
		if err == nil {
			return nil
		}
		if !mongo.IsDuplicateKeyError(err) {
			return domain.NewInternal("failed to create user: " + err.Error())
		}
	}
	return domain.NewInternal("failed to generate unique tag after retries")
}

func (s *authService) createDefaultActivities(ctx context.Context, userID primitive.ObjectID) {
	now := time.Now()
	defaults := []string{"유튜브", "릴스"}
	for _, name := range defaults {
		activity := &model.Activity{
			UserID:    userID,
			Name:      name,
			CreatedAt: now,
		}
		s.activityRepo.Create(ctx, activity)
	}
}

func generateTag(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}
