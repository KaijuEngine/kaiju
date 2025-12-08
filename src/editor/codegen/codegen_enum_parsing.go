/******************************************************************************/
/* codegen_enum_parsing.go                                                    */
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

package codegen

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

func evalConstExpr(expr ast.Expr, iotaVal int64, prevValues map[string]any) (any, error) {
	if expr == nil {
		return nil, errors.New("nil expression")
	}
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Name == "iota" {
			return iotaVal, nil
		}
		if v, ok := prevValues[e.Name]; ok {
			return v, nil
		}
		return nil, fmt.Errorf("undefined identifier: %s", e.Name)
	case *ast.BasicLit:
		switch e.Kind {
		case token.INT:
			i, err := strconv.ParseInt(e.Value, 0, 64)
			if err != nil {
				return nil, err
			}
			return i, nil
		case token.FLOAT:
			f, err := strconv.ParseFloat(e.Value, 64)
			if err != nil {
				return nil, err
			}
			return f, nil
		case token.CHAR:
			r, _, _, err := strconv.UnquoteChar(e.Value, '\'')
			if err != nil {
				return nil, err
			}
			return int64(r), nil
			// Rune is int64 for safety
		case token.STRING:
			s, err := strconv.Unquote(e.Value)
			if err != nil {
				return nil, err
			}
			return s, nil
		default:
			return nil, errors.New("unsupported basic literal")
		}
	case *ast.ParenExpr:
		return evalConstExpr(e.X, iotaVal, prevValues)
	case *ast.UnaryExpr:
		xv, err := evalConstExpr(e.X, iotaVal, prevValues)
		if err != nil {
			return nil, err
		}
		switch xv.(type) {
		case int64:
			v := xv.(int64)
			switch e.Op {
			case token.ADD:
				return v, nil
			case token.SUB:
				return -v, nil
			case token.XOR:
				// Bitwise complement
				return ^v, nil
			default:
				return nil, fmt.Errorf("unsupported unary op on int: %s", e.Op)
			}
		case float64:
			v := xv.(float64)
			switch e.Op {
			case token.ADD:
				return v, nil
			case token.SUB:
				return -v, nil
			default:
				return nil, fmt.Errorf("unsupported unary op on float: %s", e.Op)
			}
		default:
			return nil, errors.New("unary operator not supported on string/char")
		}
	case *ast.BinaryExpr:
		lv, err := evalConstExpr(e.X, iotaVal, prevValues)
		if err != nil {
			return nil, err
		}
		rv, err := evalConstExpr(e.Y, iotaVal, prevValues)
		if err != nil {
			return nil, err
		}
		// String concatenation only allowed with +
		if e.Op == token.ADD {
			if lstr, ok := lv.(string); ok {
				if rstr, ok := rv.(string); ok {
					return lstr + rstr, nil
				}
				return nil, errors.New("cannot add string and non-string")
			}
		} else {
			if _, ok := lv.(string); ok {
				return nil, errors.New("string only supports + with another string")
			}
			if _, ok := rv.(string); ok {
				return nil, errors.New("string only supports + with another string")
			}
		}
		// Number operations
		lIsInt := isInt(lv)
		rIsInt := isInt(rv)
		if lIsInt && rIsInt {
			l := toInt64(lv)
			r := toInt64(rv)
			switch e.Op {
			case token.ADD:
				return l + r, nil
			case token.SUB:
				return l - r, nil
			case token.MUL:
				return l * r, nil
			case token.QUO:
				if r == 0 {
					return nil, errors.New("division by zero")
				}
				// Integer truncate towards zero
				return l / r, nil
			case token.REM:
				if r == 0 {
					return nil, errors.New("remainder by zero")
				}
				return l % r, nil
			case token.AND:
				return l & r, nil
			case token.OR:
				return l | r, nil
			case token.XOR:
				return l ^ r, nil
			case token.SHL:
				return l << r, nil
			case token.SHR:
				return l >> r, nil
			case token.AND_NOT:
				return l &^ r, nil
			default:
				return nil, fmt.Errorf("unsupported operator on int: %s", e.Op)
			}
		}
		// Float path (at least one operand is float64)
		l := toFloat64(lv)
		r := toFloat64(rv)
		switch e.Op {
		case token.ADD:
			return l + r, nil
		case token.SUB:
			return l - r, nil
		case token.MUL:
			return l * r, nil
		case token.QUO:
			if r == 0 {
				return nil, errors.New("division by zero")
			}
			return l / r, nil
		default:
			return nil, fmt.Errorf("operator %s not supported on float", e.Op)
		}
	default:
		return nil, fmt.Errorf("unsupported expr type %T", expr)
	}
}

func isInt(v any) bool {
	_, ok := v.(int64)
	return ok
}

func toInt64(v any) int64 {
	if i, ok := v.(int64); ok {
		return i
	}
	return int64(v.(float64))
}

func toFloat64(v any) float64 {
	switch vv := v.(type) {
	case int64:
		return float64(vv)
	case float64:
		return vv
	default:
		panic("not numeric")
	}
}

func locateAllEnumValues(f *ast.File, t *ast.TypeSpec) map[string]any {
	result := make(map[string]any)
	typeName := t.Name.Name
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		iotaVal := int64(0)
		var prevInner ast.Expr = nil
		inSequence := false
		// reset for every const() group — this fixes cross-group carry-over
		for _, spec := range gd.Specs {
			vs := spec.(*ast.ValueSpec)
			if len(vs.Names) != 1 {
				iotaVal += int64(len(vs.Names))
				continue
			}
			nameIdent := vs.Names[0]
			if nameIdent.Name == "_" || strings.HasPrefix(nameIdent.Name, "_") {
				iotaVal++
				continue
			}
			name := nameIdent.Name
			var inner ast.Expr
			isEnumEntry := false
			if len(vs.Values) > 0 {
				expr := vs.Values[0]
				// Cast style: VarName = TypeAlias("value")
				if call, ok := expr.(*ast.CallExpr); ok && len(call.Args) == 1 {
					if id, ok := call.Fun.(*ast.Ident); ok && id.Name == typeName {
						inner = call.Args[0]
						isEnumEntry = true
					}
				}
				// Typed style: VarName TypeAlias = "value"
				if !isEnumEntry && vs.Type != nil {
					if id, ok := vs.Type.(*ast.Ident); ok && id.Name == typeName {
						inner = expr
						isEnumEntry = true
					}
				}
				if isEnumEntry {
					prevInner = inner
					inSequence = true
				}
			} else {
				// Implicit repeat of previous expression (only within the same const() group)
				if inSequence && prevInner != nil {
					inner = prevInner
					isEnumEntry = true
				}
			}
			if isEnumEntry {
				val, err := evalConstExpr(inner, iotaVal, result)
				if err == nil {
					result[name] = val
				}
				// silently skip on error, the source is probably invalid anyway
			} else {
				inSequence = false
				prevInner = nil
			}
			iotaVal++
		}
	}
	return result
}
