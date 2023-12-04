package collision

import "kaiju/matrix"

type Segment struct {
	A matrix.Vec3
	B matrix.Vec3
}

func LineSegmentFromRay(ray Ray, length float32) Segment {
	return Segment{ray.Origin, ray.Point(length)}
}

func (l Segment) TriangleHit(a, b, c matrix.Vec3) bool {
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
		return false
	}
	ap := p.Subtract(a)
	t := matrix.Vec3Dot(ap, n)
	if t < 0 {
		return false
	}
	e := matrix.Vec3Cross(qp, ap)
	v := matrix.Vec3Dot(ac, e)
	if v < 0 || v > d {
		return false
	}
	w := -matrix.Vec3Dot(ab, e)
	if w < 0 || v+w > d {
		return false
	}
	return true
}
