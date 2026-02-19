/******************************************************************************/
/* svg_transform.go                                                           */
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
	"math"
	"regexp"
	"strconv"
	"strings"

	"kaiju/matrix"
)

// TransformType represents the type of SVG transform operation
type TransformType int

const (
	TransformTranslate TransformType = iota
	TransformRotate
	TransformScale
	TransformSkewX
	TransformSkewY
	TransformMatrix
)

// Transform represents a single SVG transform operation
type Transform struct {
	Type TransformType
	// For translate
	TranslateX, TranslateY float64
	// For rotate
	RotateAngle float64 // in degrees
	RotateCX    float64 // rotation center X (optional)
	RotateCY    float64 // rotation center Y (optional)
	// For scale
	ScaleX, ScaleY float64
	// For skew
	SkewAngle float64 // in degrees
	// For matrix
	Matrix [6]float64 // a, b, c, d, e, f
}

// ToMat3 converts the transform to a 3x3 matrix
func (t Transform) ToMat3() matrix.Mat3 {
	switch t.Type {
	case TransformTranslate:
		// Translation matrix
		// [1 0 tx]
		// [0 1 ty]
		// [0 0 1 ]
		return matrix.Mat3{
			1, 0, float32(t.TranslateX),
			0, 1, float32(t.TranslateY),
			0, 0, 1,
		}
	case TransformRotate:
		// SVG rotation is 2D, around optional center point
		angleRad := t.RotateAngle * math.Pi / 180
		cos := float32(math.Cos(angleRad))
		sin := float32(math.Sin(angleRad))
		if t.RotateCX != 0 || t.RotateCY != 0 {
			// Translate to origin, rotate, translate back
			// T(cx,cy) * R * T(-cx,-cy)
			transToOrigin := matrix.Mat3{
				1, 0, float32(-t.RotateCX),
				0, 1, float32(-t.RotateCY),
				0, 0, 1,
			}
			rotation := matrix.Mat3{
				cos, -sin, 0,
				sin, cos, 0,
				0, 0, 1,
			}
			transBack := matrix.Mat3{
				1, 0, float32(t.RotateCX),
				0, 1, float32(t.RotateCY),
				0, 0, 1,
			}
			return transBack.Multiply(rotation).Multiply(transToOrigin)
		}
		// Simple rotation
		return matrix.Mat3{
			cos, -sin, 0,
			sin, cos, 0,
			0, 0, 1,
		}
	case TransformScale:
		// Scale matrix
		// [sx 0  0]
		// [0  sy 0]
		// [0  0  1]
		return matrix.Mat3{
			float32(t.ScaleX), 0, 0,
			0, float32(t.ScaleY), 0,
			0, 0, 1,
		}
	case TransformSkewX:
		// Skew X matrix
		// [1 tan(a) 0]
		// [0 1      0]
		// [0 0      1]
		return matrix.Mat3{
			1, float32(math.Tan(t.SkewAngle * math.Pi / 180)), 0,
			0, 1, 0,
			0, 0, 1,
		}
	case TransformSkewY:
		// Skew Y matrix
		// [1      0 0]
		// [tan(a) 1 0]
		// [0      0 1]
		return matrix.Mat3{
			1, 0, 0,
			float32(math.Tan(t.SkewAngle * math.Pi / 180)), 1, 0,
			0, 0, 1,
		}
	case TransformMatrix:
		// Custom matrix [a c e]
		//               [b d f]
		//               [0 0 1]
		return matrix.Mat3{
			float32(t.Matrix[0]), float32(t.Matrix[2]), float32(t.Matrix[4]),
			float32(t.Matrix[1]), float32(t.Matrix[3]), float32(t.Matrix[5]),
			0, 0, 1,
		}
	default:
		return matrix.Mat3Identity()
	}
}

