package vector_graphics

import (
	"kaiju/matrix"
	"math"
)

// StrokeToPolygons converts stroked paths of a VectorGraphic into filled
// polygons suitable for rasterisation. The implementation currently handles
// only straight line segments (MoveTo, LineTo, ClosePath) and butt caps. Curve
// commands are ignored (treated as straight lines) and more advanced join
// styles are approximated with simple offsets.
func (v *VectorGraphic) StrokeToPolygons() [][]matrix.Vec3 {
	var polygons [][]matrix.Vec3
	// Recursively process groups and their paths.
	var walkGroup func(g Group)
	walkGroup = func(g Group) {
		// Process paths in this group.
		for _, p := range g.Paths {
			poly := strokePathToPolygon(p)
			if len(poly) > 0 {
				polygons = append(polygons, poly)
			}
		}
		// Recurse into sub‑groups.
		for _, sub := range g.Groups {
			walkGroup(sub)
		}
	}
	for _, grp := range v.Groups {
		walkGroup(grp)
	}
	return polygons
}

// strokePathToPolygon tessellates a single Path into a polygon. It returns a
// slice of Vec3 where Z = 0. The algorithm:
//  1. Flatten the path into a list of absolute Vec2 points.
//  2. For each segment compute a perpendicular normal and offset the points
//     left/right by half the stroke width.
//  3. Build the polygon by walking the left side forward and the right side
//     backward, closing the loop.
func strokePathToPolygon(p Path) []matrix.Vec3 {
	if p.StrokeWidth <= 0 {
		return nil
	}
	pts := flattenPath(p)
	if len(pts) < 2 {
		return nil
	}
	half := float64(p.StrokeWidth) / 2.0
	// Left and right offset vertices.
	left := make([]matrix.Vec2, 0, len(pts))
	right := make([]matrix.Vec2, 0, len(pts))
	// For each segment, compute normal and offset the start point.
	for i := 0; i < len(pts)-1; i++ {
		a := pts[i]
		b := pts[i+1]
		n := segmentNormal(a, b)
		// left side = point + normal*half, right side = point - normal*half
		left = append(left, offsetPoint(a, n, half, true))
		right = append(right, offsetPoint(a, n, half, false))
		// For the last point of the path we will handle after the loop.
		if i == len(pts)-2 {
			// offset the final point using the same normal as the last segment.
			left = append(left, offsetPoint(b, n, half, true))
			right = append(right, offsetPoint(b, n, half, false))
		}
	}
	// Assemble polygon: left side forward, right side reverse.
	var poly []matrix.Vec3
	for _, v2 := range left {
		poly = append(poly, matrix.NewVec3(v2.X(), v2.Y(), 0))
	}
	// reverse right side
	for i := len(right) - 1; i >= 0; i-- {
		v2 := right[i]
		poly = append(poly, matrix.NewVec3(v2.X(), v2.Y(), 0))
	}
	// Ensure polygon is closed (first == last). If not, append first.
	if len(poly) > 0 && (poly[0][0] != poly[len(poly)-1][0] || poly[0][1] != poly[len(poly)-1][1]) {
		poly = append(poly, poly[0])
	}
	return poly
}

// flattenPath converts a Path's segment data into a slice of absolute Vec2
// points. Only MoveTo, LineTo and ClosePath are interpreted. Curves are ignored.
func flattenPath(p Path) []matrix.Vec2 {
	var pts []matrix.Vec2
	var cur matrix.Vec2
	for _, seg := range p.Data {
		switch seg.Cmd {
		case PathCmdMoveTo:
			cur = matrix.NewVec2(matrix.Float(seg.Params[0]), matrix.Float(seg.Params[1]))
			pts = append(pts, cur)
		case PathCmdLineTo:
			cur = matrix.NewVec2(matrix.Float(seg.Params[0]), matrix.Float(seg.Params[1]))
			pts = append(pts, cur)
		case PathCmdClosePath:
			// No additional point needed; polygon will be closed later.
		default:
			// Unsupported command – ignore for now.
		}
	}
	return pts
}

// segmentNormal returns a unit perpendicular vector to the segment a->b.
func segmentNormal(a, b matrix.Vec2) matrix.Vec2 {
	dx := b.X() - a.X()
	dy := b.Y() - a.Y()
	// Perpendicular ( -dy, dx )
	n := matrix.NewVec2(-dy, dx)
	length := math.Hypot(float64(n.X()), float64(n.Y()))
	if length == 0 {
		return matrix.NewVec2(0, 0)
	}
	inv := 1.0 / length
	return matrix.NewVec2(n.X()*matrix.Float(inv), n.Y()*matrix.Float(inv))
}

// offsetPoint returns the point shifted by normal*halfWidth. If left is true the
// shift is in the direction of the normal, otherwise opposite.
func offsetPoint(p, n matrix.Vec2, halfWidth float64, left bool) matrix.Vec2 {
	factor := matrix.Float(halfWidth)
	if !left {
		factor = -factor
	}
	return matrix.Vec2{p.X() + n.X()*factor, p.Y() + n.Y()*factor}
}
