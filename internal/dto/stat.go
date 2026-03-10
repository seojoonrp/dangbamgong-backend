package dto

import "time"

// GET /stats/home
type HomeStatResponse struct {
	// 오늘 잠에 든 유저
	MyTotalDurationSec *int64 `json:"myTotalDurationSec"`
	MyRank             *int   `json:"myRank"`
	TotalSleptUsers    *int   `json:"totalSleptUsers"`
	// 공통
	CurrentVoidCount int `json:"currentVoidCount"`
	TodaySleptCount  int `json:"todaySleptCount"`
}

// GET /stats/daily
type DailyStatResponse struct {
	TargetDay  string            `json:"targetDay"`
	Buckets    []BucketItem      `json:"buckets"`
	MySessions []VoidSessionItem `json:"mySessions"`
}

type BucketItem struct {
	Time   string `json:"time"`
	Count  int    `json:"count"`
	IsMine bool   `json:"isMine"`
}

type VoidSessionItem struct {
	StartedAt  time.Time `json:"startedAt"`
	EndedAt    time.Time `json:"endedAt"`
	Activities []string  `json:"activities"`
}
