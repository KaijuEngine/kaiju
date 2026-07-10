/******************************************************************************/
/* triangle.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import "kaijuengine.com/matrix"

type Triangle struct {
	P           Plane
	EdgePlaneBC Plane
	EdgePlaneCA Plane
}

type DetailedTriangle struct {
	Points   [3]matrix.Vec3
	Normal   matrix.Vec3
	Centroid matrix.Vec3
	Radius   matrix.Float
}

func (t DetailedTriangle) Bounds() AABB {
	min := t.Points[0]
	max := t.Points[0]
	for i := 1; i < 3; i++ {
		min = matrix.Vec3Min(min, t.Points[i])
		max = matrix.Vec3Max(max, t.Points[i])
	}
	return NewAABB(min.Add(max).Scale(0.5), max.Subtract(min).Scale(0.5))
}

func (t DetailedTriangle) RayIntersectTest(ray Ray, length matrix.Float, transform *matrix.Transform) (matrix.Vec3, bool) {
	p0, p1, p2 := t.Points[0], t.Points[1], t.Points[2]
	if transform != nil {
		mat := transform.WorldMatrix()
		p0 = mat.TransformPoint(p0)
		p1 = mat.TransformPoint(p1)
		p2 = mat.TransformPoint(p2)
	}
	return ray.TriangleHit(length, p0, p1, p2)
}

// DetailedTriangleFromPoints creates a detailed triangle from three points, a
// detailed triangle is different from a regular triangle in that it contains
// additional information such as the centroid and radius
func DetailedTriangleFromPoints(points [3]matrix.Vec3) DetailedTriangle {
	tri := DetailedTriangle{
		Points:   [3]matrix.Vec3{points[0], points[1], points[2]},
		Normal:   matrix.Vec3Zero(),
		Centroid: matrix.Vec3Zero(),
		Radius:   0.0,
	}
	e0 := tri.Points[2].Subtract(tri.Points[1])
	e1 := tri.Points[0].Subtract(tri.Points[2])
	tri.Normal = matrix.Vec3Cross(e0, e1).Normal()
	tri.Centroid = matrix.Vec3{
		(tri.Points[0].X() + tri.Points[1].X() + tri.Points[2].X()) / 3.0,
		(tri.Points[0].Y() + tri.Points[1].Y() + tri.Points[2].Y()) / 3.0,
		(tri.Points[0].Z() + tri.Points[1].Z() + tri.Points[2].Z()) / 3.0,
	}
	p := [3]matrix.Vec3{
		tri.Centroid.Subtract(tri.Points[0]),
		tri.Centroid.Subtract(tri.Points[1]),
		tri.Centroid.Subtract(tri.Points[2]),
	}
	tri.Radius = max(p[0].Length(), max(p[1].Length(), p[2].Length()))
	return tri
}
