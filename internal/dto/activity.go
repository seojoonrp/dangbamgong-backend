package dto

import "time"

// GET /activities
type ActivityListResponse struct {
	Activities []ActivityItem `json:"activities"`
}

type ActivityItem struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	UsageCount int        `json:"usageCount"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

// POST /activities
type CreateActivityRequest struct {
	Name string `json:"name" validate:"required,min=1,max=10"`
}

// PATCH /activities/:activity_id
type UpdateActivityRequest struct {
	Name string `json:"name" validate:"required,min=1,max=10"`
}

type CreateActivityResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	UsageCount int        `json:"usageCount"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}
