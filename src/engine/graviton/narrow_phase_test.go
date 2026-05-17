/******************************************************************************/
/* narrow_phase_test.go                                                       */
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
	"kaijuengine.com/platform/concurrent"
)

func TestNarrowPhaseSphereSphereContact(t *testing.T) {
	a := testRigidBody(Shape{}, matrix.Vec3{0, 0, 0})
	a.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	b := testRigidBody(Shape{}, matrix.Vec3{1.5, 0, 0})
	b.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	manifold, ok := CollideBodies(a, b)
	if !ok {
		t.Fatal("expected overlapping spheres to collide")
	}
	if manifold.Count != 1 {
		t.Fatalf("expected 1 contact, got %d", manifold.Count)
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Right(), 0.0001) {
		t.Fatalf("expected +X normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.5) > 0.0001 {
		t.Fatalf("expected penetration 0.5, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseSphereAABBContact(t *testing.T) {
	sphereBody := testRigidBody(Shape{}, matrix.Vec3{0, 0, 0})
	sphereBody.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	boxBody := testRigidBody(Shape{}, matrix.Vec3{1.75, 0, 0})
	boxBody.Collision.Shape.SetAABB(matrix.Vec3Zero(), matrix.Vec3{1, 1, 1})
	manifold, ok := CollideBodies(sphereBody, boxBody)
	if !ok {
		t.Fatal("expected sphere and AABB to collide")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Right(), 0.0001) {
		t.Fatalf("expected +X normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.25) > 0.0001 {
		t.Fatalf("expected penetration 0.25, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseCapsuleCapsuleContact(t *testing.T) {
	a := testRigidBody(Shape{}, matrix.Vec3{0, 0, 0})
	a.Collision.Shape.SetCapsule(matrix.Vec3Zero(), 0.5, 2, matrix.Vec3Up())
	b := testRigidBody(Shape{}, matrix.Vec3{0.75, 0, 0})
	b.Collision.Shape.SetCapsule(matrix.Vec3Zero(), 0.5, 2, matrix.Vec3Up())
	manifold, ok := CollideBodies(a, b)
	if !ok {
		t.Fatal("expected overlapping capsules to collide")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Right(), 0.0001) {
		t.Fatalf("expected +X normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.25) > 0.0001 {
		t.Fatalf("expected penetration 0.25, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseCylinderSphereContact(t *testing.T) {
	cylinder := testRigidBody(NewCylinderShape(1, 2), matrix.Vec3Zero())
	sphere := testRigidBody(NewSphereShape(1), matrix.Vec3{1.75, 0, 0})
	manifold, ok := CollideBodies(cylinder, sphere)
	if !ok {
		t.Fatal("expected cylinder and sphere to collide")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Right(), 0.0001) {
		t.Fatalf("expected +X cylinder-to-sphere normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.25) > 0.0001 {
		t.Fatalf("expected penetration 0.25, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseCylinderBoxContact(t *testing.T) {
	cylinder := testRigidBody(NewCylinderShape(1, 2), matrix.Vec3Zero())
	box := testRigidBody(NewBoxShape(matrix.Vec3One()), matrix.Vec3{1.75, 0, 0})
	manifold, ok := CollideBodies(cylinder, box)
	if !ok {
		t.Fatal("expected cylinder and box to collide")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Right(), 0.0001) {
		t.Fatalf("expected +X cylinder-to-box normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.25) > 0.0001 {
		t.Fatalf("expected penetration 0.25, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseConeSphereContact(t *testing.T) {
	cone := testRigidBody(NewConeShape(1, 2), matrix.Vec3Zero())
	sphere := testRigidBody(NewSphereShape(0.5), matrix.Vec3{0.8, 0, 0})
	expectedNormal := matrix.NewVec3(0.8944272, -0.4472136, 0)
	expectedPenetration := matrix.Float(0.2316718)
	manifold, ok := CollideBodies(cone, sphere)
	if !ok {
		t.Fatal("expected cone and sphere to collide")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, expectedNormal, 0.0001) {
		t.Fatalf("expected cone side normal %v, got %v", expectedNormal, contact.Normal)
	}
	if matrix.Abs(contact.Penetration-expectedPenetration) > 0.0001 {
		t.Fatalf("expected penetration %f, got %f", expectedPenetration, contact.Penetration)
	}
}

func TestNarrowPhaseConeBoxContact(t *testing.T) {
	cone := testRigidBody(NewConeShape(1, 2), matrix.Vec3Zero())
	box := testRigidBody(NewBoxShape(matrix.Vec3One()), matrix.Vec3{0, 1.75, 0})
	manifold, ok := CollideBodies(cone, box)
	if !ok {
		t.Fatal("expected cone and box to collide")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Up(), 0.0001) {
		t.Fatalf("expected +Y cone-to-box normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.25) > 0.0001 {
		t.Fatalf("expected penetration 0.25, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseSphereStaticMeshFloorContact(t *testing.T) {
	sphere := testRigidBody(NewSphereShape(0.5), matrix.Vec3{0, 0.45, 0})
	mesh := testStaticMeshBody(testMeshFloor())
	manifold, ok := CollideBodies(sphere, mesh)
	if !ok {
		t.Fatal("expected sphere to collide with mesh floor")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Down(), 0.0001) {
		t.Fatalf("expected downward sphere-to-mesh normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.05) > 0.0001 {
		t.Fatalf("expected penetration 0.05, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseCapsuleStaticMeshFloorContact(t *testing.T) {
	capsule := testRigidBody(NewCapsuleShape(0.5, 2), matrix.Vec3{0, 1.45, 0})
	mesh := testStaticMeshBody(testMeshFloor())
	manifold, ok := CollideBodies(capsule, mesh)
	if !ok {
		t.Fatal("expected capsule to collide with mesh floor")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Down(), 0.0001) {
		t.Fatalf("expected downward capsule-to-mesh normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.05) > 0.0001 {
		t.Fatalf("expected penetration 0.05, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseOOBBStaticMeshFloorContact(t *testing.T) {
	box := testRigidBody(NewBoxShape(matrix.NewVec3(0.5, 0.5, 0.5)), matrix.Vec3{0, 0.45, 0})
	mesh := testStaticMeshBody(testMeshFloor())
	manifold, ok := CollideBodies(box, mesh)
	if !ok {
		t.Fatal("expected box to collide with mesh floor")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Down(), 0.0001) {
		t.Fatalf("expected downward box-to-mesh normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.05) > 0.0001 {
		t.Fatalf("expected penetration 0.05, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseSphereStaticMeshSlopeContact(t *testing.T) {
	sphere := testRigidBody(NewSphereShape(0.5), matrix.Vec3{0, 0.45, 0})
	mesh := testStaticMeshBody(testSlopedMeshFloor())
	manifold, ok := CollideBodies(sphere, mesh)
	if !ok {
		t.Fatal("expected sphere to collide with sloped mesh floor")
	}
	normal := manifold.Contacts[0].Normal
	expected := matrix.NewVec3(0.4472136, -0.8944272, 0)
	if !matrix.Vec3ApproxTo(normal, expected, 0.0001) {
		t.Fatalf("expected slope normal %v, got %v", expected, normal)
	}
}

func TestNarrowPhaseSphereStaticMeshEdgeContact(t *testing.T) {
	sphere := testRigidBody(NewSphereShape(0.5), matrix.Vec3{0, 0.4, -0.2})
	mesh := testStaticMeshBody(NewMeshCollisionFromVertices([]matrix.Vec3{
		{-1, 0, 0},
		{1, 0, 0},
		{0, 0, 1},
	}, []uint32{0, 1, 2}))
	manifold, ok := CollideBodies(sphere, mesh)
	if !ok {
		t.Fatal("expected sphere to collide with triangle edge")
	}
	contact := manifold.Contacts[0]
	if contact.Normal.Y() >= 0 || contact.Normal.Z() <= 0 {
		t.Fatalf("expected edge contact normal to point down and toward triangle, got %v", contact.Normal)
	}
	if contact.Penetration <= 0 {
		t.Fatalf("expected positive edge penetration, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseSphereStaticTerrainFloorContact(t *testing.T) {
	sphere := testRigidBody(NewSphereShape(0.5), matrix.Vec3{0, 0.45, 0})
	terrain := testStaticTerrainBody(testFlatTerrain(t))
	manifold, ok := CollideBodies(sphere, terrain)
	if !ok {
		t.Fatal("expected sphere to collide with terrain floor")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Down(), 0.0001) {
		t.Fatalf("expected downward sphere-to-terrain normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.05) > 0.0001 {
		t.Fatalf("expected penetration 0.05, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseCapsuleStaticTerrainFloorContact(t *testing.T) {
	capsule := testRigidBody(NewCapsuleShape(0.5, 2), matrix.Vec3{0, 1.45, 0})
	terrain := testStaticTerrainBody(testFlatTerrain(t))
	manifold, ok := CollideBodies(capsule, terrain)
	if !ok {
		t.Fatal("expected capsule to collide with terrain floor")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Down(), 0.0001) {
		t.Fatalf("expected downward capsule-to-terrain normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.05) > 0.0001 {
		t.Fatalf("expected penetration 0.05, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseOOBBStaticTerrainFloorContact(t *testing.T) {
	box := testRigidBody(NewBoxShape(matrix.NewVec3(0.5, 0.5, 0.5)), matrix.Vec3{0, 0.45, 0})
	terrain := testStaticTerrainBody(testFlatTerrain(t))
	manifold, ok := CollideBodies(box, terrain)
	if !ok {
		t.Fatal("expected box to collide with terrain floor")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Down(), 0.0001) {
		t.Fatalf("expected downward box-to-terrain normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.05) > 0.0001 {
		t.Fatalf("expected penetration 0.05, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseSphereStaticTerrainTransformedContact(t *testing.T) {
	sphere := testRigidBody(NewSphereShape(0.5), matrix.Vec3{1, 1.45, -1})
	terrain := testStaticTerrainBody(testFlatTerrain(t))
	terrain.Transform.SetPosition(matrix.Vec3{1, 1, -1})
	manifold, ok := CollideBodies(sphere, terrain)
	if !ok {
		t.Fatal("expected sphere to collide with transformed terrain floor")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Down(), 0.0001) {
		t.Fatalf("expected downward sphere-to-terrain normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.05) > 0.0001 {
		t.Fatalf("expected penetration 0.05, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseSphereStaticTerrainUsesCurrentHeights(t *testing.T) {
	terrainCollision := testFlatTerrain(t)
	terrain := testStaticTerrainBody(terrainCollision)
	sphere := testRigidBody(NewSphereShape(0.5), matrix.Vec3{0, 0.45, 0})
	if _, ok := CollideBodies(sphere, terrain); !ok {
		t.Fatal("expected sphere to collide with initial terrain floor")
	}
	for i := range terrainCollision.Heights {
		terrainCollision.Heights[i] = -2
	}
	terrainCollision.MinHeight = -2
	terrainCollision.MaxHeight = -2
	terrainCollision.RefreshBounds()
	terrain.Collision.Shape = NewTerrainShape(terrainCollision.Bounds)
	terrain.Collision.LocalAABB = terrainCollision.Bounds
	if _, ok := CollideBodies(sphere, terrain); ok {
		t.Fatal("expected generated terrain triangles to use edited heights")
	}
}

func TestSystemDynamicSphereRestsOnStaticMeshFloor(t *testing.T) {
	system := System{}
	system.Initialize()
	sphere := system.NewBody()
	sphere.SetDynamic(1, matrix.Vec3One())
	sphere.Collision.Shape = NewSphereShape(0.5)
	sphere.Transform.SetPosition(matrix.Vec3{0, 0.5, 0})
	floor := system.NewBody()
	floor.SetStaticMesh(testMeshFloor())
	workGroup := concurrent.WorkGroup{}
	workGroup.Init()
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()
	hadContact := false
	for range 30 {
		system.Step(&workGroup, &threads, 1.0/60.0)
		hadContact = hadContact || len(system.Contacts()) > 0
	}
	if !hadContact {
		t.Fatal("expected sphere to contact mesh floor while stepping")
	}
	if sphere.Transform.WorldPosition().Y() < 0.45 {
		t.Fatalf("expected sphere to rest on mesh floor, got position %v", sphere.Transform.WorldPosition())
	}
}

func TestNarrowPhaseParallelMatchesSequential(t *testing.T) {
	pairs := make([]ActivePair, 0, 256)
	bodies := make([]*RigidBody, 0, 64)
	for x := range 12 {
		for y := range 6 {
			body := testRigidBody(Shape{}, matrix.Vec3{
				matrix.Float(x) * 0.85,
				matrix.Float(y) * 0.85,
				0,
			})
			body.Collision.Shape.SetSphere(matrix.Vec3Zero(), 0.5)
			bodies = append(bodies, body)
		}
	}
	for i := range bodies {
		for j := i + 1; j < len(bodies); j++ {
			pairs = append(pairs, ActivePair{BodyA: bodies[i], BodyB: bodies[j]})
		}
	}
	var sequential NarrowPhase
	seq := manifoldSet(sequential.Collide(pairs, nil))
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()
	var parallel NarrowPhase
	par := manifoldSet(parallel.Collide(pairs, &threads))
	if len(seq) != len(par) {
		t.Fatalf("expected %d parallel manifolds, got %d", len(seq), len(par))
	}
	for pair := range seq {
		if !par[pair] {
			t.Fatalf("parallel narrow phase missed pair %v", pair)
		}
	}
}

func TestSystemStepPublishesContacts(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	dynamic := system.NewBody()
	dynamic.Active = true
	dynamic.Simulation.Type = RigidBodyTypeDynamic
	dynamic.SetMass(1, matrix.Vec3One())
	dynamic.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	dynamic.Collision.Group = 0
	dynamic.Collision.Mask = 1
	static := system.NewBody()
	static.Active = true
	static.Simulation.Type = RigidBodyTypeStatic
	static.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	static.Collision.Group = 0
	static.Collision.Mask = 1
	static.Transform.SetPosition(matrix.Vec3{1.5, 0, 0})
	workGroup := concurrent.WorkGroup{}
	workGroup.Init()
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()
	system.Step(&workGroup, &threads, 0)
	contacts := system.Contacts()
	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact manifold, got %d", len(contacts))
	}
	if contacts[0].BodyA == nil || contacts[0].BodyB == nil || contacts[0].Count == 0 {
		t.Fatal("expected populated contact manifold")
	}
}

func testRigidBody(shape Shape, position matrix.Vec3) *RigidBody {
	body := &RigidBody{}
	body.Active = true
	body.Collision.Shape = shape
	body.Collision.Group = 0
	body.Collision.Mask = 1
	body.Transform.SetupRawTransform()
	body.Transform.SetPosition(position)
	body.SetMass(1, matrix.Vec3One())
	body.Simulation.Type = RigidBodyTypeDynamic
	return body
}

func testStaticMeshBody(mesh *MeshCollision) *RigidBody {
	body := &RigidBody{}
	body.Transform.SetupRawTransform()
	body.SetStaticMesh(mesh)
	return body
}

func testStaticTerrainBody(terrain *TerrainCollision) *RigidBody {
	body := &RigidBody{}
	body.Transform.SetupRawTransform()
	body.SetStaticTerrain(terrain)
	return body
}

func testMeshFloor() *MeshCollision {
	return NewMeshCollisionFromVertices([]matrix.Vec3{
		{-2, 0, -2},
		{2, 0, -2},
		{-2, 0, 2},
		{2, 0, 2},
	}, []uint32{0, 1, 2, 2, 1, 3})
}

func testSlopedMeshFloor() *MeshCollision {
	return NewMeshCollisionFromVertices([]matrix.Vec3{
		{-2, -1, -2},
		{2, 1, -2},
		{-2, -1, 2},
		{2, 1, 2},
	}, []uint32{0, 1, 2, 2, 1, 3})
}

func testFlatTerrain(t *testing.T) *TerrainCollision {
	t.Helper()
	return testTerrainCollision(t, 2, matrix.NewVec2(4, 4), []matrix.Float{
		0, 0,
		0, 0,
	}, 0, 0)
}

func manifoldSet(manifolds []ContactManifold) map[[2]*RigidBody]bool {
	set := make(map[[2]*RigidBody]bool, len(manifolds))
	for _, manifold := range manifolds {
		set[[2]*RigidBody{manifold.BodyA, manifold.BodyB}] = true
	}
	return set
}
