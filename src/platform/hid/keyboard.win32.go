//go:build windows

/******************************************************************************/
/* keyboard.win32.go                                                          */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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
	case 0xA4:
		return KeyboardKeyLeftAlt
	case 0xA5:
		return KeyboardKeyRightAlt
	case 0xA2:
		return KeyboardKeyLeftCtrl
	case 0xA3:
		return KeyboardKeyRightCtrl
	case 0xA0:
		return KeyboardKeyLeftShift
	case 0xA1:
		return KeyboardKeyRightShift
	case 0x41:
		return KeyboardKeyA
	case 0x42:
		return KeyboardKeyB
	case 0x43:
		return KeyboardKeyC
	case 0x44:
		return KeyboardKeyD
	case 0x45:
		return KeyboardKeyE
	case 0x46:
		return KeyboardKeyF
	case 0x47:
		return KeyboardKeyG
	case 0x48:
		return KeyboardKeyH
	case 0x49:
		return KeyboardKeyI
	case 0x4A:
		return KeyboardKeyJ
	case 0x4B:
		return KeyboardKeyK
	case 0x4C:
		return KeyboardKeyL
	case 0x4D:
		return KeyboardKeyM
	case 0x4E:
		return KeyboardKeyN
	case 0x4F:
		return KeyboardKeyO
	case 0x50:
		return KeyboardKeyP
	case 0x51:
		return KeyboardKeyQ
	case 0x52:
		return KeyboardKeyR
	case 0x53:
		return KeyboardKeyS
	case 0x54:
		return KeyboardKeyT
	case 0x55:
		return KeyboardKeyU
	case 0x56:
		return KeyboardKeyV
	case 0x57:
		return KeyboardKeyW
	case 0x58:
		return KeyboardKeyX
	case 0x59:
		return KeyboardKeyY
	case 0x5A:
		return KeyboardKeyZ
	case 0x25:
		return KeyboardKeyLeft
	case 0x26:
		return KeyboardKeyUp
	case 0x27:
		return KeyboardKeyRight
	case 0x28:
		return KeyboardKeyDown
	case 0x1B:
		return KeyboardKeyEscape
	case 0x09:
		return KeyboardKeyTab
	case 0x20:
		return KeyboardKeySpace
	case 0x08:
		return KeyboardKeyBackspace
	case 0xC0:
		return KeyboardKeyBackQuote
	case 0x2E:
		return KeyboardKeyDelete
	case 0x0D:
		return KeyboardKeyEnter
	case 0xBC:
		return KeyboardKeyComma
	case 0xBE:
		return KeyboardKeyPeriod
	case 0xDC:
		return KeyboardKeyBackSlash
	case 0xBF:
		return KeyboardKeyForwardSlash
	case 0xDB:
		return KeyboardKeyOpenBracket
	case 0xDD:
		return KeyboardKeyCloseBracket
	case 0xBA:
		return KeyboardKeySemicolon
	case 0xDE:
		return KeyboardKeyQuote
	case 0xBB:
		return KeyboardKeyEqual
	case 0xBD:
		return KeyboardKeyMinus
	case 0x30:
		return KeyboardKey0
	case 0x31:
		return KeyboardKey1
	case 0x32:
		return KeyboardKey2
	case 0x33:
		return KeyboardKey3
	case 0x34:
		return KeyboardKey4
	case 0x35:
		return KeyboardKey5
	case 0x36:
		return KeyboardKey6
	case 0x37:
		return KeyboardKey7
	case 0x38:
		return KeyboardKey8
	case 0x39:
		return KeyboardKey9
	case 0x60:
		return KeyboardNumKey0
	case 0x61:
		return KeyboardNumKey1
	case 0x62:
		return KeyboardNumKey2
	case 0x63:
		return KeyboardNumKey3
	case 0x64:
		return KeyboardNumKey4
	case 0x65:
		return KeyboardNumKey5
	case 0x66:
		return KeyboardNumKey6
	case 0x67:
		return KeyboardNumKey7
	case 0x68:
		return KeyboardNumKey8
	case 0x69:
		return KeyboardNumKey9
	case 0x6A:
		return KeyboardNumKeyMultiply
	case 0x6B:
		return KeyboardNumKeyAdd
	case 0x6D:
		return KeyboardNumKeySubtract
	case 0x6E:
		return KeyboardNumKeyPeriod
	case 0x6F:
		return KeyboardNumKeyDivide
	case 0x70:
		return KeyboardKeyF1
	case 0x71:
		return KeyboardKeyF2
	case 0x72:
		return KeyboardKeyF3
	case 0x73:
		return KeyboardKeyF4
	case 0x74:
		return KeyboardKeyF5
	case 0x75:
		return KeyboardKeyF6
	case 0x76:
		return KeyboardKeyF7
	case 0x77:
		return KeyboardKeyF8
	case 0x78:
		return KeyboardKeyF9
	case 0x79:
		return KeyboardKeyF10
	case 0x7A:
		return KeyboardKeyF11
	case 0x7B:
		return KeyboardKeyF12
	case 0x14:
		return KeyboardKeyCapsLock
	case 0x91:
		return KeyboardKeyScrollLock
	case 0x90:
		return KeyboardKeyNumLock
	case 0x2C:
		return KeyboardKeyPrintScreen
	case 0x13:
		return KeyboardKeyPause
	case 0x2D:
		return KeyboardKeyInsert
	case 0x24:
		return KeyboardKeyHome
	case 0x21:
		return KeyboardKeyPageUp
	case 0x22:
		return KeyboardKeyPageDown
	case 0x23:
		return KeyboardKeyEnd
	default:
		return KeyBoardKeyInvalid
	}
}
