package dto

import "time"

// POST /void/start
type VoidStartResponse struct {
	SessionID string    `json:"session_id"`
	StartedAt time.Time `json:"started_at"`
	TargetDay string    `json:"target_day"`
}

// POST /void/end
type VoidEndRequest struct {
	Activities []string `json:"activities" validate:"max=5"`
}

type VoidEndResponse struct {
	SessionID   string    `json:"session_id"`
	StartedAt   time.Time `json:"started_at"`
	EndedAt     time.Time `json:"ended_at"`
	DurationSec int64     `json:"duration_sec"`
	TargetDay   string    `json:"target_day"`
	Activities  []string  `json:"activities"`
}

// POST /void/test - 테스트 공백 데이터 생성
type TestVoidRequest struct {
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at"`
	Activities []string  `json:"activities"`
}

// GET /void/history
type VoidHistoryResponse struct {
	TargetDay        string        `json:"target_day"`
	Sessions         []VoidSession `json:"sessions"`
	TotalDurationSec int64         `json:"total_duration_sec"`
}

type VoidSession struct {
	SessionID   string    `json:"session_id"`
	StartedAt   time.Time `json:"started_at"`
	EndedAt     time.Time `json:"ended_at"`
	DurationSec int64     `json:"duration_sec"`
	Activities  []string  `json:"activities"`
}
