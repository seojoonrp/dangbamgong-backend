package dto

// POST /auth/login
type LoginRequest struct {
	Provider          string  `json:"provider" validate:"required,oneof=GOOGLE KAKAO APPLE"`
	IDToken           string  `json:"idToken" validate:"required"`
	AppleRefreshToken *string `json:"appleRefreshToken"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	IsNewUser   bool   `json:"isNewUser"`
}

// POST /auth/login/test
type TestLoginRequest struct {
	SocialID string `json:"socialId" validate:"required"`
}

// POST /auth/nickname
type SetNicknameRequest struct {
	Nickname string `json:"nickname" validate:"required,min=3,max=15"`
}

type SetNicknameResponse struct {
	Nickname string `json:"nickname"`
}
