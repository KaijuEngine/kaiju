/******************************************************************************/
/* css_text_transform.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
)

func normalizedTextNodeData(data string) string {
	data = strings.TrimSpace(data)
	data = strings.ReplaceAll(data, "\r", "")
	data = strings.ReplaceAll(data, "\n", " ")
	data = strings.ReplaceAll(data, "\t", " ")
	return klib.ReplaceStringRecursive(data, "  ", " ")
}

func capitalizeText(text string) string {
	out := []rune(text)
	nextWord := true
	for i, r := range out {
		if unicode.IsLetter(r) {
			if nextWord {
				out[i] = unicode.ToTitle(r)
			}
			nextWord = false
		} else {
			nextWord = !unicode.IsDigit(r)
		}
	}
	return string(out)
}

func fullWidthText(text string) string {
	out := []rune(text)
	for i, r := range out {
		if r == ' ' {
			out[i] = '\u3000'
		} else if r >= '!' && r <= '~' {
			out[i] = r + 0xFEE0
		}
	}
	return string(out)
}

func transformText(text, transform string) (string, error) {
	switch transform {
	case "none", "initial", "unset", "revert":
		return text, nil
	case "capitalize":
		return capitalizeText(text), nil
	case "uppercase":
		return strings.ToUpper(text), nil
	case "lowercase":
		return strings.ToLower(text), nil
	case "full-width":
		return fullWidthText(text), nil
	case "full-size-kana", "math-auto":
		return "", errors.New("TextTransform does not currently support " + transform)
	default:
		return "", errors.New("TextTransform received unexpected value " + transform)
	}
}

func setChildrenTextTransform(elm *document.Element, transform string) error {
	if elm.IsText() {
		text, err := transformText(normalizedTextNodeData(elm.Data), transform)
		if err != nil {
			return err
		}
		elm.UI.ToLabel().SetText(text)
		return nil
	}
	for _, child := range elm.Children {
		if err := setChildrenTextTransform(child, transform); err != nil {
			return err
		}
	}
	return nil
}

// none|capitalize|uppercase|lowercase|full-width|initial|inherit|unset
func (p TextTransform) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("TextTransform requires exactly 1 value but got %d", len(values))
	}
	if values[0].Str == "inherit" {
		return nil
	}
	return setChildrenTextTransform(elm, values[0].Str)
}
