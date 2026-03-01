/******************************************************************************/
/* vector_graphic_types.go                                                    */
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
	"strconv"
	"strings"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/vector_graphics/svg"
)

type StrokeLinecap int8
type StrokeLinejoin int8

const (
	StrokeLinecapButt StrokeLinecap = iota
	StrokeLinecapRound
	StrokeLinecapSquare
	StrokeLinecapInherit
)

const (
	StrokeLinejoinMiter StrokeLinejoin = iota
	StrokeLinejoinRound
	StrokeLinejoinBevel
	StrokeLinejoinInherit
)

type VectorGraphic struct {
	ViewBox [4]float64
	Groups  []Group
}

type Group struct {
	Transform         matrix.Transform
	Opacity           matrix.Float
	Groups            []Group
	Paths             []Path
	Ellipses          []Ellipse
	AnimateTransforms []AnimateTransform
}

type Path struct {
	Id             string
	Data           []PathSegment
	Stroke         matrix.Color
	StrokeWidth    matrix.Float
	Fill           matrix.Color
	StrokeLinecap  StrokeLinecap
	StrokeLinejoin StrokeLinejoin
	Animates       []Animate
}

type Ellipse struct {
	Id             string
	Center         matrix.Vec2
	Radius         matrix.Vec2
	Stroke         matrix.Color
	StrokeWidth    matrix.Float
	Fill           matrix.Color
	StrokeLinecap  StrokeLinecap
	StrokeLinejoin StrokeLinejoin
	Animates       []Animate
}

func VectorGraphicFromSVG(svgData svg.SVG) VectorGraphic {
	viewBox := parseViewBox(svgData.ViewBox)
	groups := convertSvgGroups(svgData.Groups)
	return VectorGraphic{
		ViewBox: viewBox,
		Groups:  groups,
	}
}

func parseViewBox(viewBoxStr string) [4]float64 {
	var viewBox [4]float64
	parts := strings.Fields(viewBoxStr)
	if len(parts) >= 4 {
		for i := 0; i < 4 && i < len(parts); i++ {
			if val, err := strconv.ParseFloat(parts[i], 64); err == nil {
				viewBox[i] = val
			}
		}
	}
	return viewBox
}

func convertSvgGroups(groups []svg.Group) []Group {
	result := make([]Group, len(groups))
	for i, g := range groups {
		result[i] = convertSvgGroup(g)
	}
	return result
}

func convertSvgGroup(g svg.Group) Group {
	out := Group{
		Opacity:           matrix.Float(g.Opacity),
		Groups:            convertSvgGroups(g.Groups),
		Paths:             convertSvgPaths(g.Paths),
		Ellipses:          convertSvgEllipses(g.Ellipses),
		AnimateTransforms: convertSvgAnimateTransforms(g.AnimateTransforms),
	}
	readGroupTransform(g.Transform, &out.Transform)
	return out
}

func readGroupTransform(transformStr string, transform *matrix.Transform) {
	if transformStr == "" {
		return
	}
	// Parse the transform string using the svg package's ParseTransform function
	transforms, err := svg.ParseTransform(transformStr)
	if err != nil || len(transforms) == 0 {
		return
	}
	// Extract translate, rotate, and scale from the first matching transform
	for i := range transforms {
		t := &transforms[i]
		switch t.Type {
		case svg.TransformTranslate:
			transform.SetPosition(matrix.NewVec3(matrix.Float(t.TranslateX), matrix.Float(t.TranslateY), 0))
		case svg.TransformRotate:
			transform.SetRotation(matrix.NewVec3(0, 0, matrix.Float(t.RotateAngle)))
		case svg.TransformScale:
			transform.SetScale(matrix.NewVec3(matrix.Float(t.ScaleX), matrix.Float(t.ScaleY), 1))
		}
	}
}

func convertSvgPaths(paths []svg.Path) []Path {
	result := make([]Path, len(paths))
	for i, p := range paths {
		result[i] = convertSvgPath(p)
	}
	return result
}

func convertSvgPath(p svg.Path) Path {
	return Path{
		Id:             p.Id,
		Data:           ParseData(p.Data),
		Stroke:         ParseColor(p.Stroke),
		StrokeWidth:    matrix.Float(p.StrokeWidth),
		Fill:           ParseColor(p.Fill),
		StrokeLinecap:  mapStrokeLinecap(p.StrokeLinecap),
		StrokeLinejoin: mapStrokeLinejoin(p.StrokeLinejoin),
		Animates:       convertSvgAnimates(p.Animates),
	}
}

