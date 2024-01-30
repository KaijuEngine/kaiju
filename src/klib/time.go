package klib

import "time"

const (
	ISO8601 = "2006-01-02T15:04:05Z"
)

func TickerWait(interval, limit time.Duration, condition func() bool) bool {
	ticker := time.NewTicker(interval)
	start := time.Now()
	for tt := range ticker.C {
		if res := condition(); res || tt.Sub(start) > limit {
			ticker.Stop()
			return res
		}
	}
	return false
}
