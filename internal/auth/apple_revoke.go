package auth

import (
	"context"
	"dangbamgong-backend/internal/domain"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sideshow/apns2/token"
)

func GenerateAppleClientSecret() (string, error) {
	authKey, err := token.AuthKeyFromFile(os.Getenv("APPLE_KEY_PATH"))
	if err != nil {
		return "", domain.NewInternal("failed to get auth key from file: " + err.Error())
	}

	now := time.Now()
	t := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": os.Getenv("APPLE_TEAM_ID"),
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": os.Getenv("APPLE_TOPIC"),
	})
	t.Header["kid"] = os.Getenv("APPLE_KEY_ID")

	signedToken, err := t.SignedString(authKey)
	if err != nil {
		return "", domain.NewInternal("failed to sign token: " + err.Error())
	}

	return signedToken, nil
}

// RevokeAppleToken revokes an Apple refresh token.
func RevokeAppleToken(ctx context.Context, refreshToken string) error {
	clientSecret, err := GenerateAppleClientSecret()
	if err != nil {
		return err
	}

	data := url.Values{
		"client_id":       {clientID()},
		"client_secret":   {clientSecret},
		"token":           {refreshToken},
		"token_type_hint": {"refresh_token"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://appleid.apple.com/auth/revoke",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("apple revoke failed with status %d", resp.StatusCode)
	}
	return nil
}

// ExchangeAppleAuthCode exchanges an Apple authorization code for a refresh token.
// iOS 앱에서 전달받은 one-time authorization code를 Apple의 /auth/token 엔드포인트에서 refresh token으로 교환한다.
func ExchangeAppleAuthCode(ctx context.Context, authCode string) (string, error) {
	clientSecret, err := GenerateAppleClientSecret()
	if err != nil {
		return "", err
	}

	data := url.Values{
		"client_id":     {clientID()},
		"client_secret": {clientSecret},
		"code":          {authCode},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://appleid.apple.com/auth/token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		RefreshToken string `json:"refresh_token"`
		Error        string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", domain.NewInternal("failed to decode apple token response: " + err.Error())
	}

	if result.Error != "" {
		return "", fmt.Errorf("apple token exchange failed: %s", result.Error)
	}
	if result.RefreshToken == "" {
		return "", fmt.Errorf("apple token exchange returned empty refresh token")
	}

	return result.RefreshToken, nil
}

func clientID() string {
	return os.Getenv("APPLE_TOPIC")
}
