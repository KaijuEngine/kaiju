/******************************************************************************/
/* aabb_test.go                                                               */
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

package collision

import (
	"kaiju/matrix"
	"testing"
)

func TestAABBHit(t *testing.T) {
	box := AABB{matrix.Vec3Zero(), matrix.Vec3{0.5, 0.5, 0.5}}
	r := Ray{matrix.Vec3Right(), matrix.Vec3Left()}
	if _, ok := box.RayHit(r); !ok {
		t.Error("Expected hit")
	}
	r = Ray{matrix.Vec3Left(), matrix.Vec3Right()}
	if _, ok := box.RayHit(r); !ok {
		t.Error("Expected hit")
	}
	r = Ray{matrix.Vec3Up(), matrix.Vec3Down()}
	if _, ok := box.RayHit(r); !ok {
		t.Error("Expected hit")
	}
	r = Ray{matrix.Vec3Down(), matrix.Vec3Up()}
	if _, ok := box.RayHit(r); !ok {
		t.Error("Expected hit")
	}
	r = Ray{matrix.Vec3Forward(), matrix.Vec3Backward()}
	if _, ok := box.RayHit(r); !ok {
		t.Error("Expected hit")
	}
	r = Ray{matrix.Vec3Backward(), matrix.Vec3Forward()}
	if _, ok := box.RayHit(r); !ok {
		t.Error("Expected hit")
	}
}

func TestAABBMiss(t *testing.T) {
	box := AABB{matrix.Vec3Zero(), matrix.Vec3{0.5, 0.5, 0.5}}
	r := Ray{matrix.Vec3Right(), matrix.Vec3Up()}
	if _, ok := box.RayHit(r); ok {
		t.Error("Expected miss")
	}
	r.Direction = matrix.Vec3Down()
	if _, ok := box.RayHit(r); ok {
		t.Error("Expected miss")
	}
	r.Direction = matrix.Vec3Right()
	if _, ok := box.RayHit(r); ok {
		t.Error("Expected miss")
	}
	r.Direction = matrix.Vec3Forward()
	if _, ok := box.RayHit(r); ok {
		t.Error("Expected miss")
	}
	r.Direction = matrix.Vec3Backward()
	if _, ok := box.RayHit(r); ok {
		t.Error("Expected miss")
	}
}

func TestTriangleIntersect(t *testing.T) {
	box := AABB{matrix.Vec3Zero(), matrix.Vec3{0.5, 0.5, 0.5}}
	points0 := [3]matrix.Vec3{
		{-0.25, 0.0, 0.25},
		{0.0, 0.0, -0.25},
		{0.25, 0.0, 0.25},
	}
	if !box.TriangleIntersect(DetailedTriangleFromPoints(points0)) {
		t.Error("Expected intersect")
	}
}

func TestAABBUnion(t *testing.T) {
	a := AABB{
		Center: matrix.Vec3{1, 0, 0},
		Extent: matrix.Vec3{1, 1, 1},
	}
	b := AABB{
		Center: matrix.Vec3{0, 0, 0},
		Extent: matrix.Vec3{2, 2, 2},
	}
	c := AABBUnion(a, b)
	if !c.ContainsAABB(a) {
		t.Fail()
	}
	if !c.ContainsAABB(b) {
		t.Fail()
	}
}
