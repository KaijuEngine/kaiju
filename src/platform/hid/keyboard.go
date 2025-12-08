/******************************************************************************/
/* keyboard.go                                                                */
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

import "strings"

type KeyState = uint8
type KeyboardKey = int
type KeyCallbackId int

const (
	KeyStateIdle KeyState = iota
	KeyStateDown
	KeyStateHeld
	KeyStateUp
	KeyStatePressedAndReleased
)

const (
	KeyBoardKeyInvalid KeyboardKey = -1 + iota
	KeyboardKeyLeftAlt
	KeyboardKeyRightAlt
	KeyboardKeyLeftCtrl
	KeyboardKeyRightCtrl
	KeyboardKeyLeftShift
	KeyboardKeyRightShift
	KeyboardKeyA
	KeyboardKeyB
	KeyboardKeyC
	KeyboardKeyD
	KeyboardKeyE
	KeyboardKeyF
	KeyboardKeyG
	KeyboardKeyH
	KeyboardKeyI
	KeyboardKeyJ
	KeyboardKeyK
	KeyboardKeyL
	KeyboardKeyM
	KeyboardKeyN
	KeyboardKeyO
	KeyboardKeyP
	KeyboardKeyQ
	KeyboardKeyR
	KeyboardKeyS
	KeyboardKeyT
	KeyboardKeyU
	KeyboardKeyV
	KeyboardKeyW
	KeyboardKeyX
	KeyboardKeyY
	KeyboardKeyZ
	KeyboardKeyLeft
	KeyboardKeyUp
	KeyboardKeyRight
	KeyboardKeyDown
	KeyboardKeyEscape
	KeyboardKeyTab
	KeyboardKeySpace
	KeyboardKeyBackspace
	KeyboardKeyBackQuote
	KeyboardKeyDelete
	KeyboardKeyReturn
	KeyboardKeyEnter
	KeyboardKeyComma
	KeyboardKeyPeriod
	KeyboardKeyBackSlash
	KeyboardKeyForwardSlash
	KeyboardKeyOpenBracket
	KeyboardKeyCloseBracket
	KeyboardKeySemicolon
	KeyboardKeyQuote
	KeyboardKeyEqual
	KeyboardKeyMinus
	KeyboardKey0
	KeyboardKey1
	KeyboardKey2
	KeyboardKey3
	KeyboardKey4
	KeyboardKey5
	KeyboardKey6
	KeyboardKey7
	KeyboardKey8
	KeyboardKey9
	KeyboardNumKey0
	KeyboardNumKey1
	KeyboardNumKey2
	KeyboardNumKey3
	KeyboardNumKey4
	KeyboardNumKey5
	KeyboardNumKey6
	KeyboardNumKey7
	KeyboardNumKey8
	KeyboardNumKey9
	KeyboardNumKeyDivide
	KeyboardNumKeyMultiply
	KeyboardNumKeyAdd
	KeyboardNumKeySubtract
	KeyboardNumKeyPeriod
	KeyboardKeyF1
	KeyboardKeyF2
	KeyboardKeyF3
	KeyboardKeyF4
	KeyboardKeyF5
	KeyboardKeyF6
	KeyboardKeyF7
	KeyboardKeyF8
	KeyboardKeyF9
	KeyboardKeyF10
	KeyboardKeyF11
	KeyboardKeyF12
	KeyboardKeyCapsLock
	KeyboardKeyScrollLock
	KeyboardKeyNumLock
	KeyboardKeyPrintScreen
	KeyboardKeyPause
	KeyboardKeyInsert
	KeyboardKeyHome
	KeyboardKeyPageUp
	KeyboardKeyPageDown
	KeyboardKeyEnd
	KeyboardKeyMaximum
)

type keyCallback struct {
	id KeyCallbackId
	fn func(keyId int, keyState KeyState)
}

type Keyboard struct {
	keyStates      [KeyboardKeyMaximum]KeyState
	nextCallbackId KeyCallbackId
	keyCallbacks   []keyCallback
}

