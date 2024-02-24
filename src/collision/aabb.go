/******************************************************************************/
/* aabb.go                                                                    */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package collision

import (
	"kaiju/matrix"
	"math"
	"unsafe"
)

type AABB struct {
	Center matrix.Vec3
	Extent matrix.Vec3
}

func AABBFromMinMax(min, max matrix.Vec3) AABB {
	return AABB{
		Center: min.Add(max).Scale(0.5),
		Extent: max.Subtract(min).Scale(0.5),
	}
}

func (box *AABB) Min() matrix.Vec3 { return box.Center.Subtract(box.Extent) }
func (box *AABB) Max() matrix.Vec3 { return box.Center.Add(box.Extent) }

func (box *AABB) RayHit(ray Ray) (matrix.Vec3, bool) {
	tMin := matrix.Float(0)
	tMax := matrix.Inf(1)
	o := ray.Origin
	d := ray.Direction
	c := box.Center
	e := box.Extent
	for i := 0; i < 3; i++ {
		bMin := c[i] - e[i]
		bMax := c[i] + e[i]
		if matrix.Abs(d[i]) < math.SmallestNonzeroFloat64 {
			if o[i] < bMin || o[i] > bMax {
				return matrix.Vec3{}, false
			}
		} else {
			ood := 1.0 / d[i]
			t1 := (bMin - o[i]) * ood
			t2 := (bMax - o[i]) * ood
			if t1 > t2 {
				t1, t2 = t2, t1
			}
			tMin = max(tMin, t1)
			tMax = min(tMax, t2)
			if tMin > tMax {
				return matrix.Vec3{}, false
			}
		}
	}
	hit := ray.Direction.Scale(tMin).Add(ray.Origin)
	return hit, true
}

func (box *AABB) Contains(point matrix.Vec3) bool {
	return point.X() >= (box.Center.X()-box.Extent.X()) &&
		point.X() <= (box.Center.X()+box.Extent.X()) &&
		point.Y() >= (box.Center.Y()-box.Extent.Y()) &&
		point.Y() <= (box.Center.Y()+box.Extent.Y()) &&
		point.Z() >= (box.Center.Z()-box.Extent.Z()) &&
		point.Z() <= (box.Center.Z()+box.Extent.Z())
}

func (a *AABB) AABBIntersect(b AABB) bool {
	return matrix.Abs(a.Center.X()-b.Center.X()) <= (a.Extent.X()+b.Extent.X()) &&
		matrix.Abs(a.Center.Y()-b.Center.Y()) <= (a.Extent.Y()+b.Extent.Y()) &&
		matrix.Abs(a.Center.Z()-b.Center.Z()) <= (a.Extent.Z()+b.Extent.Z())
}

func (box *AABB) PlaneIntersect(plane Plane) bool {
	r := box.Extent.X()*matrix.Abs(plane.Normal.X()) +
		box.Extent.Y()*matrix.Abs(plane.Normal.Y()) +
		box.Extent.Z()*matrix.Abs(plane.Normal.Z())
	dist := matrix.Vec3Dot(plane.Normal, box.Center) - plane.Dot
	return matrix.Abs(dist) <= r
}

