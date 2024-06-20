//go:build linux && !android

/******************************************************************************/
/* keyboard.x11.go                                                           */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package hid

func ToKeyboardKey(nativeKey int) KeyboardKey {
	switch nativeKey {
	case 0xFFE9:
		return KeyboardKeyLeftAlt
	case 0xFFEA:
		return KeyboardKeyRightAlt
	case 0xFFE3:
		return KeyboardKeyLeftCtrl
	case 0xFFE4:
		return KeyboardKeyRightCtrl
	case 0xFFE1:
		return KeyboardKeyLeftShift
	case 0xFFE2:
		return KeyboardKeyRightShift
	case 0x0061:
		return KeyboardKeyA
	case 0x0062:
		return KeyboardKeyB
	case 0x0063:
		return KeyboardKeyC
	case 0x0064:
		return KeyboardKeyD
	case 0x0065:
		return KeyboardKeyE
	case 0x0066:
		return KeyboardKeyF
	case 0x0067:
		return KeyboardKeyG
	case 0x0068:
		return KeyboardKeyH
	case 0x0069:
		return KeyboardKeyI
	case 0x006A:
		return KeyboardKeyJ
	case 0x006B:
		return KeyboardKeyK
	case 0x006C:
		return KeyboardKeyL
	case 0x006D:
		return KeyboardKeyM
	case 0x006E:
		return KeyboardKeyN
	case 0x006F:
		return KeyboardKeyO
	case 0x0070:
		return KeyboardKeyP
	case 0x0071:
		return KeyboardKeyQ
	case 0x0072:
		return KeyboardKeyR
	case 0x0073:
		return KeyboardKeyS
	case 0x0074:
		return KeyboardKeyT
	case 0x0075:
		return KeyboardKeyU
	case 0x0076:
		return KeyboardKeyV
	case 0x0077:
		return KeyboardKeyW
	case 0x0078:
		return KeyboardKeyX
	case 0x0079:
		return KeyboardKeyY
	case 0x007A:
		return KeyboardKeyZ
	case 0x08FB:
		return KeyboardKeyLeft
	case 0x08FC:
		return KeyboardKeyUp
	case 0x08FD:
		return KeyboardKeyRight
	case 0x08FE:
		return KeyboardKeyDown
	case 0xFF1B:
		return KeyboardKeyEscape
	case 0xFF09:
		return KeyboardKeyTab
	case 0x0020:
		return KeyboardKeySpace
	case 0xFF08:
		return KeyboardKeyBackspace
	case 0x0060:
		return KeyboardKeyBackQuote
	case 0xFFFF:
		return KeyboardKeyDelete
	case 0xFF8D:
		return KeyboardKeyReturn
	case 0xFF0D:
		return KeyboardKeyEnter
	case 0x002C:
		return KeyboardKeyComma
	case 0x002E:
		return KeyboardKeyPeriod
	case 0x005C:
		return KeyboardKeyBackSlash
	case 0x002F:
		return KeyboardKeyForwardSlash
	case 0x005B:
		return KeyboardKeyOpenBracket
	case 0x005D:
		return KeyboardKeyCloseBracket
	case 0x003B:
		return KeyboardKeySemicolon
	case 0x0027:
		return KeyboardKeyQuote
	case 0x003D:
		return KeyboardKeyEqual
	case 0x002D:
		return KeyboardKeyMinus
	case 0x0030:
		return KeyboardKey0
	case 0x0031:
		return KeyboardKey1
	case 0x0032:
		return KeyboardKey2
	case 0x0033:
		return KeyboardKey3
	case 0x0034:
		return KeyboardKey4
	case 0x0035:
		return KeyboardKey5
	case 0x0036:
		return KeyboardKey6
	case 0x0037:
		return KeyboardKey7
	case 0x0038:
		return KeyboardKey8
	case 0x0039:
		return KeyboardKey9
	case 0xFFB0:
		return KeyboardNumKey0
	case 0xFFB1:
		return KeyboardNumKey1
	case 0xFFB2:
		return KeyboardNumKey2
	case 0xFFB3:
		return KeyboardNumKey3
	case 0xFFB4:
		return KeyboardNumKey4
	case 0xFFB5:
		return KeyboardNumKey5
	case 0xFFB6:
		return KeyboardNumKey6
	case 0xFFB7:
		return KeyboardNumKey7
	case 0xFFB8:
		return KeyboardNumKey8
	case 0xFFB9:
		return KeyboardNumKey9
	case 0xFFBE:
		return KeyboardKeyF1
	case 0xFFBF:
		return KeyboardKeyF2
	case 0xFFC0:
		return KeyboardKeyF3
	case 0xFFC1:
		return KeyboardKeyF4
	case 0xFFC2:
		return KeyboardKeyF5
	case 0xFFC3:
		return KeyboardKeyF6
	case 0xFFC4:
		return KeyboardKeyF7
	case 0xFFC5:
		return KeyboardKeyF8
	case 0xFFC6:
		return KeyboardKeyF9
	case 0xFFC7:
		return KeyboardKeyF10
	case 0xFFC8:
		return KeyboardKeyF11
	case 0xFFC9:
		return KeyboardKeyF12
	case 0xFFE5:
		return KeyboardKeyCapsLock
	case 0xFF14:
		return KeyboardKeyScrollLock
	case 0xFF7F:
		return KeyboardKeyNumLock
	case 0xFD1D:
		return KeyboardKeyPrintScreen
	case 0xFF13:
		return KeyboardKeyPause
	case 0xFF63:
		return KeyboardKeyInsert
	case 0xFF50:
		return KeyboardKeyHome
	case 0xFF55:
		return KeyboardKeyPageUp
	case 0xFF56:
		return KeyboardKeyPageDown
	case 0xFF57:
		return KeyboardKeyEnd
	default:
		return KeyBoardKeyInvalid
	}
}
