/******************************************************************************/
/* keyboard.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package hid

import (
	"log/slog"
	"runtime"
	"strings"
	"time"
)

type KeyState = uint8
type KeyboardKey = int
type KeyCallbackId int

const (
	KeyStateIdle KeyState = iota
	KeyStateDown
	KeyStateHeld
	KeyStateUp
	KeyStatePressedAndReleased
	KeyStateToggled
)

const (
	KeyBoardKeyInvalid KeyboardKey = -1 + iota
	KeyboardKeyLeftAlt
	KeyboardKeyRightAlt
	KeyboardKeyLeftCtrl
	KeyboardKeyRightCtrl
	KeyboardKeyLeftShift
	KeyboardKeyRightShift
	KeyboardKeyLeftMeta
	KeyboardKeyRightMeta
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
	lastClicked    [KeyboardKeyMaximum]time.Time
	nextCallbackId KeyCallbackId
	keyCallbacks   []keyCallback
}

func NewKeyboard() Keyboard {
	lastClicked := [KeyboardKeyMaximum]time.Time{}
	t := time.Now()
	for i := 0; i < KeyboardKeyMaximum; i++ {
		lastClicked[i] = t
	}

	return Keyboard{
		nextCallbackId: 1,
		keyCallbacks:   []keyCallback{},
		lastClicked:    lastClicked,
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

// HasCtrlOrMeta checks if the meta key is pressed on Darwin, or the ctrl key on
// Windows and Linux.
func (k Keyboard) HasCtrlOrMeta() bool {
	if runtime.GOOS == "darwin" {
		return k.HasMeta()
	}

	return k.HasCtrl()
}

func (k Keyboard) HasShift() bool {
	return k.KeyHeld(KeyboardKeyLeftShift) || k.KeyHeld(KeyboardKeyRightShift)
}

func (k Keyboard) HasAlt() bool {
	return k.KeyHeld(KeyboardKeyLeftAlt) || k.KeyHeld(KeyboardKeyRightAlt)
}

func (k Keyboard) HasMeta() bool {
	return k.KeyHeld(KeyboardKeyLeftMeta) || k.KeyHeld(KeyboardKeyRightMeta)
}

func (k Keyboard) HasModifier() bool {
	return k.HasCtrl() || k.HasMeta() || k.HasShift() || k.HasAlt()
}

func (k Keyboard) IsToggleKey(key KeyboardKey) bool {
	return key == KeyboardKeyCapsLock ||
		key == KeyboardKeyNumLock ||
		key == KeyboardKeyScrollLock
}

func (k Keyboard) IsToggleKeyOn(key KeyboardKey) bool {
	if k.IsToggleKey(key) {
		return k.keyStates[key] == KeyStateToggled
	}
	return false
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
		k.lastClicked[key] = time.Now()
	}
}

func (k *Keyboard) SetKeyUp(key KeyboardKey) {
	if key != KeyBoardKeyInvalid {
		k.keyStates[key] = KeyStateUp
		k.doKeyCallbacks(key, KeyStateUp)
	}
}

func (k *Keyboard) ToggleKey(key KeyboardKey) {
	if k.IsToggleKey(key) {
		isOn := k.IsToggleKeyOn(key)
		if isOn {
			k.keyStates[key] = KeyStateIdle
		} else {
			k.keyStates[key] = KeyStateToggled
		}
	}
}

func (k *Keyboard) SetToggleKeyState(key KeyboardKey, state KeyState) {
	if key != KeyBoardKeyInvalid && k.IsToggleKey(key) {
		k.keyStates[key] = state
	}
}

func (k *Keyboard) KeyToRune(key KeyboardKey) rune {
	c := ""
	isNumpadKey := false
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
		isNumpadKey = true
		c = "0"
	case KeyboardNumKey1:
		isNumpadKey = true
		c = "1"
	case KeyboardNumKey2:
		isNumpadKey = true
		c = "2"
	case KeyboardNumKey3:
		isNumpadKey = true
		c = "3"
	case KeyboardNumKey4:
		isNumpadKey = true
		c = "4"
	case KeyboardNumKey5:
		isNumpadKey = true
		c = "5"
	case KeyboardNumKey6:
		isNumpadKey = true
		c = "6"
	case KeyboardNumKey7:
		isNumpadKey = true
		c = "7"
	case KeyboardNumKey8:
		isNumpadKey = true
		c = "8"
	case KeyboardNumKey9:
		isNumpadKey = true
		c = "9"
	case KeyboardNumKeyMultiply:
		isNumpadKey = true
		c = "*"
	case KeyboardNumKeyAdd:
		isNumpadKey = true
		c = "+"
	case KeyboardNumKeySubtract:
		isNumpadKey = true
		c = "-"
	case KeyboardNumKeyPeriod:
		isNumpadKey = true
		c = "."
	case KeyboardNumKeyDivide:
		isNumpadKey = true
		c = "/"
	default:
		c = ""
	}
	f := '\000'
	if c != "" {
		f = rune(c[0])
		if k.HasShift() && !isNumpadKey {
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

func (k *Keyboard) GetKeyLastClicked(keyId int) time.Time {
	if keyId < 0 || keyId > KeyboardKeyMaximum {
		slog.Error("inside keyboard::GetKeyLastClicked", "keyId", keyId)
		return time.Now()
	}

	return k.lastClicked[keyId]
}
