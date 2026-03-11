package push

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

type PushClient interface {
	Send(ctx context.Context, deviceToken string, title string, body string, data map[string]string) error
}

type apnsClient struct {
	client *apns2.Client
	topic  string
}

func NewAPNsClient() PushClient {
	keyPath := os.Getenv("APNS_KEY_PATH")
	keyID := os.Getenv("APNS_KEY_ID")
	teamID := os.Getenv("APNS_TEAM_ID")
	topic := os.Getenv("APNS_TOPIC")
	env := os.Getenv("APNS_ENV")

	if keyPath == "" || keyID == "" || teamID == "" || topic == "" {
		log.Println("[PUSH] APNs not configured, using noop client")
		return &noopClient{}
	}

	authKey, err := token.AuthKeyFromFile(keyPath)
	if err != nil {
		log.Printf("[PUSH] Failed to load APNs auth key: %v, using noop client\n", err)
		return &noopClient{}
	}

	authToken := &token.Token{
		AuthKey: authKey,
		KeyID:   keyID,
		TeamID:  teamID,
	}

	var client *apns2.Client
	if env == "production" {
		client = apns2.NewTokenClient(authToken).Production()
	} else {
		client = apns2.NewTokenClient(authToken).Development()
	}

	return &apnsClient{
		client: client,
		topic:  topic,
	}
}

func (c *apnsClient) Send(ctx context.Context, deviceToken string, title string, body string, data map[string]string) error {
	p := payload.NewPayload().AlertTitle(title).AlertBody(body).Sound("default")

	if len(data) > 0 {
		for k, v := range data {
			p.Custom(k, v)
		}
	}

	payloadBytes, err := json.Marshal(p)
	if err != nil {
		return err
	}

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       c.topic,
		Payload:     payloadBytes,
		PushType:    apns2.PushTypeAlert,
	}

	resp, err := c.client.PushWithContext(ctx, notification)
	if err != nil {
		return err
	}

	if !resp.Sent() {
		log.Printf("[PUSH] APNs push not sent: %d %s\n", resp.StatusCode, resp.Reason)
	}

	return nil
}

// APN 설정되지 않은 경우 로그용 noop
type noopClient struct{}

func (c *noopClient) Send(ctx context.Context, deviceToken string, title string, body string, data map[string]string) error {
	log.Printf("[PUSH] noop: title=%s body=%s token=%s\n", title, body, deviceToken)
	return nil
}
