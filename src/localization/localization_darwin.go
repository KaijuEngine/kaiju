/******************************************************************************/
/* localization_darwin.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

//go:build !ios

package localization

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#include <Foundation/Foundation.h>
#include <stdlib.h>
#include <string.h>

static char* kaiju_current_localization(void) {
	@autoreleasepool {
		NSString *locale = nil;
		NSArray *languages = [NSLocale preferredLanguages];
		if ([languages count] > 0) {
			locale = [languages objectAtIndex:0];
		}
		if (locale == nil || [locale length] == 0) {
			locale = [[NSLocale currentLocale] localeIdentifier];
		}
		if (locale == nil) {
			return NULL;
		}
		const char *utf8 = [locale UTF8String];
		if (utf8 == NULL) {
			return NULL;
		}
		return strdup(utf8);
	}
}
*/
import "C"

import "unsafe"

func currentLocalization() string {
	loc := C.kaiju_current_localization()
	if loc == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(loc))
	return C.GoString(loc)
}