func NewKeyboard() Keyboard {
	return Keyboard{
		nextCallbackId: 1,
		keyCallbacks:   []keyCallback{},
	}
}

func (k *Keyboard) AddKeyCallback(cb func(keyId int, keyState KeyState)) KeyCallbackId {
	id := k.nextCallbackId
	k.keyCallbacks = append(k.keyCallbacks, keyCallback{
		id: id,
		fn: cb,
	})
	k.nextCallbackId++
	return id
}

func (k *Keyboard) RemoveKeyCallback(id KeyCallbackId) {
	for i, cb := range k.keyCallbacks {
		if cb.id == id {
			last := len(k.keyCallbacks) - 1
			k.keyCallbacks[last], k.keyCallbacks[i] = k.keyCallbacks[i], k.keyCallbacks[last]
			k.keyCallbacks = k.keyCallbacks[:last]
			return
		}
	}
}

func (k Keyboard) KeyUp(key KeyboardKey) bool {
	return k.keyStates[key] == KeyStateUp
}

func (k Keyboard) KeyDown(key KeyboardKey) bool {
	return k.keyStates[key] == KeyStateDown
}

func (k Keyboard) KeyHeld(key KeyboardKey) bool {
	return k.keyStates[key] == KeyStateDown || k.keyStates[key] == KeyStateHeld ||
		k.keyStates[key] == KeyStatePressedAndReleased
}

func (k Keyboard) HasCtrl() bool {
	return k.KeyHeld(KeyboardKeyLeftCtrl) || k.KeyHeld(KeyboardKeyRightCtrl)
}

func (k Keyboard) HasShift() bool {
	return k.KeyHeld(KeyboardKeyLeftShift) || k.KeyHeld(KeyboardKeyRightShift)
}

func (k Keyboard) HasAlt() bool {
	return k.KeyHeld(KeyboardKeyLeftAlt) || k.KeyHeld(KeyboardKeyRightAlt)
}

func (k *Keyboard) EndUpdate() {
	for i := 0; i < KeyboardKeyMaximum; i++ {
		switch k.keyStates[i] {
		case KeyStateDown:
			k.keyStates[i] = KeyStateHeld
		case KeyStateHeld:
			k.doKeyCallbacks(i, KeyStateHeld)
		case KeyStateUp:
			k.keyStates[i] = KeyStateIdle
		case KeyStatePressedAndReleased:
			k.keyStates[i] = KeyStateIdle
		}
	}
}

func (k *Keyboard) doKeyCallbacks(key int, state KeyState) {
	for i := range k.keyCallbacks {
		k.keyCallbacks[i].fn(key, state)
	}
}

func (k *Keyboard) SetKeyDownUp(key KeyboardKey) {
	if key != KeyBoardKeyInvalid {
		k.keyStates[key] = KeyStatePressedAndReleased
		k.doKeyCallbacks(key, KeyStatePressedAndReleased)
	}
}

func (k *Keyboard) SetKeyDown(key KeyboardKey) {
	if key != KeyBoardKeyInvalid && k.keyStates[key] != KeyStateHeld {
		k.keyStates[key] = KeyStateDown
		k.doKeyCallbacks(key, KeyStateDown)
	}
}

func (k *Keyboard) SetKeyUp(key KeyboardKey) {
	if key != KeyBoardKeyInvalid {
		k.keyStates[key] = KeyStateUp
		k.doKeyCallbacks(key, KeyStateUp)
	}
}

