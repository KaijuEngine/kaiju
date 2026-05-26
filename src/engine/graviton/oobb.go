/******************************************************************************/
/* oobb.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"kaijuengine.com/matrix"
)

type OOBB Shape

func (s *Shape) SetOOBB(center, extent matrix.Vec3, orientation matrix.Mat3) {
	s.Type = ShapeTypeOOBB
	s.Center = center
	s.Extent = extent
	s.Orientation = orientation
}

func NewOOBB(center, extent matrix.Vec3, orientation matrix.Mat3) OOBB {
	s := Shape{}
	s.SetOOBB(center, extent, orientation)
	return OOBB(s)
}

func OOBBFromAABB(aabb AABB) OOBB {
	return NewOOBB(aabb.Center, aabb.Extent, matrix.Mat3Identity())
}

func OOBBFromTransform(baseAABB AABB, transform *matrix.Transform) OOBB {
	worldMat := transform.WorldMatrix()
	center := worldMat.TransformPoint(baseAABB.Center)
	orientation := worldMat.Mat3()
	extent := baseAABB.Extent.Multiply(transform.WorldScale())
	return NewOOBB(center, extent, orientation)
}

func (o OOBB) ContainsPoint(point matrix.Vec3) bool {
	localPoint := o.Orientation.Transpose().MultiplyVec3(point.Subtract(o.Center))
	if matrix.Abs(localPoint.X()) <= o.Extent.X() &&
		matrix.Abs(localPoint.Y()) <= o.Extent.Y() &&
		matrix.Abs(localPoint.Z()) <= o.Extent.Z() {
		return true
	}
	return false
}

func (o OOBB) Intersect(other OOBB) bool {
	axes := make([]matrix.Vec3, 6, 15)
	for i := 0; i < 3; i++ {
		axes[i] = o.Orientation.ColumnVector(i)
		axes[i+3] = other.Orientation.ColumnVector(i)
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			cross := matrix.Vec3Cross(o.Orientation.ColumnVector(i), other.Orientation.ColumnVector(j))
			if cross.Length() > 1e-6 {
				cross.Normalize()
				axes = append(axes, cross)
			}
		}
	}
	for _, axis := range axes {
		min1, max1 := o.projectInterval(axis)
		min2, max2 := other.projectInterval(axis)
		if !intervalsOverlap(min1, max1, min2, max2) {
			return false
		}
	}
	return true
}

func (o OOBB) RayIntersect(ray Ray, length float32) bool {
	localRayOrigin := o.Orientation.Transpose().MultiplyVec3(ray.Origin.Subtract(o.Center))
	localRayDir := o.Orientation.Transpose().MultiplyVec3(ray.Direction)
	localRay := Ray{
		Origin:    localRayOrigin,
		Direction: localRayDir,
	}
	localAABB := NewAABB(matrix.Vec3Zero(), o.Extent)
	_, hit := localAABB.RayHit(localRay)
	return hit
}

func (o OOBB) Bounds() AABB {
	corners := o.Corners()
	min := matrix.Vec3Largest()
	max := matrix.NewVec3(-min.X(), -min.Y(), -min.Z())
	for _, c := range corners {
		min = matrix.Vec3Min(min, c)
		max = matrix.Vec3Max(max, c)
	}
	return AABBFromMinMax(min, max)
}

func (o OOBB) Corners() [8]matrix.Vec3 {
	var corners [8]matrix.Vec3
	signs := [8][3]float32{
		{-1, -1, -1}, {1, -1, -1}, {-1, 1, -1}, {1, 1, -1},
		{-1, -1, 1}, {1, -1, 1}, {-1, 1, 1}, {1, 1, 1},
	}
	for i := 0; i < 8; i++ {
		local := matrix.Vec3{
			signs[i][0] * o.Extent.X(),
			signs[i][1] * o.Extent.Y(),
			signs[i][2] * o.Extent.Z(),
		}
		corners[i] = o.Orientation.MultiplyVec3(local).Add(o.Center)
	}
	return corners
}

func intervalsOverlap(min1, max1, min2, max2 float32) bool {
	const epsilon = 1e-6
	return max1 >= (min2-epsilon) && max2 >= (min1-epsilon)
}

func (o OOBB) projectInterval(axis matrix.Vec3) (float32, float32) {
	p := matrix.Vec3Dot(o.Center, axis)
	r := matrix.Abs(matrix.Vec3Dot(o.Orientation.ColumnVector(0), axis))*o.Extent.X() +
		matrix.Abs(matrix.Vec3Dot(o.Orientation.ColumnVector(1), axis))*o.Extent.Y() +
		matrix.Abs(matrix.Vec3Dot(o.Orientation.ColumnVector(2), axis))*o.Extent.Z()
	minProj := p - r
	maxProj := p + r
	return minProj, maxProj
}
