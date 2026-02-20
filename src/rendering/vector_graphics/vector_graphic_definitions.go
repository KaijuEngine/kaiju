/******************************************************************************/
/* vector_graphic_definitions.go                                              */
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
	"strconv"
	"strings"

	"kaiju/matrix"
	"kaiju/rendering/vector_graphics/svg"
)

type GradientUnits int8
type SpreadMethod int8

const (
	GradientUnitsUnset GradientUnits = iota
	GradientUnitsUserSpaceOnUse
	GradientUnitsObjectBoundingBox
)

const (
	SpreadMethodUnset SpreadMethod = iota
	SpreadMethodPad
	SpreadMethodReflect
	SpreadMethodRepeat
)

// Defs represents the <defs> element of an SVG. It contains reusable resources
// such as gradients that can be referenced by other elements in the SVG.
type Defs struct {
	LinearGradients []LinearGradient
	RadialGradients []RadialGradient
}

// LinearGradient maps to the <linearGradient> SVG element.
type LinearGradient struct {
	Id     string
	Point1 matrix.Vec2
	Point2 matrix.Vec2
	GradientCommon
}

// RadialGradient maps to the <radialGradient> SVG element.
type RadialGradient struct {
	Id         string
	Center     matrix.Vec2
	Radius     matrix.Float
	FocalPoint matrix.Vec2
	GradientCommon
}

type GradientCommon struct {
	LinkedId      string
	GradientUnits GradientUnits
	SpreadMethod  SpreadMethod
	Stops         []GradientStop
}

// GradientStop maps to the <stop> element inside a gradient definition.
type GradientStop struct {
	Offset      float64
	StopColor   string
	StopOpacity float64
}

// ColorStop is a convenient representation of a gradient stop using the engine's
// matrix.Color type.
type ColorStop struct {
	Offset float64
	Color  matrix.Color
}

func DefsFromSvg(defs svg.Defs) Defs {
	// Convert SVG definitions into the engine's internal representation.
	var out Defs
	// Helper to map SVG GradientUnits (string) to internal enum.
	mapUnits := func(u svg.GradientUnits) GradientUnits {
		switch u {
		case svg.GradientUnitsUserSpaceOnUse:
			return GradientUnitsUserSpaceOnUse
		case svg.GradientUnitsObjectBoundingBox:
			return GradientUnitsObjectBoundingBox
		default:
			return GradientUnitsUnset
		}
	}
	// Helper to map SVG SpreadMethod (string) to internal enum.
	mapSpread := func(s svg.SpreadMethod) SpreadMethod {
		switch s {
		case svg.SpreadMethodPad:
			return SpreadMethodPad
		case svg.SpreadMethodReflect:
			return SpreadMethodReflect
		case svg.SpreadMethodRepeat:
			return SpreadMethodRepeat
		default:
			return SpreadMethodUnset
		}
	}
	// Helper to convert SVG gradient stops to internal GradientStop slice.
	convertStops := func(src []svg.GradientStop) []GradientStop {
		outStops := make([]GradientStop, len(src))
		for i, s := range src {
			outStops[i] = GradientStop{Offset: s.Offset, StopColor: s.StopColor, StopOpacity: s.StopOpacity}
		}
		return outStops
	}
	// Linear gradients
	for _, lg := range defs.LinearGradients {
		var conv LinearGradient
		conv.Id = lg.Id
		conv.Point1 = matrix.NewVec2(matrix.Float(lg.X1), matrix.Float(lg.Y1))
		conv.Point2 = matrix.NewVec2(matrix.Float(lg.X2), matrix.Float(lg.Y2))
		conv.GradientCommon = GradientCommon{
			LinkedId:      lg.XLinkHref,
			GradientUnits: mapUnits(lg.GradientUnits),
			SpreadMethod:  mapSpread(lg.SpreadMethod),
			Stops:         convertStops(lg.Stops),
		}
		out.LinearGradients = append(out.LinearGradients, conv)
	}
	// Radial gradients
	for _, rg := range defs.RadialGradients {
		var conv RadialGradient
		conv.Id = rg.Id
		conv.Center = matrix.NewVec2(matrix.Float(rg.CX), matrix.Float(rg.CY))
		conv.Radius = matrix.Float(rg.R)
		conv.FocalPoint = matrix.NewVec2(matrix.Float(rg.FX), matrix.Float(rg.FY))
		conv.GradientCommon = GradientCommon{
			LinkedId:      rg.XLinkHref,
			GradientUnits: mapUnits(rg.GradientUnits),
			SpreadMethod:  mapSpread(rg.SpreadMethod),
			Stops:         []GradientStop{},
		}
		out.RadialGradients = append(out.RadialGradients, conv)
	}
	return out
}

