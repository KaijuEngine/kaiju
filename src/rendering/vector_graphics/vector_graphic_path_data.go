/******************************************************************************/
/* vector_graphic_path_data.go                                                */
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

package vector_graphics

/*
| Command | Meaning | Parameters (absolute) | Example |
|---------|---------|-----------------------|---------|
| `M` / `m` | moveto (start a new sub‑path) | `x y` | `M100 200` |
| `L` / `l` | lineto | `x y` | `L150 250` |
| `H` / `h` | horizontal lineto | `x` | `H200` |
| `V` / `v` | vertical lineto | `y` | `V150` |
| `C` / `c` | cubic Bézier curve | `x1 y1 x2 y2 x y` | `C10 20 30 40 50 60` |
| `S` / `s` | smooth cubic Bézier (reflect previous control point) | `x2 y2 x y` | `S70 80 90 100` |
| `Q` / `q` | quadratic Bézier curve | `x1 y1 x y` | `Q110 120 130 140` |
| `T` / `t` | smooth quadratic Bézier | `x y` | `T150 160` |
| `A` / `a` | elliptical arc | `rx ry x‑axis‑rotation large‑arc‑flag sweep‑flag x y` | `A30 30 0 0 1 200 200` |
| `Z` / `z` | close path (draw a line back to the start of the sub‑path) | *none* | `Z` |
*/

import (
	"regexp"
	"strconv"
)

var tokenRe = regexp.MustCompile(`([AaCcHhLlMmQqSsTtVvZz])|([-+]?[0-9]*\.?[0-9]+(?:[eE][-+]?[0-9]+)?)`)

// PathCommandType enumerates SVG path command letters.
type PathCommandType uint8

type PathRelativity uint8

// PathData is a slice of PathSegment.
type PathData []PathSegment

const (
	// MoveTo command (M or m)
	PathCmdMoveTo PathCommandType = iota
	// LineTo command (L or l)
	PathCmdLineTo
	// Horizontal line (H or h)
	PathCmdHorizontalLineTo
	// Vertical line (V or v)
	PathCmdVerticalLineTo
	// Cubic Bézier curve (C or c)
	PathCmdCubicBezier
	// Smooth cubic Bézier (S or s)
	PathCmdSmoothCubicBezier
	// Quadratic Bézier curve (Q or q)
	PathCmdQuadraticBezier
	// Smooth quadratic Bézier (T or t)
	PathCmdSmoothQuadraticBezier
	// Elliptical arc (A or a)
	PathCmdArc
	// ClosePath (Z or z)
	PathCmdClosePath
)

const (
	// Absolute coordinates (uppercase command letter)
	PathAbs PathRelativity = iota
	// Relative coordinates (lowercase command letter)
	PathRel
)

type PathSegment struct {
	Cmd        PathCommandType // The command type (M, L, C, etc.)
	Rel        PathRelativity  // Absolute or relative
	Params     [6]float64      // Parameter values (x, y, …)
	ParamCount uint8           // Number of valid entries in Params
}

// X returns the first parameter (commonly the X coordinate).
func (s *PathSegment) X() float64 { return s.Params[0] }

// SetX sets the first parameter.
func (s *PathSegment) SetX(x float64) { s.Params[0] = x }

// Y returns the second parameter (commonly the Y coordinate).
func (s *PathSegment) Y() float64 { return s.Params[1] }

// SetY sets the second parameter.
func (s *PathSegment) SetY(y float64) { s.Params[1] = y }

// X1 returns the first control point X.
func (s *PathSegment) X1() float64 { return s.Params[0] }

// SetX1 sets the first control point X.
func (s *PathSegment) SetX1(x float64) { s.Params[0] = x }

// Y1 returns the first control point Y.
func (s *PathSegment) Y1() float64 { return s.Params[1] }

// SetY1 sets the first control point Y.
func (s *PathSegment) SetY1(y float64) { s.Params[1] = y }

// X2 returns the second control point X.
func (s *PathSegment) X2() float64 { return s.Params[2] }

// SetX2 sets the second control point X.
func (s *PathSegment) SetX2(x float64) { s.Params[2] = x }

// Y2 returns the second control point Y.
func (s *PathSegment) Y2() float64 { return s.Params[3] }

