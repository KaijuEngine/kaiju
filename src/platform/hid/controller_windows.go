/******************************************************************************/
/* controller_windows.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package hid

func ToControllerButton(nativeButton int) (ControllerButton, error) {
	return ControllerButton(nativeButton), nil
}
