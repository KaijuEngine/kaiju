/******************************************************************************/
/* terrain_collision_test.go                                                  */
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

package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestTerrainCollisionSamplesTerrainLocalCoordinates(t *testing.T) {
	collision := testTerrainCollision(t, 3, matrix.NewVec2(2, 2), []matrix.Float{
		0, 1, 2,
		3, 4, 5,
		6, 7, 8,
	}, 0, 8)
	if got := collision.Height(1, 2); got != 7 {
		t.Fatalf("expected height 7, got %f", got)
	}
	if got := collision.SampleGrid(0.5, 0.5); !matrix.ApproxTo(got, 2, 0.0001) {
		t.Fatalf("expected interpolated grid height 2, got %f", got)
	}
	if got := collision.HeightAtLocal(matrix.NewVec2(0, 0)); !matrix.ApproxTo(got, 4, 0.0001) {
		t.Fatalf("expected center local height 4, got %f", got)
	}
	local := collision.GridToLocal(2, 0)
	if !matrix.Vec3ApproxTo(local, matrix.NewVec3(1, 2, -1), 0.0001) {
		t.Fatalf("expected grid 2,0 to map to local 1,2,-1, got %v", local)
	}
	x, z := collision.LocalToGrid(matrix.NewVec2(-1, 1))
	if !matrix.ApproxTo(x, 0, 0.0001) || !matrix.ApproxTo(z, 2, 0.0001) {
		t.Fatalf("expected local -1,1 to map to grid 0,2, got %f,%f", x, z)
	}
}

func TestTerrainCollisionBoundsAndNormals(t *testing.T) {
	collision := testTerrainCollision(t, 2, matrix.NewVec2(4, 6), []matrix.Float{
		1, 1,
		1, 1,
	}, -2, 8)
	bounds := collision.LocalBounds()
	if !matrix.Vec3ApproxTo(bounds.Center, matrix.NewVec3(0, 3, 0), 0.0001) {
		t.Fatalf("expected bounds center 0,3,0, got %v", bounds.Center)
	}
	if !matrix.Vec3ApproxTo(bounds.Extent, matrix.NewVec3(2, 5, 3), 0.0001) {
		t.Fatalf("expected bounds extent 2,5,3, got %v", bounds.Extent)
	}
	normal := collision.NormalAtLocal(matrix.NewVec2(0, 0))
	if !matrix.Vec3ApproxTo(normal, matrix.Vec3Up(), 0.0001) {
		t.Fatalf("expected flat terrain normal to point up, got %v", normal)
	}
}

func TestTerrainCollisionCellRangeForLocalAABB(t *testing.T) {
	collision := testTerrainCollision(t, 3, matrix.NewVec2(2, 2), make([]matrix.Float, 9), -1, 1)
	bounds := AABBFromMinMax(matrix.NewVec3(-0.1, -0.1, -0.1), matrix.NewVec3(0.1, 0.1, 0.1))
	minX, minZ, maxX, maxZ, ok := collision.CellRangeForLocalAABB(bounds)
	if !ok {
		t.Fatal("expected local bounds to overlap terrain")
	}
	if minX != 0 || minZ != 0 || maxX != 1 || maxZ != 1 {
		t.Fatalf("expected center range [0,0]-[1,1], got [%d,%d]-[%d,%d]", minX, minZ, maxX, maxZ)
	}
	outside := AABBFromMinMax(matrix.NewVec3(3, -0.1, 3), matrix.NewVec3(4, 0.1, 4))
	if _, _, _, _, ok := collision.CellRangeForLocalAABB(outside); ok {
		t.Fatal("expected outside bounds to miss terrain")
	}
}

