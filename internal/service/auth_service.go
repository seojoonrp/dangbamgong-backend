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
	userRepo       repository.UserRepository
	socialVerifier auth.SocialVerifier
}

func NewAuthService(ur repository.UserRepository, socialVerifier auth.SocialVerifier) AuthService {
	return &authService{
		userRepo:       ur,
		socialVerifier: socialVerifier,
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
	appleRefreshToken string,
) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindBySocial(ctx, provider, socialID)
	if err != nil {
		return nil, domain.NewInternal("failed to find user: " + err.Error())
	}

	isNewUser := false
	if user == nil {
		isNewUser = true
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

	// TODO: revoke social account (Apple uses apple_refresh_token)
	// TODO: delete related data (activities, friends, void sessions, etc.)

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
