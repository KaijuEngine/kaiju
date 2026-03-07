package vector_graphics

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func PolygonToMesh(name string, points []matrix.Vec2) *rendering.Mesh {
	verts := make([]rendering.Vertex, len(points))
	for i := range points {
		verts[i] = rendering.Vertex{
			Position: matrix.NewVec3(points[i].X(), points[i].Y(), 0),
			Color:    matrix.ColorWhite(),
			Normal:   matrix.Vec3{0.0, 0.0, 1.0},
		}
	}
	return rendering.NewMesh(name, verts, pointsToPolygonIndices(points))
}

// pointsToPolygonIndices triangulates a polygon defined by a slice of 3D
// points. It returns an index buffer suitable for rendering a triangle list.
// The implementation uses an ear‑clipping algorithm which works for both convex
// and non‑convex simple polygons (no self‑intersections). The points are
// assumed to lie on a common plane; the algorithm projects them onto the XY
// plane for the geometric tests. The winding order of the input determines the
// orientation – the algorithm automatically detects clockwise or
// counter‑clockwise winding.
func pointsToPolygonIndices(points []matrix.Vec2) []uint32 {
	n := len(points)
	if n < 3 {
		return nil
	}

	// Helper: compute signed area (2 * area) of the polygon projected onto XY.
	signedArea := func(pts []matrix.Vec2) float32 {
		var area float32
		for i := 0; i < len(pts); i++ {
			j := (i + 1) % len(pts)
			area += pts[i][0]*pts[j][1] - pts[j][0]*pts[i][1]
		}
		return area * 0.5
	}

	// Determine orientation: positive area => CCW, negative => CW.
	ccw := signedArea(points) > 0

	// Helper: check if point p is inside triangle (a,b,c) using barycentric.
	pointInTri := func(p, a, b, c matrix.Vec2) bool {
		// Vectors
		v0 := matrix.Vec2{c[0] - a[0], c[1] - a[1]}
		v1 := matrix.Vec2{b[0] - a[0], b[1] - a[1]}
		v2 := matrix.Vec2{p[0] - a[0], p[1] - a[1]}

		dot00 := v0[0]*v0[0] + v0[1]*v0[1]
		dot01 := v0[0]*v1[0] + v0[1]*v1[1]
		dot02 := v0[0]*v2[0] + v0[1]*v2[1]
		dot11 := v1[0]*v1[0] + v1[1]*v1[1]
		dot12 := v1[0]*v2[0] + v1[1]*v2[1]

		invDenom := 1.0 / (dot00*dot11 - dot01*dot01)
		u := (dot11*dot02 - dot01*dot12) * invDenom
		v := (dot00*dot12 - dot01*dot02) * invDenom
		return u >= 0 && v >= 0 && u+v <= 1
	}

	// Helper: determine if the triangle (prev, cur, next) is an ear.
	isEar := func(prevIdx, curIdx, nextIdx int, idxList []int) bool {
		a := points[prevIdx]
		b := points[curIdx]
		c := points[nextIdx]

		// Compute cross product Z to test convexity respecting winding.
		cross := (b[0]-a[0])*(c[1]-a[1]) - (b[1]-a[1])*(c[0]-a[0])
		if ccw {
			if cross <= 0 {
				return false // not a convex corner
			}
		} else {
			if cross >= 0 {
				return false
			}
		}

		// Ensure no other vertex lies inside the triangle.
		for _, vi := range idxList {
			if vi == prevIdx || vi == curIdx || vi == nextIdx {
				continue
			}
			if pointInTri(points[vi], a, b, c) {
				return false
			}
		}
		return true
	}

	// Initialize a list of vertex indices.
	idxList := make([]int, n)
	for i := 0; i < n; i++ {
		idxList[i] = i
	}

	var triangles []uint32
	// addTriangle appends a triangle ensuring the winding is counter‑clockwise.
	// If the original polygon is clockwise (ccw == false) we swap the last two
	// indices to reverse the winding.
	addTriangle := func(a, b, c uint32) {
		if ccw {
			triangles = append(triangles, a, b, c)
		} else {
			// Reverse order to make it CCW.
			triangles = append(triangles, a, c, b)
		}
	}
	// Ear clipping loop.
	for len(idxList) > 3 {
		earFound := false
		// iterate over vertices to find an ear
		for i := 0; i < len(idxList); i++ {
			prev := idxList[(i-1+len(idxList))%len(idxList)]
			cur := idxList[i]
			next := idxList[(i+1)%len(idxList)]
			if isEar(prev, cur, next, idxList) {
				// Append triangle indices (as uint32) ensuring CCW winding
				addTriangle(uint32(prev), uint32(cur), uint32(next))
				// Remove ear vertex (cur) from the list
				idxList = append(idxList[:i], idxList[i+1:]...)
				earFound = true
				break
			}
		}
		if !earFound {
			// Polygon may be degenerate; break to avoid infinite loop.
			break
		}
	}
	// Remaining triangle
	if len(idxList) == 3 {
		addTriangle(uint32(idxList[0]), uint32(idxList[1]), uint32(idxList[2]))
	}
	return triangles
}