// ParseColor parses an SVG color string (hex, rgb(), rgba()) into a matrix.Color.
func ParseColor(s string) matrix.Color {
	s = strings.TrimSpace(s)
	// rgb(r,g,b)
	if strings.HasPrefix(s, "rgb(") && strings.HasSuffix(s, ")") {
		inner := s[4 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 3 {
			r, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			g, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			b, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			return matrix.Color{matrix.Float(r) / 255, matrix.Float(g) / 255,
				matrix.Float(b) / 255, 1}
		}
	}
	// rgba(r,g,b,a)
	if strings.HasPrefix(s, "rgba(") && strings.HasSuffix(s, ")") {
		inner := s[5 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 4 {
			r, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			g, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			b, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			a, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
			return matrix.Color{matrix.Float(r) / 255, matrix.Float(g) / 255,
				matrix.Float(b) / 255, matrix.Float(a)}
		}
	}
	// hex format
	if strings.HasPrefix(s, "#") {
		return parseHexColor(s[1:])
	}
	// fallback black
	return matrix.Color{0, 0, 0, 1.0}
}

func parseHexColor(hex string) matrix.Color {
	c := matrix.Color{0, 0, 0, 1.0}
	switch len(hex) {
	case 3:
		// short form #rgb
		r, _ := strconv.ParseInt(strings.Repeat(string(hex[0]), 2), 16, 64)
		g, _ := strconv.ParseInt(strings.Repeat(string(hex[1]), 2), 16, 64)
		b, _ := strconv.ParseInt(strings.Repeat(string(hex[2]), 2), 16, 64)
		c.SetR(matrix.Float(r) / 255)
		c.SetG(matrix.Float(g) / 255)
		c.SetB(matrix.Float(b) / 255)
	case 6:
		r, _ := strconv.ParseInt(hex[0:2], 16, 64)
		g, _ := strconv.ParseInt(hex[2:4], 16, 64)
		b, _ := strconv.ParseInt(hex[4:6], 16, 64)
		c.SetR(matrix.Float(r) / 255)
		c.SetG(matrix.Float(g) / 255)
		c.SetB(matrix.Float(b) / 255)
	}
	return c
}

// FindLinearGradientByID returns a pointer to the LinearGradient with the given id.
func (d *Defs) FindLinearGradientByID(id string) *LinearGradient {
	if d == nil {
		return nil
	}
	for i := range d.LinearGradients {
		if d.LinearGradients[i].Id == id {
			return &d.LinearGradients[i]
		}
	}
	return nil
}

// FindRadialGradientByID returns a pointer to the RadialGradient with the given id.
func (d *Defs) FindRadialGradientByID(id string) *RadialGradient {
	if d == nil {
		return nil
	}
	for i := range d.RadialGradients {
		if d.RadialGradients[i].Id == id {
			return &d.RadialGradients[i]
		}
	}
	return nil
}

// ResolveLinearGradient resolves a linear gradient, following any xlink:href reference.
func (d *Defs) ResolveLinearGradient(id string) *LinearGradient {
	gradient := d.FindLinearGradientByID(id)
	if gradient == nil {
		return nil
	}
	if gradient.LinkedId != "" {
		refID := strings.TrimPrefix(gradient.LinkedId, "#")
		base := d.FindLinearGradientByID(refID)
		if base != nil {
			// Merge base into gradient where values are missing
			result := *gradient
			if result.GradientUnits == GradientUnitsUnset {
				result.GradientUnits = base.GradientUnits
			}
			if len(result.Stops) == 0 {
				result.Stops = base.Stops
			}
			return &result
		}
	}
	return gradient
}

// ResolveRadialGradient resolves a radial gradient, following any xlink:href reference.
func (d *Defs) ResolveRadialGradient(id string) *RadialGradient {
	gradient := d.FindRadialGradientByID(id)
	if gradient == nil {
		return nil
	}
	if gradient.LinkedId != "" {
		refID := strings.TrimPrefix(gradient.LinkedId, "#")
		// radial gradients may reference linear gradients for stops
		base := d.FindLinearGradientByID(refID)
		if base != nil {
			result := *gradient
			if result.GradientUnits == GradientUnitsUnset {
				result.GradientUnits = base.GradientUnits
			}
			if len(result.Stops) == 0 {
				result.Stops = base.Stops
			}
			return &result
		}
	}
	return gradient
}
