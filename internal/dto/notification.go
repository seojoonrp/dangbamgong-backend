package dto

import "time"

// GET /notifications
type NotificationItem struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	IsRead    bool      `json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}

type NotificationListResponse struct {
	Notifications []NotificationItem `json:"notifications"`
	HasMore       bool               `json:"hasMore"`
}

// GET /notifications/unread-count
type UnreadCountResponse struct {
	Count int `json:"count"`
}

// PUT /devices/token
type RegisterDeviceRequest struct {
	Token string `json:"token" validate:"required"`
}