func TestTerrainCollisionForEachTriangleInLocalAABB(t *testing.T) {
	collision := testTerrainCollision(t, 2, matrix.NewVec2(2, 2), []matrix.Float{
		0, 0,
		0, 0,
	}, -1, 1)
	bounds := AABBFromMinMax(matrix.NewVec3(-1, -0.1, -1), matrix.NewVec3(1, 0.1, 1))
	triangles := make([]DetailedTriangle, 0, 2)
	collision.ForEachTriangleInLocalAABB(bounds, func(triangle DetailedTriangle) bool {
		triangles = append(triangles, triangle)
		return true
	})
	if len(triangles) != 2 {
		t.Fatalf("expected 2 terrain cell triangles, got %d", len(triangles))
	}
	for i := range triangles {
		if !matrix.Vec3ApproxTo(triangles[i].Normal, matrix.Vec3Up(), 0.0001) {
			t.Fatalf("expected triangle %d normal to point up, got %v", i, triangles[i].Normal)
		}
	}
}

func TestTerrainCollisionRaycastHitsHeightField(t *testing.T) {
	collision := testTerrainCollision(t, 2, matrix.NewVec2(4, 4), []matrix.Float{
		0, 0,
		0, 0,
	}, -1, 1)
	ray := Ray{
		Origin:    matrix.NewVec3(0, 2, 0),
		Direction: matrix.Vec3Down(),
	}
	hit, ok := collision.Raycast(ray, 4, nil)
	if !ok {
		t.Fatal("expected raycast to hit terrain")
	}
	if !matrix.ApproxTo(hit.Distance, 2, 0.001) {
		t.Fatalf("expected hit distance 2, got %f", hit.Distance)
	}
	if !matrix.Vec3ApproxTo(hit.Point, matrix.Vec3Zero(), 0.001) {
		t.Fatalf("expected terrain hit at origin, got %v", hit.Point)
	}
	if !matrix.Vec3ApproxTo(hit.Normal, matrix.Vec3Up(), 0.001) {
		t.Fatalf("expected terrain normal up, got %v", hit.Normal)
	}
}

func TestTerrainCollisionRaycastRespectsLength(t *testing.T) {
	collision := testTerrainCollision(t, 2, matrix.NewVec2(4, 4), []matrix.Float{
		0, 0,
		0, 0,
	}, -1, 1)
	ray := Ray{
		Origin:    matrix.NewVec3(0, 3, 0),
		Direction: matrix.Vec3Down(),
	}
	if hit, ok := collision.Raycast(ray, 1, nil); ok {
		t.Fatalf("expected short raycast to miss terrain, got hit %+v", hit)
	}
}

func TestSystemRaycastHitsStaticTerrain(t *testing.T) {
	system := System{}
	system.Initialize()
	body := system.NewBody()
	body.Transform.SetPosition(matrix.NewVec3(3, 0, -2))
	body.SetStaticTerrain(testTerrainCollision(t, 2, matrix.NewVec2(4, 4), []matrix.Float{
		0, 0,
		0, 0,
	}, -1, 1))
	hit, ok := system.Raycast(matrix.NewVec3(3, 2, -2), matrix.NewVec3(3, -2, -2))
	if !ok {
		t.Fatal("expected system raycast to hit terrain")
	}
	if hit.Body != body {
		t.Fatalf("expected hit body %p, got %p", body, hit.Body)
	}
	if !matrix.ApproxTo(hit.Distance, 2, 0.001) {
		t.Fatalf("expected hit distance 2, got %f", hit.Distance)
	}
	if !matrix.Vec3ApproxTo(hit.Point, matrix.NewVec3(3, 0, -2), 0.001) {
		t.Fatalf("expected translated terrain hit point, got %v", hit.Point)
	}
	if !matrix.Vec3ApproxTo(hit.Normal, matrix.Vec3Up(), 0.001) {
		t.Fatalf("expected terrain normal up, got %v", hit.Normal)
	}
}

