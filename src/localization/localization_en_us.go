package localization

import (
	"strings"

	"kaijuengine.com/platform/hid"
)

type English struct{}

func (English) KeyToRune(keyboard *hid.Keyboard, key hid.KeyboardKey) rune {
	c := ""
	isNumpadKey := false
	switch key {
	case hid.KeyboardKeyA:
		c = "a"
	case hid.KeyboardKeyB:
		c = "b"
	case hid.KeyboardKeyC:
		c = "c"
	case hid.KeyboardKeyD:
		c = "d"
	case hid.KeyboardKeyE:
		c = "e"
	case hid.KeyboardKeyF:
		c = "f"
	case hid.KeyboardKeyG:
		c = "g"
	case hid.KeyboardKeyH:
		c = "h"
	case hid.KeyboardKeyI:
		c = "i"
	case hid.KeyboardKeyJ:
		c = "j"
	case hid.KeyboardKeyK:
		c = "k"
	case hid.KeyboardKeyL:
		c = "l"
	case hid.KeyboardKeyM:
		c = "m"
	case hid.KeyboardKeyN:
		c = "n"
	case hid.KeyboardKeyO:
		c = "o"
	case hid.KeyboardKeyP:
		c = "p"
	case hid.KeyboardKeyQ:
		c = "q"
	case hid.KeyboardKeyR:
		c = "r"
	case hid.KeyboardKeyS:
		c = "s"
	case hid.KeyboardKeyT:
		c = "t"
	case hid.KeyboardKeyU:
		c = "u"
	case hid.KeyboardKeyV:
		c = "v"
	case hid.KeyboardKeyW:
		c = "w"
	case hid.KeyboardKeyX:
		c = "x"
	case hid.KeyboardKeyY:
		c = "y"
	case hid.KeyboardKeyZ:
		c = "z"
	//case KEYBOARD_KEY_TAB:			c = "\t"; break;
	case hid.KeyboardKeySpace:
		c = " "
	case hid.KeyboardKeyBackQuote:
		c = "`"
	case hid.KeyboardKeyComma:
		c = ","
	case hid.KeyboardKeyPeriod:
		c = "."
	case hid.KeyboardKeyBackSlash:
		c = "\\"
	case hid.KeyboardKeyForwardSlash:
		c = "/"
	case hid.KeyboardKeyOpenBracket:
		c = "["
	case hid.KeyboardKeyCloseBracket:
		c = "]"
	case hid.KeyboardKeySemicolon:
		c = ";"
	case hid.KeyboardKeyQuote:
		c = "'"
	case hid.KeyboardKeyEqual:
		c = "="
	case hid.KeyboardKeyMinus:
		c = "-"
	case hid.KeyboardKey0:
		c = "0"
	case hid.KeyboardKey1:
		c = "1"
	case hid.KeyboardKey2:
		c = "2"
	case hid.KeyboardKey3:
		c = "3"
	case hid.KeyboardKey4:
		c = "4"
	case hid.KeyboardKey5:
		c = "5"
	case hid.KeyboardKey6:
		c = "6"
	case hid.KeyboardKey7:
		c = "7"
	case hid.KeyboardKey8:
		c = "8"
	case hid.KeyboardKey9:
		c = "9"
	case hid.KeyboardNumKey0:
		isNumpadKey = true
		c = "0"
	case hid.KeyboardNumKey1:
		isNumpadKey = true
		c = "1"
	case hid.KeyboardNumKey2:
		isNumpadKey = true
		c = "2"
	case hid.KeyboardNumKey3:
		isNumpadKey = true
		c = "3"
	case hid.KeyboardNumKey4:
		isNumpadKey = true
		c = "4"
	case hid.KeyboardNumKey5:
		isNumpadKey = true
		c = "5"
	case hid.KeyboardNumKey6:
		isNumpadKey = true
		c = "6"
	case hid.KeyboardNumKey7:
		isNumpadKey = true
		c = "7"
	case hid.KeyboardNumKey8:
		isNumpadKey = true
		c = "8"
	case hid.KeyboardNumKey9:
		isNumpadKey = true
		c = "9"
	case hid.KeyboardNumKeyMultiply:
		isNumpadKey = true
		c = "*"
	case hid.KeyboardNumKeyAdd:
		isNumpadKey = true
		c = "+"
	case hid.KeyboardNumKeySubtract:
		isNumpadKey = true
		c = "-"
	case hid.KeyboardNumKeyPeriod:
		isNumpadKey = true
		c = "."
	case hid.KeyboardNumKeyDivide:
		isNumpadKey = true
		c = "/"
	default:
		c = ""
	}
	f := '\000'
	if c != "" {
		f = rune(c[0])
		if keyboard.HasShift() && !isNumpadKey {
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
