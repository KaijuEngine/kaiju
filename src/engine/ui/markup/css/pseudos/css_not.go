/******************************************************************************/
/* css_not.go                                                                 */
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

package pseudos

import (
	"errors"
	"strings"

	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p Not) Process(elm *document.Element, value rules.SelectorPart) ([]*document.Element, error) {
	if len(value.Args) == 0 {
		return []*document.Element{}, errors.New(":not requires a selector argument")
	}
	selectors := splitNotSelectors(value.Args)
	for i := range selectors {
		if notSelectorMatches(elm, selectors[i]) {
			return []*document.Element{}, nil
		}
	}
	return []*document.Element{elm}, nil
}

func splitNotSelectors(args []string) [][]string {
	selectors := make([][]string, 0, 1)
	current := make([]string, 0, len(args))
	depth := 0
	for i := range args {
		token := strings.TrimSpace(args[i])
		if token == "[" || strings.HasSuffix(token, "(") {
			depth++
		} else if (token == "]" || token == ")") && depth > 0 {
			depth--
		}
		if token == "," && depth == 0 {
			if len(current) > 0 {
				selectors = append(selectors, current)
			}
			current = make([]string, 0, len(args)-i-1)
		} else {
			current = append(current, args[i])
		}
	}
	if len(current) > 0 {
		selectors = append(selectors, current)
	}
	return selectors
}

func notSelectorMatches(elm *document.Element, args []string) bool {
	matched := false
	for i := 0; i < len(args); i++ {
		token := strings.TrimSpace(args[i])
		if token == "" {
			continue
		}
		switch token {
		case ".":
			i++
			if i >= len(args) || !elm.HasClass(strings.TrimSpace(args[i])) {
				return false
			}
			matched = true
		case "#":
			i++
			if i >= len(args) || elm.Attribute("id") != strings.TrimSpace(args[i]) {
				return false
			}
			matched = true
		case "[":
			end := i + 1
			for end < len(args) && strings.TrimSpace(args[end]) != "]" {
				end++
			}
			if end >= len(args) || !notAttributeSelectorMatches(elm, args[i+1:end]) {
				return false
			}
			i = end
			matched = true
		default:
			if strings.HasPrefix(token, "#") {
				if elm.Attribute("id") != strings.TrimPrefix(token, "#") {
					return false
				}
				matched = true
			} else if strings.HasPrefix(token, ".") {
				if !elm.HasClass(strings.TrimPrefix(token, ".")) {
					return false
				}
				matched = true
			} else if token == ":" || strings.HasSuffix(token, "(") || token == ")" {
				return false
			} else if !strings.EqualFold(elm.Data, token) {
				return false
			} else {
				matched = true
			}
		}
	}
	return matched
}

func notAttributeSelectorMatches(elm *document.Element, args []string) bool {
	key := ""
	operator := ""
	value := ""
	for i := range args {
		token := strings.TrimSpace(args[i])
		if token == "" {
			continue
		}
		if key == "" {
			key = token
		} else if operator == "" {
			operator = token
		} else if value == "" {
			value = strings.Trim(token, `"`)
		}
	}
	if key == "" {
		return false
	}
	attr := elm.Attribute(key)
	switch operator {
	case "":
		return attr != ""
	case "=":
		return attr == value
	default:
		return false
	}
}
