//go:build android

/******************************************************************************/
/* keyboard.android.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package hid

func ToKeyboardKey(nativeKey int) KeyboardKey {
	return KeyBoardKeyInvalid
}
