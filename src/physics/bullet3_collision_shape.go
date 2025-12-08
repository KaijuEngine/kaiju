/******************************************************************************/
/* bullet3_collision_shape.go                                                 */
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

package physics

/*
#cgo CXXFLAGS: -std=c++11
#cgo LDFLAGS: -L../../libs -lBulletDynamics -lBulletCollision -lLinearMath -lstdc++ -lm
#include "bullet3_wrapper.h"
#cgo noescape btCollisionShape_calculateLocalInertia
#cgo nocallback btCollisionShape_calculateLocalInertia
#cgo noescape destroy_btCollisionShape
#cgo nocallback destroy_btCollisionShape
#cgo noescape new_btBoxShape
#cgo nocallback new_btBoxShape
#cgo noescape new_btSphereShape
#cgo nocallback new_btSphereShape
#cgo noescape new_btCapsuleShape
#cgo nocallback new_btCapsuleShape
#cgo noescape new_btCylinderShape
#cgo nocallback new_btCylinderShape
#cgo noescape new_btConeShape
#cgo nocallback new_btConeShape
#cgo noescape new_btStaticPlaneShape
#cgo nocallback new_btStaticPlaneShape
#cgo noescape new_btCompoundShape
#cgo nocallback new_btCompoundShape
#cgo noescape new_btConvexHullShape
#cgo nocallback new_btConvexHullShape
#cgo noescape new_btEmptyShape
#cgo nocallback new_btEmptyShape
#cgo noescape new_btMultiSphereShape
#cgo nocallback new_btMultiSphereShape
#cgo noescape new_btUniformScalingShape
#cgo nocallback new_btUniformScalingShape
*/
import "C"
import (
	"kaiju/matrix"
	"runtime"
	"unsafe"
)

type CollisionShape struct{ ptr *C.btCollisionShape }
type BoxShape struct{ CollisionShape }
type SphereShape struct{ CollisionShape }
type EmptyShape struct{ CollisionShape }
type CapsuleShape struct{ CollisionShape }
type CylinderShape struct{ CollisionShape }
type ConeShape struct{ CollisionShape }
type StaticPlaneShape struct{ CollisionShape }
type CompoundShape struct{ CollisionShape }
type ConvexShape struct{ CollisionShape }
type ConvexHullShape struct{ ConvexShape }
type MultiSphereShape struct{ CollisionShape }
type UniformScalingShape struct{ CollisionShape }

func NewEmptyShape(size matrix.Vec3) *EmptyShape {
	s := &EmptyShape{
		CollisionShape: CollisionShape{
			ptr: (*C.btCollisionShape)(C.new_btEmptyShape()),
		},
	}
	runtime.AddCleanup(s, func(ptr *C.btCollisionShape) {
		C.destroy_btCollisionShape(ptr)
	}, s.ptr)
	return s
}

func NewBoxShape(size matrix.Vec3) *BoxShape {
	s := &BoxShape{
		CollisionShape: CollisionShape{
			ptr: (*C.btCollisionShape)(C.new_btBoxShape(C.float(size.X()),
				C.float(size.Y()), C.float(size.Z()))),
		},
	}
	runtime.AddCleanup(s, func(ptr *C.btCollisionShape) {
		C.destroy_btCollisionShape(ptr)
	}, s.ptr)
	return s
}

func NewSphereShape(radius float32) *SphereShape {
	s := &SphereShape{
		CollisionShape: CollisionShape{
			ptr: (*C.btCollisionShape)(C.new_btSphereShape(C.float(radius))),
		},
	}
	runtime.AddCleanup(s, func(ptr *C.btCollisionShape) {
		C.destroy_btCollisionShape(ptr)
	}, s.ptr)
	return s
}

func NewCapsuleShape(radius, height float32) *CapsuleShape {
	s := &CapsuleShape{
		CollisionShape: CollisionShape{
			ptr: (*C.btCollisionShape)(C.new_btCapsuleShape(C.float(radius), C.float(height))),
		},
	}
	runtime.AddCleanup(s, func(ptr *C.btCollisionShape) {
		C.destroy_btCollisionShape(ptr)
	}, s.ptr)
	return s
}

