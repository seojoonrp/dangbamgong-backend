package auth

import (
	"context"

	"dangbamgong-backend/internal/domain"
)

type SocialVerifyResult struct {
	SocialID string
}

type SocialVerifier interface {
	Verify(ctx context.Context, provider string, idToken string) (*SocialVerifyResult, error)
}

type defaultSocialVerifier struct{}

func NewSocialVerifier() SocialVerifier {
	return &defaultSocialVerifier{}
}

func (v *defaultSocialVerifier) Verify(ctx context.Context, provider string, idToken string) (*SocialVerifyResult, error) {
	// TODO: implement GOOGLE, KAKAO, APPLE verification
	return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "social login not yet implemented for "+provider)
}
