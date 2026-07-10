/******************************************************************************/
/* ray.go                                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"kaijuengine.com/matrix"
)

type Ray struct {
	Origin    matrix.Vec3
	Direction matrix.Vec3
}

// Point returns the point at the given distance along the ray
func (r Ray) Point(distance matrix.Float) matrix.Vec3 {
	return r.Origin.Add(r.Direction.Scale(distance))
}

// TriangleHit returns true if the ray hits the triangle defined by the three points
func (r Ray) TriangleHit(rayLen matrix.Float, a, b, c matrix.Vec3) (matrix.Vec3, bool) {
	s := Segment{r.Origin, r.Point(rayLen)}
	return s.TriangleHit(a, b, c)
}

// PlaneHit returns the point of intersection with the plane and true if the ray hits the plane
func (r Ray) PlaneHit(planePosition, planeNormal matrix.Vec3) (hit matrix.Vec3, success bool) {
	hit = matrix.Vec3{}
	success = false
	d := matrix.Vec3Dot(planeNormal, r.Direction)
	if matrix.Abs(d) < matrix.FloatSmallestNonzero {
		return
	}
	diff := planePosition.Subtract(r.Origin)
	distance := matrix.Vec3Dot(diff, planeNormal) / d
	if distance < 0 {
		return
	}
	return r.Point(distance), true
}

// SphereHit returns true if the ray hits the sphere
func (r Ray) SphereHit(center matrix.Vec3, radius, maxLen matrix.Float) bool {
	delta := center.Subtract(r.Origin)
	lenght := matrix.Vec3Dot(r.Direction, delta)
	if lenght < 0 || lenght > (maxLen+radius) {
		return false
	}
	d2 := matrix.Vec3Dot(delta, delta) - lenght*lenght
	return d2 <= (radius * radius)
}
