package dto

// GET /stats/live
type LiveStatResponse struct {
	CurrentVoidCount int `json:"currentVoidCount"`
	TodaySleptCount  int `json:"todaySleptCount"`
}

// GET /stats/daily
type DailyStatResponse struct {
	TargetDay  string       `json:"targetDay"`
	Buckets    []BucketItem `json:"buckets"`
	MyRank     *int         `json:"myRank"`
	TotalUsers *int         `json:"totalUsers"`
}

type BucketItem struct {
	Time   string `json:"time"`
	Count  int    `json:"count"`
	IsMine bool   `json:"isMine"`
}
