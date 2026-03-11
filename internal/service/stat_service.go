package service

import (
	"context"
	"fmt"
	"time"

	"dangbamgong-backend/internal/config"
	"dangbamgong-backend/internal/domain"
	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/model"
	"dangbamgong-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const bucketInterval = 20 * time.Minute

type StatService interface {
	GetHomeStat(ctx context.Context, userID string) (*dto.HomeStatResponse, error)
	GetDailyStat(ctx context.Context, userID string, targetDay string) (*dto.DailyStatResponse, error)
	GetMyVoidStat(ctx context.Context, userID string) (*dto.MyVoidStatResponse, error)
}

type statService struct {
	statRepo        repository.StatRepository
	voidSessionRepo repository.VoidSessionRepository
}

func NewStatService(
	sr repository.StatRepository,
	vr repository.VoidSessionRepository,
) StatService {
	return &statService{
		statRepo:        sr,
		voidSessionRepo: vr,
	}
}

func (s *statService) GetHomeStat(ctx context.Context, userID string) (*dto.HomeStatResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	currentCount, err := s.statRepo.CountCurrentVoid(ctx)
	if err != nil {
		return nil, domain.NewInternal("failed to count current void: " + err.Error())
	}

	today := config.CalcTargetDay(time.Now())
	sleptCount, err := s.statRepo.CountDistinctUsersForDay(ctx, today)
	if err != nil {
		return nil, domain.NewInternal("failed to count today slept: " + err.Error())
	}

	resp := &dto.HomeStatResponse{
		CurrentVoidCount: currentCount,
		TodaySleptCount:  sleptCount,
	}

	// 유저의 오늘 랭킹 계산
	durations, err := s.statRepo.GetUserDurations(ctx, today)
	if err != nil {
		return nil, domain.NewInternal("failed to get user durations: " + err.Error())
	}

	for i, d := range durations {
		if d.UserID == oid {
			rank := i + 1
			total := len(durations)
			resp.MyTotalDurationSec = &d.TotalDurSec
			resp.MyRank = &rank
			resp.TotalSleptUsers = &total
			break
		}
	}

	return resp, nil
}

func (s *statService) GetDailyStat(ctx context.Context, userID string, targetDay string) (*dto.DailyStatResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	if targetDay == "" {
		return nil, domain.NewBadRequest(domain.ErrBadRequest, "target_day is required")
	}

	now := time.Now()

	// 필요한 버킷 목록 생성
	expectedBuckets := generateBuckets(targetDay, now)

	// 캐시 조회
	cached, err := s.statRepo.GetBucketCache(ctx, targetDay)
	if err != nil {
		return nil, domain.NewInternal("failed to get bucket cache: " + err.Error())
	}

	cachedSet := make(map[string]int, len(cached))
	for _, c := range cached {
		cachedSet[c.Bucket] = c.Count
	}

	// 미싱 버킷 확인
	var missingBuckets []string
	for _, b := range expectedBuckets {
		if _, ok := cachedSet[b]; !ok {
			missingBuckets = append(missingBuckets, b)
		}
	}

	// 미싱 버킷이 있으면 세션에서 계산 후 캐싱
	if len(missingBuckets) > 0 {
		sessions, err := s.voidSessionRepo.FindByTargetDay(ctx, targetDay)
		if err != nil {
			return nil, domain.NewInternal("failed to find sessions: " + err.Error())
		}

		computed := computeBucketCounts(targetDay, missingBuckets, sessions)
		if err := s.statRepo.UpsertBucketCache(ctx, computed); err != nil {
			return nil, domain.NewInternal("failed to upsert cache: " + err.Error())
		}

		for _, c := range computed {
			cachedSet[c.Bucket] = c.Count
		}
	}

	// 유저 본인의 세션으로 is_mine 계산 + 세션 목록 반환
	mySessions, err := s.voidSessionRepo.FindByUserIDAndTargetDay(ctx, oid, targetDay)
	if err != nil {
		return nil, domain.NewInternal("failed to find user sessions: " + err.Error())
	}

	bucketItems := make([]dto.BucketItem, len(expectedBuckets))
	for i, b := range expectedBuckets {
		bucketItems[i] = dto.BucketItem{
			Time:   bucketToDisplayTime(b),
			Count:  cachedSet[b],
			IsMine: isUserInBucket(b, mySessions),
		}
	}

	// 세션 목록을 DTO로 변환
	sessionItems := make([]dto.VoidSessionItem, len(mySessions))
	for i, s := range mySessions {
		sessionItems[i] = dto.VoidSessionItem{
			StartedAt:  s.StartedAt,
			EndedAt:    s.EndedAt,
			Activities: s.Activities,
		}
	}

	return &dto.DailyStatResponse{
		TargetDay:  targetDay,
		Buckets:    bucketItems,
		MySessions: sessionItems,
	}, nil
}

func (s *statService) GetMyVoidStat(ctx context.Context, userID string) (*dto.MyVoidStatResponse, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, domain.NewUnauthorized(domain.ErrUnauthorized, "invalid user id")
	}

	stats, err := s.voidSessionRepo.AggregateUserStats(ctx, oid)
	if err != nil {
		return nil, domain.NewInternal("failed to aggregate void stats: " + err.Error())
	}

	if stats == nil {
		return &dto.MyVoidStatResponse{
			TotalDurationSec:   0,
			AverageDurationSec: 0,
			MaxDurationSec:     0,
		}, nil
	}

	var avg int64
	if stats.SessionCount > 0 {
		avg = stats.TotalDurationSec / int64(stats.SessionCount)
	}

	return &dto.MyVoidStatResponse{
		TotalDurationSec:   stats.TotalDurationSec,
		AverageDurationSec: avg,
		MaxDurationSec:     stats.MaxDurationSec,
	}, nil
}

