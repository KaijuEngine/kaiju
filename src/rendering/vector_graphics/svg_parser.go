/******************************************************************************/
/* svg_parser.go                                                              */
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
	"encoding/xml"
	"fmt"
	"os"
)

// ParseSVGFile parses an SVG file and returns the SVG structure
func ParseSVGFile(filename string) (*SVG, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read SVG file: %w", err)
	}
	return ParseSVG(data)
}

// ParseSVG parses SVG XML data and returns the SVG structure
func ParseSVG(data []byte) (*SVG, error) {
	var svg SVG
	err := xml.Unmarshal(data, &svg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SVG: %w", err)
	}
	return &svg, nil
}

// ParseSVGString parses an SVG string and returns the SVG structure
func ParseSVGString(s string) (*SVG, error) {
	return ParseSVG([]byte(s))
}

// SVGParser provides a streaming parser for large SVG files
type SVGParser struct {
	decoder *xml.Decoder
}

// NewSVGParser creates a new streaming SVG parser
func NewSVGParser(data []byte) *SVGParser {
	return &SVGParser{
		decoder: xml.NewDecoder(nil), // Will be set in Parse
	}
}

// ParseFile parses an SVG file using streaming approach
// Useful for very large SVG files that shouldn't be loaded entirely into memory
func (p *SVGParser) ParseFile(filename string) (*SVG, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open SVG file: %w", err)
	}
	defer file.Close()
	p.decoder = xml.NewDecoder(file)
	return p.parse()
}

func (p *SVGParser) parse() (*SVG, error) {
	var svg SVG
	err := p.decoder.Decode(&svg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SVG: %w", err)
	}
	return &svg, nil
}

// FindAllPaths recursively finds all paths in the SVG
func (s *SVG) FindAllPaths() []Path {
	var paths []Path
	for i := range s.Groups {
		paths = append(paths, s.Groups[i].findPaths()...)
	}
	return paths
}

func (g *Group) findPaths() []Path {
	var paths []Path
	paths = append(paths, g.Paths...)
	for i := range g.Groups {
		paths = append(paths, g.Groups[i].findPaths()...)
	}
	return paths
}

// FindAllEllipses recursively finds all ellipses in the SVG
func (s *SVG) FindAllEllipses() []Ellipse {
	var ellipses []Ellipse
	for i := range s.Groups {
		ellipses = append(ellipses, s.Groups[i].findEllipses()...)
	}
	return ellipses
}

func (g *Group) findEllipses() []Ellipse {
	var ellipses []Ellipse
	ellipses = append(ellipses, g.Ellipses...)
	for i := range g.Groups {
		ellipses = append(ellipses, g.Groups[i].findEllipses()...)
	}
	return ellipses
}

// FindAllGroups recursively finds all groups in the SVG
func (s *SVG) FindAllGroups() []Group {
	var groups []Group
	for i := range s.Groups {
		groups = append(groups, s.Groups[i])
		groups = append(groups, s.Groups[i].findGroups()...)
	}
	return groups
}

func (g *Group) findGroups() []Group {
	var groups []Group
	groups = append(groups, g.Groups...)
	for i := range g.Groups {
		groups = append(groups, g.Groups[i].findGroups()...)
	}
	return groups
}

// GetViewBox parses the viewBox attribute into its components
// Returns min-x, min-y, width, height
func (s *SVG) GetViewBox() (float64, float64, float64, float64, error) {
	if s.ViewBox == "" {
		return 0, 0, 0, 0, fmt.Errorf("no viewBox defined")
	}
	parts := make([]float64, 0, 4)
	// Simple parsing - split by whitespace or comma
	for _, part := range splitViewBox(s.ViewBox) {
		var val float64
		_, err := fmt.Sscanf(part, "%f", &val)
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("invalid viewBox value: %s", part)
		}
		parts = append(parts, val)
	}
	if len(parts) != 4 {
		return 0, 0, 0, 0, fmt.Errorf("viewBox must have 4 values, got %d", len(parts))
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

// splitViewBox splits a viewBox string into parts
func splitViewBox(s string) []string {
	var parts []string
	var current string
	for _, r := range s {
		if r == ' ' || r == ',' || r == '\t' || r == '\n' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