func (k *Keyboard) KeyToRune(key KeyboardKey) rune {
	c := ""
	switch key {
	case KeyboardKeyA:
		c = "a"
	case KeyboardKeyB:
		c = "b"
	case KeyboardKeyC:
		c = "c"
	case KeyboardKeyD:
		c = "d"
	case KeyboardKeyE:
		c = "e"
	case KeyboardKeyF:
		c = "f"
	case KeyboardKeyG:
		c = "g"
	case KeyboardKeyH:
		c = "h"
	case KeyboardKeyI:
		c = "i"
	case KeyboardKeyJ:
		c = "j"
	case KeyboardKeyK:
		c = "k"
	case KeyboardKeyL:
		c = "l"
	case KeyboardKeyM:
		c = "m"
	case KeyboardKeyN:
		c = "n"
	case KeyboardKeyO:
		c = "o"
	case KeyboardKeyP:
		c = "p"
	case KeyboardKeyQ:
		c = "q"
	case KeyboardKeyR:
		c = "r"
	case KeyboardKeyS:
		c = "s"
	case KeyboardKeyT:
		c = "t"
	case KeyboardKeyU:
		c = "u"
	case KeyboardKeyV:
		c = "v"
	case KeyboardKeyW:
		c = "w"
	case KeyboardKeyX:
		c = "x"
	case KeyboardKeyY:
		c = "y"
	case KeyboardKeyZ:
		c = "z"
	//case KEYBOARD_KEY_TAB:			c = "\t"; break;
	case KeyboardKeySpace:
		c = " "
	case KeyboardKeyBackQuote:
		c = "`"
	case KeyboardKeyComma:
		c = ","
	case KeyboardKeyPeriod:
		c = "."
	case KeyboardKeyBackSlash:
		c = "\\"
	case KeyboardKeyForwardSlash:
		c = "/"
	case KeyboardKeyOpenBracket:
		c = "["
	case KeyboardKeyCloseBracket:
		c = "]"
	case KeyboardKeySemicolon:
		c = ";"
	case KeyboardKeyQuote:
		c = "'"
	case KeyboardKeyEqual:
		c = "="
	case KeyboardKeyMinus:
		c = "-"
	case KeyboardKey0:
		c = "0"
	case KeyboardKey1:
		c = "1"
	case KeyboardKey2:
		c = "2"
	case KeyboardKey3:
		c = "3"
	case KeyboardKey4:
		c = "4"
	case KeyboardKey5:
		c = "5"
	case KeyboardKey6:
		c = "6"
	case KeyboardKey7:
		c = "7"
	case KeyboardKey8:
		c = "8"
	case KeyboardKey9:
		c = "9"
	case KeyboardNumKey0:
		c = "0"
	case KeyboardNumKey1:
		c = "1"
	case KeyboardNumKey2:
		c = "2"
	case KeyboardNumKey3:
		c = "3"
	case KeyboardNumKey4:
		c = "4"
	case KeyboardNumKey5:
		c = "5"
	case KeyboardNumKey6:
		c = "6"
	case KeyboardNumKey7:
		c = "7"
	case KeyboardNumKey8:
		c = "8"
	case KeyboardNumKey9:
		c = "9"
	case KeyboardNumKeyMultiply:
		c = "*"
	case KeyboardNumKeyAdd:
		c = "+"
	case KeyboardNumKeySubtract:
		c = "-"
	case KeyboardNumKeyPeriod:
		c = "."
	case KeyboardNumKeyDivide:
		c = "/"
	default:
		c = ""
	}
	f := '\000'
	if c != "" {
		f = rune(c[0])
		if k.HasShift() {
			switch f {
			case '1':
				f = '!'
			case '2':
				f = '@'
			case '3':
				f = '#'
			case '4':
				f = '$'
			case '5':
				f = '%'
			case '6':
				f = '^'
			case '7':
				f = '&'
			case '8':
				f = '*'
			case '9':
				f = '('
			case '0':
				f = ')'
			case '-':
				f = '_'
			case '=':
				f = '+'
			case '`':
				f = '~'
			case '[':
				f = '{'
			case ']':
				f = '}'
			case '\\':
				f = '|'
			case ';':
				f = ':'
			case '\'':
				f = '"'
			case ',':
				f = '<'
			case '.':
				f = '>'
			case '/':
				f = '?'
			default:
				f = []rune(strings.ToUpper(c))[0]
				if f >= 'a' && f <= 'z' {
					f -= 32
				}
			}
		}
	}
	return f
}

func (k *Keyboard) Reset() {
	for i := 0; i < KeyboardKeyMaximum; i++ {
		if k.keyStates[i] == KeyStateDown || k.keyStates[i] == KeyStateHeld {
			k.keyStates[i] = KeyStateUp
		}
	}
}
