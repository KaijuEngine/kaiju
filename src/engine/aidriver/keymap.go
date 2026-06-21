/******************************************************************************/
/* keymap.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package aidriver

import (
	"strings"

	"kaijuengine.com/platform/hid"
)

// keyNames maps friendly, case-insensitive key names used by the HTTP API onto
// the engine's hid.KeyboardKey enum. It is the single source of truth for the
// "key" field of key_down / key_up / key_press actions.
var keyNames = buildKeyNames()

// buttonNames maps friendly mouse button names onto the engine's mouse button
// index constants.
var buttonNames = map[string]int{
	"left":   hid.MouseButtonLeft,
	"middle": hid.MouseButtonMiddle,
	"right":  hid.MouseButtonRight,
	"x1":     hid.MouseButtonX1,
	"x2":     hid.MouseButtonX2,
}

// lookupKey resolves a friendly key name to its engine key. The match is
// case-insensitive and tolerant of surrounding whitespace.
func lookupKey(name string) (hid.KeyboardKey, bool) {
	key, ok := keyNames[strings.ToLower(strings.TrimSpace(name))]
	return key, ok
}

// lookupButton resolves a friendly mouse button name to its engine index.
func lookupButton(name string) (int, bool) {
	idx, ok := buttonNames[strings.ToLower(strings.TrimSpace(name))]
	return idx, ok
}

// shiftedRune describes how an ASCII rune is produced on a US-QWERTY layout: the
// engine key to press and whether the shift modifier must be held.
type shiftedRune struct {
	key   hid.KeyboardKey
	shift bool
}

// runeToKey maps an ASCII rune to the key + shift needed to type it on a
// US-QWERTY keyboard. It returns ok=false for anything outside that table
// (callers surface those as warnings and skip them).
func runeToKey(r rune) (hid.KeyboardKey, bool, bool) {
	if sr, ok := runeMap[r]; ok {
		return sr.key, sr.shift, true
	}
	return hid.KeyBoardKeyInvalid, false, false
}

func buildKeyNames() map[string]hid.KeyboardKey {
	m := map[string]hid.KeyboardKey{
		"escape":       hid.KeyboardKeyEscape,
		"esc":          hid.KeyboardKeyEscape,
		"return":       hid.KeyboardKeyReturn,
		"enter":        hid.KeyboardKeyReturn,
		"numenter":     hid.KeyboardKeyEnter,
		"space":        hid.KeyboardKeySpace,
		"spacebar":     hid.KeyboardKeySpace,
		"tab":          hid.KeyboardKeyTab,
		"backspace":    hid.KeyboardKeyBackspace,
		"delete":       hid.KeyboardKeyDelete,
		"del":          hid.KeyboardKeyDelete,
		"insert":       hid.KeyboardKeyInsert,
		"home":         hid.KeyboardKeyHome,
		"end":          hid.KeyboardKeyEnd,
		"pageup":       hid.KeyboardKeyPageUp,
		"pagedown":     hid.KeyboardKeyPageDown,
		"left":         hid.KeyboardKeyLeft,
		"right":        hid.KeyboardKeyRight,
		"up":           hid.KeyboardKeyUp,
		"down":         hid.KeyboardKeyDown,
		"capslock":     hid.KeyboardKeyCapsLock,
		"numlock":      hid.KeyboardKeyNumLock,
		"scrolllock":   hid.KeyboardKeyScrollLock,
		"printscreen":  hid.KeyboardKeyPrintScreen,
		"pause":        hid.KeyboardKeyPause,
		"shift":        hid.KeyboardKeyLeftShift,
		"leftshift":    hid.KeyboardKeyLeftShift,
		"rightshift":   hid.KeyboardKeyRightShift,
		"ctrl":         hid.KeyboardKeyLeftCtrl,
		"control":      hid.KeyboardKeyLeftCtrl,
		"leftctrl":     hid.KeyboardKeyLeftCtrl,
		"rightctrl":    hid.KeyboardKeyRightCtrl,
		"alt":          hid.KeyboardKeyLeftAlt,
		"leftalt":      hid.KeyboardKeyLeftAlt,
		"rightalt":     hid.KeyboardKeyRightAlt,
		"meta":         hid.KeyboardKeyLeftMeta,
		"cmd":          hid.KeyboardKeyLeftMeta,
		"command":      hid.KeyboardKeyLeftMeta,
		"super":        hid.KeyboardKeyLeftMeta,
		"win":          hid.KeyboardKeyLeftMeta,
		"leftmeta":     hid.KeyboardKeyLeftMeta,
		"rightmeta":    hid.KeyboardKeyRightMeta,
		"minus":        hid.KeyboardKeyMinus,
		"equal":        hid.KeyboardKeyEqual,
		"comma":        hid.KeyboardKeyComma,
		"period":       hid.KeyboardKeyPeriod,
		"semicolon":    hid.KeyboardKeySemicolon,
		"quote":        hid.KeyboardKeyQuote,
		"backquote":    hid.KeyboardKeyBackQuote,
		"backslash":    hid.KeyboardKeyBackSlash,
		"slash":        hid.KeyboardKeyForwardSlash,
		"forwardslash": hid.KeyboardKeyForwardSlash,
		"openbracket":  hid.KeyboardKeyOpenBracket,
		"closebracket": hid.KeyboardKeyCloseBracket,
	}
	// Letters a..z (contiguous in the enum).
	for i := range 26 {
		m[string(rune('a'+i))] = hid.KeyboardKeyA + hid.KeyboardKey(i)
	}
	// Top-row digits 0..9 (contiguous in the enum).
	for i := range 10 {
		m[string(rune('0'+i))] = hid.KeyboardKey0 + hid.KeyboardKey(i)
	}
	// Function keys f1..f12 (contiguous in the enum).
	for i := range 12 {
		m["f"+itoa(i+1)] = hid.KeyboardKeyF1 + hid.KeyboardKey(i)
	}
	return m
}

// runeMap is the US-QWERTY rune -> (key, shift) table used by type_text.
var runeMap = buildRuneMap()

func buildRuneMap() map[rune]shiftedRune {
	m := map[rune]shiftedRune{
		' ':  {hid.KeyboardKeySpace, false},
		'\n': {hid.KeyboardKeyReturn, false},
		'\r': {hid.KeyboardKeyReturn, false},
		'\t': {hid.KeyboardKeyTab, false},
		'-':  {hid.KeyboardKeyMinus, false},
		'_':  {hid.KeyboardKeyMinus, true},
		'=':  {hid.KeyboardKeyEqual, false},
		'+':  {hid.KeyboardKeyEqual, true},
		'[':  {hid.KeyboardKeyOpenBracket, false},
		'{':  {hid.KeyboardKeyOpenBracket, true},
		']':  {hid.KeyboardKeyCloseBracket, false},
		'}':  {hid.KeyboardKeyCloseBracket, true},
		'\\': {hid.KeyboardKeyBackSlash, false},
		'|':  {hid.KeyboardKeyBackSlash, true},
		';':  {hid.KeyboardKeySemicolon, false},
		':':  {hid.KeyboardKeySemicolon, true},
		'\'': {hid.KeyboardKeyQuote, false},
		'"':  {hid.KeyboardKeyQuote, true},
		',':  {hid.KeyboardKeyComma, false},
		'<':  {hid.KeyboardKeyComma, true},
		'.':  {hid.KeyboardKeyPeriod, false},
		'>':  {hid.KeyboardKeyPeriod, true},
		'/':  {hid.KeyboardKeyForwardSlash, false},
		'?':  {hid.KeyboardKeyForwardSlash, true},
		'`':  {hid.KeyboardKeyBackQuote, false},
		'~':  {hid.KeyboardKeyBackQuote, true},
	}
	// Lowercase + uppercase letters.
	for i := range 26 {
		key := hid.KeyboardKeyA + hid.KeyboardKey(i)
		m[rune('a'+i)] = shiftedRune{key, false}
		m[rune('A'+i)] = shiftedRune{key, true}
	}
	// Digits and their shifted symbols on the number row.
	shifted := []rune{')', '!', '@', '#', '$', '%', '^', '&', '*', '('}
	for i := range 10 {
		key := hid.KeyboardKey0 + hid.KeyboardKey(i)
		m[rune('0'+i)] = shiftedRune{key, false}
		m[shifted[i]] = shiftedRune{key, true}
	}
	return m
}

// itoa is a tiny helper to avoid importing strconv just for small key indices.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [4]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
