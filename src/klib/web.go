/******************************************************************************/
/* web.go                                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"os/exec"
	"runtime"
)

var OpenWebsiteAndroidFunc func(url string)

func OpenWebsite(url string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url).Run()
	case "android":
		OpenWebsiteAndroidFunc(url)
	default:
		exec.Command("open", url).Run()
	}
}