func TestNewTerrainShapeSetup(t *testing.T) {
	bounds := AABBFromMinMax(matrix.NewVec3(-2, -1, -3), matrix.NewVec3(2, 5, 3))
	shape := NewTerrainShape(bounds)
	if shape.Type != ShapeTypeTerrain {
		t.Fatalf("expected terrain shape, got %v", shape.Type)
	}
	if !matrix.Vec3ApproxTo(shape.Center, bounds.Center, 0.0001) {
		t.Fatalf("expected terrain shape center %v, got %v", bounds.Center, shape.Center)
	}
	if !matrix.Vec3ApproxTo(shape.Extent, bounds.Extent, 0.0001) {
		t.Fatalf("expected terrain shape extent %v, got %v", bounds.Extent, shape.Extent)
	}
}

func TestStaticTerrainBodyGeneratesBroadPhaseAABB(t *testing.T) {
	system := System{}
	system.Initialize()
	body := system.NewBody()
	body.Transform.SetPosition(matrix.NewVec3(3, 0, -2))
	terrain := testTerrainCollision(t, 2, matrix.NewVec2(4, 6), []matrix.Float{
		0, 0,
		0, 0,
	}, -1, 5)
	body.SetStaticTerrain(terrain)
	if body.Collision.Terrain != terrain {
		t.Fatal("expected body to store terrain collision")
	}
	if body.Collision.Mesh != nil {
		t.Fatal("expected terrain body to clear mesh collision")
	}
	if body.Collision.Shape.Type != ShapeTypeTerrain {
		t.Fatalf("expected terrain shape, got %v", body.Collision.Shape.Type)
	}
	if !body.IsStatic() {
		t.Fatal("expected terrain body to be static")
	}
	system.broadPhase.Rebuild(&system.bodies)
	if len(system.broadPhase.proxies) != 1 {
		t.Fatalf("expected 1 broad phase proxy, got %d", len(system.broadPhase.proxies))
	}
	proxy := system.broadPhase.proxies[0]
	if proxy.body != body {
		t.Fatal("expected proxy to reference terrain body")
	}
	if !matrix.Approx(proxy.bounds[matrix.Vx].min, 1) || !matrix.Approx(proxy.bounds[matrix.Vx].max, 5) {
		t.Fatalf("expected terrain proxy X bounds [1,5], got [%f,%f]",
			proxy.bounds[matrix.Vx].min, proxy.bounds[matrix.Vx].max)
	}
	if !matrix.Approx(proxy.bounds[matrix.Vy].min, -1) || !matrix.Approx(proxy.bounds[matrix.Vy].max, 5) {
		t.Fatalf("expected terrain proxy Y bounds [-1,5], got [%f,%f]",
			proxy.bounds[matrix.Vy].min, proxy.bounds[matrix.Vy].max)
	}
	if !matrix.Approx(proxy.bounds[matrix.Vz].min, -5) || !matrix.Approx(proxy.bounds[matrix.Vz].max, 1) {
		t.Fatalf("expected terrain proxy Z bounds [-5,1], got [%f,%f]",
			proxy.bounds[matrix.Vz].min, proxy.bounds[matrix.Vz].max)
	}
}

func TestSystemDynamicSphereRestsOnStaticTerrain(t *testing.T) {
	system := System{}
	system.Initialize()
	sphere := system.NewBody()
	sphere.SetDynamic(1, matrix.Vec3One())
	sphere.Collision.Shape = NewSphereShape(0.5)
	sphere.Transform.SetPosition(matrix.NewVec3(0, 0.5, 0))
	terrain := system.NewBody()
	terrain.SetStaticTerrain(testTerrainCollision(t, 2, matrix.NewVec2(4, 4), []matrix.Float{
		0, 0,
		0, 0,
	}, 0, 0))
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	hadContact := false
	for range 30 {
		system.Step(workGroup, threads, 1.0/60.0)
		hadContact = hadContact || len(system.Contacts()) > 0
	}
	if !hadContact {
		t.Fatal("expected sphere to contact terrain while stepping")
	}
	if sphere.Transform.WorldPosition().Y() < 0.45 {
		t.Fatalf("expected sphere to rest on terrain, got position %v", sphere.Transform.WorldPosition())
	}
}

