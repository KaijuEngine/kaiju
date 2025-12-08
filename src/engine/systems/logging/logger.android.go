//go:build android

/******************************************************************************/
/* logger.android.go                                                          */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package logging

/*
#cgo noescape log_verbose
#cgo noescape log_info
#cgo noescape log_warn
#cgo noescape log_error
#include <stdlib.h>
#include <android/log.h>

void log_verbose(const char* message) {
	__android_log_print(ANDROID_LOG_VERBOSE, "KaijuEngineLog", message);
}

void log_info(const char* message) {
	__android_log_print(ANDROID_LOG_INFO, "KaijuEngineLog", message);
}

void log_warn(const char* message) {
	__android_log_print(ANDROID_LOG_WARN, "KaijuEngineLog", message);
}

void log_error(const char* message) {
	__android_log_print(ANDROID_LOG_ERROR, "KaijuEngineLog", message);
}
*/
import "C"
import "unsafe"

func ExtPlatformLogVerbose(message string) {
	msg := C.CString(message)
	defer C.free(unsafe.Pointer(msg))
	C.log_verbose(msg)
}

func ExtPlatformLogInfo(message string) {
	msg := C.CString(message)
	defer C.free(unsafe.Pointer(msg))
	C.log_info(msg)
}

func ExtPlatformLogWarn(message string) {
	msg := C.CString(message)
	defer C.free(unsafe.Pointer(msg))
	C.log_warn(msg)
}

func ExtPlatformLogError(message string) {
	msg := C.CString(message)
	defer C.free(unsafe.Pointer(msg))
	C.log_error(msg)
}
