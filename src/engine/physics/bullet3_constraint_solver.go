/******************************************************************************/
/* bullet3_constraint_solver.go                                               */
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
#include "bullet3_wrapper.h"
#cgo noescape new_btSequentialImpulseConstraintSolver
#cgo nocallback new_btSequentialImpulseConstraintSolver
#cgo noescape destroy_btSequentialImpulseConstraintSolver
#cgo nocallback destroy_btSequentialImpulseConstraintSolver
*/
import "C"
import "runtime"

type SequentialImpulseConstraintSolver struct {
	ptr *C.btSequentialImpulseConstraintSolver
}

func NewSequentialImpulseConstraintSolver() *SequentialImpulseConstraintSolver {
	s := &SequentialImpulseConstraintSolver{
		ptr: C.new_btSequentialImpulseConstraintSolver(),
	}
	runtime.AddCleanup(s, func(ptr *C.btSequentialImpulseConstraintSolver) {
		C.destroy_btSequentialImpulseConstraintSolver(ptr)
	}, s.ptr)
	return s
}
