//go:build !windows && !linux && !darwin && !android

/******************************************************************************/
/* query_unsupported.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package power

func Query() (Status, error) {
	return Status{Source: SourceUnknown, BatteryPercent: -1}, nil
}
