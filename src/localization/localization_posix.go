//go:build linux || android

/******************************************************************************/
/* localization_posix.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package localization

import "os"

func currentLocalization() string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if loc := os.Getenv(key); loc != "" {
			return loc
		}
	}
	return ""
}