func convertSvgEllipses(ellipses []svg.Ellipse) []Ellipse {
	result := make([]Ellipse, len(ellipses))
	for i, e := range ellipses {
		result[i] = convertSvgEllipse(e)
	}
	return result
}

func convertSvgEllipse(e svg.Ellipse) Ellipse {
	return Ellipse{
		Id:             e.Id,
		Center:         matrix.NewVec2(matrix.Float(e.CX), matrix.Float(e.CY)),
		Radius:         matrix.NewVec2(matrix.Float(e.RX), matrix.Float(e.RY)),
		Stroke:         ParseColor(e.Stroke),
		StrokeWidth:    matrix.Float(e.StrokeWidth),
		Fill:           ParseColor(e.Fill),
		StrokeLinecap:  mapStrokeLinecap(e.StrokeLinecap),
		StrokeLinejoin: mapStrokeLinejoin(e.StrokeLinejoin),
		Animates:       convertSvgAnimates(e.Animates),
	}
}

func convertSvgAnimates(animates []svg.Animate) []Animate {
	result := make([]Animate, len(animates))
	for i, a := range animates {
		result[i] = AnimateFromSvg(a)
	}
	return result
}

func convertSvgAnimateTransforms(transforms []svg.AnimateTransform) []AnimateTransform {
	result := make([]AnimateTransform, len(transforms))
	for i, t := range transforms {
		result[i] = AnimateTransformFromSvg(t)
	}
	return result
}

func mapStrokeLinecap(s string) StrokeLinecap {
	switch s {
	case "round":
		return StrokeLinecapRound
	case "square":
		return StrokeLinecapSquare
	case "inherit":
		return StrokeLinecapInherit
	default:
		return StrokeLinecapButt
	}
}

func mapStrokeLinejoin(s string) StrokeLinejoin {
	switch s {
	case "round":
		return StrokeLinejoinRound
	case "bevel":
		return StrokeLinejoinBevel
	case "inherit":
		return StrokeLinejoinInherit
	default:
		return StrokeLinejoinMiter
	}
}

// ToVerts will Generate vertices for both the fill and the outline (stroke).
// The fill uses the ellipse's radius, while the outline creates an inner
// and outer ring based on the stroke width.
//
// A good starting value for `segments` is 32
func (e Ellipse) ToVerts(segments int) []rendering.Vertex {
	// Capacity includes fill vertices plus inner and outer outline vertices.
	verts := make([]rendering.Vertex, 0, (segments+1)*3)
	// Helper to generate a ring of vertices given radii.
	addRing := func(radiusX, radiusY float32) {
		for i := 0; i <= segments; i++ {
			angle := float32(i) * 2.0 * math.Pi / matrix.Float(segments)
			x := float32(math.Cos(float64(angle))) * radiusX
			y := float32(math.Sin(float64(angle))) * radiusY
			position := matrix.Vec3{x, y, 0}
			normal := matrix.Vec3{x / radiusX, y / radiusY, 0}.Normal()
			uv := matrix.Vec2{float32(i) / matrix.Float(segments), 0}
			verts = append(verts, rendering.Vertex{
				Position:     position,
				Normal:       normal,
				Tangent:      matrix.Vec4{0, 0, 0, 1},
				UV0:          uv,
				Color:        matrix.ColorWhite(),
				JointIds:     matrix.Vec4i{},
				JointWeights: matrix.Vec4{},
				MorphTarget:  matrix.Vec3{},
			})
		}
	}
	// Fill ring (original radius).
	addRing(float32(e.Radius.X()), float32(e.Radius.Y()))
	// Outline rings if stroke width is greater than zero.
	if e.StrokeWidth > 0 {
		half := float32(e.StrokeWidth) / 2.0
		// Outer ring expands outward.
		outerX := float32(e.Radius.X()) + half
		outerY := float32(e.Radius.Y()) + half
		// Inner ring contracts inward, clamped to non‑negative.
		innerX := float32(e.Radius.X()) - half
		if innerX < 0 {
			innerX = 0
		}
		innerY := float32(e.Radius.Y()) - half
		if innerY < 0 {
			innerY = 0
		}
		addRing(outerX, outerY)
		addRing(innerX, innerY)
	}
	return verts
}
