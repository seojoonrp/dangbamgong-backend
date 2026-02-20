package dto

// GET /stats/live
type LiveStatResponse struct {
	CurrentVoidCount int `json:"current_void_count"`
	TodaySleptCount  int `json:"today_slept_count"`
}

// GET /stats/daily
type DailyStatResponse struct {
	TargetDay  string       `json:"target_day"`
	Buckets    []BucketItem `json:"buckets"`
	MyRank     *int         `json:"my_rank"`
	TotalUsers *int         `json:"total_users"`
}

type BucketItem struct {
	Time   string `json:"time"`
	Count  int    `json:"count"`
	IsMine bool   `json:"is_mine"`
}
