/******************************************************************************/
/* bullet3_world.go                                                           */
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
#cgo windows,amd64 LDFLAGS: -L../../libs -lBulletDynamics_win_amd64 -lBulletCollision_win_amd64 -lLinearMath_win_amd64 -lstdc++ -lm
#cgo linux,amd64 LDFLAGS: -L../../libs -lBulletDynamics_nix_amd64 -lBulletCollision_nix_amd64 -lLinearMath_nix_amd64 -lstdc++ -lm
#cgo darwin,arm64 LDFLAGS: -L../libs -lBulletDynamics_darwin_arm64 -lBulletCollision_darwin_arm64 -lLinearMath_darwin_arm64 -lstdc++ -lm
#cgo darwin,amd64 LDFLAGS: -L../libs -lBulletDynamics_darwin_amd64 -lBulletCollision_darwin_amd64 -lLinearMath_darwin_amd64 -lstdc++ -lm
#include "bullet3_wrapper.h"
#cgo noescape new_btDiscreteDynamicsWorld
#cgo nocallback new_btDiscreteDynamicsWorld
#cgo noescape destroy_btDiscreteDynamicsWorld
#cgo nocallback destroy_btDiscreteDynamicsWorld
#cgo noescape btDiscreteDynamicsWorld_setGravity
#cgo nocallback btDiscreteDynamicsWorld_setGravity
#cgo noescape btDiscreteDynamicsWorld_stepSimulation
#cgo nocallback btDiscreteDynamicsWorld_stepSimulation
#cgo noescape btDiscreteDynamicsWorld_addRigidBody
#cgo nocallback btDiscreteDynamicsWorld_addRigidBody
#cgo noescape btDiscreteDynamicsWorld_removeRigidBody
#cgo nocallback btDiscreteDynamicsWorld_removeRigidBody
#cgo noescape btDiscreteDynamicsWorld_rayTest
#cgo nocallback btDiscreteDynamicsWorld_rayTest
#cgo noescape btDiscreteDynamicsWorld_sphereSweep
#cgo nocallback btDiscreteDynamicsWorld_sphereSweep
*/
import "C"
import (
	"kaiju/matrix"
	"runtime"
)

type World struct{ ptr *C.btDiscreteDynamicsWorld }

func NewDiscreteDynamicsWorld(dispatcher *CollisionDispatcher, broadphase *BroadphaseInterface, solver *SequentialImpulseConstraintSolver, collisionConfig *DefaultCollisionConfiguration) *World {
	w := &World{
		ptr: C.new_btDiscreteDynamicsWorld(dispatcher.ptr, broadphase.ptr, solver.ptr, collisionConfig.ptr),
	}
	runtime.AddCleanup(w, func(ptr *C.btDiscreteDynamicsWorld) {
		C.destroy_btDiscreteDynamicsWorld(ptr)
	}, w.ptr)
	return w
}

func (w *World) SetGravity(v matrix.Vec3) {
	C.btDiscreteDynamicsWorld_setGravity(w.ptr,
		C.float(v.X()), C.float(v.Y()), C.float(v.Z()))
}

func (w *World) AddRigidBody(body *RigidBody) {
	C.btDiscreteDynamicsWorld_addRigidBody(w.ptr, body.ptr)
}

func (w *World) RemoveRigidBody(body *RigidBody) {
	C.btDiscreteDynamicsWorld_removeRigidBody(w.ptr, body.ptr)
}

func (w *World) StepSimulation(timeStep float32) {
	C.btDiscreteDynamicsWorld_stepSimulation(w.ptr, C.float(timeStep))
}

func (w *World) Raycast(from, to matrix.Vec3) CollisionHit {
	return CollisionHit(C.btDiscreteDynamicsWorld_rayTest(w.ptr,
		C.float(from.X()), C.float(from.Y()), C.float(from.Z()),
		C.float(to.X()), C.float(to.Y()), C.float(to.Z())))
}

func (w *World) SphereSweep(from, to matrix.Vec3, radius float32) CollisionHit {
	return CollisionHit(C.btDiscreteDynamicsWorld_sphereSweep(w.ptr,
		C.float(from.X()), C.float(from.Y()), C.float(from.Z()),
		C.float(to.X()), C.float(to.Y()), C.float(to.Z()), C.float(radius)))
}
