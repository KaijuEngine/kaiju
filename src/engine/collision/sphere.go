/******************************************************************************/
/* sphere.go                                                                  */
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

import (
	"kaiju/klib"
	"kaiju/matrix"
)

type Sphere struct {
	Position matrix.Vec3
	Radius   float32
}

func (a Sphere) Overlap(b Sphere) bool {
	distSq := a.Position.SquareDistance(b.Position)
	radiusSum := a.Radius + b.Radius
	radiusSumSq := radiusSum * radiusSum
	return distSq <= radiusSumSq
}

func (s Sphere) IntersectsAABB(b AABB) bool {
	var sqDist float32
	minX := b.Center.X() - b.Extent.X()
	maxX := b.Center.X() + b.Extent.X()
	x := s.Position.X()
	if x < minX {
		d := x - minX
		sqDist += d * d
	} else if x > maxX {
		d := x - maxX
		sqDist += d * d
	}
	minY := b.Center.Y() - b.Extent.Y()
	maxY := b.Center.Y() + b.Extent.Y()
	y := s.Position.Y()
	if y < minY {
		d := y - minY
		sqDist += d * d
	} else if y > maxY {
		d := y - maxY
		sqDist += d * d
	}
	minZ := b.Center.Z() - b.Extent.Z()
	maxZ := b.Center.Z() + b.Extent.Z()
	z := s.Position.Z()
	if z < minZ {
		d := z - minZ
		sqDist += d * d
	} else if z > maxZ {
		d := z - maxZ
		sqDist += d * d
	}
	return sqDist <= s.Radius*s.Radius
}

func (s Sphere) IntersectsOOBB(b OOBB) bool {
	p := s.Position.Subtract(b.Center)
	local := b.Orientation.Transpose().MultiplyVec3(p)
	clamped := matrix.NewVec3(
		klib.Clamp(local.X(), -b.Extent.X(), b.Extent.X()),
		klib.Clamp(local.Y(), -b.Extent.Y(), b.Extent.Y()),
		klib.Clamp(local.Z(), -b.Extent.Z(), b.Extent.Z()),
	)
	diff := local.Subtract(clamped)
	distSq := diff.X()*diff.X() + diff.Y()*diff.Y() + diff.Z()*diff.Z()
	return distSq <= s.Radius*s.Radius
}

func (s Sphere) IntersectsRay(r Ray) (bool, float32) {
	m := r.Origin.Subtract(s.Position)
	a := r.Direction.X()*r.Direction.X() +
		r.Direction.Y()*r.Direction.Y() +
		r.Direction.Z()*r.Direction.Z()
	b := 2.0 * (r.Direction.X()*m.X() + r.Direction.Y()*m.Y() + r.Direction.Z()*m.Z())
	c := m.X()*m.X() + m.Y()*m.Y() + m.Z()*m.Z() - s.Radius*s.Radius
	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return false, 0
	}
	sqrtDisc := float32(matrix.Sqrt(discriminant))
	t := (-b - sqrtDisc) / (2 * a)
	if t < 0 {
		t = (-b + sqrtDisc) / (2 * a)
		if t < 0 {
			return false, 0
		}
	}
	return true, t
}

func (s Sphere) IntersectsPlane(p Plane) (bool, float32) {
	dist := matrix.Vec3Dot(s.Position, p.Normal) + p.Dot
	if dist < 0 {
		dist = -dist
	}
	if dist <= s.Radius {
		return true, s.Radius - dist
	}
	return false, 0
}

func (s Sphere) IntersectsFrustum(f Frustum) bool {
	for _, p := range f.Planes {
		dist := matrix.Vec3Dot(s.Position, p.Normal) + p.Dot
		if dist < -s.Radius {
			return false
		}
	}
	return true
}
