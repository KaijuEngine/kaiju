//go:build !windows

/******************************************************************************/
/* render_thread_id_other.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine

func currentOSThreadID() uint64 { return 0 }
