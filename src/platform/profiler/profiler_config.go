/******************************************************************************/
/* profiler_config.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package profiler

const (
	pprofCPUFile   = "cpu.prof"
	pprofHeapFile  = "heap.prof"
	traceFile      = "trace.out"
	pprofMergeFile = "default.pgo"
	pprofWebPort   = "9382"

	pprofCtxDataKey   = "pprofWebCtx"
	traceCtxDataKey   = "traceWebCtx"
	pprofFileKey      = "pprofFile"
	pprofWebOpenedKey = "pprofWebOpened"
)
