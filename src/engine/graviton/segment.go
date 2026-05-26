/******************************************************************************/
/* segment.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import "kaijuengine.com/matrix"

type Segment struct {
	A matrix.Vec3
	B matrix.Vec3
}

// LineSegmentFromRay creates a line segment from a ray
func LineSegmentFromRay(ray Ray, length float32) Segment {
	return Segment{ray.Origin, ray.Point(length)}
}

// TriangleHit returns true if the segment hits the triangle defined by the three points
func (l Segment) TriangleHit(a, b, c matrix.Vec3) (matrix.Vec3, bool) {
	p := l.A
	q := l.B
	ab := b.Subtract(a)
	ac := c.Subtract(a)
	qp := p.Subtract(q)
	// Compute triangle normal, can be pre-calculated or cached if
	// intersecting multiple segments against the same triangle
	n := matrix.Vec3Cross(ab, ac)
	d := matrix.Vec3Dot(qp, n)
	if d <= 0 {
		return matrix.Vec3{}, false
	}
	ap := p.Subtract(a)
	t := matrix.Vec3Dot(ap, n)
	if t < 0 {
		return matrix.Vec3{}, false
	}
	e := matrix.Vec3Cross(qp, ap)
	v := matrix.Vec3Dot(ac, e)
	if v < 0 || v > d {
		return matrix.Vec3{}, false
	}
	w := -matrix.Vec3Dot(ab, e)
	if w < 0 || v+w > d {
		return matrix.Vec3{}, false
	}
	s := t / d
	dir := q.Subtract(p)
	hit := p.Add(dir.Scale(s))
	return hit, true
}
