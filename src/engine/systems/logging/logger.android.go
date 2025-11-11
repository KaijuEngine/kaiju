//go:build android

package logging

/*
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
