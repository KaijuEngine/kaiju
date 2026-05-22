/******************************************************************************/
/* system_test.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import (
	"math"
	"testing"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

func TestSystemRemoveBodyReleasesBody(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	body := system.NewBody()
	if system.bodies.ElementCount() != 1 {
		t.Fatalf("expected 1 pooled body, got %d", system.bodies.ElementCount())
	}
	system.RemoveBody(body)
	if system.bodies.ElementCount() != 0 {
		t.Fatalf("expected removed body to release its pool slot, got %d bodies", system.bodies.ElementCount())
	}
	if body.Active || body.pooled {
		t.Fatal("expected removed body reference to be inactive and detached from the pool")
	}
	system.RemoveBody(body)
	if system.bodies.ElementCount() != 0 {
		t.Fatalf("expected removing an already removed body to be safe, got %d bodies", system.bodies.ElementCount())
	}
}

func TestSystemAddRemoveConstraint(t *testing.T) {
	system := System{}
	system.Initialize()
	bodyA := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	bodyB := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic)
	constraint := system.NewConstraint(ConstraintTypeGeneric, bodyA, bodyB)
	if system.constraints.ElementCount() != 1 {
		t.Fatalf("expected 1 pooled constraint, got %d", system.constraints.ElementCount())
	}
	if !constraint.IsValid() {
		t.Fatal("expected new body-body constraint to be valid")
	}
	if !constraint.IsBodyBody() {
		t.Fatal("expected two body endpoints to create a body-body constraint")
	}
	system.RemoveConstraint(constraint)
	if system.constraints.ElementCount() != 0 {
		t.Fatalf("expected removed constraint to release its pool slot, got %d constraints", system.constraints.ElementCount())
	}
	if constraint.Active || constraint.Enabled || constraint.pooled {
		t.Fatal("expected removed constraint reference to be disabled and detached from the pool")
	}
	worldConstraint := system.NewConstraint(ConstraintTypeGeneric, bodyA, nil)
	if !worldConstraint.IsValid() {
		t.Fatal("expected body-world constraint to be valid")
	}
	if !worldConstraint.IsBodyWorld() {
		t.Fatal("expected nil second endpoint to create a body-world constraint")
	}
}

func TestSystemClearRemovesConstraints(t *testing.T) {
	system := System{}
	system.Initialize()
	bodyA := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	bodyB := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic)
	system.NewConstraint(ConstraintTypeGeneric, bodyA, bodyB)
	system.NewConstraint(ConstraintTypeGeneric, bodyA, nil)
	system.ClearConstraints()
	if system.constraints.ElementCount() != 0 {
		t.Fatalf("expected clear constraints to release all constraints, got %d", system.constraints.ElementCount())
	}
	if system.bodies.ElementCount() != 2 {
		t.Fatalf("expected clearing constraints not to affect bodies, got %d", system.bodies.ElementCount())
	}
	system.NewConstraint(ConstraintTypeGeneric, bodyA, bodyB)
	system.Clear()
	if system.constraints.ElementCount() != 0 {
		t.Fatalf("expected system clear to release all constraints, got %d", system.constraints.ElementCount())
	}
	if system.bodies.ElementCount() != 0 {
		t.Fatalf("expected system clear to release all bodies, got %d", system.bodies.ElementCount())
	}
}

func TestSystemConstraintsReturnsStoredConstraints(t *testing.T) {
	system := System{}
	system.Initialize()
	bodyA := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	bodyB := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic)
	constraint := system.NewConstraint(ConstraintTypeGeneric, bodyA, bodyB)
	constraints := system.Constraints()
	if len(constraints) != 1 {
		t.Fatalf("expected 1 stored constraint, got %d", len(constraints))
	}
	if constraints[0] != constraint {
		t.Fatal("expected Constraints to return the stored constraint")
	}
}

func TestSystemRemoveBodyDisablesAttachedConstraints(t *testing.T) {
	system := System{}
	system.Initialize()
	bodyA := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	bodyB := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic)
	bodyC := addSystemSphere(&system, matrix.Vec3{4, 0, 0}, RigidBodyTypeDynamic)
	bodyBody := system.NewConstraint(ConstraintTypeGeneric, bodyA, bodyB)
	bodyWorld := system.NewConstraint(ConstraintTypeGeneric, bodyA, nil)
	unattached := system.NewConstraint(ConstraintTypeGeneric, bodyB, bodyC)
	system.RemoveBody(bodyA)
	if system.constraints.ElementCount() != 3 {
		t.Fatalf("expected removing a body to leave constraints in storage, got %d", system.constraints.ElementCount())
	}
	if bodyBody.Active || bodyBody.Enabled || bodyBody.IsValid() {
		t.Fatal("expected body-body constraint attached to removed body to be disabled")
	}
	if bodyWorld.Active || bodyWorld.Enabled || bodyWorld.IsValid() {
		t.Fatal("expected body-world constraint attached to removed body to be disabled")
	}
	if bodyBody.BodyA != nil || bodyWorld.BodyA != nil {
		t.Fatal("expected removed body endpoint to be cleared from attached constraints")
	}
	if !unattached.IsValid() {
		t.Fatal("expected unrelated constraint to remain valid")
	}
}

func TestSystemStepSolvesDistanceJoint(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintPositionIterations = 8
	bodyA := addSystemSphere(&system, matrix.Vec3{-3, 0, 0}, RigidBodyTypeDynamic)
	bodyB := addSystemSphere(&system, matrix.Vec3{3, 0, 0}, RigidBodyTypeDynamic)
	bodyA.Collision.Mask = 0
	bodyB.Collision.Mask = 0
	bodyA.Simulation.SleepThreshold = 10000
	bodyB.Simulation.SleepThreshold = 10000
	system.NewDistanceJoint(bodyA, bodyB, matrix.Vec3Zero(), matrix.Vec3Zero()).SetRestLength(2)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 120 {
		system.Step(workGroup, threads, 1.0/60.0)
	}
	if len(system.Constraints()) != 1 {
		t.Fatalf("expected distance joint to be visible as 1 constraint, got %d", len(system.Constraints()))
	}
	distance := bodyA.Transform.WorldPosition().Distance(bodyB.Transform.WorldPosition())
	if matrix.Abs(distance-2) > 0.02 {
		t.Fatalf("expected Step to solve distance joint rest length 2, got %f", distance)
	}
}

func TestSystemConstraintAndContactSolveTogether(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintVelocityIterations = 8
	system.ConstraintPositionIterations = 8
	bodyA := addSystemSphere(&system, matrix.Vec3{0, 0, 0}, RigidBodyTypeDynamic)
	bodyB := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic)
	wall := addSystemSphere(&system, matrix.Vec3{3.5, 0, 0}, RigidBodyTypeStatic)
	bodyA.Simulation.SleepThreshold = 10000
	bodyB.Simulation.SleepThreshold = 10000
	system.NewDistanceJoint(bodyA, bodyB, matrix.Vec3Zero(), matrix.Vec3Zero()).SetRestLength(2)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 30 {
		system.Step(workGroup, threads, 1.0/60.0)
	}
	if len(system.Contacts()) == 0 {
		t.Fatal("expected bodyB and wall to generate a contact during Step")
	}
	distance := bodyA.Transform.WorldPosition().Distance(bodyB.Transform.WorldPosition())
	if matrix.Abs(distance-2) > 0.02 {
		t.Fatalf("expected constraint to preserve rest length while contacts solve, got %f", distance)
	}
	separation := bodyB.Transform.WorldPosition().Distance(wall.Transform.WorldPosition())
	if separation < 1.99 {
		t.Fatalf("expected contact solve to separate constrained body from wall, got %f", separation)
	}
}

func TestSystemStepExcludesRemovedBodyFromContacts(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	dynamic := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	static := addSystemSphere(&system, matrix.Vec3{1.5, 0, 0}, RigidBodyTypeStatic)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 0)
	if len(system.Contacts()) != 1 {
		t.Fatalf("expected overlapping bodies to create 1 contact, got %d", len(system.Contacts()))
	}
	system.RemoveBody(dynamic)
	system.Step(workGroup, threads, 0)
	if len(system.Contacts()) != 0 {
		t.Fatalf("expected removed body to be absent from contacts, got %d", len(system.Contacts()))
	}
	system.broadPhase.Rebuild(&system.bodies)
	if len(system.broadPhase.proxies) != 1 || system.broadPhase.proxies[0].body != static {
		t.Fatal("expected broad phase rebuild to include only the remaining body")
	}
}

func TestSystemClearRemovesBodiesAndContacts(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	addSystemSphere(&system, matrix.Vec3{1.5, 0, 0}, RigidBodyTypeStatic)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 0)
	if len(system.Contacts()) != 1 {
		t.Fatalf("expected overlapping bodies to create 1 contact, got %d", len(system.Contacts()))
	}
	system.Clear()
	if system.bodies.ElementCount() != 0 {
		t.Fatalf("expected clear to release all bodies, got %d", system.bodies.ElementCount())
	}
	if len(system.Contacts()) != 0 {
		t.Fatalf("expected clear to reset contact manifolds, got %d", len(system.Contacts()))
	}
	if len(system.broadPhase.proxies) != 0 {
		t.Fatalf("expected clear to reset broad phase proxies, got %d", len(system.broadPhase.proxies))
	}
	system.Step(workGroup, threads, 0)
	if len(system.Contacts()) != 0 {
		t.Fatalf("expected empty system step to have no contacts, got %d", len(system.Contacts()))
	}
}

func TestSystemRaycastReturnsClosestHit(t *testing.T) {
	system := System{}
	system.Initialize()
	farBody := addSystemSphere(&system, matrix.Vec3{4, 0, 0}, RigidBodyTypeStatic)
	nearBody := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeStatic)
	hit, ok := system.Raycast(matrix.Vec3Zero(), matrix.Vec3{10, 0, 0})
	if !ok {
		t.Fatal("expected raycast to hit a body")
	}
	if hit.Body != nearBody {
		t.Fatalf("expected raycast to return closest body %p, got %p", nearBody, hit.Body)
	}
	if hit.Body == farBody {
		t.Fatal("expected far body not to be selected")
	}
	if !matrix.Approx(hit.Distance, 1) {
		t.Fatalf("expected hit distance 1, got %f", hit.Distance)
	}
	if !matrix.Vec3ApproxTo(hit.Point, matrix.Vec3{1, 0, 0}, 0.0001) {
		t.Fatalf("expected hit point at 1,0,0, got %v", hit.Point)
	}
	if !matrix.Vec3ApproxTo(hit.Normal, matrix.Vec3Left(), 0.0001) {
		t.Fatalf("expected hit normal -X, got %v", hit.Normal)
	}
}

func TestSystemRaycastNoHit(t *testing.T) {
	system := System{}
	system.Initialize()
	addSystemSphere(&system, matrix.Vec3{0, 3, 0}, RigidBodyTypeStatic)
	if hit, ok := system.Raycast(matrix.Vec3Zero(), matrix.Vec3{10, 0, 0}); ok {
		t.Fatalf("expected raycast to miss, got hit %+v", hit)
	}
}

func TestSystemSphereSweepNoHit(t *testing.T) {
	system := System{}
	system.Initialize()
	addSystemSphere(&system, matrix.Vec3{0, 3, 0}, RigidBodyTypeStatic)
	if hit, ok := system.SphereSweep(matrix.Vec3Zero(), matrix.Vec3{10, 0, 0}, 0.5); ok {
		t.Fatalf("expected sphere sweep to miss, got hit %+v", hit)
	}
}

func TestSystemSphereSweepReturnsClosestHit(t *testing.T) {
	system := System{}
	system.Initialize()
	farBody := addSystemSphere(&system, matrix.Vec3{4, 0, 0}, RigidBodyTypeStatic)
	nearBody := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeStatic)
	hit, ok := system.SphereSweep(matrix.Vec3Zero(), matrix.Vec3{10, 0, 0}, 0.5)
	if !ok {
		t.Fatal("expected sphere sweep to hit a body")
	}
	if hit.Body != nearBody {
		t.Fatalf("expected sphere sweep to return closest body %p, got %p", nearBody, hit.Body)
	}
	if hit.Body == farBody {
		t.Fatal("expected far body not to be selected")
	}
	if !matrix.Approx(hit.Distance, 0.5) {
		t.Fatalf("expected hit distance 0.5, got %f", hit.Distance)
	}
	if !matrix.Vec3ApproxTo(hit.Point, matrix.Vec3{1, 0, 0}, 0.0001) {
		t.Fatalf("expected hit point at 1,0,0, got %v", hit.Point)
	}
	if !matrix.Vec3ApproxTo(hit.Normal, matrix.Vec3Left(), 0.0001) {
		t.Fatalf("expected hit normal -X, got %v", hit.Normal)
	}
}

func TestSystemSphereSweepStartOverlap(t *testing.T) {
	system := System{}
	system.Initialize()
	body := addSystemSphere(&system, matrix.Vec3{0.75, 0, 0}, RigidBodyTypeStatic)
	hit, ok := system.SphereSweep(matrix.Vec3Zero(), matrix.Vec3{10, 0, 0}, 0.5)
	if !ok {
		t.Fatal("expected sphere sweep to report start overlap")
	}
	if hit.Body != body {
		t.Fatalf("expected sphere sweep to return overlapping body %p, got %p", body, hit.Body)
	}
	if !matrix.Approx(hit.Distance, 0) {
		t.Fatalf("expected start-overlap distance 0, got %f", hit.Distance)
	}
	if !matrix.Vec3ApproxTo(hit.Normal, matrix.Vec3Left(), 0.0001) {
		t.Fatalf("expected start-overlap normal -X, got %v", hit.Normal)
	}
}

func TestSystemDynamicBodySleepsAtRest(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	body := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.Simulation.SleepThreshold = 0.2
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 0.1)
	if body.Simulation.IsSleeping {
		t.Fatal("expected body to remain awake before sleep threshold")
	}
	system.Step(workGroup, threads, 0.1)

	if !body.Simulation.IsSleeping {
		t.Fatalf("expected resting body to sleep after threshold, timer %f", body.Simulation.SleepTimer)
	}
}

func TestSystemDoesNotAutoSleepStaticOrKinematicBodies(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	staticBody := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeStatic)
	kinematicBody := addSystemSphere(&system, matrix.Vec3{3, 0, 0}, RigidBodyTypeKinematic)
	staticBody.Simulation.SleepThreshold = 0.1
	kinematicBody.Simulation.SleepThreshold = 0.1
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 1)
	if staticBody.Simulation.IsSleeping {
		t.Fatal("expected static body not to auto sleep")
	}
	if kinematicBody.Simulation.IsSleeping {
		t.Fatal("expected kinematic body not to auto sleep")
	}
}

func TestSystemTransformChangeWakesSleepingBody(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	body := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.Sleep()
	body.Transform.SetPosition(matrix.Vec3{1, 0, 0})
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 0)
	if body.Simulation.IsSleeping {
		t.Fatal("expected transform change to wake sleeping body")
	}
}

func TestSystemContactWithAwakeBodyWakesSleepingBody(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	sleeping := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	awake := addSystemSphere(&system, matrix.Vec3{1.5, 0, 0}, RigidBodyTypeDynamic)
	sleeping.Sleep()
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 0)
	if sleeping.Simulation.IsSleeping {
		t.Fatal("expected contact with awake dynamic body to wake sleeping body")
	}
	if awake.Simulation.IsSleeping {
		t.Fatal("expected awake body to remain awake")
	}
}

func TestConstraintCreationWakesBodies(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	bodyA := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	bodyB := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic)
	bodyA.Sleep()
	bodyB.Sleep()
	constraint := system.NewConstraint(ConstraintTypeGeneric, bodyA, bodyB)
	if bodyA.Simulation.IsSleeping {
		t.Fatal("expected constraint creation to wake body A")
	}
	if bodyB.Simulation.IsSleeping {
		t.Fatal("expected constraint creation to wake body B")
	}
	constraint.SetEnabled(false)
	bodyA.Sleep()
	bodyB.Sleep()
	constraint.SetEnabled(true)
	if bodyA.Simulation.IsSleeping {
		t.Fatal("expected constraint enabling to wake body A")
	}
	if bodyB.Simulation.IsSleeping {
		t.Fatal("expected constraint enabling to wake body B")
	}
}

func TestConstraintWithMovingBodyWakesSleepingBody(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	sleeping := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	moved := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic)
	sleeping.Collision.Mask = 0
	moved.Collision.Mask = 0
	system.NewDistanceJoint(sleeping, moved, matrix.Vec3Zero(), matrix.Vec3Zero())
	sleeping.Sleep()
	moved.Sleep()
	moved.Transform.SetPosition(matrix.Vec3{3, 0, 0})
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 0)
	if sleeping.Simulation.IsSleeping {
		t.Fatal("expected moved constrained body to wake its sleeping partner")
	}
	if moved.Simulation.IsSleeping {
		t.Fatal("expected moved constrained body to wake itself")
	}
}

func TestStableConstraintIslandSleeps(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	bodyA := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	bodyB := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic)
	bodyA.Collision.Mask = 0
	bodyB.Collision.Mask = 0
	bodyA.Simulation.SleepThreshold = 0.1
	bodyB.Simulation.SleepThreshold = 0.1
	system.NewDistanceJoint(bodyA, bodyB, matrix.Vec3Zero(), matrix.Vec3Zero())
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 3 {
		system.Step(workGroup, threads, 0.05)
	}
	if !bodyA.Simulation.IsSleeping {
		t.Fatalf("expected stable constrained body A to sleep, timer %f", bodyA.Simulation.SleepTimer)
	}
	if !bodyB.Simulation.IsSleeping {
		t.Fatalf("expected stable constrained body B to sleep, timer %f", bodyB.Simulation.SleepTimer)
	}
}

func TestStretchedConstraintPreventsSleep(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintVelocityIterations = 0
	system.ConstraintPositionIterations = 0
	bodyA := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	bodyB := addSystemSphere(&system, matrix.Vec3{3, 0, 0}, RigidBodyTypeDynamic)
	bodyA.Collision.Mask = 0
	bodyB.Collision.Mask = 0
	bodyA.Simulation.SleepThreshold = 0.1
	bodyB.Simulation.SleepThreshold = 0.1
	system.NewDistanceJoint(bodyA, bodyB, matrix.Vec3Zero(), matrix.Vec3Zero()).SetRestLength(1)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 3 {
		system.Step(workGroup, threads, 0.05)
	}
	if bodyA.Simulation.IsSleeping {
		t.Fatal("expected stretched constrained body A to stay awake")
	}
	if bodyB.Simulation.IsSleeping {
		t.Fatal("expected stretched constrained body B to stay awake")
	}
}

func TestSystemStepIntegratesAngularVelocityAsRadians(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	body := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.MotionState.AngularVelocity = matrix.NewVec3(matrix.Float(math.Pi), 0, 0)
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 0.5)
	if !matrix.Vec3ApproxTo(body.Transform.Rotation(), matrix.NewVec3(90, 0, 0), 0.0001) {
		t.Fatalf("expected half second at pi rad/s to rotate 90 degrees, got %v", body.Transform.Rotation())
	}
}

func TestIntegrateAngularVelocityUsesWorldSpaceAxis(t *testing.T) {
	rotation := matrix.NewVec3(0, 90, 0)
	angularVelocity := matrix.NewVec3(matrix.Float(math.Pi), 0, 0)
	actual := matrix.QuaternionFromEuler(integrateAngularVelocity(rotation, angularVelocity, 0.5))
	expected := matrix.QuaternionAxisAngle(matrix.Vec3Right(), matrix.Float(math.Pi/2)).
		Multiply(matrix.QuaternionFromEuler(rotation))

	v := matrix.NewVec3(0.25, 0.5, -1)
	actualDirection := actual.MultiplyVec3(v)
	expectedDirection := expected.MultiplyVec3(v)
	if !matrix.Vec3ApproxTo(actualDirection, expectedDirection, 0.0001) {
		t.Fatalf("expected angular velocity to rotate around world axis, got %v, want %v", actualDirection, expectedDirection)
	}
}

func addSystemSphere(system *System, position matrix.Vec3, bodyType RigidBodyType) *RigidBody {
	body := system.NewBody()
	body.Active = true
	body.Simulation.Type = bodyType
	body.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	body.Collision.Group = 0
	body.Collision.Mask = 1
	body.Transform.SetPosition(position)
	if bodyType == RigidBodyTypeDynamic {
		body.SetMass(1, matrix.Vec3One())
	}
	return body
}

func testStepWorkers(t *testing.T) (*concurrent.WorkGroup, *concurrent.Threads, func()) {
	t.Helper()
	workGroup := concurrent.WorkGroup{}
	workGroup.Init()
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	return &workGroup, &threads, threads.Stop
}
