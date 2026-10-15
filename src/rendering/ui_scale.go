/******************************************************************************/
/* ui_scale.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"os"
	"strconv"
)

// UIScale multiplies UI pixel lengths and default font sizes. It is read once
// from the KAIJU_UI_SCALE environment variable (e.g. KAIJU_UI_SCALE=2) as a
// workaround for HiDPI displays where the system scale is not applied.
var UIScale = uiScaleFromEnv()

func uiScaleFromEnv() float32 {
	raw := os.Getenv("KAIJU_UI_SCALE")
	if raw == "" {
		return 1
	}
	scale, err := strconv.ParseFloat(raw, 32)
	if err != nil || scale <= 0 {
		return 1
	}
	return float32(scale)
}