func (box *AABB) TriangleIntersect(tri DetailedTriangle) bool {
	var p0, p1, p2, r matrix.Float

	// Translate triangle as conceptually moving AABB to origin
	tri.Points[0].SubtractAssign(box.Center)
	tri.Points[1].SubtractAssign(box.Center)
	tri.Points[2].SubtractAssign(box.Center)
	tri.Centroid.SubtractAssign(box.Center)
	t := tri.Points

	// Quick radius check to exit early
	bRad := max(box.Extent.X(), box.Extent.Y(), box.Extent.Z())
	cLen := tri.Centroid.Length()
	if cLen > (bRad + tri.Radius) {
		return false
	}

	// Compute edge vectors for triangle
	e0 := t[1].Subtract(t[0])
	e1 := t[2].Subtract(t[1])
	e2 := t[0].Subtract(t[2])

	a00 := matrix.Vec3Cross(matrix.Vec3Right(), e0)
	a01 := matrix.Vec3Cross(matrix.Vec3Right(), e1)
	a02 := matrix.Vec3Cross(matrix.Vec3Right(), e2)
	a10 := matrix.Vec3Cross(matrix.Vec3Up(), e0)
	a11 := matrix.Vec3Cross(matrix.Vec3Up(), e1)
	a12 := matrix.Vec3Cross(matrix.Vec3Up(), e2)
	a20 := matrix.Vec3Cross(matrix.Vec3Forward(), e0)
	a21 := matrix.Vec3Cross(matrix.Vec3Forward(), e1)
	a22 := matrix.Vec3Cross(matrix.Vec3Forward(), e2)

	// p0 == p1 due to AABB
	p0 = matrix.Vec3Dot(t[0], a00)
	p1 = matrix.Vec3Dot(t[1], a00)
	p2 = matrix.Vec3Dot(t[2], a00)
	r = box.Extent.X()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Right(), a00)) +
		box.Extent.Y()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Up(), a00)) +
		box.Extent.Z()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Forward(), a00))
	// TODO:  The above simplifies to the below comment
	//r = extents.Y() * matrix.Abs(e0.Z()) + extents.Z() * matrix.Abs(e0.Y());
	if max(-max(p0, max(p1, p2)), min(p0, min(p1, p2))) > r {
		return false
	}

	// TODO:  p0 = p1 so we can remove one of them, this holds true with different
	// combinations of p0, p1, and p2 in the remaining similar blocks
	p0 = matrix.Vec3Dot(t[0], a01)
	p1 = matrix.Vec3Dot(t[1], a01)
	p2 = matrix.Vec3Dot(t[2], a01)
	r = box.Extent.X()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Right(), a01)) +
		box.Extent.Y()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Up(), a01)) +
		box.Extent.Z()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Forward(), a01))
	if max(-max(p0, max(p1, p2)), min(p0, min(p1, p2))) > r {
		return false
	}

	p0 = matrix.Vec3Dot(t[0], a02)
	p1 = matrix.Vec3Dot(t[1], a02)
	p2 = matrix.Vec3Dot(t[2], a02)
	r = box.Extent.X()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Right(), a02)) +
		box.Extent.Y()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Up(), a02)) +
		box.Extent.Z()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Forward(), a02))
	if max(-max(p0, max(p1, p2)), min(p0, min(p1, p2))) > r {
		return false
	}

	p0 = matrix.Vec3Dot(t[0], a10)
	p1 = matrix.Vec3Dot(t[1], a10)
	p2 = matrix.Vec3Dot(t[2], a10)
	r = box.Extent.X()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Right(), a10)) +
		box.Extent.Y()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Up(), a10)) +
		box.Extent.Z()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Forward(), a10))
	if max(-max(p0, max(p1, p2)), min(p0, min(p1, p2))) > r {
		return false
	}

	p0 = matrix.Vec3Dot(t[0], a11)
	p1 = matrix.Vec3Dot(t[1], a11)
	p2 = matrix.Vec3Dot(t[2], a11)
	r = box.Extent.X()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Right(), a11)) +
		box.Extent.Y()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Up(), a11)) +
		box.Extent.Z()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Forward(), a11))
	if max(-max(p0, max(p1, p2)), min(p0, min(p1, p2))) > r {
		return false
	}

	p0 = matrix.Vec3Dot(t[0], a12)
	p1 = matrix.Vec3Dot(t[1], a12)
	p2 = matrix.Vec3Dot(t[2], a12)
	r = box.Extent.X()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Right(), a12)) +
		box.Extent.Y()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Up(), a12)) +
		box.Extent.Z()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Forward(), a12))
	if max(-max(p0, max(p1, p2)), min(p0, min(p1, p2))) > r {
		return false
	}

	p0 = matrix.Vec3Dot(t[0], a20)
	p1 = matrix.Vec3Dot(t[1], a20)
	p2 = matrix.Vec3Dot(t[2], a20)
	r = box.Extent.X()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Right(), a20)) +
		box.Extent.Y()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Up(), a20)) +
		box.Extent.Z()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Forward(), a20))
	if max(-max(p0, max(p1, p2)), min(p0, min(p1, p2))) > r {
		return false
	}

	p0 = matrix.Vec3Dot(t[0], a21)
	p1 = matrix.Vec3Dot(t[1], a21)
	p2 = matrix.Vec3Dot(t[2], a21)
	r = box.Extent.X()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Right(), a21)) +
		box.Extent.Y()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Up(), a21)) +
		box.Extent.Z()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Forward(), a21))
	if max(-max(p0, max(p1, p2)), min(p0, min(p1, p2))) > r {
		return false
	}

	p0 = matrix.Vec3Dot(t[0], a22)
	p1 = matrix.Vec3Dot(t[1], a22)
	p2 = matrix.Vec3Dot(t[2], a22)
	r = box.Extent.X()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Right(), a22)) +
		box.Extent.Y()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Up(), a22)) +
		box.Extent.Z()*matrix.Abs(matrix.Vec3Dot(matrix.Vec3Forward(), a22))
	if max(-max(p0, max(p1, p2)), min(p0, min(p1, p2))) > r {
		return false
	}

	if max(t[0].X(), max(t[1].X(), t[2].X())) < -box.Extent.X() ||
		min(t[0].X(), min(t[1].X(), t[2].X())) > box.Extent.X() {
		return false
	} else if max(t[0].Y(), max(t[1].Y(), t[2].Y())) < -box.Extent.Y() ||
		min(t[0].Y(), min(t[1].Y(), t[2].Y())) > box.Extent.Y() {
		return false
	} else if max(t[0].Z(), max(t[1].Z(), t[2].Z())) < -box.Extent.Z() ||
		min(t[0].Z(), min(t[1].Z(), t[2].Z())) > box.Extent.Z() {
		return false
	}

	p := Plane{
		Normal: matrix.Vec3Cross(e0, e1),
	}
	p.Dot = matrix.Vec3Dot(p.Normal, t[0])
	return box.PlaneIntersect(p)
}

