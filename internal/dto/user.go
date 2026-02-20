package dto

import "time"

// GET /users/me
type UserMeResponse struct {
	ID                   string               `json:"id"`
	Tag                  string               `json:"tag"`
	Nickname             string               `json:"nickname"`
	IsInVoid             bool                 `json:"is_in_void"`
	CurrentVoidStartedAt *time.Time           `json:"current_void_started_at"`
	NotificationSettings NotificationSettings `json:"notification_settings"`
}

type NotificationSettings struct {
	VoidReminder  bool `json:"void_reminder"`
	ReminderHours int  `json:"reminder_hours"`
	FriendNudge   bool `json:"friend_nudge"`
}

// PATCH /users/me/settings
type UpdateSettingsRequest struct {
	VoidReminder  *bool `json:"void_reminder"`
	ReminderHours *int  `json:"reminder_hours"`
	FriendNudge   *bool `json:"friend_nudge"`
}

type UpdateSettingsResponse struct {
	VoidReminder  bool `json:"void_reminder"`
	ReminderHours int  `json:"reminder_hours"`
	FriendNudge   bool `json:"friend_nudge"`
}

// GET /users/blocks
type BlockListResponse struct {
	Blocks []BlockItem `json:"blocks"`
}

type BlockItem struct {
	UserID    string    `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Tag       string    `json:"tag"`
	BlockedAt time.Time `json:"blocked_at"`
}
