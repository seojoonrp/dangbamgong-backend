package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"

	"dangbamgong-backend/internal/domain"

	"github.com/golang-jwt/jwt/v5"
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
	switch provider {
	case "GOOGLE":
		return v.verifyGoogle(ctx, idToken)
	case "KAKAO":
		return v.verifyKakao(ctx, idToken)
	case "APPLE":
		return v.verifyApple(ctx, idToken)
	default:
		return nil, domain.NewBadRequest(domain.ErrInvalidToken, "unsupported provider: "+provider)
	}
}

// Google: ID token을 tokeninfo 엔드포인트로 검증
func (v *defaultSocialVerifier) verifyGoogle(ctx context.Context, idToken string) (*SocialVerifyResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://oauth2.googleapis.com/tokeninfo?id_token="+idToken, nil)
	if err != nil {
		return nil, domain.NewInternal("failed to create google verify request: " + err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, domain.NewInternal("failed to verify google token: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "invalid google id token")
	}

	var result struct {
		Sub string `json:"sub"`
		Aud string `json:"aud"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, domain.NewInternal("failed to decode google token info: " + err.Error())
	}

	clientID := os.Getenv("GOOGLE_WEB_CLIENT_ID")
	if result.Aud != clientID {
		return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "google token audience mismatch")
	}

	if result.Sub == "" {
		return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "google token missing sub")
	}

	return &SocialVerifyResult{SocialID: result.Sub}, nil
}

// Kakao: access token으로 사용자 정보 조회
func (v *defaultSocialVerifier) verifyKakao(ctx context.Context, accessToken string) (*SocialVerifyResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://kapi.kakao.com/v2/user/me", nil)
	if err != nil {
		return nil, domain.NewInternal("failed to create kakao verify request: " + err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, domain.NewInternal("failed to verify kakao token: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "invalid kakao access token")
	}

	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, domain.NewInternal("failed to decode kakao user info: " + err.Error())
	}

	if result.ID == 0 {
		return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "kakao user id not found")
	}

	return &SocialVerifyResult{SocialID: strconv.FormatInt(result.ID, 10)}, nil
}

// Apple: ID token (JWT)을 Apple 공개 키로 검증
func (v *defaultSocialVerifier) verifyApple(ctx context.Context, idToken string) (*SocialVerifyResult, error) {
	// JWT header에서 kid 추출 (검증 없이 파싱)
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "failed to parse apple id token: "+err.Error())
	}

	kid, ok := token.Header["kid"].(string)
	if !ok || kid == "" {
		return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "apple token missing kid header")
	}

	// Apple JWKS에서 공개 키 가져오기
	pubKey, err := fetchApplePublicKey(ctx, kid)
	if err != nil {
		return nil, err
	}

	// 서명 검증 + claims 파싱
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(idToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "failed to verify apple id token: "+err.Error())
	}

	// issuer 검증
	iss, _ := claims["iss"].(string)
	if iss != "https://appleid.apple.com" {
		return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "apple token issuer mismatch")
	}

	// audience 검증
	aud, _ := claims["aud"].(string)
	if aud != os.Getenv("APPLE_TOPIC") {
		return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "apple token audience mismatch")
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "apple token missing sub")
	}

	return &SocialVerifyResult{SocialID: sub}, nil
}

// Apple JWKS에서 kid에 매칭되는 RSA 공개 키 구성
func fetchApplePublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://appleid.apple.com/auth/keys", nil)
	if err != nil {
		return nil, domain.NewInternal("failed to create apple jwks request: " + err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, domain.NewInternal("failed to fetch apple jwks: " + err.Error())
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, domain.NewInternal("failed to decode apple jwks: " + err.Error())
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid {
			return parseRSAPublicKey(key.N, key.E)
		}
	}

	return nil, domain.NewUnauthorized(domain.ErrInvalidToken, "apple public key not found for kid: "+kid)
}

// base64url로 인코딩된 n, e 값으로 RSA 공개 키 생성
func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, domain.NewInternal("failed to decode apple key n: " + err.Error())
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, domain.NewInternal("failed to decode apple key e: " + err.Error())
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}
