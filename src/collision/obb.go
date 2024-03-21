package collision

import "kaiju/matrix"

type OBB struct {
	Center      matrix.Vec3
	Extent      matrix.Vec3
	Orientation matrix.Mat3
}

func OBBFromAABB(aabb AABB) OBB {
	return OBB{
		Center:      aabb.Center,
		Extent:      aabb.Extent,
		Orientation: matrix.Mat3Identity(),
	}
}

func (o OBB) ContainsPoint(point matrix.Vec3) bool {
	localPoint := o.Orientation.Transpose().MultiplyVec3(point.Subtract(o.Center))
	if matrix.Abs(localPoint.X()) <= o.Extent.X() &&
		matrix.Abs(localPoint.Y()) <= o.Extent.Y() &&
		matrix.Abs(localPoint.Z()) <= o.Extent.Z() {
		return true
	}
	return false
}

func (o OBB) ProjectOntoAxis(axis matrix.Vec3) OBB {
	projection := OBB{
		Center: o.Center,
		Extent: matrix.Vec3{
			matrix.Abs(matrix.Vec3Dot(o.Extent, axis)),
			matrix.Abs(matrix.Vec3Dot(o.Extent, axis)),
			matrix.Abs(matrix.Vec3Dot(o.Extent, axis)),
		},
		Orientation: o.Orientation,
	}
	return projection
}

func (o OBB) Overlaps(other OBB) bool {
	if matrix.Abs(o.Center.X()-other.Center.X()) > o.Extent.X()+other.Extent.X() {
		return false
	}
	if matrix.Abs(o.Center.Y()-other.Center.Y()) > o.Extent.Y()+other.Extent.Y() {
		return false
	}
	if matrix.Abs(o.Center.Z()-other.Center.Z()) > o.Extent.Z()+other.Extent.Z() {
		return false
	}
	return true
}

func (o OBB) Intersect(other OBB) bool {
	for i := 0; i < 3; i++ {
		axisA := o.Orientation.ColumnVector(i)
		axisB := other.Orientation.ColumnVector(i)
		projectionA := o.ProjectOntoAxis(axisA)
		projectionB := other.ProjectOntoAxis(axisA)
		if !projectionA.Overlaps(projectionB) {
			return false
		}
		projectionA = o.ProjectOntoAxis(axisB)
		projectionB = other.ProjectOntoAxis(axisB)
		if !projectionA.Overlaps(projectionB) {
			return false
		}
	}
	return true
}
