/******************************************************************************/
/* bullet3_rigid_body.go                                                      */
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

package physics

/*
#cgo CXXFLAGS: -std=c++11
#cgo LDFLAGS: -L../../libs -lBulletDynamics -lBulletCollision -lLinearMath -lstdc++ -lm
#include "bullet3_wrapper.h"
#cgo noescape new_btRigidBody
#cgo nocallback new_btRigidBody
#cgo noescape destroy_btRigidBody
#cgo nocallback destroy_btRigidBody
#cgo noescape btRigidBody_getPosition
#cgo nocallback btRigidBody_getPosition
#cgo noescape btRigidBody_getRotation
#cgo nocallback btRigidBody_getRotation
#cgo noescape btRigidBody_applyForceAtPoint
#cgo nocallback btRigidBody_applyForceAtPoint
#cgo noescape btRigidBody_applyImpulseAtPoint
#cgo nocallback btRigidBody_applyImpulseAtPoint
*/
import "C"
import (
	"runtime"
	"unsafe"

	"github.com/KaijuEngine/kaiju/matrix"
)

type RigidBody struct {
	ptr   *C.btRigidBody
	shape *CollisionShape
}

func NewRigidBody(mass float32, motion *MotionState, shape *CollisionShape, inertia matrix.Vec3) *RigidBody {
	b := &RigidBody{
		ptr: C.new_btRigidBody(C.float(mass),
			motion.ptr, shape.ptr,
			C.float(inertia.X()), C.float(inertia.Y()), C.float(inertia.Z())),
		shape: shape,
	}
	runtime.AddCleanup(b, func(ptr *C.btRigidBody) {
		C.destroy_btRigidBody(ptr)
	}, b.ptr)
	return b
}

func (r *RigidBody) Shape() *CollisionShape { return r.shape }

func (r *RigidBody) IsCollisionObject(obj CollisionObject) bool {
	return unsafe.Pointer(r.ptr) == unsafe.Pointer(obj.ptr)
}

func (r *RigidBody) Position() matrix.Vec3 {
	out := matrix.Vec3{}
	C.btRigidBody_getPosition(r.ptr,
		(*C.float)(out.PX()), (*C.float)(out.PY()), (*C.float)(out.PZ()))
	return out
}

func (r *RigidBody) Rotation() matrix.Quaternion {
	out := matrix.Vec4{}
	C.btRigidBody_getRotation(r.ptr,
		(*C.float)(out.PX()), (*C.float)(out.PY()),
		(*C.float)(out.PZ()), (*C.float)(out.PW()))
	return matrix.QuaternionFromVec4(out)
}

func (r *RigidBody) ApplyForceAtPoint(force, point matrix.Vec3) {
	C.btRigidBody_applyForceAtPoint(r.ptr,
		C.float(force.X()), C.float(force.Y()), C.float(force.Z()),
		C.float(point.X()), C.float(point.Y()), C.float(point.Z()))
}

func (r *RigidBody) ApplyImpulseAtPoint(force, point matrix.Vec3) {
	C.btRigidBody_applyImpulseAtPoint(r.ptr,
		C.float(force.X()), C.float(force.Y()), C.float(force.Z()),
		C.float(point.X()), C.float(point.Y()), C.float(point.Z()))
}
