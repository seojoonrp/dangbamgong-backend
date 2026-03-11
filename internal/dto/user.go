package dto

import "time"

// GET /users/me
type UserMeResponse struct {
	ID                   string               `json:"id"`
	Tag                  string               `json:"tag"`
	Nickname             string               `json:"nickname"`
	IsInVoid             bool                 `json:"isInVoid"`
	CurrentVoidStartedAt *time.Time           `json:"currentVoidStartedAt"`
	NotificationSettings NotificationSettings `json:"notificationSettings"`
}

type NotificationSettings struct {
	VoidReminder  bool `json:"voidReminder"`
	ReminderHours int  `json:"reminderHours"`
	FriendRequest bool `json:"friendRequest"`
	FriendNudge   bool `json:"friendNudge"`
}

// PATCH /users/me/settings
type UpdateSettingsRequest struct {
	VoidReminder  *bool `json:"voidReminder"`
	ReminderHours *int  `json:"reminderHours"`
	FriendRequest *bool `json:"friendRequest"`
	FriendNudge   *bool `json:"friendNudge"`
}

type UpdateSettingsResponse struct {
	VoidReminder  bool `json:"voidReminder"`
	ReminderHours int  `json:"reminderHours"`
	FriendRequest bool `json:"friendRequest"`
	FriendNudge   bool `json:"friendNudge"`
}

// GET /users/blocks
type BlockListResponse struct {
	Blocks []BlockItem `json:"blocks"`
}

// PATCH /users/me/nickname
type ChangeNicknameRequest struct {
	Nickname string `json:"nickname" validate:"required,min=3,max=15"`
}

type ChangeNicknameResponse struct {
	Nickname string `json:"nickname"`
}

// GET /users/search
type UserSearchResponse struct {
	Users []UserSearchItem `json:"users"`
}

type UserSearchItem struct {
	UserID    string `json:"userId"`
	Nickname  string `json:"nickname"`
	Tag       string `json:"tag"`
	IsBlocked bool   `json:"isBlocked"`
}

type BlockItem struct {
	UserID    string    `json:"userId"`
	Nickname  string    `json:"nickname"`
	Tag       string    `json:"tag"`
	BlockedAt time.Time `json:"blockedAt"`
}
