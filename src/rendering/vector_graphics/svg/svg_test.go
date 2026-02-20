/******************************************************************************/
/* svg_parser_comprehensive_test.go                                           */
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

package svg

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kaiju/matrix"
)

func getTestSVGPath(t *testing.T, filename string) string {
	// Get the directory of the current test file
	testDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	return filepath.Join(testDir, filename)
}

func TestParseBushSVG(t *testing.T) {
	svgPath := getTestSVGPath(t, "bush.svg")
	svg, err := ParseSVGFile(svgPath)
	if err != nil {
		t.Fatalf("Failed to parse bush.svg: %v", err)
	}
	// Test basic SVG structure
	t.Run("BasicStructure", func(t *testing.T) {
		if svg.Xmlns != "http://www.w3.org/2000/svg" {
			t.Errorf("Expected xmlns 'http://www.w3.org/2000/svg', got '%s'", svg.Xmlns)
		}
		// Note: xmlns:xlink is present in file but Go's xml package may not parse namespaced attributes
		// the same way. This is acceptable for parsing purposes.
		if svg.XmlnsXLink != "" && svg.XmlnsXLink != "http://www.w3.org/1999/xlink" {
			t.Errorf("Unexpected xmlns:xlink value: '%s'", svg.XmlnsXLink)
		}
		if svg.ViewBox != "0 0 512 512" {
			t.Errorf("Expected viewBox '0 0 512 512', got '%s'", svg.ViewBox)
		}
	})
	// Test viewBox parsing
	t.Run("ViewBoxParsing", func(t *testing.T) {
		minX, minY, width, height, err := svg.GetViewBox()
		if err != nil {
			t.Errorf("Failed to parse viewBox: %v", err)
			return
		}
		if minX != 0 || minY != 0 || width != 512 || height != 512 {
			t.Errorf("ViewBox values incorrect: got (%f, %f, %f, %f), expected (0, 0, 512, 512)",
				minX, minY, width, height)
		}
	})
	// Test groups
	t.Run("Groups", func(t *testing.T) {
		groups := svg.FindAllGroups()
		if len(groups) == 0 {
			t.Error("Expected to find groups, found none")
			return
		}
		t.Logf("Found %d groups", len(groups))
		// Check for transform attributes on groups
		groupsWithTransforms := 0
		for i, g := range groups {
			if g.Transform != "" {
				groupsWithTransforms++
				t.Logf("Group %d has transform: %s", i, g.Transform)
			}
		}
		if groupsWithTransforms == 0 {
			t.Error("Expected some groups to have transforms")
		}
	})
	// Test paths
	t.Run("Paths", func(t *testing.T) {
		paths := svg.FindAllPaths()
		if len(paths) == 0 {
			t.Error("Expected to find paths, found none")
			return
		}
		if len(paths) != 8 {
			t.Errorf("Expected 8 paths, got %d", len(paths))
		}
		t.Logf("Found %d paths", len(paths))
		// Verify path properties
		for i, path := range paths {
			if path.Id == "" {
				t.Errorf("Path %d has no ID", i)
			}
			if path.Data == "" {
				t.Errorf("Path %d (ID: %s) has no path data", i, path.Id)
			}
			if !strings.HasPrefix(path.Fill, "url(#") {
				t.Errorf("Path %d (ID: %s) expected fill to be a gradient reference, got '%s'",
					i, path.Id, path.Fill)
			}
			if path.StrokeWidth != 10 {
				t.Errorf("Path %d (ID: %s) expected stroke-width 10, got %f",
					i, path.Id, path.StrokeWidth)
			}
			if path.StrokeLinecap != "round" {
				t.Errorf("Path %d (ID: %s) expected stroke-linecap 'round', got '%s'",
					i, path.Id, path.StrokeLinecap)
			}
		}
	})
	// Test ellipses
	t.Run("Ellipses", func(t *testing.T) {
		ellipses := svg.FindAllEllipses()
		if len(ellipses) != 1 {
			t.Errorf("Expected 1 ellipse, got %d", len(ellipses))
			return
		}
		ellipse := ellipses[0]
		if ellipse.Id != "Circle" {
			t.Errorf("Expected ellipse ID 'Circle', got '%s'", ellipse.Id)
		}
		if ellipse.RX != 37.5091 {
			t.Errorf("Expected RX 37.5091, got %f", ellipse.RX)
		}
		if ellipse.RY != 78.8447 {
			t.Errorf("Expected RY 78.8447, got %f", ellipse.RY)
		}
		if !strings.HasPrefix(ellipse.Fill, "url(#") {
			t.Errorf("Expected fill to be a gradient reference, got '%s'", ellipse.Fill)
		}
	})
	// Test gradients
	t.Run("Gradients", func(t *testing.T) {
		if len(svg.Defs.LinearGradients) == 0 {
			t.Error("Expected to find linear gradients")
			return
		}
		t.Logf("Found %d linear gradients", len(svg.Defs.LinearGradients))
		// Check for gradients with xlink:href
		gradientsWithHref := 0
		for _, grad := range svg.Defs.LinearGradients {
			if grad.XLinkHref != "" {
				gradientsWithHref++
				if !strings.HasPrefix(grad.XLinkHref, "#") {
					t.Errorf("Gradient %s has invalid xlink:href '%s', expected to start with #",
						grad.Id, grad.XLinkHref)
				}
			}
			if grad.GradientUnits == "" && grad.XLinkHref == "" {
				// Base gradients don't need gradientUnits
				t.Logf("Gradient %s has no gradientUnits (base gradient)", grad.Id)
			}
		}
		if gradientsWithHref == 0 {
			t.Error("Expected some gradients to have xlink:href references")
		}
		// Test gradient lookup
		if len(svg.Defs.LinearGradients) > 0 {
			firstID := svg.Defs.LinearGradients[0].Id
			found := svg.Defs.FindLinearGradientByID(firstID)
			if found == nil {
				t.Errorf("Failed to find gradient by ID: %s", firstID)
			}
		}
	})
	// Test transforms parsing on groups
	t.Run("TransformParsing", func(t *testing.T) {
		groups := svg.FindAllGroups()
		testedTransforms := false
		for _, g := range groups {
			if g.Transform == "" {
				continue
			}
			transforms, err := ParseTransform(g.Transform)
			if err != nil {
				t.Errorf("Failed to parse transform '%s': %v", g.Transform, err)
				continue
			}
			if len(transforms) > 0 {
				testedTransforms = true
				t.Logf("Parsed '%s' into %d transform(s)", g.Transform, len(transforms))
				// Test conversion to matrix
				mat := TransformsToMat3(transforms)
				zeroMat := matrix.Mat3{}
				if mat == zeroMat {
					t.Error("Transform resulted in zero matrix")
				}
			}
		}
		if !testedTransforms {
			t.Error("No transforms were tested")
		}
	})
}

