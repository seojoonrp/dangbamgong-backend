package config

import "time"

const DayStartHour = 16 // 하루의 시작 기준 시각

var KST = time.FixedZone("KST", 9*60*60)