// SetY2 sets the second control point Y.
func (s *PathSegment) SetY2(y float64) { s.Params[3] = y }

// Rx returns the X radius for arc commands.
func (s *PathSegment) Rx() float64 { return s.Params[0] }

// SetRx sets the X radius.
func (s *PathSegment) SetRx(rx float64) { s.Params[0] = rx }

// Ry returns the Y radius for arc commands.
func (s *PathSegment) Ry() float64 { return s.Params[1] }

// SetRy sets the Y radius.
func (s *PathSegment) SetRy(ry float64) { s.Params[1] = ry }

// XAxisRotation returns the rotation of the arc's x‑axis.
func (s *PathSegment) XAxisRotation() float64 { return s.Params[2] }

// SetXAxisRotation sets the rotation of the arc's x‑axis.
func (s *PathSegment) SetXAxisRotation(r float64) { s.Params[2] = r }

// LargeArcFlag returns the large‑arc flag (0 or 1).
func (s *PathSegment) LargeArcFlag() float64 { return s.Params[3] }

// SetLargeArcFlag sets the large‑arc flag.
func (s *PathSegment) SetLargeArcFlag(v float64) { s.Params[3] = v }

// SweepFlag returns the sweep flag (0 or 1).
func (s *PathSegment) SweepFlag() float64 { return s.Params[4] }

// SetSweepFlag sets the sweep flag.
func (s *PathSegment) SetSweepFlag(v float64) { s.Params[4] = v }

// Tokenize the input string into command letters and numeric values.
// Whitespace and commas are treated as separators.
// Example: "M10,10L20 20" → ["M","10","10","L","20","20"]
// Regular expression matches a single command letter or a number.
// Numbers may be integer or floating point, with optional sign and exponent.
func ParseData(s string) []PathSegment {
	tokens := tokenRe.FindAllString(s, -1)
	type cmdInfo struct {
		typ   PathCommandType
		rel   PathRelativity
		count int
	}
	var current *cmdInfo
	result := make([]PathSegment, 0, len(s)/8)
	paramBuf := make([]float64, 0, 6)
	flush := func() {
		if current == nil {
			return
		}
		if current.count == 0 {
			result = append(result, PathSegment{Cmd: current.typ, Rel: current.rel})
			return
		}
		for len(paramBuf) >= current.count {
			seg := PathSegment{Cmd: current.typ, Rel: current.rel, ParamCount: uint8(current.count)}
			for i := 0; i < current.count && i < len(seg.Params); i++ {
				seg.Params[i] = paramBuf[i]
			}
			result = append(result, seg)
			paramBuf = paramBuf[current.count:]
		}
	}
	for _, tk := range tokens {
		if len(tk) == 1 && ((tk[0] >= 'A' && tk[0] <= 'Z') || (tk[0] >= 'a' && tk[0] <= 'z')) {
			flush()
			var info cmdInfo
			switch tk[0] {
			case 'M', 'm':
				info.typ = PathCmdMoveTo
				info.count = 2
			case 'L', 'l':
				info.typ = PathCmdLineTo
				info.count = 2
			case 'H', 'h':
				info.typ = PathCmdHorizontalLineTo
				info.count = 1
			case 'V', 'v':
				info.typ = PathCmdVerticalLineTo
				info.count = 1
			case 'C', 'c':
				info.typ = PathCmdCubicBezier
				info.count = 6
			case 'S', 's':
				info.typ = PathCmdSmoothCubicBezier
				info.count = 4
			case 'Q', 'q':
				info.typ = PathCmdQuadraticBezier
				info.count = 4
			case 'T', 't':
				info.typ = PathCmdSmoothQuadraticBezier
				info.count = 2
			case 'A', 'a':
				info.typ = PathCmdArc
				info.count = 6
			case 'Z', 'z':
				info.typ = PathCmdClosePath
				info.count = 0
			}
			if tk[0] >= 'a' && tk[0] <= 'z' {
				info.rel = PathRel
			} else {
				info.rel = PathAbs
			}
			current = &info
			paramBuf = paramBuf[:0]
			if info.count == 0 {
				flush()
				current = nil
			}
			continue
		}
		if current == nil {
			continue
		}
		if val, err := strconv.ParseFloat(tk, 64); err == nil {
			paramBuf = append(paramBuf, val)
		}
	}
	flush()
	return result
}
