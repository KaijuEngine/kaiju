//go:build linux && !android

/******************************************************************************/
/* query_linux.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package power

func Query() (Status, error) {
	return queryLinuxPowerSupply("/sys/class/power_supply")
}