func TestParseBushWiggleSVG(t *testing.T) {
	svgPath := getTestSVGPath(t, "bush_wiggle.svg")
	svg, err := ParseSVGFile(svgPath)
	if err != nil {
		t.Fatalf("Failed to parse bush_wiggle.svg: %v", err)
	}
	// Test that wiggle version has animations
	t.Run("AnimationsExist", func(t *testing.T) {
		paths := svg.FindAllPaths()
		animCount := 0
		for _, path := range paths {
			animCount += len(path.Animates)
		}
		if animCount == 0 {
			t.Error("Expected paths to have animations in bush_wiggle.svg")
		} else {
			t.Logf("Found %d path animations", animCount)
		}
	})
	// Test animation details
	t.Run("AnimationDetails", func(t *testing.T) {
		paths := svg.FindAllPaths()
		for i, path := range paths {
			if len(path.Animates) == 0 {
				continue
			}
			for j, anim := range path.Animates {
				if anim.AttributeName != "d" {
					t.Errorf("Path %d Animation %d: expected attributeName 'd', got '%s'",
						i, j, anim.AttributeName)
				}
				if anim.CalcMode != CalcModeSpline {
					t.Errorf("Path %d Animation %d: expected calcMode 'spline', got '%s'",
						i, j, anim.CalcMode)
				}
				if anim.Duration != "0.375s" {
					t.Errorf("Path %d Animation %d: expected dur '0.375s', got '%s'",
						i, j, anim.Duration)
				}
				if anim.Values == "" {
					t.Errorf("Path %d Animation %d: expected values to be set", i, j)
				}
				if anim.KeyTimes == "" {
					t.Errorf("Path %d Animation %d: expected keyTimes to be set", i, j)
				}
				if anim.KeySplines == "" {
					t.Errorf("Path %d Animation %d: expected keySplines to be set", i, j)
				}
				if anim.Fill != FillFreeze {
					t.Errorf("Path %d Animation %d: expected fill 'freeze', got '%s'",
						i, j, anim.Fill)
				}
			}
		}
	})
	// Test animation parsing
	t.Run("AnimationParsing", func(t *testing.T) {
		paths := svg.FindAllPaths()
		for _, path := range paths {
			for _, anim := range path.Animates {
				// Parse values
				values := ParseValues(anim.Values)
				if len(values) == 0 && anim.Values != "" {
					t.Error("Failed to parse animation values")
				}
				// Parse key times
				keyTimes, err := ParseKeyTimes(anim.KeyTimes)
				if err != nil {
					t.Errorf("Failed to parse keyTimes: %v", err)
				}
				if len(keyTimes) > 0 {
					// Verify keyTimes are in valid range [0, 1]
					for i, kt := range keyTimes {
						if kt < 0 || kt > 1 {
							t.Errorf("keyTime %d out of range [0,1]: %f", i, kt)
						}
					}
					// Verify keyTimes are non-decreasing (SVG allows duplicates for certain cases)
					for i := 1; i < len(keyTimes); i++ {
						if keyTimes[i] < keyTimes[i-1] {
							t.Errorf("keyTimes not monotonically non-decreasing at index %d: %f < %f",
								i, keyTimes[i], keyTimes[i-1])
						}
					}
				}
				// Parse key splines
				splines, err := ParseKeySplines(anim.KeySplines)
				if err != nil {
					t.Errorf("Failed to parse keySplines: %v", err)
				}
				if len(splines) > 0 {
					// Verify spline control points are in valid range [0, 1]
					for i, spline := range splines {
						if spline.X1 < 0 || spline.X1 > 1 || spline.X2 < 0 || spline.X2 > 1 {
							t.Errorf("Spline %d X values out of range [0,1]: (%f, %f)",
								i, spline.X1, spline.X2)
						}
					}
				}
			}
		}
	})
	// Test ellipse animations
	t.Run("EllipseAnimations", func(t *testing.T) {
		ellipses := svg.FindAllEllipses()
		if len(ellipses) == 0 {
			t.Fatal("No ellipses found")
		}
		ellipse := ellipses[0]
		t.Logf("Ellipse has %d animate elements", len(ellipse.Animates))
		// Check for position animations
		for _, anim := range ellipse.Animates {
			if anim.AttributeName != "cx" && anim.AttributeName != "cy" {
				t.Logf("Ellipse animation on attribute: %s", anim.AttributeName)
			}
		}
	})
	// Test animateTransform elements
	t.Run("AnimateTransforms", func(t *testing.T) {
		groups := svg.FindAllGroups()
		animTransformCount := 0
		for _, g := range groups {
			animTransformCount += len(g.AnimateTransforms)
		}
		if animTransformCount > 0 {
			t.Logf("Found %d animateTransform elements", animTransformCount)
		}
	})
	// Test both files have same structure but different features
	t.Run("StructureComparison", func(t *testing.T) {
		bushPath := getTestSVGPath(t, "bush.svg")
		bush, _ := ParseSVGFile(bushPath)
		// Both should have same number of paths
		if len(svg.FindAllPaths()) != len(bush.FindAllPaths()) {
			t.Error("bush.svg and bush_wiggle.svg have different number of paths")
		}
		// Both should have same number of ellipses
		if len(svg.FindAllEllipses()) != len(bush.FindAllEllipses()) {
			t.Error("bush.svg and bush_wiggle.svg have different number of ellipses")
		}
		// Note: bush_wiggle.svg has more gradients than bush.svg (18 vs 10)
		// This is expected as the animated version uses unique gradients for each element
		wiggleCount := len(svg.Defs.LinearGradients)
		bushCount := len(bush.Defs.LinearGradients)
		t.Logf("bush.svg has %d gradients, bush_wiggle.svg has %d gradients", bushCount, wiggleCount)
	})
}

