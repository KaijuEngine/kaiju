package common_workspace

import "kaiju/platform/hid"

type HotKey struct {
	Keys  []hid.KeyboardKey
	Ctrl  bool
	Shift bool
	Alt   bool
	Call  func()
}
