/******************************************************************************/
/* css_text_transform.go                                                      */
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
