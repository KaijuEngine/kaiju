//go:build windows

/******************************************************************************/
/* high_resolution_timer.win.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package chrono

/*
#cgo noescape get_freq
#cgo nocallback get_freq
#cgo noescape get_counter
#cgo nocallback get_counter
#include <stdint.h>
#include <windows.h>

int64_t get_freq() {
	LARGE_INTEGER freq;
	QueryPerformanceFrequency(&freq);
	return freq.QuadPart;
}

int64_t get_counter() {
	LARGE_INTEGER counter;
	QueryPerformanceCounter(&counter);
	return counter.QuadPart;
}
*/
import "C"

type HighResolutionTimer struct {
	freq  int64
	begin int64
}

func (t *HighResolutionTimer) start() {
	t.freq = int64(C.get_freq())
	t.begin = int64(C.get_counter())
}

func (t *HighResolutionTimer) stop() (seconds float64) {
	end := int64(C.get_counter())
	elapsedTicks := end - t.begin
	return float64(elapsedTicks) / float64(t.freq)
}
