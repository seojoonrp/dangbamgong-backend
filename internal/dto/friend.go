package dto

import "time"

// GET /friends/search
type FriendSearchResponse struct {
	Users []FriendSearchItem `json:"users"`
}

type FriendSearchItem struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Tag      string `json:"tag"`
}

// GET /friends
type FriendListResponse struct {
	Friends []FriendItem `json:"friends"`
}

type FriendItem struct {
	UserID    string    `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Tag       string    `json:"tag"`
	IsInVoid  bool      `json:"is_in_void"`
	CreatedAt time.Time `json:"created_at"`
}

// GET /friends/requests?type=received
type ReceivedRequestsResponse struct {
	Requests []ReceivedRequestItem `json:"requests"`
}

type ReceivedRequestItem struct {
	RequestID string           `json:"request_id"`
	Sender    FriendSearchItem `json:"sender"`
	CreatedAt time.Time        `json:"created_at"`
}

// GET /friends/requests?type=sent
type SentRequestsResponse struct {
	Requests []SentRequestItem `json:"requests"`
}

type SentRequestItem struct {
	RequestID string           `json:"request_id"`
	Receiver  FriendSearchItem `json:"receiver"`
	Status    string           `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
}

// POST /friends/requests
type SendFriendRequestRequest struct {
	ReceiverID string `json:"receiver_id" validate:"required"`
}

type SendFriendRequestResponse struct {
	RequestID string `json:"request_id"`
}
