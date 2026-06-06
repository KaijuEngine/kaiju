//go:build windows

/******************************************************************************/
/* render_thread_id_windows.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine

import "syscall"

var kernel32GetCurrentThreadId = syscall.NewLazyDLL("kernel32.dll").NewProc("GetCurrentThreadId")

func currentOSThreadID() uint64 {
	id, _, _ := kernel32GetCurrentThreadId.Call()
	return uint64(id)
}
