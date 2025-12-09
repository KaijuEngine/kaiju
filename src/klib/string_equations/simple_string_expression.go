/******************************************************************************/
/* simple_string_expression.go                                                */
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

package string_equations

import (
	"fmt"
	"strconv"
	"strings"
)

// CalculateSimpleStringExpression will take an input string math equation,
// using only simple expressions (+, -, *, /) (and parenthesis) to generate
// a resulting value
func CalculateSimpleStringExpression(expression string) (float64, error) {
	tokens, err := tokenize(expression)
	if err != nil {
		return 0, err
	}
	if len(tokens) == 0 {
		return 0, fmt.Errorf("empty expression")
	}
	result, index, err := evaluateExpression(tokens, 0)
	if err != nil {
		return 0, err
	}
	if index != len(tokens) {
		return 0, fmt.Errorf("invalid expression, unexpected token at position %d: %v", index, tokens[index:])
	}
	return result, nil
}

func tokenize(expression string) ([]string, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	tokens := []string{}
	var currentNumber string

	for _, char := range expression {
		s := string(char)
		if isDigit(s) || s == "." {
			currentNumber += s
		} else {
			if currentNumber != "" {
				tokens = append(tokens, currentNumber)
				currentNumber = ""
			}
			switch s {
			case "+", "-", "*", "/", "(", ")":
				tokens = append(tokens, s)
			default:
				return nil, fmt.Errorf("invalid character in expression: %s", s)
			}
		}
	}
	if currentNumber != "" {
		tokens = append(tokens, currentNumber)
	}
	return tokens, nil
}

func isDigit(s string) bool {
	return s >= "0" && s <= "9"
}

func evaluateExpression(tokens []string, index int) (float64, int, error) {
	leftValue, nextIndex, err := evaluateTerm(tokens, index)
	if err != nil {
		return 0, index, err
	}
	index = nextIndex
	for index < len(tokens) {
		operator := tokens[index]
		if operator == "+" || operator == "-" {
			index++
			rightValue, nextIndex, err := evaluateTerm(tokens, index)
			if err != nil {
				return 0, index, err
			}
			index = nextIndex

			if operator == "+" {
				leftValue += rightValue
			} else if operator == "-" {
				leftValue -= rightValue
			}
		} else {
			break
		}
	}
	return leftValue, index, nil
}

func evaluateTerm(tokens []string, index int) (float64, int, error) {
	leftValue, nextIndex, err := evaluateFactor(tokens, index)
	if err != nil {
		return 0, index, err
	}
	index = nextIndex
	for index < len(tokens) {
		operator := tokens[index]
		if operator == "*" || operator == "/" {
			index++
			rightValue, nextIndex, err := evaluateFactor(tokens, index)
			if err != nil {
				return 0, index, err
			}
			index = nextIndex

			if operator == "*" {
				leftValue *= rightValue
			} else if operator == "/" {
				if rightValue == 0 {
					return 0, index, fmt.Errorf("division by zero")
				}
				leftValue /= rightValue
			}
		} else {
			break
		}
	}
	return leftValue, index, nil
}

func evaluateFactor(tokens []string, index int) (float64, int, error) {
	if index >= len(tokens) {
		return 0, index, fmt.Errorf("unexpected end of expression")
	}
	token := tokens[index]
	if token == "(" {
		index++
		result, nextIndex, err := evaluateExpression(tokens, index)
		if err != nil {
			return 0, index, err
		}
		index = nextIndex
		if index >= len(tokens) || tokens[index] != ")" {
			return 0, index, fmt.Errorf("mismatched parenthesis, expecting ')'")
		}
		index++
		return result, index, nil
	} else if isDigit(string(token[0])) || (len(token) > 1 && (token[0] == '-' || token[0] == '+') && isDigit(string(token[1]))) {
		number, err := strconv.ParseFloat(token, 64)
		if err != nil {
			return 0, index, fmt.Errorf("invalid number format: %s", token)
		}
		index++
		return number, index, nil
	} else {
		return 0, index, fmt.Errorf("unexpected token at position %d: %s", index, token)
	}
}
