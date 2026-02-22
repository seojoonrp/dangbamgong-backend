package config

import "time"

const DayStartHour = 16 // 하루의 시작 기준 시각

var KST = time.FixedZone("KST", 9*60*60)

// 주어진 시간이 속하는 대상 날짜 반환
// 16시(KST) 이전이면 전날, 16시 이후이면 당일
func CalcTargetDay(t time.Time) string {
	kst := t.In(KST)
	if kst.Hour() < DayStartHour {
		kst = kst.AddDate(0, 0, -1)
	}
	return kst.Format("2006-01-02")
}
