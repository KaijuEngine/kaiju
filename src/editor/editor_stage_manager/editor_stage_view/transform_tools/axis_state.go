/******************************************************************************/
/* axis_state.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package transform_tools

type AxisState uint8

const (
	AxisStateNone AxisState = iota
	AxisStateX
	AxisStateY
	AxisStateZ
)

func (a *AxisState) Toggle(axis AxisState) {
	if *a == axis {
		*a = AxisStateNone
	} else {
		*a = axis
	}
}