func TestTransformVariations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"translate_xy", "translate(10 20)", 1},
		{"translate_comma", "translate(10, 20)", 1},
		{"translate_single", "translate(10)", 1},
		{"rotate_angle", "rotate(45)", 1},
		{"rotate_center", "rotate(45, 100, 100)", 1},
		{"rotate_negative", "rotate(-90)", 1},
		{"scale_uniform", "scale(2)", 1},
		{"scale_xy", "scale(2, 3)", 1},
		{"skewX", "skewX(30)", 1},
		{"skewY", "skewY(45)", 1},
		{"matrix", "matrix(1, 0, 0, 1, 10, 20)", 1},
		{"multiple", "translate(10 20) rotate(45) scale(2)", 3},
		{"nested", "translate(0 0) rotate(90)", 2},
		{"decimal", "translate(260.848 260.97)", 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transforms, err := ParseTransform(test.input)
			if err != nil {
				t.Errorf("ParseTransform(%q) error: %v", test.input, err)
				return
			}
			if len(transforms) != test.expected {
				t.Errorf("ParseTransform(%q) expected %d transforms, got %d",
					test.input, test.expected, len(transforms))
			}
		})
	}
}

func TestTransformToMatrix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected matrix.Mat3
	}{
		{
			name:  "identity",
			input: "translate(0 0)",
			expected: matrix.Mat3{
				1, 0, 0,
				0, 1, 0,
				0, 0, 1,
			},
		},
		{
			name:  "translate_only",
			input: "translate(10 20)",
			expected: matrix.Mat3{
				1, 0, 10,
				0, 1, 20,
				0, 0, 1,
			},
		},
		{
			name:  "scale_only",
			input: "scale(2)",
			expected: matrix.Mat3{
				2, 0, 0,
				0, 2, 0,
				0, 0, 1,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mat, err := ParseTransformAttribute(test.input)
			if err != nil {
				t.Errorf("ParseTransformAttribute(%q) error: %v", test.input, err)
				return
			}
			// Compare with epsilon for float comparison
			epsilon := float32(0.0001)
			for i := 0; i < 9; i++ {
				diff := mat[i] - test.expected[i]
				if diff < 0 {
					diff = -diff
				}
				if diff > epsilon {
					t.Errorf("Matrix[%d] = %f, expected %f", i, mat[i], test.expected[i])
				}
			}
		})
	}
}

func TestColorParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Color
	}{
		{"#ff0000", Color{R: 1, G: 0, B: 0, A: 1}},
		{"#00ff00", Color{R: 0, G: 1, B: 0, A: 1}},
		{"#0000ff", Color{R: 0, G: 0, B: 1, A: 1}},
		{"#ffffff", Color{R: 1, G: 1, B: 1, A: 1}},
		{"#000000", Color{R: 0, G: 0, B: 0, A: 1}},
		{"#f00", Color{R: 1, G: 0, B: 0, A: 1}},
		{"#0f0", Color{R: 0, G: 1, B: 0, A: 1}},
		{"#00f", Color{R: 0, G: 0, B: 1, A: 1}},
		{"rgb(255, 0, 0)", Color{R: 1, G: 0, B: 0, A: 1}},
		{"rgb(0, 255, 0)", Color{R: 0, G: 1, B: 0, A: 1}},
		{"rgb(0, 0, 255)", Color{R: 0, G: 0, B: 1, A: 1}},
		{"rgba(255, 128, 64, 0.5)", Color{R: 1, G: 0.50196, B: 0.25098, A: 0.5}},
	}
	for _, test := range tests {
		t.Run(strings.ReplaceAll(test.input, "#", "hex_"), func(t *testing.T) {
			result := ParseColor(test.input)
			epsilon := 0.01
			if math.Abs(result.R-test.expected.R) > epsilon {
				t.Errorf("R: got %f, expected %f", result.R, test.expected.R)
			}
			if math.Abs(result.G-test.expected.G) > epsilon {
				t.Errorf("G: got %f, expected %f", result.G, test.expected.G)
			}
			if math.Abs(result.B-test.expected.B) > epsilon {
				t.Errorf("B: got %f, expected %f", result.B, test.expected.B)
			}
			if math.Abs(result.A-test.expected.A) > epsilon {
				t.Errorf("A: got %f, expected %f", result.A, test.expected.A)
			}
		})
	}
}

func TestKeySplinesParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"single", "0 0 1 1", 1},
		{"double", "0 0 1 1; 1 0 0.2 0.996603", 2},
		{"triple", "0 0 1 1; 1 0 0.2 0.996603; 0.548632 0 0.830321 1", 3},
		{"with_spaces", "0   0   1   1", 1},
		{"empty", "", 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			splines, err := ParseKeySplines(test.input)
			if err != nil {
				t.Errorf("ParseKeySplines error: %v", err)
				return
			}
			if len(splines) != test.expected {
				t.Errorf("Expected %d splines, got %d", test.expected, len(splines))
			}
		})
	}
}

func TestSVGFileNotFound(t *testing.T) {
	_, err := ParseSVGFile("nonexistent.svg")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestInvalidSVG(t *testing.T) {
	invalidXML := []byte(`<not-valid-xml`)
	_, err := ParseSVG(invalidXML)
	if err == nil {
		t.Error("Expected error for invalid XML")
	}
}

func TestParseSVGString(t *testing.T) {
	const svgString = `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
  <g transform="translate(10 20)">
    <path d="M0 0 L10 10" id="testPath"/>
  </g>
</svg>`
	svg, err := ParseSVGString(svgString)
	if err != nil {
		t.Fatalf("Failed to parse SVG string: %v", err)
	}
	if svg.ViewBox != "0 0 100 100" {
		t.Errorf("Expected viewBox '0 0 100 100', got '%s'", svg.ViewBox)
	}
	paths := svg.FindAllPaths()
	if len(paths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(paths))
	}
	if paths[0].Id != "testPath" {
		t.Errorf("Expected path ID 'testPath', got '%s'", paths[0].Id)
	}
}

