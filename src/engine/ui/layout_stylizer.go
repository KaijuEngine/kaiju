/******************************************************************************/
/* layout_stylizer.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

type LayoutStylizer interface {
	ProcessStyle(layout *Layout) []error
}

// PendingLayoutStylizer separates a stylesheet refresh from a layout refresh.
// Markup stylizers use this to apply changed paint properties before UI.Clean
// decides whether the tree actually needs layout stabilization.
type PendingLayoutStylizer interface {
	LayoutStylizer
	HasPendingStyle() bool
	ProcessPendingStyle(layout *Layout) []error
}
