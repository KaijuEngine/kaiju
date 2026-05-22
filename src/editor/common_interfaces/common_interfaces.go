/******************************************************************************/
/* common_interfaces.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package common_interfaces

type Focusable interface {
	FocusInterface()
	BlurInterface()
}
