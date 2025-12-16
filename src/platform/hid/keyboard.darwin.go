//go:build darwin && !ios

/******************************************************************************/
/* keyboard.darwin.go                                                         */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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

// ToKeyboardKey converts macOS virtual key codes to engine KeyboardKey values.
// macOS key codes are defined in HIToolbox/Events.h (kVK_* constants).
func ToKeyboardKey(nativeKey int) KeyboardKey {
	switch nativeKey {
	// Letters (QWERTY layout)
	case 0x00:
		return KeyboardKeyA
	case 0x0B:
		return KeyboardKeyB
	case 0x08:
		return KeyboardKeyC
	case 0x02:
		return KeyboardKeyD
	case 0x0E:
		return KeyboardKeyE
	case 0x03:
		return KeyboardKeyF
	case 0x05:
		return KeyboardKeyG
	case 0x04:
		return KeyboardKeyH
	case 0x22:
		return KeyboardKeyI
	case 0x26:
		return KeyboardKeyJ
	case 0x28:
		return KeyboardKeyK
	case 0x25:
		return KeyboardKeyL
	case 0x2E:
		return KeyboardKeyM
	case 0x2D:
		return KeyboardKeyN
	case 0x1F:
		return KeyboardKeyO
	case 0x23:
		return KeyboardKeyP
	case 0x0C:
		return KeyboardKeyQ
	case 0x0F:
		return KeyboardKeyR
	case 0x01:
		return KeyboardKeyS
	case 0x11:
		return KeyboardKeyT
	case 0x20:
		return KeyboardKeyU
	case 0x09:
		return KeyboardKeyV
	case 0x0D:
		return KeyboardKeyW
	case 0x07:
		return KeyboardKeyX
	case 0x10:
		return KeyboardKeyY
	case 0x06:
		return KeyboardKeyZ

	// Numbers (top row)
	case 0x1D:
		return KeyboardKey0
	case 0x12:
		return KeyboardKey1
	case 0x13:
		return KeyboardKey2
	case 0x14:
		return KeyboardKey3
	case 0x15:
		return KeyboardKey4
	case 0x17:
		return KeyboardKey5
	case 0x16:
		return KeyboardKey6
	case 0x1A:
		return KeyboardKey7
	case 0x1C:
		return KeyboardKey8
	case 0x19:
		return KeyboardKey9

	// Keypad
	case 0x52:
		return KeyboardNumKey0
	case 0x53:
		return KeyboardNumKey1
	case 0x54:
		return KeyboardNumKey2
	case 0x55:
		return KeyboardNumKey3
	case 0x56:
		return KeyboardNumKey4
	case 0x57:
		return KeyboardNumKey5
	case 0x58:
		return KeyboardNumKey6
	case 0x59:
		return KeyboardNumKey7
	case 0x5B:
		return KeyboardNumKey8
	case 0x5C:
		return KeyboardNumKey9
	case 0x43:
		return KeyboardNumKeyMultiply
	case 0x45:
		return KeyboardNumKeyAdd
	case 0x4E:
		return KeyboardNumKeySubtract
	case 0x41:
		return KeyboardNumKeyPeriod
	case 0x4B:
		return KeyboardNumKeyDivide

	// Function keys
	case 0x7A:
		return KeyboardKeyF1
	case 0x78:
		return KeyboardKeyF2
	case 0x63:
		return KeyboardKeyF3
	case 0x76:
		return KeyboardKeyF4
	case 0x60:
		return KeyboardKeyF5
	case 0x61:
		return KeyboardKeyF6
	case 0x62:
		return KeyboardKeyF7
	case 0x64:
		return KeyboardKeyF8
	case 0x65:
		return KeyboardKeyF9
	case 0x6D:
		return KeyboardKeyF10
	case 0x67:
		return KeyboardKeyF11
	case 0x6F:
		return KeyboardKeyF12

	// Arrow keys
	case 0x7B:
		return KeyboardKeyLeft
	case 0x7C:
		return KeyboardKeyRight
	case 0x7D:
		return KeyboardKeyDown
	case 0x7E:
		return KeyboardKeyUp

	// Control keys
	case 0x35:
		return KeyboardKeyEscape
	case 0x30:
		return KeyboardKeyTab
	case 0x31:
		return KeyboardKeySpace
	case 0x24:
		return KeyboardKeyReturn
	case 0x4C:
		return KeyboardKeyEnter // Keypad enter
	case 0x33:
		return KeyboardKeyBackspace
	case 0x75:
		return KeyboardKeyDelete
	case 0x73:
		return KeyboardKeyHome
	case 0x77:
		return KeyboardKeyEnd
	case 0x74:
		return KeyboardKeyPageUp
	case 0x79:
		return KeyboardKeyPageDown

	// Modifiers
	case 0x38:
		return KeyboardKeyLeftShift
	case 0x3C:
		return KeyboardKeyRightShift
	case 0x3B:
		return KeyboardKeyLeftCtrl
	case 0x3E:
		return KeyboardKeyRightCtrl
	case 0x3A:
		return KeyboardKeyLeftAlt
	case 0x3D:
		return KeyboardKeyRightAlt
	// macOS Command keys map to Ctrl for shortcuts (Cmd+C, Cmd+V, etc.)
	// This allows engine's Ctrl+C/V/X clipboard shortcuts to work with Cmd key
	case 0x37:
		return KeyboardKeyLeftCtrl // Left Command → Left Ctrl
	case 0x36:
		return KeyboardKeyRightCtrl // Right Command → Right Ctrl

	// Special keys
	case 0x39:
		return KeyboardKeyCapsLock
	case 0x47:
		return KeyboardKeyNumLock
	case 0x6B:
		return KeyboardKeyScrollLock
	case 0x71:
		return KeyboardKeyPrintScreen
	case 0x72:
		return KeyboardKeyInsert
	case 0x7F:
		return KeyboardKeyPause

	// Punctuation
	case 0x32:
		return KeyboardKeyBackQuote
	case 0x2B:
		return KeyboardKeyComma
	case 0x2F:
		return KeyboardKeyPeriod
	case 0x2C:
		return KeyboardKeyForwardSlash
	case 0x2A:
		return KeyboardKeyBackSlash
	case 0x21:
		return KeyboardKeyOpenBracket
	case 0x1E:
		return KeyboardKeyCloseBracket
	case 0x29:
		return KeyboardKeySemicolon
	case 0x27:
		return KeyboardKeyQuote
	case 0x18:
		return KeyboardKeyEqual
	case 0x1B:
		return KeyboardKeyMinus

	default:
		return KeyBoardKeyInvalid
	}
}
