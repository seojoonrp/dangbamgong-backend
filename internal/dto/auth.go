package dto

// POST /auth/login
type LoginRequest struct {
	Provider          string  `json:"provider" validate:"required,oneof=GOOGLE KAKAO APPLE"`
	IDToken           string  `json:"id_token" validate:"required"`
	AppleRefreshToken *string `json:"apple_refresh_token"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	IsNewUser   bool   `json:"is_new_user"`
}

// POST /auth/login/test
type TestLoginRequest struct {
	SocialID string `json:"social_id" validate:"required"`
}

// POST /auth/nickname
type SetNicknameRequest struct {
	Nickname string `json:"nickname" validate:"required,min=3,max=20"`
}

type SetNicknameResponse struct {
	Nickname string `json:"nickname"`
}
