/******************************************************************************/
/* layout_stylizer.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

type LayoutStylizer interface {
	ProcessStyle(layout *Layout) []error
}
