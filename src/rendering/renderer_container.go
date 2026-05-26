/******************************************************************************/
/* renderer_container.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "unsafe"

type RenderingContainer interface {
	GetDrawableSize() (int32, int32)
	GetInstanceExtensions() []string
	PlatformWindow() unsafe.Pointer
	PlatformInstance() unsafe.Pointer
}
