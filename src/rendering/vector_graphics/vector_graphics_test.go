/******************************************************************************/
/* vector_graphic_test.go                                                     */
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
	_ "embed"
	"math"
	"strings"
	"testing"

	"kaiju/matrix"
	"kaiju/rendering/vector_graphics/svg"
)

//go:embed svg/ball.svg
var svgBall string

//go:embed svg/bush.svg
var svgBush string

//go:embed svg/bush_wiggle.svg
var svgBallWiggle string

// Test basic VectorGraphic creation from minimal SVG
func TestVectorGraphicFromSVGBasic(t *testing.T) {
	svgData := svg.SVG{
		ViewBox: "0 0 100 100",
		Groups:  []svg.Group{},
	}
	result := VectorGraphicFromSVG(svgData)
	// Verify ViewBox parsing
	if result.ViewBox[0] != 0 || result.ViewBox[1] != 0 || result.ViewBox[2] != 100 || result.ViewBox[3] != 100 {
		t.Errorf("Expected ViewBox [0 0 100 100], got [%v %v %v %v]",
			result.ViewBox[0], result.ViewBox[1], result.ViewBox[2], result.ViewBox[3])
	}
	// Verify empty groups
	if len(result.Groups) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(result.Groups))
	}
}

// Test ViewBox parsing with various formats
func TestParseViewBox(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected [4]float64
	}{
		{"Standard", "0 0 100 100", [4]float64{0, 0, 100, 100}},
		{"WithNegative", "-10 -20 50 75", [4]float64{-10, -20, 50, 75}},
		{"WithDecimals", "0.5 1.5 100.25 200.75", [4]float64{0.5, 1.5, 100.25, 200.75}},
		{"WithExtraSpaces", "0   0   100   100", [4]float64{0, 0, 100, 100}},
		{"Empty", "", [4]float64{0, 0, 0, 0}},
		{"PartialValues", "0 0 100", [4]float64{0, 0, 0, 0}}, // Requires exactly 4 values, returns zeros if less
		{"NonNumeric", "a b c d", [4]float64{0, 0, 0, 0}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := parseViewBox(test.input)
			for i := 0; i < 4; i++ {
				if result[i] != test.expected[i] {
					t.Errorf("ViewBox[%d]: expected %f, got %f", i, test.expected[i], result[i])
				}
			}
		})
	}
}

// Test group conversion with simple structure
func TestConvertSvgGroupsSimple(t *testing.T) {
	svgGroups := []svg.Group{
		{
			Transform: "",
			Opacity:   1.0,
			Groups:    []svg.Group{},
			Paths:     []svg.Path{},
			Ellipses:  []svg.Ellipse{},
		},
	}
	result := convertSvgGroups(svgGroups)
	if len(result) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(result))
	}
	group := result[0]
	if group.Opacity != 1.0 {
		t.Errorf("Expected opacity 1.0, got %v", group.Opacity)
	}
	if group.Transform.Position.X() != 0 || group.Transform.Position.Y() != 0 {
		t.Error("Expected position (0, 0), got default transform")
	}
}

// Test group with transform parsing
func TestConvertGroupTransform(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		exposeX  float64
		exposeY  float64
		expScale float64
	}{
		{"Translate", "translate(10 20)", 10, 20, 1},
		{"Scale", "scale(2 3)", 0, 0, 2},
		{"Empty", "", 0, 0, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convertGroupTransform(test.input)

			epsilon := 0.0001
			if !floatEqual(float64(result.Position.X()), test.exposeX, epsilon) {
				t.Errorf("X: expected %f, got %f", test.exposeX, result.Position.X())
			}
			if !floatEqual(float64(result.Position.Y()), test.exposeY, epsilon) {
				t.Errorf("Y: expected %f, got %f", test.exposeY, result.Position.Y())
			}
		})
	}
}

