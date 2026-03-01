/******************************************************************************/
/* svg_types.go                                                               */
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
)

// SVG represents the root <svg> element
type SVG struct {
	XMLName    xml.Name   `xml:"svg"`
	Xmlns      string     `xml:"xmlns,attr"`
	XmlnsXLink string     `xml:"xmlns:xlink,attr"`
	ViewBox    string     `xml:"viewBox,attr"`
	Groups     []Group    `xml:"g"`
	Defs       Defs       `xml:"defs"`
	Paths      []Path     `xml:"path"`
	Ellipses   []Ellipse  `xml:"ellipse"`
	Rects      []Rect     `xml:"rect"`
	Circles    []Circle   `xml:"circle"`
	Lines      []Line     `xml:"line"`
	Polygons   []Polygon  `xml:"polygon"`
	Polylines  []Polyline `xml:"polyline"`
	ClipPaths  []ClipPath `xml:"clipPath"`
}

// Group represents <g> with transforms
type Group struct {
	XMLName           xml.Name           `xml:"g"`
	Transform         string             `xml:"transform,attr"`
	Opacity           float64            `xml:"opacity,attr"`
	ClipPath          string             `xml:"clip-path,attr"`
	ClipRule          string             `xml:"clip-rule,attr"`
	Mask              string             `xml:"mask,attr"`
	Groups            []Group            `xml:"g"`
	Paths             []Path             `xml:"path"`
	Ellipses          []Ellipse          `xml:"ellipse"`
	Rects             []Rect             `xml:"rect"`
	Circles           []Circle           `xml:"circle"`
	Lines             []Line             `xml:"line"`
	Polygons          []Polygon          `xml:"polygon"`
	Polylines         []Polyline         `xml:"polyline"`
	AnimateTransforms []AnimateTransform `xml:"animateTransform"`
}

// Path represents <path> elements
type Path struct {
	XMLName        xml.Name  `xml:"path"`
	Id             string    `xml:"id,attr"`
	Data           string    `xml:"d,attr"`
	Stroke         string    `xml:"stroke,attr"`
	StrokeWidth    float64   `xml:"stroke-width,attr"`
	Fill           string    `xml:"fill,attr"`
	StrokeLinecap  string    `xml:"stroke-linecap,attr"`
	StrokeLinejoin string    `xml:"stroke-linejoin,attr"`
	Animates       []Animate `xml:"animate"`
}

// Ellipse represents <ellipse> elements
type Ellipse struct {
	XMLName        xml.Name  `xml:"ellipse"`
	Id             string    `xml:"id,attr"`
	CX             float64   `xml:"cx,attr"`
	CY             float64   `xml:"cy,attr"`
	RX             float64   `xml:"rx,attr"`
	RY             float64   `xml:"ry,attr"`
	Stroke         string    `xml:"stroke,attr"`
	StrokeWidth    float64   `xml:"stroke-width,attr"`
	Fill           string    `xml:"fill,attr"`
	StrokeLinecap  string    `xml:"stroke-linecap,attr"`
	StrokeLinejoin string    `xml:"stroke-linejoin,attr"`
	Animates       []Animate `xml:"animate"`
}

// Rect represents <rect> elements
type Rect struct {
	XMLName        xml.Name  `xml:"rect"`
	Id             string    `xml:"id,attr"`
	X              float64   `xml:"x,attr"`
	Y              float64   `xml:"y,attr"`
	Width          float64   `xml:"width,attr"`
	Height         float64   `xml:"height,attr"`
	RX             float64   `xml:"rx,attr"`
	RY             float64   `xml:"ry,attr"`
	Stroke         string    `xml:"stroke,attr"`
	StrokeWidth    float64   `xml:"stroke-width,attr"`
	Fill           string    `xml:"fill,attr"`
	StrokeLinecap  string    `xml:"stroke-linecap,attr"`
	StrokeLinejoin string    `xml:"stroke-linejoin,attr"`
	Animates       []Animate `xml:"animate"`
}

// Circle represents <circle> elements
type Circle struct {
	XMLName        xml.Name  `xml:"circle"`
	Id             string    `xml:"id,attr"`
	CX             float64   `xml:"cx,attr"`
	CY             float64   `xml:"cy,attr"`
	R              float64   `xml:"r,attr"`
	Stroke         string    `xml:"stroke,attr"`
	StrokeWidth    float64   `xml:"stroke-width,attr"`
	Fill           string    `xml:"fill,attr"`
	StrokeLinecap  string    `xml:"stroke-linecap,attr"`
	StrokeLinejoin string    `xml:"stroke-linejoin,attr"`
	Animates       []Animate `xml:"animate"`
}

// Line represents <line> elements
type Line struct {
	XMLName       xml.Name  `xml:"line"`
	Id            string    `xml:"id,attr"`
	X1            float64   `xml:"x1,attr"`
	Y1            float64   `xml:"y1,attr"`
	X2            float64   `xml:"x2,attr"`
	Y2            float64   `xml:"y2,attr"`
	Stroke        string    `xml:"stroke,attr"`
	StrokeWidth   float64   `xml:"stroke-width,attr"`
	StrokeLinecap string    `xml:"stroke-linecap,attr"`
	Animates      []Animate `xml:"animate"`
}

// Polygon represents <polygon> elements
type Polygon struct {
	XMLName        xml.Name  `xml:"polygon"`
	Id             string    `xml:"id,attr"`
	Points         string    `xml:"points,attr"`
	Stroke         string    `xml:"stroke,attr"`
	StrokeWidth    float64   `xml:"stroke-width,attr"`
	Fill           string    `xml:"fill,attr"`
	StrokeLinecap  string    `xml:"stroke-linecap,attr"`
	StrokeLinejoin string    `xml:"stroke-linejoin,attr"`
	Animates       []Animate `xml:"animate"`
}

// Polyline represents <polyline> elements
type Polyline struct {
	XMLName        xml.Name  `xml:"polyline"`
	Id             string    `xml:"id,attr"`
	Points         string    `xml:"points,attr"`
	Stroke         string    `xml:"stroke,attr"`
	StrokeWidth    float64   `xml:"stroke-width,attr"`
	Fill           string    `xml:"fill,attr"`
	StrokeLinecap  string    `xml:"stroke-linecap,attr"`
	StrokeLinejoin string    `xml:"stroke-linejoin,attr"`
	Animates       []Animate `xml:"animate"`
}
