/******************************************************************************/
/* svg_animation.go                                                           */
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

// CalcMode represents animation calculation modes for SVG SMIL animations
type CalcMode string

// Additive represents additive behavior for animations
type Additive string

// Accumulate represents accumulation behavior for repeated animations
type Accumulate string

// FillMode represents the fill behavior when animation completes
type FillMode string

// RestartMode controls when animations can restart
type RestartMode string

// AttributeName represents the name of an animatable SVG attribute.
type AttributeName string

type AttributeType string

const (
	// CalcModeDiscrete jumps between values without interpolation
	CalcModeDiscrete CalcMode = "discrete"
	// CalcModeLinear uses simple linear interpolation between values (default)
	CalcModeLinear CalcMode = "linear"
	// CalcModePaced produces even pace across animation, ignores keyTimes/keySplines
	CalcModePaced CalcMode = "paced"
	// CalcModeSpline uses cubic bezier spline interpolation
	CalcModeSpline CalcMode = "spline"
)

const (
	// AdditiveReplace overrides the underlying value (default)
	AdditiveReplace Additive = "replace"
	// AdditiveSum adds animation values to underlying value
	AdditiveSum Additive = "sum"
)

const (
	// AccumulateNone means repeat iterations are not cumulative (default)
	AccumulateNone Accumulate = "none"
	// AccumulateSum means each repeat builds upon the previous iteration
	AccumulateSum Accumulate = "sum"
)

const (
	// FillFreeze keeps the final animation value
	FillFreeze FillMode = "freeze"
	// FillRemove reverts to underlying value (default)
	FillRemove FillMode = "remove"
)

const (
	// RestartAlways allows restart at any time (default)
	RestartAlways RestartMode = "always"
	// RestartWhenNotActive prevents restart while running
	RestartWhenNotActive RestartMode = "whenNotActive"
	// RestartNever prevents any restart
	RestartNever RestartMode = "never"
)

// List of all animatable attribute names for SVG SMIL animations.
// The list follows the SVG specification's “Animatable attributes” table.
const (
	// Geometry / shape attributes
	AttrX          AttributeName = "x"
	AttrY          AttributeName = "y"
	AttrDX         AttributeName = "dx"
	AttrDY         AttributeName = "dy"
	AttrWidth      AttributeName = "width"
	AttrHeight     AttributeName = "height"
	AttrRX         AttributeName = "rx"
	AttrRY         AttributeName = "ry"
	AttrCX         AttributeName = "cx"
	AttrCY         AttributeName = "cy"
	AttrR          AttributeName = "r"
	AttrFX         AttributeName = "fx"
	AttrFY         AttributeName = "fy"
	AttrX1         AttributeName = "x1"
	AttrY1         AttributeName = "y1"
	AttrX2         AttributeName = "x2"
	AttrY2         AttributeName = "y2"
	AttrPoints     AttributeName = "points"
	AttrD          AttributeName = "d"
	AttrPathLength AttributeName = "pathLength"

	// Transform attributes (generic)
	AttrTransform AttributeName = "transform"

	// Presentation attributes
	AttrFill                AttributeName = "fill"
	AttrStroke              AttributeName = "stroke"
	AttrStrokeWidth         AttributeName = "stroke-width"
	AttrStrokeDasharray     AttributeName = "stroke-dasharray"
	AttrStrokeDashoffset    AttributeName = "stroke-dashoffset"
	AttrStrokeOpacity       AttributeName = "stroke-opacity"
	AttrFillOpacity         AttributeName = "fill-opacity"
	AttrOpacity             AttributeName = "opacity"
	AttrVisibility          AttributeName = "visibility"
	AttrDisplay             AttributeName = "display"
	AttrColor               AttributeName = "color"
	AttrFontFamily          AttributeName = "font-family"
	AttrFontSize            AttributeName = "font-size"
	AttrFontWeight          AttributeName = "font-weight"
	AttrFontStyle           AttributeName = "font-style"
	AttrTextAnchor          AttributeName = "text-anchor"
	AttrLetterSpacing       AttributeName = "letter-spacing"
	AttrWordSpacing         AttributeName = "word-spacing"
	AttrKerning             AttributeName = "kerning"
	AttrTextLength          AttributeName = "textLength"
	AttrLengthAdjust        AttributeName = "lengthAdjust"
	AttrStopColor           AttributeName = "stop-color"
	AttrStopOpacity         AttributeName = "stop-opacity"
	AttrOffset              AttributeName = "offset"
	AttrGradientTransform   AttributeName = "gradientTransform"
	AttrMask                AttributeName = "mask"
	AttrClipPath            AttributeName = "clip-path"
	AttrFilter              AttributeName = "filter"
	AttrMaskContentUnits    AttributeName = "maskContentUnits"
	AttrMaskUnits           AttributeName = "maskUnits"
	AttrClipPathUnits       AttributeName = "clipPathUnits"
	AttrPatternUnits        AttributeName = "patternUnits"
	AttrPatternContentUnits AttributeName = "patternContentUnits"
	AttrPatternTransform    AttributeName = "patternTransform"
	AttrViewBox             AttributeName = "viewBox"
	AttrPreserveAspectRatio AttributeName = "preserveAspectRatio"
	AttrHref                AttributeName = "href"
	AttrXlinkHref           AttributeName = "xlink:href"
)