// Test path conversion
func TestConvertSvgPath(t *testing.T) {
	pathData := "M10 20 L30 40"
	svgPath := svg.Path{
		Id:             "testPath",
		Data:           pathData,
		Stroke:         "#ff0000",
		StrokeWidth:    2.0,
		Fill:           "#00ff00",
		StrokeLinecap:  "round",
		StrokeLinejoin: "bevel",
		Animates:       []svg.Animate{},
	}
	result := convertSvgPath(svgPath)
	if result.Id != "testPath" {
		t.Errorf("Expected ID 'testPath', got '%s'", result.Id)
	}
	if result.StrokeWidth != 2.0 {
		t.Errorf("Expected stroke width 2.0, got %f", result.StrokeWidth)
	}
	if result.StrokeLinecap != StrokeLinecapRound {
		t.Errorf("Expected stroke linecap 'round', got %v", result.StrokeLinecap)
	}
	if result.StrokeLinejoin != StrokeLinejoinBevel {
		t.Errorf("Expected stroke linejoin 'bevel', got %v", result.StrokeLinejoin)
	}
	// Verify stroke color (red)
	if !colorEqual(result.Stroke, matrix.Color{1, 0, 0, 1}, 0.01) {
		t.Errorf("Expected red stroke, got %v", result.Stroke)
	}
	// Verify fill color (green)
	if !colorEqual(result.Fill, matrix.Color{0, 1, 0, 1}, 0.01) {
		t.Errorf("Expected green fill, got %v", result.Fill)
	}
}

// Test ellipse conversion
func TestConvertSvgEllipse(t *testing.T) {
	svgEllipse := svg.Ellipse{
		Id:             "testEllipse",
		CX:             50.0,
		CY:             75.0,
		RX:             25.0,
		RY:             15.0,
		Stroke:         "#0000ff",
		StrokeWidth:    1.5,
		Fill:           "none",
		StrokeLinecap:  "square",
		StrokeLinejoin: "miter",
		Animates:       []svg.Animate{},
	}
	result := convertSvgEllipse(svgEllipse)
	if result.Id != "testEllipse" {
		t.Errorf("Expected ID 'testEllipse', got '%s'", result.Id)
	}
	if !vec2Equal(result.Center, matrix.NewVec2(50, 75), 0.0001) {
		t.Errorf("Expected center (50, 75), got (%v, %v)", result.Center.X(), result.Center.Y())
	}
	if !vec2Equal(result.Radius, matrix.NewVec2(25, 15), 0.0001) {
		t.Errorf("Expected radius (25, 15), got (%v, %v)", result.Radius.X(), result.Radius.Y())
	}
	if result.StrokeWidth != 1.5 {
		t.Errorf("Expected stroke width 1.5, got %f", result.StrokeWidth)
	}
	if result.StrokeLinecap != StrokeLinecapSquare {
		t.Errorf("Expected stroke linecap 'square', got %v", result.StrokeLinecap)
	}
}

// Test nested groups conversion
func TestConvertNestedGroups(t *testing.T) {
	innerPath := svg.Path{
		Id:   "innerPath",
		Data: "M0 0 L10 10",
	}
	innerGroup := svg.Group{
		Transform: "translate(5 5)",
		Opacity:   0.8,
		Paths:     []svg.Path{innerPath},
		Groups:    []svg.Group{},
	}
	outerGroup := svg.Group{
		Transform: "translate(10 10)",
		Opacity:   1.0,
		Groups:    []svg.Group{innerGroup},
		Paths:     []svg.Path{},
	}
	result := convertSvgGroup(outerGroup)
	if len(result.Groups) != 1 {
		t.Fatalf("Expected 1 nested group, got %d", len(result.Groups))
	}
	nestedGroup := result.Groups[0]
	if nestedGroup.Opacity != 0.8 {
		t.Errorf("Expected nested group opacity 0.8, got %v", nestedGroup.Opacity)
	}
	if len(nestedGroup.Paths) != 1 {
		t.Fatalf("Expected 1 path in nested group, got %d", len(nestedGroup.Paths))
	}
	if nestedGroup.Paths[0].Id != "innerPath" {
		t.Errorf("Expected path ID 'innerPath', got '%s'", nestedGroup.Paths[0].Id)
	}
}

