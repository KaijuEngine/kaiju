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
	"kaiju/matrix"
	"kaiju/rendering/vector_graphics/svg"
	"strconv"
	"strings"
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

type Transform struct {
	Position matrix.Vec2
	Rotation matrix.Float
	Scale    matrix.Vec2
}

type Group struct {
	Transform         Transform
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
	return Group{
		Transform:         convertGroupTransform(g.Transform),
		Opacity:           matrix.Float(g.Opacity),
		Groups:            convertSvgGroups(g.Groups),
		Paths:             convertSvgPaths(g.Paths),
		Ellipses:          convertSvgEllipses(g.Ellipses),
		AnimateTransforms: convertSvgAnimateTransforms(g.AnimateTransforms),
	}
}

func convertGroupTransform(transformStr string) Transform {
	t := Transform{
		Position: matrix.NewVec2(0, 0),
		Rotation: 0,
		Scale:    matrix.NewVec2(1, 1),
	}
	if transformStr == "" {
		return t
	}
	// Parse the transform string using the svg package's ParseTransform function
	transforms, err := svg.ParseTransform(transformStr)
	if err != nil || len(transforms) == 0 {
		return t
	}
	// Extract translate, rotate, and scale from the first matching transform
	for _, transform := range transforms {
		switch transform.Type {
		case svg.TransformTranslate:
			t.Position = matrix.NewVec2(matrix.Float(transform.TranslateX), matrix.Float(transform.TranslateY))
		case svg.TransformRotate:
			t.Rotation = matrix.Float(transform.RotateAngle)
		case svg.TransformScale:
			t.Scale = matrix.NewVec2(matrix.Float(transform.ScaleX), matrix.Float(transform.ScaleY))
		}
	}
	return t
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
