/******************************************************************************/
/* css_not.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
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
