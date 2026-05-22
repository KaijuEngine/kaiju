/******************************************************************************/
/* common_workspace_hotkey.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package common_workspace

import "kaijuengine.com/platform/hid"

type HotKey struct {
	Keys  []hid.KeyboardKey
	Ctrl  bool
	Meta  bool
	Shift bool
	Alt   bool
	Call  func()
}