func TestNarrowPhaseSphereCollidesWithRaisedTerrainHeight(t *testing.T) {
	sphere := testRigidBody(NewSphereShape(0.5), matrix.NewVec3(0, 1.45, 0))
	terrain := testStaticTerrainBody(testTerrainCollision(t, 2, matrix.NewVec2(4, 4), []matrix.Float{
		1, 1,
		1, 1,
	}, 0, 2))
	manifold, ok := CollideBodies(sphere, terrain)
	if !ok {
		t.Fatal("expected sphere to collide with raised terrain")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Down(), 0.0001) {
		t.Fatalf("expected downward sphere-to-terrain normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.05) > 0.0001 {
		t.Fatalf("expected penetration 0.05 against raised terrain, got %f", contact.Penetration)
	}
}

func TestNarrowPhasePrimitivesCollideWithSlopedTerrain(t *testing.T) {
	tests := []struct {
		name     string
		body     *RigidBody
		maxY     matrix.Float
		minDepth matrix.Float
	}{
		{
			name:     "sphere",
			body:     testRigidBody(NewSphereShape(0.5), matrix.NewVec3(0, 0.45, 0)),
			maxY:     -0.5,
			minDepth: 0.0001,
		},
		{
			name:     "capsule",
			body:     testRigidBody(NewCapsuleShape(0.5, 2), matrix.NewVec3(0, 1.45, 0)),
			maxY:     -0.5,
			minDepth: 0.0001,
		},
		{
			name:     "oobb",
			body:     testRigidBody(NewBoxShape(matrix.NewVec3(0.5, 0.5, 0.5)), matrix.NewVec3(0, 0.45, 0)),
			maxY:     -0.5,
			minDepth: 0.0001,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terrain := testStaticTerrainBody(testSlopedTerrain(t))
			manifold, ok := CollideBodies(tt.body, terrain)
			if !ok {
				t.Fatal("expected primitive to collide with sloped terrain")
			}
			contact := manifold.Contacts[0]
			horizontal := matrix.NewVec2(contact.Normal.X(), contact.Normal.Z()).Length()
			if contact.Normal.Y() > tt.maxY || horizontal < 0.2 {
				t.Fatalf("expected non-flat slope contact normal pointing down, got %v", contact.Normal)
			}
			if contact.Penetration <= tt.minDepth {
				t.Fatalf("expected positive slope penetration, got %f", contact.Penetration)
			}
		})
	}
}

func TestTerrainTransformScaleAffectsCollision(t *testing.T) {
	sphere := testRigidBody(NewSphereShape(0.5), matrix.NewVec3(4, 7.45, -3))
	terrain := testStaticTerrainBody(testTerrainCollision(t, 2, matrix.NewVec2(4, 4), []matrix.Float{
		1, 1,
		1, 1,
	}, 1, 1))
	terrain.Transform.SetPosition(matrix.NewVec3(4, 5, -3))
	terrain.Transform.SetScale(matrix.NewVec3(2, 2, 0.5))
	manifold, ok := CollideBodies(sphere, terrain)
	if !ok {
		t.Fatal("expected sphere to collide with translated and scaled terrain")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Down(), 0.0001) {
		t.Fatalf("expected downward normal from scaled terrain, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.05) > 0.0001 {
		t.Fatalf("expected penetration 0.05 against scaled terrain, got %f", contact.Penetration)
	}
}

func testTerrainCollision(t *testing.T, resolution int, worldSize matrix.Vec2, heights []matrix.Float, minHeight, maxHeight matrix.Float) *TerrainCollision {
	t.Helper()
	collision, err := NewTerrainCollision(resolution, worldSize, heights, minHeight, maxHeight)
	if err != nil {
		t.Fatal(err)
	}
	return collision
}

func testSlopedTerrain(t *testing.T) *TerrainCollision {
	t.Helper()
	return testTerrainCollision(t, 2, matrix.NewVec2(4, 4), []matrix.Float{
		-1, 1,
		-1, 1,
	}, -1, 1)
}
