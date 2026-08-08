//go:build darwin

/******************************************************************************/
/* main.darwin.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package main

import (
	"runtime"

	"kaijuengine.com/platform/windowing"
)

func main() {
	runtime.LockOSThread()
	go func() {
		_main(nil)
		// Teardown (including Vulkan destroy) is done; leave [NSApp run].
		windowing.CocoaStopApp()
	}()
	windowing.CocoaRunApp()
}
