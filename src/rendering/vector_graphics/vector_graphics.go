package vector_graphics

import (
	"fmt"
	"strconv"
	"strings"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering/vector_graphics/svg"
)

type VectorGraphic struct {
	Shapes     []Shape
	Animations []Animation
}

// VectorGraphicFromSVG converts an SVG structure into the engine's VectorGraphic.
// It walks all supported SVG elements, creates matching shape structs, copies
// visual attributes, and extracts simple SMIL <animate> and <animateTransform>
// animations into the VectorGraphic's Animation slice.
func VectorGraphicFromSVG(svg svg.SVG) VectorGraphic {
	var vg VectorGraphic

	// ----- Circles -------------------------------------------------
	for _, c := range svg.FindAllCircles() {
		shape := &Circle{
			Center: matrix.NewVec3(matrix.Float(c.CX), matrix.Float(c.CY), 0),
			Radius: matrix.Float(c.R),
		}
		// Base visual attributes
		shape.Stroke = Color(parseColor(c.Stroke))
		shape.Fill = Color(parseColor(c.Fill))
		shape.StrokeWidth = matrix.Float(c.StrokeWidth)
		vg.Shapes = append(vg.Shapes, shape)

		// Animations attached to this circle
		for _, a := range c.Animates {
			if anim, ok := buildAnimation(shape, a); ok {
				vg.Animations = append(vg.Animations, anim)
			}
		}
	}

	// ----- Ellipses ------------------------------------------------
	for _, e := range svg.FindAllEllipses() {
		shape := &Ellipse{
			Center: matrix.NewVec3(matrix.Float(e.CX), matrix.Float(e.CY), 0),
			Radius: matrix.NewVec2(matrix.Float(e.RX), matrix.Float(e.RY)),
		}
		shape.Stroke = Color(parseColor(e.Stroke))
		shape.Fill = Color(parseColor(e.Fill))
		shape.StrokeWidth = matrix.Float(e.StrokeWidth)
		vg.Shapes = append(vg.Shapes, shape)
		for _, a := range e.Animates {
			if anim, ok := buildAnimation(shape, a); ok {
				vg.Animations = append(vg.Animations, anim)
			}
		}
	}

	// ----- Rectangles -----------------------------------------------
	for _, r := range svg.FindAllRects() {
		// Compute centre of rectangle
		cx := r.X + r.Width/2
		cy := r.Y + r.Height/2
		shape := &Rectangle{
			Center: matrix.NewVec2(matrix.Float(cx), matrix.Float(cy)),
			Size:   matrix.NewVec2(matrix.Float(r.Width), matrix.Float(r.Height)),
		}
		shape.Stroke = Color(parseColor(r.Stroke))
		shape.Fill = Color(parseColor(r.Fill))
		shape.StrokeWidth = matrix.Float(r.StrokeWidth)
		vg.Shapes = append(vg.Shapes, shape)
		for _, a := range r.Animates {
			if anim, ok := buildAnimation(shape, a); ok {
				vg.Animations = append(vg.Animations, anim)
			}
		}
	}

	// ----- Polygons ------------------------------------------------
	for _, p := range svg.FindAllPolygons() {
		pts := parsePoints(p.Points)
		shape := &Polygon{Points: pts}
		shape.Stroke = Color(parseColor(p.Stroke))
		shape.Fill = Color(parseColor(p.Fill))
		shape.StrokeWidth = matrix.Float(p.StrokeWidth)
		vg.Shapes = append(vg.Shapes, shape)
		for _, a := range p.Animates {
			if anim, ok := buildAnimation(shape, a); ok {
				vg.Animations = append(vg.Animations, anim)
			}
		}
	}

	// ----- Lines (as LineSegment) -----------------------------------
	for _, l := range svg.FindAllLines() {
		shape := &LineSegment{
			From: matrix.NewVec2(matrix.Float(l.X1), matrix.Float(l.Y1)),
			To:   matrix.NewVec2(matrix.Float(l.X2), matrix.Float(l.Y2)),
		}
		shape.Stroke = Color(parseColor(l.Stroke))
		shape.StrokeWidth = matrix.Float(l.StrokeWidth)
		vg.Shapes = append(vg.Shapes, shape)
		for _, a := range l.Animates {
			if anim, ok := buildAnimation(shape, a); ok {
				vg.Animations = append(vg.Animations, anim)
			}
		}
	}

	// Note: Path elements are not yet supported.
	return vg
}

// func (g *VectorGraphic) ToMesh() *rendering.Mesh {
// 	for i := range g.Shapes {
// 		switch s := g.Shapes[i].(type) {
// 		case *LineSegment:
// 		case *Circle:
// 		case *Ellipse:
// 		case *Rectangle:
// 		case *Polygon:
// 		}
// 	}
// }

// Helper: parse a float from string, returning matrix.Float and success flag.
func parseFloat(s string) (matrix.Float, bool) {
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, false
	}
	return matrix.Float(v), true
}

