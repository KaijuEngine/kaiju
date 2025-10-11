/******************************************************************************/
/* ray.go                                                                     */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
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

import (
	"kaiju/matrix"
)

type Ray struct {
	Origin    matrix.Vec3
	Direction matrix.Vec3
}

// Point returns the point at the given distance along the ray
func (r Ray) Point(distance float32) matrix.Vec3 {
	return r.Origin.Add(r.Direction.Scale(distance))
}

// TriangleHit returns true if the ray hits the triangle defined by the three points
func (r Ray) TriangleHit(rayLen float32, a, b, c matrix.Vec3) bool {
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
func (r Ray) SphereHit(center matrix.Vec3, radius, maxLen float32) bool {
	delta := center.Subtract(r.Origin)
	lenght := matrix.Vec3Dot(r.Direction, delta)
	if lenght < 0 || lenght > (maxLen+radius) {
		return false
	}
	d2 := matrix.Vec3Dot(delta, delta) - lenght*lenght
	return d2 <= (radius * radius)
}