const (
	AttrXml AttributeType = "XML"
	AttrCSS AttributeType = "XML"
)

// SharedAnimation contains fields common to both <animate> and <animateTransform> elements.
type SharedAnimation struct {
	From       string   `xml:"from,attr"`
	To         string   `xml:"to,attr"`
	By         string   `xml:"by,attr"`
	Values     string   `xml:"values,attr"`
	KeyTimes   string   `xml:"keyTimes,attr"`
	KeySplines string   `xml:"keySplines,attr"`
	CalcMode   CalcMode `xml:"calcMode,attr"`
	// Duration corresponds to the "dur" attribute used by both elements.
	Duration    string      `xml:"dur,attr"`
	Begin       string      `xml:"begin,attr"`
	End         string      `xml:"end,attr"`
	RepeatCount string      `xml:"repeatCount,attr"`
	RepeatDur   string      `xml:"repeatDur,attr"`
	Fill        FillMode    `xml:"fill,attr"`
	Restart     RestartMode `xml:"restart,attr"`
	Additive    Additive    `xml:"additive,attr"`
	Accumulate  Accumulate  `xml:"accumulate,attr"`
}

// Animate represents <animate> element for animating attributes over time
type Animate struct {
	XMLName       xml.Name      `xml:"animate"`
	AttributeName AttributeName `xml:"attributeName,attr"`
	AttributeType AttributeType `xml:"attributeType,attr"` // "XML" or "CSS"
	SharedAnimation
	Min string `xml:"min,attr"`
	Max string `xml:"max,attr"`
}

// AnimateTransform represents <animateTransform> for animating transformations
type AnimateTransform struct {
	XMLName xml.Name `xml:"animateTransform"`
	Type    string   `xml:"type,attr"` // "translate", "rotate", "scale", "skewX", "skewY"
	SharedAnimation
}

// KeySpline represents a cubic bezier spline control points for interpolation
type KeySpline struct {
	X1 float64
	Y1 float64
	X2 float64
	Y2 float64
}

// ParseKeySplines parses a keySplines attribute string into slice of KeySpline
// Format: "x1 y1 x2 y2; x1 y1 x2 y2; ..."
func ParseKeySplines(s string) ([]KeySpline, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ";")
	splines := make([]KeySpline, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		values := strings.Fields(part)
		if len(values) != 4 {
			continue
		}
		x1, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			continue
		}
		y1, err := strconv.ParseFloat(values[1], 64)
		if err != nil {
			continue
		}
		x2, err := strconv.ParseFloat(values[2], 64)
		if err != nil {
			continue
		}
		y2, err := strconv.ParseFloat(values[3], 64)
		if err != nil {
			continue
		}
		splines = append(splines, KeySpline{X1: x1, Y1: y1, X2: x2, Y2: y2})
	}
	return splines, nil
}

// ParseValues parses a semicolon-separated values string into slice
func ParseValues(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ";")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// ParseKeyTimes parses a semicolon-separated keyTimes string into float slice
func ParseKeyTimes(s string) ([]float64, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ";")
	times := make([]float64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		t, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, err
		}
		times = append(times, t)
	}
	return times, nil
}