// ParseTransform parses an SVG transform attribute string into a slice of Transform
// Supports: translate(x,y), rotate(a,cx,cy), scale(x,y), skewX(a), skewY(a), matrix(a,b,c,d,e,f)
func ParseTransform(s string) ([]Transform, error) {
	if s == "" {
		return nil, nil
	}
	var transforms []Transform
	// Regular expression to match transform functions
	// Matches: functionName(possibly nested parens content)
	re := regexp.MustCompile(`(\w+)\s*\(([^)]+)\)`)
	matches := re.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		funcName := strings.ToLower(strings.TrimSpace(match[1]))
		params := parseTransformParams(match[2])
		transform, err := createTransform(funcName, params)
		if err != nil {
			continue // Skip invalid transforms
		}
		transforms = append(transforms, transform)
	}
	return transforms, nil
}

// parseTransformParams parses the content inside transform function parentheses
func parseTransformParams(s string) []float64 {
	// Replace commas and spaces with a single delimiter
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)
	params := make([]float64, 0, len(parts))
	for _, part := range parts {
		val, err := strconv.ParseFloat(part, 64)
		if err != nil {
			continue
		}
		params = append(params, val)
	}
	return params
}

// createTransform creates a Transform from function name and parameters
func createTransform(name string, params []float64) (Transform, error) {
	switch name {
	case "translate":
		return createTranslateTransform(params), nil
	case "rotate":
		return createRotateTransform(params), nil
	case "scale":
		return createScaleTransform(params), nil
	case "skewx":
		return createSkewXTransform(params), nil
	case "skewy":
		return createSkewYTransform(params), nil
	case "matrix":
		return createMatrixTransform(params), nil
	default:
		return Transform{}, nil
	}
}

func createTranslateTransform(params []float64) Transform {
	t := Transform{Type: TransformTranslate}
	if len(params) >= 1 {
		t.TranslateX = params[0]
	}
	if len(params) >= 2 {
		t.TranslateY = params[1]
	}
	return t
}

func createRotateTransform(params []float64) Transform {
	t := Transform{Type: TransformRotate}
	if len(params) >= 1 {
		t.RotateAngle = params[0]
	}
	if len(params) >= 3 {
		t.RotateCX = params[1]
		t.RotateCY = params[2]
	}
	return t
}

func createScaleTransform(params []float64) Transform {
	t := Transform{Type: TransformScale}
	if len(params) >= 1 {
		t.ScaleX = params[0]
		t.ScaleY = params[0] // Uniform scale if only one parameter
	}
	if len(params) >= 2 {
		t.ScaleY = params[1]
	}
	return t
}

func createSkewXTransform(params []float64) Transform {
	t := Transform{Type: TransformSkewX}
	if len(params) >= 1 {
		t.SkewAngle = params[0]
	}
	return t
}

func createSkewYTransform(params []float64) Transform {
	t := Transform{Type: TransformSkewY}
	if len(params) >= 1 {
		t.SkewAngle = params[0]
	}
	return t
}

func createMatrixTransform(params []float64) Transform {
	t := Transform{Type: TransformMatrix}
	if len(params) >= 6 {
		copy(t.Matrix[:], params[:6])
	}
	return t
}

// TransformsToMat3 combines multiple transforms into a single matrix
// Transforms are applied right to left (matrix multiplication order)
func TransformsToMat3(transforms []Transform) matrix.Mat3 {
	result := matrix.Mat3Identity()
	// Apply transforms in order (left to right in SVG means right to left in matrix multiplication)
	for _, t := range transforms {
		transformMatrix := t.ToMat3()
		result = transformMatrix.Multiply(result)
	}
	return result
}

// ParseTransformAttribute parses a transform string and returns the combined matrix
func ParseTransformAttribute(s string) (matrix.Mat3, error) {
	transforms, err := ParseTransform(s)
	if err != nil {
		return matrix.Mat3Identity(), err
	}
	return TransformsToMat3(transforms), nil
}