func NewnCylinderShape(halfExtents matrix.Vec3) *CylinderShape {
	s := &CylinderShape{
		CollisionShape: CollisionShape{
			ptr: (*C.btCollisionShape)(C.new_btCylinderShape(C.float(halfExtents.X()),
				C.float(halfExtents.Y()), C.float(halfExtents.Z()))),
		},
	}
	runtime.AddCleanup(s, func(ptr *C.btCollisionShape) {
		C.destroy_btCollisionShape(ptr)
	}, s.ptr)
	return s
}

func NewConeShape(radius, height float32) *ConeShape {
	s := &ConeShape{
		CollisionShape: CollisionShape{
			ptr: (*C.btCollisionShape)(C.new_btConeShape(
				C.float(radius), C.float(height))),
		},
	}
	runtime.AddCleanup(s, func(ptr *C.btCollisionShape) {
		C.destroy_btCollisionShape(ptr)
	}, s.ptr)
	return s
}

func NewStaticPlaneShape(normal matrix.Vec3, constant float32) *StaticPlaneShape {
	s := &StaticPlaneShape{
		CollisionShape: CollisionShape{
			ptr: (*C.btCollisionShape)(C.new_btStaticPlaneShape(
				C.float(normal.X()), C.float(normal.Y()),
				C.float(normal.Z()), C.float(constant))),
		},
	}
	runtime.AddCleanup(s, func(ptr *C.btCollisionShape) {
		C.destroy_btCollisionShape(ptr)
	}, s.ptr)
	return s
}

func NewCompoundShape(initialChildCapacity int, enableDynamicAABBTree bool) *CompoundShape {
	s := &CompoundShape{
		CollisionShape: CollisionShape{
			ptr: (*C.btCollisionShape)(C.new_btCompoundShape(
				C.int(initialChildCapacity), C.bool(enableDynamicAABBTree))),
		},
	}
	runtime.AddCleanup(s, func(ptr *C.btCollisionShape) {
		C.destroy_btCollisionShape(ptr)
	}, s.ptr)
	return s
}

func NewConvexHullShape(points []float32, stride int) *ConvexHullShape {
	s := &ConvexHullShape{
		ConvexShape{
			CollisionShape: CollisionShape{
				ptr: (*C.btCollisionShape)(C.new_btConvexHullShape(
					(*C.float)(&points[0]), C.int(len(points)), C.int(stride))),
			},
		},
	}
	runtime.AddCleanup(s, func(ptr *C.btCollisionShape) {
		C.destroy_btCollisionShape(ptr)
	}, s.ptr)
	return s
}

func NewMultiSphereShape(positions []matrix.Vec3, radii []float32) *MultiSphereShape {
	s := &MultiSphereShape{
		CollisionShape: CollisionShape{
			ptr: (*C.btCollisionShape)(C.new_btMultiSphereShape(
				(*C.float)((*float32)(unsafe.Pointer(&positions[0]))), (*C.float)(&radii[0]), C.int(len(radii)))),
		},
	}
	runtime.AddCleanup(s, func(ptr *C.btCollisionShape) {
		C.destroy_btCollisionShape(ptr)
	}, s.ptr)
	return s
}

func NewUniformScalingShape(convexChildShape *ConvexShape, scaleFactor float32) *UniformScalingShape {
	s := &UniformScalingShape{
		CollisionShape: CollisionShape{
			ptr: (*C.btCollisionShape)(C.new_btUniformScalingShape(
				(*C.btConvexShape)(convexChildShape.ptr), C.float(scaleFactor))),
		},
	}
	runtime.AddCleanup(s, func(ptr *C.btCollisionShape) {
		C.destroy_btCollisionShape(ptr)
	}, s.ptr)
	return s
}

func (s *CollisionShape) CalculateLocalInertia(mass float32) matrix.Vec3 {
	out := matrix.Vec3{}
	C.btCollisionShape_calculateLocalInertia(s.ptr, C.float(mass),
		(*C.float)(out.PX()), (*C.float)(out.PY()), (*C.float)(out.PZ()))
	return out
}
