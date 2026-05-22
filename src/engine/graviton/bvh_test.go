/******************************************************************************/
/* bvh_test.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
)

func testTransform(position matrix.Vec3) *matrix.Transform {
	transform := &matrix.Transform{}
	transform.Initialize(nil)
	transform.SetPosition(position)
	return transform
}

func TestAddSubBVHInsertsProxyLeaf(t *testing.T) {
	transform := testTransform(matrix.Vec3Zero())
	sub := NewBVH([]HitObject{AABBFromWidth(matrix.Vec3Zero(), 1)}, transform, "hit")
	var world *BVH
	proxy := AddSubBVH(&world, sub, transform)
	if proxy == nil {
		t.Fatal("expected proxy node")
	}
	if proxy == sub {
		t.Fatal("expected sub BVH to be inserted behind a proxy leaf")
	}
	if world != proxy {
		t.Fatal("expected first proxy node to become the world root")
	}
	if proxy.Item.HitCheck != sub {
		t.Fatal("expected proxy hit check to reference the sub BVH")
	}
	if proxy.Item.Data != sub {
		t.Fatal("expected proxy data to reference the sub BVH")
	}
	if sub.Parent != nil {
		t.Fatal("sub BVH should not be spliced into the world tree")
	}
	ray := Ray{
		Origin:    matrix.NewVec3(0, 0, -5),
		Direction: matrix.NewVec3(0, 0, 1),
	}
	data, _, ok := world.RayIntersect(ray, 20)
	if !ok {
		t.Fatal("expected ray to hit proxied sub BVH")
	}
	if data != "hit" {
		t.Fatalf("expected proxied sub BVH data, got %v", data)
	}
}

func TestProxyLeafRefitUpdatesWorldBounds(t *testing.T) {
	transform := testTransform(matrix.Vec3Zero())
	sub := NewBVH([]HitObject{AABBFromWidth(matrix.Vec3Zero(), 1)}, transform, "moved")
	var world *BVH
	proxy := AddSubBVH(&world, sub, transform)
	transform.SetPosition(matrix.NewVec3(10, 0, 0))
	sub.Refit()
	proxy.RefitUpwards()
	oldRay := Ray{
		Origin:    matrix.NewVec3(0, 0, -5),
		Direction: matrix.NewVec3(0, 0, 1),
	}
	if _, _, ok := world.RayIntersect(oldRay, 20); ok {
		t.Fatal("expected old world bounds to miss after proxy refit")
	}
	newRay := Ray{
		Origin:    matrix.NewVec3(10, 0, -5),
		Direction: matrix.NewVec3(0, 0, 1),
	}
	data, _, ok := world.RayIntersect(newRay, 20)
	if !ok {
		t.Fatal("expected updated proxy bounds to hit")
	}
	if data != "moved" {
		t.Fatalf("expected moved sub BVH data, got %v", data)
	}
}

func TestRemoveBVHNodeRemovesProxyLeaf(t *testing.T) {
	transform := testTransform(matrix.Vec3Zero())
	sub := NewBVH([]HitObject{AABBFromWidth(matrix.Vec3Zero(), 1)}, transform, "hit")
	var world *BVH
	proxy := AddSubBVH(&world, sub, transform)
	RemoveBVHNode(&world, proxy)
	if world != nil {
		t.Fatal("expected world BVH to be empty after removing only proxy")
	}
}
