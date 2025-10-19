/******************************************************************************/
/* bullet3_collision_object.go                                                */
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
*/
import "C"
import "kaiju/matrix"

type CollisionHit C.HitWrapper
type CollisionObject struct{ ptr *C.btCollisionObject }

func (c *CollisionHit) IsValid() bool { return c.object != nil }

func (c *CollisionHit) Point() matrix.Vec3 {
	return matrix.NewVec3(matrix.Float(c.point[0]),
		matrix.Float(c.point[1]), matrix.Float(c.point[2]))
}

func (c *CollisionHit) Normal() matrix.Vec3 {
	return matrix.NewVec3(matrix.Float(c.normal[0]),
		matrix.Float(c.normal[1]), matrix.Float(c.normal[2]))
}

func (c *CollisionHit) Object() CollisionObject {
	return CollisionObject{ptr: c.object}
}
