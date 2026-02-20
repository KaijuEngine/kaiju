/******************************************************************************/
/* svg_defs.go                                                                */
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
	"encoding/xml"
	"strconv"
	"strings"
)

// GradientUnits represents the coordinate system for gradients
type GradientUnits string

// SpreadMethod defines how a gradient is extended beyond its start/end points.
// It mirrors the SVG `spreadMethod` attribute.
type SpreadMethod string

const (
	// GradientUnitsUserSpaceOnUse uses absolute coordinates
	GradientUnitsUserSpaceOnUse GradientUnits = "userSpaceOnUse"
	// GradientUnitsObjectBoundingBox uses relative coordinates (0-1)
	GradientUnitsObjectBoundingBox GradientUnits = "objectBoundingBox"
)

const (
	// SpreadMethodPad (default) clamps the gradient colors beyond the bounds.
	SpreadMethodPad SpreadMethod = "pad"
	// SpreadMethodReflect mirrors the gradient pattern when it repeats.
	SpreadMethodReflect SpreadMethod = "reflect"
	// SpreadMethodRepeat tiles the gradient pattern beyond the bounds.
	SpreadMethodRepeat SpreadMethod = "repeat"
)

// Defs represents the <defs> section containing reusable resources
type Defs struct {
	XMLName         xml.Name         `xml:"defs"`
	LinearGradients []LinearGradient `xml:"linearGradient"`
	RadialGradients []RadialGradient `xml:"radialGradient"`
}

// LinearGradient represents <linearGradient> element
type LinearGradient struct {
	XMLName       xml.Name       `xml:"linearGradient"`
	Id            string         `xml:"id,attr"`
	X1            float64        `xml:"x1,attr"`
	Y1            float64        `xml:"y1,attr"`
	X2            float64        `xml:"x2,attr"`
	Y2            float64        `xml:"y2,attr"`
	GradientUnits GradientUnits  `xml:"gradientUnits,attr"`
	SpreadMethod  SpreadMethod   `xml:"spreadMethod,attr"`
	XLinkHref     string         `xml:"http://www.w3.org/1999/xlink href,attr"`
	Stops         []GradientStop `xml:"stop"`
}

// RadialGradient represents <radialGradient> element
type RadialGradient struct {
	XMLName       xml.Name      `xml:"radialGradient"`
	Id            string        `xml:"id,attr"`
	CX            float64       `xml:"cx,attr"`
	CY            float64       `xml:"cy,attr"`
	R             float64       `xml:"r,attr"`
	FX            float64       `xml:"fx,attr"`
	FY            float64       `xml:"fy,attr"`
	GradientUnits GradientUnits `xml:"gradientUnits,attr"`
	SpreadMethod  SpreadMethod  `xml:"spreadMethod,attr"`
	XLinkHref     string        `xml:"http://www.w3.org/1999/xlink href,attr"`
}

// GradientStop represents <stop> element for gradients
type GradientStop struct {
	XMLName     xml.Name `xml:"stop"`
	Offset      float64  `xml:"offset,attr"`
	StopColor   string   `xml:"stop-color,attr"`
	StopOpacity float64  `xml:"stop-opacity,attr"`
}

// ColorStop represents a parsed color stop with offset and color values
type ColorStop struct {
	Offset float64
	Color  Color
}

// Color represents an RGBA color
type Color struct {
	R float64
	G float64
	B float64
	A float64
}

// ParseColor parses SVG color strings including rgb() and rgba() formats
func ParseColor(s string) Color {
	s = strings.TrimSpace(s)
	// Handle rgb(r,g,b) format
	if strings.HasPrefix(s, "rgb(") && strings.HasSuffix(s, ")") {
		inner := s[4 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 3 {
			r, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			g, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			b, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			return Color{R: r / 255.0, G: g / 255.0, B: b / 255.0, A: 1.0}
		}
	}
	// Handle rgba(r,g,b,a) format
	if strings.HasPrefix(s, "rgba(") && strings.HasSuffix(s, ")") {
		inner := s[5 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 4 {
			r, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			g, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			b, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			a, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
			return Color{R: r / 255.0, G: g / 255.0, B: b / 255.0, A: a}
		}
	}
	// Handle hex colors
	if strings.HasPrefix(s, "#") {
		return parseHexColor(s[1:])
	}
	// Default to black for unknown formats
	return Color{R: 0, G: 0, B: 0, A: 1.0}
}

func parseHexColor(hex string) Color {
	c := Color{A: 1.0}
	switch len(hex) {
	case 3:
		// RGB short format
		r, _ := strconv.ParseInt(hex[0:1]+hex[0:1], 16, 64)
		g, _ := strconv.ParseInt(hex[1:2]+hex[1:2], 16, 64)
		b, _ := strconv.ParseInt(hex[2:3]+hex[2:3], 16, 64)
		c.R = float64(r) / 255.0
		c.G = float64(g) / 255.0
		c.B = float64(b) / 255.0
	case 6:
		// RGB format
		r, _ := strconv.ParseInt(hex[0:2], 16, 64)
		g, _ := strconv.ParseInt(hex[2:4], 16, 64)
		b, _ := strconv.ParseInt(hex[4:6], 16, 64)
		c.R = float64(r) / 255.0
		c.G = float64(g) / 255.0
		c.B = float64(b) / 255.0
	}
	return c
}

// FindLinearGradientByID finds a linear gradient by its ID in the defs
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

// FindRadialGradientByID finds a radial gradient by its ID in the defs
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

// ResolveLinearGradient resolves a linear gradient including xlink:href references
func (d *Defs) ResolveLinearGradient(id string) *LinearGradient {
	gradient := d.FindLinearGradientByID(id)
	if gradient == nil {
		return nil
	}
	// Handle xlink:href reference
	if gradient.XLinkHref != "" {
		refID := strings.TrimPrefix(gradient.XLinkHref, "#")
		baseGradient := d.FindLinearGradientByID(refID)
		if baseGradient != nil {
			// Merge with base gradient (base takes precedence for unset values)
			result := *gradient
			if result.GradientUnits == "" {
				result.GradientUnits = baseGradient.GradientUnits
			}
			if len(result.Stops) == 0 {
				result.Stops = baseGradient.Stops
			}
			return &result
		}
	}
	return gradient
}

// ResolveRadialGradient resolves a radial gradient including xlink:href references
func (d *Defs) ResolveRadialGradient(id string) *RadialGradient {
	gradient := d.FindRadialGradientByID(id)
	if gradient == nil {
		return nil
	}
	// Handle xlink:href reference - radial gradients can reference linear gradients for stops
	if gradient.XLinkHref != "" {
		refID := strings.TrimPrefix(gradient.XLinkHref, "#")
		// Try to find referenced gradient (could be linear or radial)
		baseGradient := d.FindLinearGradientByID(refID)
		if baseGradient != nil {
			// Merge properties
			result := *gradient
			if result.GradientUnits == "" {
				result.GradientUnits = baseGradient.GradientUnits
			}
			return &result
		}
	}
	return gradient
}
