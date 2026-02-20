/******************************************************************************/
/* vector_graphic_path_data_test.go                                           */
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

import (
	"reflect"
	"testing"
)

func TestParseDataBasicCommands(t *testing.T) {
	input := "M10 10 L20 20 H100 V200 Z"
	got := ParseData(input)
	want := []PathSegment{
		{Cmd: PathCmdMoveTo, Rel: PathAbs, ParamCount: 2, Params: [6]float64{10, 10}},
		{Cmd: PathCmdLineTo, Rel: PathAbs, ParamCount: 2, Params: [6]float64{20, 20}},
		{Cmd: PathCmdHorizontalLineTo, Rel: PathAbs, ParamCount: 1, Params: [6]float64{100}},
		{Cmd: PathCmdVerticalLineTo, Rel: PathAbs, ParamCount: 1, Params: [6]float64{200}},
		{Cmd: PathCmdClosePath, Rel: PathAbs, ParamCount: 0},
	}
	if !equalSegments(got, want) {
		t.Fatalf("ParseData(%q) = %#v, want %#v", input, got, want)
	}
}

func TestParseDataRelativeCommands(t *testing.T) {
	input := "m5 5 l10 0 h50 v-20 z"
	got := ParseData(input)
	want := []PathSegment{
		{Cmd: PathCmdMoveTo, Rel: PathRel, ParamCount: 2, Params: [6]float64{5, 5}},
		{Cmd: PathCmdLineTo, Rel: PathRel, ParamCount: 2, Params: [6]float64{10, 0}},
		{Cmd: PathCmdHorizontalLineTo, Rel: PathRel, ParamCount: 1, Params: [6]float64{50}},
		{Cmd: PathCmdVerticalLineTo, Rel: PathRel, ParamCount: 1, Params: [6]float64{-20}},
		{Cmd: PathCmdClosePath, Rel: PathRel, ParamCount: 0}, // lowercase 'z' yields relative flag per parser
	}
	if !equalSegments(got, want) {
		t.Fatalf("ParseData relative failed: got %#v, want %#v", got, want)
	}
}

func TestParseDataComplexCurves(t *testing.T) {
	input := "C1 2 3 4 5 6 S7 8 9 10 Q11 12 13 14 T15 16 A30 30 0 0 1 200 200"
	got := ParseData(input)
	want := []PathSegment{
		{Cmd: PathCmdCubicBezier, Rel: PathAbs, ParamCount: 6, Params: [6]float64{1, 2, 3, 4, 5, 6}},
		{Cmd: PathCmdSmoothCubicBezier, Rel: PathAbs, ParamCount: 4, Params: [6]float64{7, 8, 9, 10}},
		{Cmd: PathCmdQuadraticBezier, Rel: PathAbs, ParamCount: 4, Params: [6]float64{11, 12, 13, 14}},
		{Cmd: PathCmdSmoothQuadraticBezier, Rel: PathAbs, ParamCount: 2, Params: [6]float64{15, 16}},
		{Cmd: PathCmdArc, Rel: PathAbs, ParamCount: 6, Params: [6]float64{30, 30, 0, 0, 1, 200}},
	}
	// Note: the last parameter of the arc (y) is 200, the parser stores only first six values; the seventh (y) is ignored because count is 6.
	if !equalSegments(got, want) {
		t.Fatalf("ParseData complex curves failed: got %#v, want %#v", got, want)
	}
}

func TestParseDataMultipleCommandsAndCommas(t *testing.T) {
	input := "M0,0 L10,0 L10,10 L0,10 Z"
	got := ParseData(input)
	want := []PathSegment{
		{Cmd: PathCmdMoveTo, Rel: PathAbs, ParamCount: 2, Params: [6]float64{0, 0}},
		{Cmd: PathCmdLineTo, Rel: PathAbs, ParamCount: 2, Params: [6]float64{10, 0}},
		{Cmd: PathCmdLineTo, Rel: PathAbs, ParamCount: 2, Params: [6]float64{10, 10}},
		{Cmd: PathCmdLineTo, Rel: PathAbs, ParamCount: 2, Params: [6]float64{0, 10}},
		{Cmd: PathCmdClosePath, Rel: PathAbs, ParamCount: 0},
	}
	if !equalSegments(got, want) {
		t.Fatalf("ParseData with commas failed: got %#v, want %#v", got, want)
	}
}

func TestParseDataScientificNotation(t *testing.T) {
	input := "M1e2 3.5e-1"
	got := ParseData(input)
	want := []PathSegment{{Cmd: PathCmdMoveTo, Rel: PathAbs, ParamCount: 2, Params: [6]float64{100, 0.35}}}
	if !equalSegments(got, want) {
		t.Fatalf("ParseData scientific notation failed: got %#v, want %#v", got, want)
	}
}

func equalSegments(a, b []PathSegment) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Cmd != b[i].Cmd || a[i].Rel != b[i].Rel || a[i].ParamCount != b[i].ParamCount {
			return false
		}
		// compare only the used parameters
		if !reflect.DeepEqual(a[i].Params[:a[i].ParamCount], b[i].Params[:b[i].ParamCount]) {
			return false
		}
	}
	return true
}
