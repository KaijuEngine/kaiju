/******************************************************************************/
/* host_frame_rate_test.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine

import (
	"testing"
	"time"
)

func TestSetFrameRateLimitReplacesAndStopsOldTicker(t *testing.T) {
	host := &Host{}
	host.SetFrameRateLimit(200)
	oldTicker := host.frameRateLimit
	if oldTicker == nil {
		t.Fatal("expected frame-rate ticker")
	}
	<-oldTicker.C
	host.SetFrameRateLimit(100)
	if host.frameRateLimit == nil {
		t.Fatal("expected replacement frame-rate ticker")
	}
	if host.frameRateLimit == oldTicker {
		t.Fatal("expected frame-rate ticker to be replaced")
	}
	defer host.SetFrameRateLimit(0)
	select {
	case <-oldTicker.C:
	default:
	}
	select {
	case <-oldTicker.C:
		t.Fatal("old frame-rate ticker still produced ticks after replacement")
	case <-time.After(20 * time.Millisecond):
	}
}

func TestSetFrameRateLimitZeroClearsTicker(t *testing.T) {
	host := &Host{}
	host.SetFrameRateLimit(60)
	if host.frameRateLimit == nil {
		t.Fatal("expected frame-rate ticker")
	}
	host.SetFrameRateLimit(0)
	if host.frameRateLimit != nil {
		t.Fatal("expected frame-rate ticker to be cleared")
	}
}