// generateBuckets 는 targetDay의 16:00부터 현재 시간 직전 완료된 20분 버킷까지의 버킷 키를 생성한다.
func generateBuckets(targetDay string, now time.Time) []string {
	dayStart, err := time.ParseInLocation("2006-01-02", targetDay, config.KST)
	if err != nil {
		return nil
	}
	dayStart = dayStart.Add(time.Duration(config.DayStartHour) * time.Hour)

	// 마지막 완료 버킷: now를 20분 단위로 내림 후 20분 빼기
	nowKST := now.In(config.KST)
	lastComplete := nowKST.Truncate(bucketInterval).Add(-bucketInterval)

	// 과거 날짜면 다음날 15:40까지 (최대 72버킷)
	dayEnd := dayStart.Add(24 * time.Hour).Add(-bucketInterval)
	if lastComplete.After(dayEnd) {
		lastComplete = dayEnd
	}

	if lastComplete.Before(dayStart) {
		return nil
	}

	var buckets []string
	for t := dayStart; !t.After(lastComplete); t = t.Add(bucketInterval) {
		buckets = append(buckets, formatBucketKey(t))
	}
	return buckets
}

// computeBucketCounts 는 세션 목록으로부터 각 버킷의 유저 수를 계산한다.
func computeBucketCounts(targetDay string, buckets []string, sessions []model.VoidSession) []model.VoidStatCache {
	bucketUsers := make(map[string]map[primitive.ObjectID]struct{})
	for _, b := range buckets {
		bucketUsers[b] = make(map[primitive.ObjectID]struct{})
	}

	for _, session := range sessions {
		for _, b := range buckets {
			bucketTime := parseBucketKey(b)
			if bucketTime.IsZero() {
				continue
			}
			bucketEnd := bucketTime.Add(bucketInterval)

			// 세션이 버킷과 겹치는지: started_at < bucket_end AND ended_at > bucket_start
			if session.StartedAt.Before(bucketEnd) && session.EndedAt.After(bucketTime) {
				bucketUsers[b][session.UserID] = struct{}{}
			}
		}
	}

	now := time.Now()
	caches := make([]model.VoidStatCache, len(buckets))
	for i, b := range buckets {
		caches[i] = model.VoidStatCache{
			TargetDay: targetDay,
			Bucket:    b,
			Count:     len(bucketUsers[b]),
			UpdatedAt: now,
		}
	}
	return caches
}

// isUserInBucket 는 유저의 세션이 해당 버킷과 겹치는지 확인한다.
func isUserInBucket(bucket string, sessions []model.VoidSession) bool {
	bucketTime := parseBucketKey(bucket)
	if bucketTime.IsZero() {
		return false
	}
	bucketEnd := bucketTime.Add(bucketInterval)

	for _, session := range sessions {
		if session.StartedAt.Before(bucketEnd) && session.EndedAt.After(bucketTime) {
			return true
		}
	}
	return false
}

// formatBucketKey 는 "2006-01-02T15:04" 형식의 버킷 키를 생성한다.
func formatBucketKey(t time.Time) string {
	return t.In(config.KST).Format("2006-01-02T15:04")
}

// parseBucketKey 는 버킷 키를 time.Time으로 파싱한다.
func parseBucketKey(key string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02T15:04", key, config.KST)
	return t
}

// bucketToDisplayTime 는 버킷 키에서 "HH:MM" 시간 부분만 추출한다.
func bucketToDisplayTime(bucket string) string {
	if len(bucket) >= 16 {
		return fmt.Sprintf("%s:%s", bucket[11:13], bucket[14:16])
	}
	return bucket
}