func (box *AABB) FromTriangle(triangle DetailedTriangle) AABB {
	tMin := matrix.Vec3Min(triangle.Points[0],
		triangle.Points[1], triangle.Points[2])
	tMax := matrix.Vec3Max(triangle.Points[0],
		triangle.Points[1], triangle.Points[2])
	mid := tMax.Add(tMin).Scale(0.5)
	e := tMax.Subtract(mid)
	return AABB{mid, e}
}

func (a *AABB) FromAABB(b AABB) AABB {
	center := a.Center.Add(b.Center).Scale(0.5)
	aMax := a.Center.Add(a.Extent)
	bMax := b.Center.Add(b.Extent)
	aMin := a.Center.Subtract(a.Extent)
	bMin := b.Center.Subtract(b.Extent)
	mMax := matrix.Vec3MaxAbs(aMax, bMax)
	mMin := matrix.Vec3MinAbs(aMin, bMin)
	e := matrix.Vec3MaxAbs(mMax, mMin)
	return AABB{center, e}
}

func (box *AABB) InFrustum(frustum Frustum) bool {
	min := box.Min()
	max := box.Max()
	for i := 0; i < 6; i++ {
		out := 0
		pv := *(*matrix.Vec4)(unsafe.Pointer(&frustum.Planes[i].Normal))
		if matrix.Vec4Dot(pv, matrix.Vec4{min.X(), min.Y(), min.Z(), 1}) < 0 {
			out += 1
		}
		if matrix.Vec4Dot(pv, matrix.Vec4{max.X(), min.Y(), min.Z(), 1}) < 0 {
			out += 1
		}
		if matrix.Vec4Dot(pv, matrix.Vec4{min.X(), max.Y(), min.Z(), 1}) < 0 {
			out += 1
		}
		if matrix.Vec4Dot(pv, matrix.Vec4{max.X(), max.Y(), min.Z(), 1}) < 0 {
			out += 1
		}
		if matrix.Vec4Dot(pv, matrix.Vec4{min.X(), min.Y(), max.Z(), 1}) < 0 {
			out += 1
		}
		if matrix.Vec4Dot(pv, matrix.Vec4{max.X(), min.Y(), max.Z(), 1}) < 0 {
			out += 1
		}
		if matrix.Vec4Dot(pv, matrix.Vec4{min.X(), max.Y(), max.Z(), 1}) < 0 {
			out += 1
		}
		if matrix.Vec4Dot(pv, matrix.Vec4{max.X(), max.Y(), max.Z(), 1}) < 0 {
			out += 1
		}
		if out == 8 {
			return false
		}
	}
	// TODO:  Uncomment for large object calculations
	// check frustum outside/inside box
	//int out;
	//out = 0;
	//for (int i = 0; i < 8; ++i) out += ((frustum->planePositions[i].x > max.x) ? 1 : 0);
	//if (out == 8)
	//	return false;
	//out = 0;
	//for (int i = 0; i < 8; ++i) out += ((frustum->planePositions[i].x < min.x) ? 1 : 0);
	//if (out == 8)
	//	return false;
	//out = 0;
	//for (int i = 0; i < 8; ++i) out += ((frustum->planePositions[i].y > max.y) ? 1 : 0);
	//if (out == 8)
	//	return false;
	//out = 0;
	//for (int i = 0; i < 8; ++i) out += ((frustum->planePositions[i].y < min.y) ? 1 : 0);
	//if (out == 8)
	//	return false;
	//out = 0;
	//for (int i = 0; i < 8; ++i) out += ((frustum->planePositions[i].z > max.z) ? 1 : 0);
	//if (out == 8)
	//	return false;
	//out = 0;
	//for (int i = 0; i < 8; ++i) out += ((frustum->planePositions[i].z < min.z) ? 1 : 0);
	//if (out == 8)
	//	return false;
	return true
}
