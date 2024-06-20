/******************************************************************************/
/* obb.go                                                                     */
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
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

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
