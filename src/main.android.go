//go:build android

/******************************************************************************/
/* main.android.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package main

import (
	"C"
	"unsafe"

	"kaijuengine.com/engine/systems/logging"
)

//export AndroidMain
func AndroidMain(platformState unsafe.Pointer) {
	logging.ExtPlatformLogInfo("Launching Kaiju Engine!")
	_main(platformState)
}

func main() {}
