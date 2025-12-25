/******************************************************************************/
/* triangle.go                                                                */
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

package collision

import "kaiju/matrix"

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
	return AABB{
		Center: min.Add(max).Scale(0.5),
		Extent: max.Subtract(min).Scale(0.5),
	}
}

func (t DetailedTriangle) RayIntersectTest(ray Ray, length float32, transform *matrix.Transform) (matrix.Vec3, bool) {
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