// Test stroke linecap mapping
func TestMapStrokeLinecap(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected StrokeLinecap
	}{
		{"Butt", "butt", StrokeLinecapButt},
		{"Round", "round", StrokeLinecapRound},
		{"Square", "square", StrokeLinecapSquare},
		{"Inherit", "inherit", StrokeLinecapInherit},
		{"Default", "unknown", StrokeLinecapButt},
		{"Empty", "", StrokeLinecapButt},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := mapStrokeLinecap(test.input)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

// Test stroke linejoin mapping
func TestMapStrokeLinejoin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected StrokeLinejoin
	}{
		{"Miter", "miter", StrokeLinejoinMiter},
		{"Round", "round", StrokeLinejoinRound},
		{"Bevel", "bevel", StrokeLinejoinBevel},
		{"Inherit", "inherit", StrokeLinejoinInherit},
		{"Default", "unknown", StrokeLinejoinMiter},
		{"Empty", "", StrokeLinejoinMiter},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := mapStrokeLinejoin(test.input)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

// Test conversion of a real SVG file (ball.svg) to VectorGraphic and verify fields
func TestVectorGraphicFromSVG(t *testing.T) {
	tests := []string{svgBall, svgBush, svgBallWiggle}
	for _, svgData := range tests {
		// Parse the embedded SVG string
		svgData, err := svg.ParseSVGString(svgData)
		if err != nil {
			t.Fatalf("Failed to parse ball.svg: %v", err)
		}
		// Convert to VectorGraphic
		vg := VectorGraphicFromSVG(*svgData)
		// Verify ViewBox parsing
		expectedVB := [4]float64{0, 0, 512, 512}
		for i := 0; i < 4; i++ {
			if vg.ViewBox[i] != expectedVB[i] {
				t.Errorf("ViewBox[%d] expected %v, got %v", i, expectedVB[i], vg.ViewBox[i])
			}
		}
		// The SVG has two top-level groups
		if len(vg.Groups) != len(svgData.Groups) {
			t.Fatalf("Expected %d top-level groups, got %d", len(svgData.Groups), len(vg.Groups))
		}
		for i := range len(svgData.Groups) {
			if len(vg.Groups[i].Paths) != len(svgData.Groups[i].Paths) {
				t.Fatalf("Expected %d top-level groups, got %d",
					len(svgData.Groups[i].Paths), len(vg.Groups[i].Paths))
			}
			ga, gb := vg.Groups[i], svgData.Groups[i]
			for j := range len(ga.Paths) {
				pa, pb := ga.Paths[j], gb.Paths[j]
				if pa.Id != pb.Id {
					t.Errorf("Path id expected '%s', got %s", pb.Id, pa.Id)
				}
				// Verify stroke width
				if float64(pa.StrokeWidth) != pb.StrokeWidth {
					t.Errorf("Path %s: expected stroke width %v, got %v", pa.Id, pb.StrokeWidth, pa.StrokeWidth)
				}
				// Verify stroke color
				expectedStroke := ParseColor(pb.Stroke)
				if !colorEqual(pa.Stroke, expectedStroke, 0.01) {
					t.Errorf("Path %s: stroke color mismatch", pa.Id)
				}
				// Verify fill color (if not a URL reference)
				if !strings.HasPrefix(pb.Fill, "url(") {
					expectedFill := ParseColor(pb.Fill)
					if !colorEqual(pa.Fill, expectedFill, 0.01) {
						t.Errorf("Path %s: fill color mismatch", pa.Id)
					}
				}
				// Verify linecap and linejoin mappings
				if pa.StrokeLinecap != mapStrokeLinecap(pb.StrokeLinecap) {
					t.Errorf("Path %s: stroke linecap mismatch", pa.Id)
				}
				if pa.StrokeLinejoin != mapStrokeLinejoin(pb.StrokeLinejoin) {
					t.Errorf("Path %s: stroke linejoin mismatch", pa.Id)
				}
			}
		}
	}
}

// Test color parsing (hex format)
func TestParseColorHex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected matrix.Color
	}{
		{"Red", "#ff0000", matrix.Color{1, 0, 0, 1}},
		{"Green", "#00ff00", matrix.Color{0, 1, 0, 1}},
		{"Blue", "#0000ff", matrix.Color{0, 0, 1, 1}},
		{"White", "#ffffff", matrix.Color{1, 1, 1, 1}},
		{"Black", "#000000", matrix.Color{0, 0, 0, 1}},
		{"ShortRed", "#f00", matrix.Color{1, 0, 0, 1}},
		{"ShortGreen", "#0f0", matrix.Color{0, 1, 0, 1}},
		{"ShortBlue", "#00f", matrix.Color{0, 0, 1, 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ParseColor(test.input)
			if !colorEqual(result, test.expected, 0.01) {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

// Test color parsing (rgb format)
func TestParseColorRGB(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected matrix.Color
	}{
		{"Red", "rgb(255, 0, 0)", matrix.Color{1, 0, 0, 1}},
		{"Green", "rgb(0, 255, 0)", matrix.Color{0, 1, 0, 1}},
		{"Blue", "rgb(0, 0, 255)", matrix.Color{0, 0, 1, 1}},
		{"Custom", "rgb(128, 64, 192)", matrix.Color{128.0 / 255, 64.0 / 255, 192.0 / 255, 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ParseColor(test.input)
			if !colorEqual(result, test.expected, 0.01) {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

// Test color parsing (rgba format)
func TestParseColorRGBA(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected matrix.Color
	}{
		{"RedTransparent", "rgba(255, 0, 0, 0.5)", matrix.Color{1, 0, 0, 0.5}},
		{"GreenOpaque", "rgba(0, 255, 0, 1.0)", matrix.Color{0, 1, 0, 1.0}},
		{"BlueTranslucent", "rgba(0, 0, 255, 0.25)", matrix.Color{0, 0, 1, 0.25}},
		{"Custom", "rgba(128, 64, 192, 0.75)", matrix.Color{128.0 / 255, 64.0 / 255, 192.0 / 255, 0.75}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ParseColor(test.input)
			if !colorEqual(result, test.expected, 0.01) {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

// Test color parsing edge cases
func TestParseColorEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected matrix.Color
	}{
		{"EmptyString", "", matrix.Color{0, 0, 0, 1}},
		{"None", "none", matrix.Color{0, 0, 0, 1}},
		{"WhitespaceLeading", "  #ff0000", matrix.Color{1, 0, 0, 1}},
		{"InvalidRGB", "rgb(invalid)", matrix.Color{0, 0, 0, 1}},
		{"MissingAlpha", "rgba(255, 0, 0)", matrix.Color{0, 0, 0, 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ParseColor(test.input)
			if !colorEqual(result, test.expected, 0.01) {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

// Test complete SVG to VectorGraphic conversion with multiple elements
func TestVectorGraphicComplexConversion(t *testing.T) {
	svgData := svg.SVG{
		ViewBox: "0 0 200 200",
		Groups: []svg.Group{
			{
				Transform: "translate(10 10)",
				Opacity:   1.0,
				Paths: []svg.Path{
					{
						Id:          "path1",
						Data:        "M0 0 L50 50",
						Stroke:      "#ff0000",
						StrokeWidth: 2.0,
						Fill:        "none",
						Animates:    []svg.Animate{},
					},
				},
				Ellipses: []svg.Ellipse{
					{
						Id:   "ellipse1",
						CX:   100,
						CY:   100,
						RX:   30,
						RY:   20,
						Fill: "#00ff00",
					},
				},
				Groups: []svg.Group{},
			},
		},
	}
	result := VectorGraphicFromSVG(svgData)
	// Verify ViewBox
	if result.ViewBox[2] != 200 || result.ViewBox[3] != 200 {
		t.Errorf("ViewBox dimensions incorrect")
	}
	// Verify groups
	if len(result.Groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(result.Groups))
	}
	group := result.Groups[0]
	// Verify paths
	if len(group.Paths) != 1 {
		t.Fatalf("Expected 1 path, got %d", len(group.Paths))
	}
	if group.Paths[0].Id != "path1" {
		t.Errorf("Expected path ID 'path1', got '%s'", group.Paths[0].Id)
	}
	// Verify ellipses
	if len(group.Ellipses) != 1 {
		t.Fatalf("Expected 1 ellipse, got %d", len(group.Ellipses))
	}
	if group.Ellipses[0].Id != "ellipse1" {
		t.Errorf("Expected ellipse ID 'ellipse1', got '%s'", group.Ellipses[0].Id)
	}
}

// Test empty SVG
func TestVectorGraphicFromEmptySVG(t *testing.T) {
	svgData := svg.SVG{
		ViewBox: "",
		Groups:  []svg.Group{},
	}
	result := VectorGraphicFromSVG(svgData)
	for i := 0; i < 4; i++ {
		if result.ViewBox[i] != 0 {
			t.Errorf("Expected ViewBox[%d] to be 0, got %f", i, result.ViewBox[i])
		}
	}
	if len(result.Groups) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(result.Groups))
	}
}

// Test multiple paths in a group
func TestMultiplePathsInGroup(t *testing.T) {
	svgData := svg.SVG{
		ViewBox: "0 0 100 100",
		Groups: []svg.Group{
			{
				Transform: "",
				Opacity:   1.0,
				Paths: []svg.Path{
					{Id: "path1", Data: "M0 0 L10 10", Fill: "#ff0000"},
					{Id: "path2", Data: "M20 20 L30 30", Fill: "#00ff00"},
					{Id: "path3", Data: "M40 40 L50 50", Fill: "#0000ff"},
				},
				Groups:   []svg.Group{},
				Ellipses: []svg.Ellipse{},
			},
		},
	}
	result := VectorGraphicFromSVG(svgData)
	if len(result.Groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(result.Groups))
	}
	if len(result.Groups[0].Paths) != 3 {
		t.Fatalf("Expected 3 paths, got %d", len(result.Groups[0].Paths))
	}
	expectedIDs := []string{"path1", "path2", "path3"}
	for i, expectedID := range expectedIDs {
		if result.Groups[0].Paths[i].Id != expectedID {
			t.Errorf("Path %d: expected ID '%s', got '%s'", i, expectedID, result.Groups[0].Paths[i].Id)
		}
	}
}

// Test multiple ellipses in a group
func TestMultipleEllipsesInGroup(t *testing.T) {
	svgData := svg.SVG{
		ViewBox: "0 0 100 100",
		Groups: []svg.Group{
			{
				Transform: "",
				Opacity:   1.0,
				Ellipses: []svg.Ellipse{
					{Id: "ellipse1", CX: 10, CY: 10, RX: 5, RY: 5},
					{Id: "ellipse2", CX: 50, CY: 50, RX: 10, RY: 20},
					{Id: "ellipse3", CX: 90, CY: 90, RX: 15, RY: 10},
				},
				Groups: []svg.Group{},
				Paths:  []svg.Path{},
			},
		},
	}
	result := VectorGraphicFromSVG(svgData)
	if len(result.Groups[0].Ellipses) != 3 {
		t.Fatalf("Expected 3 ellipses, got %d", len(result.Groups[0].Ellipses))
	}
	expectedIDs := []string{"ellipse1", "ellipse2", "ellipse3"}
	for i, expectedID := range expectedIDs {
		if result.Groups[0].Ellipses[i].Id != expectedID {
			t.Errorf("Ellipse %d: expected ID '%s', got '%s'", i, expectedID, result.Groups[0].Ellipses[i].Id)
		}
	}
}

// Test deeply nested groups
func TestDeeplyNestedGroups(t *testing.T) {
	// Create a 3-level deep group hierarchy
	level3Group := svg.Group{
		Transform: "translate(3 3)",
		Opacity:   0.3,
		Paths:     []svg.Path{{Id: "deepPath", Data: "M0 0"}},
		Groups:    []svg.Group{},
	}
	level2Group := svg.Group{
		Transform: "translate(2 2)",
		Opacity:   0.6,
		Groups:    []svg.Group{level3Group},
		Paths:     []svg.Path{},
	}
	level1Group := svg.Group{
		Transform: "translate(1 1)",
		Opacity:   1.0,
		Groups:    []svg.Group{level2Group},
		Paths:     []svg.Path{},
	}
	result := convertSvgGroup(level1Group)
	// Navigate the hierarchy
	if len(result.Groups) != 1 {
		t.Fatal("Level 1: expected 1 group")
	}
	level2 := result.Groups[0]
	if len(level2.Groups) != 1 {
		t.Fatal("Level 2: expected 1 group")
	}
	level3 := level2.Groups[0]
	if len(level3.Paths) != 1 {
		t.Fatal("Level 3: expected 1 path")
	}
	if level3.Paths[0].Id != "deepPath" {
		t.Errorf("Expected path ID 'deepPath', got '%s'", level3.Paths[0].Id)
	}
	if level3.Opacity != 0.3 {
		t.Errorf("Expected opacity 0.3, got %v", level3.Opacity)
	}
}

// Test group opacity preservation
func TestGroupOpacityPreservation(t *testing.T) {
	tests := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	for _, opacity := range tests {
		t.Run(string(rune(int(opacity*100))), func(t *testing.T) {
			svgGroup := svg.Group{
				Opacity:  opacity,
				Groups:   []svg.Group{},
				Paths:    []svg.Path{},
				Ellipses: []svg.Ellipse{},
			}
			result := convertSvgGroup(svgGroup)
			if float64(result.Opacity) != opacity {
				t.Errorf("Expected opacity %v, got %v", opacity, result.Opacity)
			}
		})
	}
}

// Helper function to compare floats with epsilon
func floatEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// Helper function to compare Vec2
func vec2Equal(a, b matrix.Vec2, epsilon float64) bool {
	return floatEqual(float64(a.X()), float64(b.X()), epsilon) &&
		floatEqual(float64(a.Y()), float64(b.Y()), epsilon)
}

// Helper function to compare Color
func colorEqual(a, b matrix.Color, epsilon float64) bool {
	return floatEqual(float64(a.R()), float64(b.R()), epsilon) &&
		floatEqual(float64(a.G()), float64(b.G()), epsilon) &&
		floatEqual(float64(a.B()), float64(b.B()), epsilon) &&
		floatEqual(float64(a.A()), float64(b.A()), epsilon)
}
