package klib

import (
	"context"
	"time"
)

const (
	ISO8601 = "2006-01-02T15:04:05Z"
)

func DelayCall(d time.Duration, f func(), ctx context.Context) {
	go func() {
		c, cf := context.WithTimeout(ctx, d)
		defer cf()
		<-c.Done()
		if ctx.Err() == nil {
			f()
		}
	}()
}

func TickerWait(interval, limit time.Duration, condition func() bool) bool {
	// TODO:  This will likely need to be bound to the host, if a host
	// is destroyed then this could run after the fact and cause a panic
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
