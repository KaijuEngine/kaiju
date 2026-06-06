//go:build !windows

/******************************************************************************/
/* high_resolution_timer.std.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package chrono

import "time"

type HighResolutionTimer struct {
	begin time.Time
}

func (t *HighResolutionTimer) start() {
	t.begin = time.Now()
}

func (t *HighResolutionTimer) stop() (seconds float64) {
	return time.Since(t.begin).Seconds()
}
