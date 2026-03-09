package vector_graphics

import "kaijuengine.com/matrix"

type Polygon struct {
	ShapeBase
	Points []PolygonPoint
}

type PolygonPoint struct {
	Handle PointCurve
	Point  matrix.Vec2
}

func (e *Polygon) Animate(animType AnimatedValueType, value float64) {
	if animType < AnimatedValueTypePolygonPoint {
		e.ShapeBase.Animate(animType, value)
		return
	}
	// The point is encoded into AnimatedValueTypePolygonPoint, see comment
	// under the declaration of AnimatedValueTypePolygonPoint
	position := animType - AnimatedValueTypePolygonPoint
	index := position / 2
	component := position % 2
	e.Points[index].Point[component] = matrix.Float(value)
}

func (e *Polygon) ToPolygon(density int) []matrix.Vec2 {
	// Generate a polygonal approximation by sampling points along the cubic
	// Bézier curves defined by each point's handle characters.
	// `density` defines the number of points per curve segment (including the
	// start and end points). A minimum of 2 points per segment is enforced.
	if len(e.Points) == 0 {
		return []matrix.Vec2{}
	}
	if density < 2 {
		density = 2
	}
	// Pre‑allocate with an estimated capacity. We will generate `density`
	// points for each segment, but we skip the duplicate end point of each
	// segment (except the final one).
	capacity := density * len(e.Points)
	out := make([]matrix.Vec2, 0, capacity)
	segments := len(e.Points)
	for i := 0; i < segments; i++ {
		next := (i + 1) % segments
		p0 := e.Points[i].Point
		p1 := p0.Add(e.Points[i].Handle.RightCharacter)
		p3 := e.Points[next].Point
		p2 := p3.Add(e.Points[next].Handle.LeftCharacter)
		for j := 0; j < density; j++ {
			// t ranges from 0 to 1 inclusive.
			t := matrix.Float(j) / matrix.Float(density-1)
			// Cubic Bézier formula.
			oneMinusT := matrix.Float(1) - t
			a := p0.Scale(oneMinusT * oneMinusT * oneMinusT)
			b := p1.Scale(3 * oneMinusT * oneMinusT * t)
			c := p2.Scale(3 * oneMinusT * t * t)
			d := p3.Scale(t * t * t)
			pt := a.Add(b).Add(c).Add(d)
			// Skip the last point of the segment to avoid duplicates, unless this
			// is the final segment.
			if j == density-1 && i != segments-1 {
				continue
			}
			out = append(out, pt)
		}
	}
	return out
}
