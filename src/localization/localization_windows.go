//go:build windows

/******************************************************************************/
/* localization_windows.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package localization

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const localeNameMaxLength = 85

var (
	kernel32                   = windows.NewLazySystemDLL("kernel32.dll")
	getUserDefaultLocaleName   = kernel32.NewProc("GetUserDefaultLocaleName")
	getSystemDefaultLocaleName = kernel32.NewProc("GetSystemDefaultLocaleName")
)

func currentLocalization() string {
	locale := make([]uint16, localeNameMaxLength)
	ret, _, _ := getUserDefaultLocaleName.Call(
		uintptr(unsafe.Pointer(&locale[0])),
		uintptr(localeNameMaxLength),
	)
	if ret == 0 {
		getSystemDefaultLocaleName.Call(
			uintptr(unsafe.Pointer(&locale[0])),
			uintptr(localeNameMaxLength),
		)
	}
	return windows.UTF16ToString(locale)
}
