package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const expoPushURL = "https://exp.host/--/api/v2/push/send"

type PushClient interface {
	Send(ctx context.Context, deviceToken string, title string, body string, data map[string]string) error
}

// ExpoPushMessage は Expo Push API のリクエストボディ
type ExpoPushMessage struct {
	To    string            `json:"to"`
	Title string            `json:"title,omitempty"`
	Body  string            `json:"body,omitempty"`
	Sound string            `json:"sound,omitempty"`
	Data  map[string]string `json:"data,omitempty"`
}

type expoClient struct {
	httpClient  *http.Client
	accessToken string // optional
}

func NewExpoPushClient() PushClient {
	accessToken := os.Getenv("EXPO_ACCESS_TOKEN")

	if accessToken != "" {
		log.Println("[PUSH] Expo push client initialized with access token")
	} else {
		log.Println("[PUSH] Expo push client initialized without access token")
	}

	return &expoClient{
		httpClient:  &http.Client{},
		accessToken: accessToken,
	}
}

func (c *expoClient) Send(ctx context.Context, deviceToken string, title string, body string, data map[string]string) error {
	// TODO: 배치 전송 최적화 — 현재는 단건 전송이지만,
	// 여러 디바이스에 동시에 보낼 경우 []ExpoPushMessage 슬라이스로
	// 한 번의 HTTP 요청에 최대 100개까지 묶어 보내는 방식을 고려해볼 것.
	// Expo Push API는 배열 형태의 요청을 지원함.
	msg := ExpoPushMessage{
		To:    deviceToken,
		Title: title,
		Body:  body,
		Sound: "default",
		Data:  data,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal expo push message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, expoPushURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create expo push request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send expo push: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("[PUSH] Expo push HTTP error: status=%d body=%s\n", resp.StatusCode, string(respBody))
		return fmt.Errorf("expo push returned status %d", resp.StatusCode)
	}

	// Expo 응답에서 개별 ticket 에러 확인
	var result struct {
		Data []struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Details struct {
				Error string `json:"error"`
			} `json:"details"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && len(result.Data) > 0 {
		ticket := result.Data[0]
		if ticket.Status == "error" {
			log.Printf("[PUSH] Expo push ticket error: %s — %s (token=%s)\n",
				ticket.Details.Error, ticket.Message, deviceToken)
		} else {
			log.Printf("[PUSH] Expo push sent successfully (token=%s)\n", deviceToken)
		}
	}

	return nil
}

// Expo 설정되지 않은 경우 로그용 noop
type noopClient struct{}

func (c *noopClient) Send(ctx context.Context, deviceToken string, title string, body string, data map[string]string) error {
	log.Printf("[PUSH] noop: title=%s body=%s token=%s\n", title, body, deviceToken)
	return nil
}