// Helper: parse a CSS color string into a matrix.Color. Supports hex (#RRGGBB, #RRGGBBAA)
// and simple "none" (transparent). For unsupported formats returns transparent.
func parseColor(s string) matrix.Color {
	s = strings.TrimSpace(s)
	if s == "" || s == "none" {
		return matrix.Color{}
	}
	// Hex format (e.g., #RRGGBB or #RRGGBBAA, also short #RGB)
	if strings.HasPrefix(s, "#") {
		hex := strings.TrimPrefix(s, "#")
		// Expand short form #RGB to #RRGGBB
		if len(hex) == 3 {
			hex = fmt.Sprintf("%c%c%c%c%c%c", hex[0], hex[0], hex[1], hex[1], hex[2], hex[2])
		}
		// Accept 6 or 8 hex digits
		if len(hex) == 6 || len(hex) == 8 {
			// Parse each component as uint8
			var r, g, b, a uint8 = 0, 0, 0, 255
			if v, err := strconv.ParseUint(hex[0:2], 16, 8); err == nil {
				r = uint8(v)
			}
			if v, err := strconv.ParseUint(hex[2:4], 16, 8); err == nil {
				g = uint8(v)
			}
			if v, err := strconv.ParseUint(hex[4:6], 16, 8); err == nil {
				b = uint8(v)
			}
			if len(hex) == 8 {
				if v, err := strconv.ParseUint(hex[6:8], 16, 8); err == nil {
					a = uint8(v)
				}
			}
			// Use matrix helper to build a Color from 0‑255 ints
			return matrix.ColorRGBAInt(int(r), int(g), int(b), int(a))
		}
	}
	// Fallback – transparent (zero value)
	return matrix.Color{}
}

// Helper: parse a points attribute (e.g., "0,0 10,0 10,10") into PolygonPoints.
func parsePoints(s string) []PolygonPoint {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	// Split on whitespace or commas, then pair.
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' })
	var pts []PolygonPoint
	for i := 0; i+1 < len(fields); i += 2 {
		x, ok1 := parseFloat(fields[i])
		y, ok2 := parseFloat(fields[i+1])
		if !ok1 || !ok2 {
			continue
		}
		pt := PolygonPoint{Point: matrix.NewVec2(x, y)}
		pts = append(pts, pt)
	}
	return pts
}

// Map SVG attribute names to engine AnimatedValueType constants.
func attributeToAnimType(attr svg.AttributeName) (AnimatedValueType, bool) {
	switch attr {
	case svg.AttrX:
		return AnimatedValueTypePositionX, true
	case svg.AttrY:
		return AnimatedValueTypePositionY, true
	case svg.AttrWidth:
		return AnimatedValueTypeWidth, true
	case svg.AttrHeight:
		return AnimatedValueTypeHeight, true
	case svg.AttrR:
		return AnimatedValueTypeRadius, true
	case svg.AttrCX:
		return AnimatedValueTypePositionX, true
	case svg.AttrCY:
		return AnimatedValueTypePositionY, true
	case svg.AttrStrokeWidth:
		return AnimatedValueTypeStrokeWidth, true
	case svg.AttrFill:
		return AnimatedValueTypeFillR, true // color animation will be handled by Color.Animate
	case svg.AttrStroke:
		return AnimatedValueTypeStrokeR, true
	default:
		return AnimatedValueTypeNone, false
	}
}

// Build an Animation object from an SVG <animate> element.
func buildAnimation(target Shape, a svg.Animate) (Animation, bool) {
	animType, ok := attributeToAnimType(a.AttributeName)
	if !ok {
		return Animation{}, false
	}
	// Simple handling: from/to or values list.
	var keys []AnimationKeyFrame
	// Duration parsing – assume seconds if ends with 's'.
	durSec := 0.0
	if a.Duration != "" {
		if strings.HasSuffix(a.Duration, "s") {
			v, err := strconv.ParseFloat(strings.TrimSuffix(a.Duration, "s"), 64)
			if err == nil {
				durSec = v
			}
		}
	}
	// Helper to add a keyframe at given time with a value.
	addKey := func(t float64, valStr string) {
		v, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return
		}
		keys = append(keys, AnimationKeyFrame{TimeCode: t, Value: v})
	}
	if a.From != "" && a.To != "" {
		addKey(0, a.From)
		if durSec > 0 {
			addKey(durSec, a.To)
		} else {
			addKey(1, a.To) // fallback to 1s
		}
	} else if a.Values != "" {
		vals := strings.Split(a.Values, ";")
		// Determine key times – if not provided, distribute evenly.
		var times []float64
		if a.KeyTimes != "" {
			kt, err := svg.ParseKeyTimes(a.KeyTimes)
			if err == nil {
				times = kt
			}
		}
		if len(times) == 0 {
			// Even distribution over duration (or 1s)
			total := durSec
			if total == 0 {
				total = 1
			}
			step := total / float64(len(vals)-1)
			for i := range vals {
				times = append(times, step*float64(i))
			}
		}
		for i, vStr := range vals {
			if i < len(times) {
				addKey(times[i], vStr)
			}
		}
	}
	if len(keys) == 0 {
		return Animation{}, false
	}
	return Animation{Target: target, Type: animType, Keys: keys}, true
}