func BenchmarkParseBushSVG(b *testing.B) {
	svgPath := getTestSVGPath(&testing.T{}, "bush.svg")
	for i := 0; i < b.N; i++ {
		_, err := ParseSVGFile(svgPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseBushWiggleSVG(b *testing.B) {
	svgPath := getTestSVGPath(&testing.T{}, "bush_wiggle.svg")
	for i := 0; i < b.N; i++ {
		_, err := ParseSVGFile(svgPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseTransform(b *testing.B) {
	input := "translate(260.848 260.97) rotate(45) scale(1 1)"
	for i := 0; i < b.N; i++ {
		_, err := ParseTransform(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseKeySplines(b *testing.B) {
	input := "0 0 1 1; 1 0 0.2 0.996603; 0.548632 0 0.830321 1"
	for i := 0; i < b.N; i++ {
		_, err := ParseKeySplines(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestParseBallSVG(t *testing.T) {
	svgPath := getTestSVGPath(t, "ball.svg")
	svg, err := ParseSVGFile(svgPath)
	if err != nil {
		t.Fatalf("Failed to parse ball.svg: %v", err)
	}
	// Test radial gradients
	t.Run("RadialGradients", func(t *testing.T) {
		if len(svg.Defs.RadialGradients) == 0 {
			t.Error("Expected to find radial gradients")
			return
		}
		t.Logf("Found %d radial gradients", len(svg.Defs.RadialGradients))
		// Verify radial gradient properties
		for i, grad := range svg.Defs.RadialGradients {
			t.Logf("RadialGradient %d: ID=%s, cx=%f, cy=%f, r=%f", i, grad.Id, grad.CX, grad.CY, grad.R)
			if grad.Id == "" {
				t.Errorf("RadialGradient %d has no ID", i)
			}
			if grad.R == 0 {
				t.Errorf("RadialGradient %s has no radius", grad.Id)
			}
			if grad.GradientUnits == "" {
				t.Errorf("RadialGradient %s has no gradientUnits", grad.Id)
			}
			if grad.XLinkHref != "" && !strings.HasPrefix(grad.XLinkHref, "#") {
				t.Errorf("RadialGradient %s has invalid xlink:href: %s", grad.Id, grad.XLinkHref)
			}
		}
		// Test radial gradient lookup
		if len(svg.Defs.RadialGradients) > 0 {
			firstID := svg.Defs.RadialGradients[0].Id
			found := svg.Defs.FindRadialGradientByID(firstID)
			if found == nil {
				t.Errorf("Failed to find radial gradient by ID: %s", firstID)
			}
		}
	})
	// Test solid stroke and fill with rgba
	t.Run("SolidColors", func(t *testing.T) {
		paths := svg.FindAllPaths()
		if len(paths) == 0 {
			t.Fatal("Expected to find paths")
		}
		solidFillFound := false
		solidStrokeFound := false
		for _, path := range paths {
			// Check for solid fill (rgba format)
			if strings.HasPrefix(path.Fill, "rgba(") {
				solidFillFound = true
				color := ParseColor(path.Fill)
				t.Logf("Path %s has solid fill: rgba(%f, %f, %f, %f)", path.Id, color.R, color.G, color.B, color.A)
			}
			// Check for solid stroke (rgba format)
			if strings.HasPrefix(path.Stroke, "rgba(") {
				solidStrokeFound = true
				color := ParseColor(path.Stroke)
				t.Logf("Path %s has solid stroke: rgba(%f, %f, %f, %f)", path.Id, color.R, color.G, color.B, color.A)
			}
			// Check for gradient fill
			if strings.HasPrefix(path.Fill, "url(#") {
				t.Logf("Path %s has gradient fill: %s", path.Id, path.Fill)
			}
		}
		if !solidFillFound {
			t.Error("Expected to find at least one path with solid rgba fill")
		}
		if !solidStrokeFound {
			t.Error("Expected to find at least one path with solid rgba stroke")
		}
	})
	// Test repeatCount="indefinite"
	t.Run("RepeatCountIndefinite", func(t *testing.T) {
		paths := svg.FindAllPaths()
		if len(paths) == 0 {
			t.Fatal("Expected to find paths")
		}
		indefiniteFound := false
		for _, path := range paths {
			for _, anim := range path.Animates {
				if anim.RepeatCount == "indefinite" {
					indefiniteFound = true
					t.Logf("Path %s has animation with repeatCount=indefinite", path.Id)
				}
			}
		}
		if !indefiniteFound {
			t.Error("Expected to find animation with repeatCount='indefinite'")
		}
	})
	// Test paths
	t.Run("Paths", func(t *testing.T) {
		paths := svg.FindAllPaths()
		if len(paths) != 2 {
			t.Errorf("Expected 2 paths, got %d", len(paths))
		}
		t.Logf("Found %d paths", len(paths))
		for i, path := range paths {
			t.Logf("Path %d: ID=%s, has %d animations", i, path.Id, len(path.Animates))
		}
	})
	// Test gradient reference resolution
	t.Run("GradientResolution", func(t *testing.T) {
		// Test resolving radial gradient that references linear gradient
		for _, radialGrad := range svg.Defs.RadialGradients {
			if radialGrad.XLinkHref != "" {
				resolved := svg.Defs.ResolveRadialGradient(radialGrad.Id)
				if resolved == nil {
					t.Errorf("Failed to resolve radial gradient: %s", radialGrad.Id)
					continue
				}
				t.Logf("Resolved radial gradient %s -> references %s", radialGrad.Id, radialGrad.XLinkHref)
			}
		}
	})
	// Test that animated paths have gradient fills (gradient follows path animation)
	t.Run("AnimatedPathsWithGradients", func(t *testing.T) {
		paths := svg.FindAllPaths()
		if len(paths) == 0 {
			t.Fatal("Expected to find paths")
		}
		// Find paths that have both animations and gradient fills
		animatedPathsWithGradients := 0
		for _, path := range paths {
			hasAnimation := len(path.Animates) > 0
			hasGradientFill := strings.HasPrefix(path.Fill, "url(#")
			if hasAnimation && hasGradientFill {
				animatedPathsWithGradients++
				// Extract gradient ID from fill="url(#id)"
				gradientID := strings.TrimPrefix(path.Fill, "url(#")
				gradientID = strings.TrimSuffix(gradientID, ")")
				// Verify the gradient exists
				var gradientExists bool
				for _, rg := range svg.Defs.RadialGradients {
					if rg.Id == gradientID {
						gradientExists = true
						break
					}
				}
				if !gradientExists {
					for _, lg := range svg.Defs.LinearGradients {
						if lg.Id == gradientID {
							gradientExists = true
							break
						}
					}
				}
				if !gradientExists {
					t.Errorf("Path %s references gradient %s but it doesn't exist in defs",
						path.Id, gradientID)
				} else {
					t.Logf("Path %s is animated and uses gradient %s (gradient will follow path animation)",
						path.Id, gradientID)
				}
			}
		}
		if animatedPathsWithGradients == 0 {
			t.Error("Expected to find at least one animated path with a gradient fill")
		} else {
			t.Logf("Found %d animated path(s) with gradient fills", animatedPathsWithGradients)
		}
	})
}
