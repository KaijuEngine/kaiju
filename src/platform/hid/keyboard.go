/******************************************************************************/
/* keyboard.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package hid

import (
	"log/slog"
	"runtime"
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
