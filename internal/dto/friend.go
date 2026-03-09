package dto

import "time"

// GET /friends
type FriendListResponse struct {
	Friends []FriendItem `json:"friends"`
}

type FriendItem struct {
	UserID    string    `json:"userId"`
	Nickname  string    `json:"nickname"`
	Tag       string    `json:"tag"`
	IsInVoid  bool      `json:"isInVoid"`
	CreatedAt time.Time `json:"createdAt"`
}

// GET /friends/requests?type=received
type ReceivedRequestsResponse struct {
	Requests []ReceivedRequestItem `json:"requests"`
}

type ReceivedRequestItem struct {
	RequestID string         `json:"requestId"`
	Sender    UserSearchItem `json:"sender"`
	CreatedAt time.Time      `json:"createdAt"`
}

// GET /friends/requests?type=sent
type SentRequestsResponse struct {
	Requests []SentRequestItem `json:"requests"`
}

type SentRequestItem struct {
	RequestID string         `json:"requestId"`
	Receiver  UserSearchItem `json:"receiver"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
}

// POST /friends/requests
type SendFriendRequestRequest struct {
	ReceiverID string `json:"receiverId" validate:"required"`
}

type SendFriendRequestResponse struct {
	RequestID string `json:"requestId"`
}
