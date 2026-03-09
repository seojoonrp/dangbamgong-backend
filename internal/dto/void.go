package dto

import "time"

// POST /void/start
type VoidStartResponse struct {
	SessionID string    `json:"sessionId"`
	StartedAt time.Time `json:"startedAt"`
	TargetDay string    `json:"targetDay"`
}

// POST /void/end
type VoidEndRequest struct {
	Activities []string `json:"activities" validate:"max=5"`
}

type VoidEndResponse struct {
	SessionID   string    `json:"sessionId"`
	StartedAt   time.Time `json:"startedAt"`
	EndedAt     time.Time `json:"endedAt"`
	DurationSec int64     `json:"durationSec"`
	TargetDay   string    `json:"targetDay"`
	Activities  []string  `json:"activities"`
}

// POST /void/test - 테스트 공백 데이터 생성
type TestVoidRequest struct {
	StartedAt  time.Time `json:"startedAt"`
	EndedAt    time.Time `json:"endedAt"`
	Activities []string  `json:"activities"`
}

// GET /void/history
type VoidHistoryResponse struct {
	TargetDay        string        `json:"targetDay"`
	Sessions         []VoidSession `json:"sessions"`
	TotalDurationSec int64         `json:"totalDurationSec"`
}

type VoidSession struct {
	SessionID   string    `json:"sessionId"`
	StartedAt   time.Time `json:"startedAt"`
	EndedAt     time.Time `json:"endedAt"`
	DurationSec int64     `json:"durationSec"`
	Activities  []string  `json:"activities"`
}
