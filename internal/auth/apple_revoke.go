package auth

import (
	"context"
	"dangbamgong-backend/internal/domain"
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

func clientID() string {
	return os.Getenv("APPLE_TOPIC")
}
