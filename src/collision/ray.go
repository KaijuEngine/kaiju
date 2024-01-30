package collision

import (
	"kaiju/matrix"
)

type Ray struct {
	Origin    matrix.Vec3
	Direction matrix.Vec3
}

func (r Ray) Point(distance float32) matrix.Vec3 {
	return r.Origin.Add(r.Direction.Scale(distance))
}

func (r Ray) TriangleHit(rayLen float32, a, b, c matrix.Vec3) bool {
	s := Segment{r.Origin, r.Point(rayLen)}
	return s.TriangleHit(a, b, c)
}

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

func (r Ray) SphereHit(center matrix.Vec3, radius, maxLen float32) bool {
	delta := center.Subtract(r.Origin)
	lenght := matrix.Vec3Dot(r.Direction, delta)
	if lenght < 0 || lenght > (maxLen+radius) {
		return false
	}
	d2 := matrix.Vec3Dot(delta, delta) - lenght*lenght
	return d2 <= (radius * radius)
}
