/******************************************************************************/
/* aabb_test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestAABBHit(t *testing.T) {
	box := NewAABB(matrix.Vec3Zero(), matrix.Vec3{0.5, 0.5, 0.5})
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
	box := NewAABB(matrix.Vec3Zero(), matrix.Vec3{0.5, 0.5, 0.5})
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
	box := NewAABB(matrix.Vec3Zero(), matrix.Vec3{0.5, 0.5, 0.5})
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
	a := NewAABB(matrix.Vec3{1, 0, 0}, matrix.Vec3{1, 1, 1})
	b := NewAABB(matrix.Vec3{0, 0, 0}, matrix.Vec3{2, 2, 2})
	c := AABBUnion(a, b)
	if !c.ContainsAABB(a) {
		t.Fail()
	}
	if !c.ContainsAABB(b) {
		t.Fail()
	}
}

func TestFrustimInAABB(t *testing.T) {
	v := matrix.Mat4Identity()
	var p matrix.Mat4
	p.Orthographic(-10*0.5, 10*0.5, -10*0.5, 10*0.5, 0.01, 100)
	vp := matrix.Mat4Multiply(v, p)
	var f Frustum
	f.ExtractPlanes(vp)
	b := NewAABB(matrix.Vec3Zero(), matrix.Vec3{50, 50, 0})
	if !b.IntersectsFrustum(f) {
		t.Fail()
	}
}
